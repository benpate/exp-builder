/******************************************
 * Fuzz Tests
 *
 * Every fuzz target in this package lives here, so
 * that the properties they enforce can be read side
 * by side.
 *
 * These assert PROPERTIES rather than outputs. The
 * package's whole job is to keep untrusted text from
 * turning into an unintended query, so the properties
 * worth pinning are the ones that describe what can
 * never come out the other end: a field nobody
 * declared, an operator nobody recognizes, or a value
 * of the wrong Go type.
 ******************************************/

package builder

import (
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/benpate/exp"
	"github.com/benpate/geo"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

/******************************************
 * Test Helpers
 ******************************************/

// predicates walks an Expression tree and returns every Predicate it contains.
// The tree is built from nested And/Or expressions, so the walk has to recurse.
func predicates(expression exp.Expression) []exp.Predicate {

	switch typed := expression.(type) {

	case exp.Predicate:
		return []exp.Predicate{typed}

	case exp.AndExpression:
		result := make([]exp.Predicate, 0)
		for _, child := range typed {
			result = append(result, predicates(child)...)
		}
		return result

	case exp.OrExpression:
		result := make([]exp.Predicate, 0)
		for _, child := range typed {
			result = append(result, predicates(child)...)
		}
		return result
	}

	// An EmptyExpression contributes nothing
	return make([]exp.Predicate, 0)
}

// canonicalOperators lists every operator that exp defines. It is written out
// by hand, rather than derived from exp, so that this package notices when the
// set it is willing to emit changes underneath it.
var canonicalOperators = map[string]bool{
	exp.OperatorEqual:          true,
	exp.OperatorNotEqual:       true,
	exp.OperatorGreaterThan:    true,
	exp.OperatorGreaterOrEqual: true,
	exp.OperatorLessThan:       true,
	exp.OperatorLessOrEqual:    true,
	exp.OperatorIn:             true,
	exp.OperatorNotIn:          true,
	exp.OperatorInAll:          true,
	exp.OperatorBeginsWith:     true,
	exp.OperatorEndsWith:       true,
	exp.OperatorContains:       true,
	exp.OperatorContainedBy:    true,
	exp.OperatorExists:         true,
	exp.OperatorGeoWithin:      true,
	exp.OperatorGeoIntersects:  true,
}

// timeRangeNames restates the ranges that parseTimeRange accepts. It is a
// deliberate second copy: a fuzz target that derived the list from the function
// under test could only ever agree with it.
var timeRangeNames = []string{
	"past-30-days", "past-60-days", "past-90-days", "past-180-days", "past-365-days",
	"next-30-days", "next-60-days", "next-90-days", "next-180-days", "next-365-days",
	"yesterday", "today", "tomorrow",
	"this-week", "next-week",
	"last-month", "this-month", "next-month",
	"last-year", "this-year", "next-year",
}

// everyDataType is the full set of data types a Field may declare.
var everyDataType = []string{
	DataTypeString, DataTypeInt, DataTypeInt64, DataTypeBool,
	DataTypeObjectID, DataTypeTime, DataTypePolygon,
}

// fuzzBuilder declares one parameter of every supported data type, plus an
// aliased field whose expression name differs from its URL name.
func fuzzBuilder() Builder {
	return NewBuilder().
		String("firstName").
		Int("publishDate").
		Int64("bignum").
		Bool("isPublished").
		ObjectID("parentId").
		Time("timeValue").
		Polygon("location").
		String("q", WithAlias("value"), WithDefaultOpContains())
}

// declaredFields returns the expression field names that a Builder is allowed
// to emit, which is the Field.Name -- NOT the URL parameter name -- because
// WithAlias may rename it.
func declaredFields(b Builder) map[string]bool {

	result := make(map[string]bool, len(b))

	for _, field := range b {
		result[field.Name] = true
	}

	return result
}

// queryStringSeeds are shared by the targets that take a whole query string.
var queryStringSeeds = []string{
	"firstName=John",
	"firstName=a&firstName=b&publishDate=ne:7",
	"firstName=&firstName=John",
	"q=???",
	"q=contains:",
	"q=exists:",
	"publishDate=MIN",
	"publishDate=lt:5",
	"bignum=MAX",
	"isPublished=TRUE",
	"parentId=123456781234567812345678",
	"timeValue=past-30-days",
	"timeValue=gt:today",
	"timeValue=2020-01-01T09:30:00Z",
	"location=1,2,3,4",
	"location=eq:1,2,3,4",
	"undeclared=whatever",
	"", "=", "&&", "a=%zz",
}

/******************************************
 * Safety Properties
 ******************************************/

// FuzzEvaluate_OnlyDeclaredFieldsAreEmitted is the package's central promise: no
// query string, however hostile, may put a field name into the expression that
// the caller did not declare. A failure here means a client can aim a query at
// a column it was never granted.
func FuzzEvaluate_OnlyDeclaredFieldsAreEmitted(f *testing.F) {

	for _, seed := range queryStringSeeds {
		f.Add(seed)
	}

	b := fuzzBuilder()
	allowed := declaredFields(b)

	f.Fuzz(func(t *testing.T, query string) {

		values, err := url.ParseQuery(query)
		if err != nil {
			return // Not a query string at all; net/url has its own fuzzers.
		}

		for _, predicate := range predicates(b.Evaluate(values)) {
			require.True(t, allowed[predicate.Field],
				"undeclared field %q escaped from query %q", predicate.Field, query)
		}
	})
}

// FuzzEvaluate_OnlyCanonicalOperatorsAreEmitted confirms that the "OP:" prefix
// cannot smuggle an arbitrary string into the operator position. Every operator
// that reaches the expression is one of exp's own constants.
func FuzzEvaluate_OnlyCanonicalOperatorsAreEmitted(f *testing.F) {

	for _, seed := range queryStringSeeds {
		f.Add(seed)
	}

	b := fuzzBuilder()

	f.Fuzz(func(t *testing.T, query string) {

		values, err := url.ParseQuery(query)
		if err != nil {
			return
		}

		for _, predicate := range predicates(b.Evaluate(values)) {
			require.True(t, canonicalOperators[predicate.Operator],
				"non-canonical operator %q escaped from query %q", predicate.Operator, query)
		}
	})
}

// FuzzEvaluate_ValuesMatchTheirDeclaredType confirms that each data type's
// parser hands back the Go type it promised. A leak here would put a raw string
// where the database expects an int or an ObjectID.
func FuzzEvaluate_ValuesMatchTheirDeclaredType(f *testing.F) {

	seeds := []string{
		"", "123", "-1", "0", "lt:5", "MIN", "MAX", "true", "FALSE", "True",
		"9223372036854775807", "9223372036854775808", "-9223372036854775808",
		"123456781234567812345678", "not-an-id", "0x7f",
		"2020-01-02", "2020-01-02T09:30:00Z", "past-30-days", "gt:today",
		"1,2,3,4", "lt:1,2", ":", "EQ:", "ne:", "exists:",
		"\x00", "\xff\xfe", "١٢٣",
	}
	for _, seed := range seeds {
		f.Add(seed)
	}

	// The type each data type is contractually required to produce.
	expectedTypes := map[string]reflect.Type{
		DataTypeString:   reflect.TypeOf(""),
		DataTypeInt:      reflect.TypeOf(int(0)),
		DataTypeInt64:    reflect.TypeOf(int64(0)),
		DataTypeBool:     reflect.TypeOf(false),
		DataTypeObjectID: reflect.TypeOf(primitive.ObjectID{}),
		DataTypeTime:     reflect.TypeOf(time.Time{}),
		DataTypePolygon:  reflect.TypeOf(geo.Polygon{}),
	}

	f.Fuzz(func(t *testing.T, value string) {

		for _, dataType := range everyDataType {

			field := NewField("field", dataType)
			result := Builder{}.EvaluateField(field, []string{value})

			for _, predicate := range predicates(result) {
				require.Equal(t, expectedTypes[dataType], reflect.TypeOf(predicate.Value),
					"%s parser produced a %T from %q", dataType, predicate.Value, value)
			}
		}
	})
}

/******************************************
 * Consistency Properties
 ******************************************/

// FuzzEvaluate_IsDeterministic confirms that the same URL always produces the
// same expression. The Builder is a map, so this holds only because Evaluate
// visits its fields in sorted order.
func FuzzEvaluate_IsDeterministic(f *testing.F) {

	for _, seed := range queryStringSeeds {
		f.Add(seed)
	}

	b := fuzzBuilder()

	f.Fuzz(func(t *testing.T, query string) {

		values, err := url.ParseQuery(query)
		if err != nil {
			return
		}

		expected := b.Evaluate(values)

		for range 8 {
			require.Equal(t, expected, b.Evaluate(values), "query=%q", query)
		}
	})
}

// FuzzEvaluate_UndeclaredParametersAreInert states the allow-list from the
// other direction: adding parameters that the Builder does not declare must not
// change the expression by so much as a byte.
func FuzzEvaluate_UndeclaredParametersAreInert(f *testing.F) {

	f.Add("firstName=John", "junk=1&other=2")
	f.Add("", "anything=at&all=here")
	f.Add("publishDate=gt:5", "publishDate2=99")
	f.Add("q=abc", "Q=ABC")

	b := fuzzBuilder()

	f.Fuzz(func(t *testing.T, query string, extra string) {

		values, err := url.ParseQuery(query)
		if err != nil {
			return
		}

		extraValues, err := url.ParseQuery(extra)
		if err != nil {
			return
		}

		expected := b.Evaluate(values)

		// Merge in only the parameters that the Builder does NOT declare.
		for name, value := range extraValues {
			if _, declared := b[name]; declared {
				continue
			}
			values[name] = value
		}

		require.Equal(t, expected, b.Evaluate(values),
			"undeclared parameters changed the result: query=%q extra=%q", query, extra)
	})
}

// FuzzEvaluateAll_AgreesWithEvaluate cross-checks the two entry points against
// each other. When every declared field is present, EvaluateAll walks the same
// fields in the same order, so it must produce exactly what Evaluate produces.
func FuzzEvaluateAll_AgreesWithEvaluate(f *testing.F) {

	for _, seed := range queryStringSeeds {
		f.Add(seed)
	}

	b := fuzzBuilder()

	f.Fuzz(func(t *testing.T, query string) {

		values, err := url.ParseQuery(query)
		if err != nil {
			return
		}

		result, evaluateAllErr := b.EvaluateAll(values)

		// A missing field is the documented outcome, and the expression that comes
		// back with it is always empty.
		if evaluateAllErr != nil {
			require.Equal(t, exp.Empty(), result, "query=%q", query)
			return
		}

		require.Equal(t, b.Evaluate(values), result, "query=%q", query)
	})
}

/******************************************
 * Parser Properties
 ******************************************/

// FuzzParseValue_Structure pins the shape of the "OP:value" split. The returned
// value is always a suffix of the input -- the parser only ever removes a
// prefix, it never invents text -- and the returned operator is always either
// the caller's default or something exp recognizes.
func FuzzParseValue_Structure(f *testing.F) {

	seeds := []string{
		"", "John", "EQ:John", "gt:7", "bogus:value", "::", ":x", "a:b:c",
		"2020-01-02T09:30:00Z", "exists:", "&GT;:1", "eq:", ":",
	}
	for _, seed := range seeds {
		f.Add(seed)
	}

	const defaultOperator = "SOME-DEFAULT"

	f.Fuzz(func(t *testing.T, input string) {

		operator, value := parseValue(input, defaultOperator)

		// Empty input yields two empty strings, and nothing else does.
		if input == "" {
			require.Equal(t, "", operator)
			require.Equal(t, "", value)
			return
		}

		// The value is always a suffix of the input.
		require.True(t, len(value) <= len(input), "value grew: input=%q value=%q", input, value)
		require.Equal(t, input[len(input)-len(value):], value,
			"value is not a suffix of the input: input=%q value=%q", input, value)

		// The default is returned untouched, or a real operator was recognized.
		if operator == defaultOperator {
			require.Equal(t, input, value, "the default operator keeps the whole input")
			return
		}

		standard, ok := exp.OperatorOk(operator)
		require.True(t, ok, "operator=%q is not one exp recognizes", operator)
		require.Equal(t, standard, operator, "operator=%q is not in canonical form", operator)

		// A recognized operator means a prefix and its colon were consumed.
		require.Equal(t, len(input), len(value)+len(input)-len(value))
		require.True(t, len(input) > len(value), "a prefix was reported but nothing was removed")
		require.Equal(t, byte(':'), input[len(input)-len(value)-1],
			"the character before the value must be the separating colon")
	})
}

// FuzzParseTimeRange_Oracle checks the parser against an independent list of the
// ranges it is supposed to know. Anything outside that list must be rejected --
// including case variants, which this parser deliberately does not accept.
func FuzzParseTimeRange_Oracle(f *testing.F) {

	seeds := append([]string{"", "garbage", "TODAY", "Today", " today", "today ", "past-31-days"}, timeRangeNames...)
	for _, seed := range seeds {
		f.Add(seed)
	}

	recognized := make(map[string]bool, len(timeRangeNames))
	for _, name := range timeRangeNames {
		recognized[name] = true
	}

	f.Fuzz(func(t *testing.T, value string) {

		begin, end := parseTimeRange(value)

		// RULE: A value outside the known list produces no range at all.
		if !recognized[value] {
			require.True(t, begin.IsZero(), "unexpected range for %q", value)
			require.True(t, end.IsZero(), "unexpected range for %q", value)
			return
		}

		// Every known range is ordered, half-open, and non-empty.
		require.False(t, begin.IsZero(), "%q should have produced a range", value)
		require.True(t, begin.Before(end), "%q produced an inverted range", value)

		// Both ends are anchored to midnight UTC, so that a range never depends on
		// the server's local timezone.
		for _, moment := range []time.Time{begin, end} {
			require.Equal(t, time.UTC, moment.Location(), "%q is not anchored to UTC", value)
			require.Equal(t, 0, moment.Hour(), "%q is not anchored to midnight", value)
			require.Equal(t, 0, moment.Minute(), "%q is not anchored to midnight", value)
			require.Equal(t, 0, moment.Second(), "%q is not anchored to midnight", value)
			require.Equal(t, 0, moment.Nanosecond(), "%q is not anchored to midnight", value)
		}

		// No named range covers more than a leap year.
		require.LessOrEqual(t, end.Sub(begin), 366*24*time.Hour, "%q spans more than a year", value)
	})
}

/******************************************
 * Robustness
 ******************************************/

// FuzzEvaluateField_NeverPanics runs one raw value through every data type's
// parser. This is the blunt target: the others describe what the output must
// look like, and this one insists there is an output at all.
func FuzzEvaluateField_NeverPanics(f *testing.F) {

	seeds := []string{
		"", "123", "lt:5", "MIN", "MAX", "true", "FALSE",
		"123456781234567812345678", "not-an-id",
		"2020-01-02", "past-30-days", "garbage",
		"1,2,3,4", "lt:1,2", ":", "EQ:", "ne:",
		"\x00\x01\x02", "\xff", "--------", "9,9,9,9,9,9,9,9,9",
	}
	for _, seed := range seeds {
		f.Add(seed)
	}

	b := NewBuilder()

	f.Fuzz(func(t *testing.T, value string) {

		for _, dataType := range everyDataType {

			field := NewField("field", dataType)

			result := b.EvaluateField(field, []string{value})
			require.NotNil(t, result, "dataType=%q value=%q", dataType, value)

			// The same value repeated must never produce more predicates than the
			// values that went in -- a time range contributes two, and nothing more.
			repeated := b.EvaluateField(field, []string{value, value, value})
			require.LessOrEqual(t, len(predicates(repeated)), 6,
				"dataType=%q value=%q produced too many predicates", dataType, value)
		}
	})
}

// FuzzBuilderEvaluate_NeverPanics throws arbitrary query strings at a Builder
// holding every supported data type, and insists that both entry points survive.
func FuzzBuilderEvaluate_NeverPanics(f *testing.F) {

	for _, seed := range queryStringSeeds {
		f.Add(seed)
	}

	b := fuzzBuilder()

	f.Fuzz(func(t *testing.T, query string) {

		values, err := url.ParseQuery(query)
		if err != nil {
			return
		}

		result := b.Evaluate(values)
		require.NotNil(t, result)

		// Fields() walks the whole tree, so it will find a nil hiding in there.
		require.NotNil(t, result.Fields())

		_, _ = b.EvaluateAll(values)
		_ = b.HasURLParams(values)
	})
}

// Fuzz me once, shame on you
