package builder

import (
	"math"
	"net/url"
	"testing"

	"github.com/benpate/exp"
	"github.com/benpate/geo"
	"github.com/benpate/rosetta/convert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestBuilder_Strings(t *testing.T) {

	b := NewBuilder().
		ObjectID("parentId").
		String("templateId").
		String("firstName").
		String("lastName").
		Int("publishDate").
		Bool("isPublished")

	u, _ := url.ParseQuery("firstName=John&firstName=Sara&firstName=ne:Kyle")
	expect := exp.OrExpression{
		exp.Predicate{Field: "firstName", Operator: "=", Value: "John"},
		exp.Predicate{Field: "firstName", Operator: "=", Value: "Sara"},
		exp.Predicate{Field: "firstName", Operator: "!=", Value: "Kyle"},
	}
	require.Equal(t, expect, b.Evaluate(u))
}

func TestBuilder_ObjectID(t *testing.T) {

	objectID, _ := primitive.ObjectIDFromHex("123456781234567812345678")

	b := NewBuilder().
		ObjectID("parentId").
		String("templateId").
		String("firstName").
		String("lastName").
		Int("publishDate").
		Bool("isPublished")

	value1, _ := url.ParseQuery("parentId=123456781234567812345678")
	expect1 := exp.Predicate{Field: "parentId", Operator: "=", Value: objectID}
	require.Equal(t, expect1, b.Evaluate(value1))

	value2, _ := url.ParseQuery("parentId=Not-An-ObjectID")
	expect2 := exp.EmptyExpression{}
	require.Equal(t, expect2, b.Evaluate(value2))
}

func TestBuilder_Int(t *testing.T) {

	b := NewBuilder().
		ObjectID("parentId").
		String("templateId").
		String("firstName").
		String("lastName").
		Int("publishDate").
		Bool("isPublished")

	{
		u, _ := url.ParseQuery("publishDate=123")
		expect := exp.Predicate{Field: "publishDate", Operator: "=", Value: 123}
		require.Equal(t, expect, b.Evaluate(u))
	}

	{
		u, _ := url.ParseQuery("publishDate=lt:123")
		expect := exp.Predicate{Field: "publishDate", Operator: "<", Value: 123}
		require.Equal(t, expect, b.Evaluate(u))
	}

	{
		u, _ := url.ParseQuery("publishDate=ge:123")
		expect := exp.Predicate{Field: "publishDate", Operator: ">=", Value: 123}
		require.Equal(t, expect, b.Evaluate(u))
	}

	{
		u, _ := url.ParseQuery("publishDate=Not-A-Number")
		expect := exp.EmptyExpression{}
		require.Equal(t, expect, b.Evaluate(u))
	}
}

func TestBuilder_Int64(t *testing.T) {

	b := NewBuilder().
		Int64("publishDate")

	{
		u, _ := url.ParseQuery("publishDate=123")
		expect := exp.Predicate{Field: "publishDate", Operator: "=", Value: int64(123)}
		require.Equal(t, expect, b.Evaluate(u))
	}

	{
		u, _ := url.ParseQuery("publishDate=lt:123")
		expect := exp.Predicate{Field: "publishDate", Operator: "<", Value: int64(123)}
		require.Equal(t, expect, b.Evaluate(u))
	}

	{
		u, _ := url.ParseQuery("publishDate=ge:123")
		expect := exp.Predicate{Field: "publishDate", Operator: ">=", Value: int64(123)}
		require.Equal(t, expect, b.Evaluate(u))
	}

	{
		u, _ := url.ParseQuery("publishDate=Not-A-Number")
		expect := exp.EmptyExpression{}
		require.Equal(t, expect, b.Evaluate(u))
	}
}

func TestBuilder_Bool(t *testing.T) {

	b := NewBuilder().
		ObjectID("parentId").
		String("templateId").
		String("firstName").
		String("lastName").
		Int("publishDate").
		Bool("isPublished")

	{
		u, _ := url.ParseQuery("isPublished=true")
		expect := exp.Predicate{Field: "isPublished", Operator: "=", Value: true}
		require.Equal(t, expect, b.Evaluate(u))
	}

	{
		u, _ := url.ParseQuery("isPublished=false")
		expect := exp.Predicate{Field: "isPublished", Operator: "=", Value: false}
		require.Equal(t, expect, b.Evaluate(u))
	}

	{
		u, _ := url.ParseQuery("isPublished=yep")
		expect := exp.EmptyExpression{}
		require.Equal(t, expect, b.Evaluate(u))
	}
}

func TestBuilder_HasURLParams(t *testing.T) {

	b := NewBuilder().
		ObjectID("parentId").
		String("templateId").
		String("firstName").
		String("lastName").
		Int("publishDate").
		Bool("isPublished")

	{
		u, _ := url.ParseQuery("isPublished=true")
		require.True(t, b.HasURLParams(u))
	}

	{
		u, _ := url.ParseQuery("isPublished=InvalidValue")
		require.True(t, b.HasURLParams(u))
	}

	{
		u, _ := url.ParseQuery("NotValid=totally")
		require.False(t, b.HasURLParams(u))
	}

}

func TestBuilder_EvaluateAll(t *testing.T) {

	b := NewBuilder().
		String("firstName")

	{
		u, _ := url.ParseQuery("firstName=John&lastName=Connor")

		expect := exp.Predicate{Field: "firstName", Operator: "=", Value: "John"}

		result, err := b.EvaluateAll(u)
		require.Nil(t, err)
		require.Equal(t, expect, result)
	}

	{
		u, _ := url.ParseQuery("")

		expect := exp.EmptyExpression{}

		result, err := b.EvaluateAll(u)
		require.NotNil(t, err)
		require.Equal(t, expect, result)
	}

	{
		u, _ := url.ParseQuery("lastName=Connor")

		expect := exp.EmptyExpression{}

		result, err := b.EvaluateAll(u)
		require.NotNil(t, err)
		require.Equal(t, expect, result)
	}
}

func TestBuilder_OMGThisSucksButImGonnaTestTheInvalidState(t *testing.T) {

	// Seriously, kids, this should never happen.  Why would you even do this?
	b := NewBuilder()
	b["firstName"] = NewField("firstName", "INVALID_VALUE")

	u, _ := url.ParseQuery("firstName=John&lastName=Connor")

	require.Equal(t, exp.EmptyExpression{}, b.Evaluate(u))
}

func TestParseValue(t *testing.T) {

	{
		operator, value := parseValue("", "=")
		require.Equal(t, "", operator)
		require.Equal(t, "", value)
	}

	{
		operator, value := parseValue("SOME Value", "=")
		require.Equal(t, "=", operator)
		require.Equal(t, "SOME Value", value)
	}

	{
		operator, value := parseValue("GT:7", "=")
		require.Equal(t, ">", operator)
		require.Equal(t, "7", value)
	}

	{
		operator, value := parseValue("GTE:7", "=")
		require.Equal(t, ">=", operator)
		require.Equal(t, "7", value)
	}
}

func TestParseValue_FailedOperatorConversion(t *testing.T) {

	operator, value := parseValue("GTE:7", "=")
	require.Equal(t, ">=", operator)
	require.Equal(t, "7", value)
}

func TestEvaluateTime(t *testing.T) {

	queryString, err := url.Parse("http://test.com?timeValue=past-365-days")
	require.Nil(t, err)

	b := NewBuilder().Time("timeValue")
	result := b.Evaluate(queryString.Query())

	require.IsType(t, exp.AndExpression{}, result)
	require.Equal(t, 2, len(result.(exp.AndExpression)))
}

func TestEvaluateTime_IndividualValue(t *testing.T) {

	b := NewBuilder().Time("timeValue")

	u, _ := url.ParseQuery("timeValue=2020-01-02")
	result := b.Evaluate(u)

	// A single date (not a range) produces a single equality predicate.
	expect := exp.Predicate{Field: "timeValue", Operator: "=", Value: convert.Time("2020-01-02")}
	require.Equal(t, expect, result)
}

func TestEvaluateTime_Invalid(t *testing.T) {

	b := NewBuilder().Time("timeValue")

	u, _ := url.ParseQuery("timeValue=not-a-date")
	require.Equal(t, exp.EmptyExpression{}, b.Evaluate(u))
}

func TestBuilder_Int_MagicValues(t *testing.T) {

	b := NewBuilder().Int("n")

	{
		u, _ := url.ParseQuery("n=MIN")
		expect := exp.Predicate{Field: "n", Operator: "=", Value: math.MinInt}
		require.Equal(t, expect, b.Evaluate(u))
	}

	{
		u, _ := url.ParseQuery("n=MAX")
		expect := exp.Predicate{Field: "n", Operator: "=", Value: math.MaxInt}
		require.Equal(t, expect, b.Evaluate(u))
	}
}

func TestBuilder_Int64_MagicValues(t *testing.T) {

	b := NewBuilder().Int64("n")

	// The Int64 "MIN"/"MAX" magic values are emitted as int64, consistent with
	// the normal strconv.ParseInt path.
	{
		u, _ := url.ParseQuery("n=MIN")
		expect := exp.Predicate{Field: "n", Operator: "=", Value: int64(math.MinInt64)}
		require.Equal(t, expect, b.Evaluate(u))
	}

	{
		u, _ := url.ParseQuery("n=MAX")
		expect := exp.Predicate{Field: "n", Operator: "=", Value: int64(math.MaxInt64)}
		require.Equal(t, expect, b.Evaluate(u))
	}
}

func TestBuilder_Polygon(t *testing.T) {

	b := NewBuilder().Polygon("location")

	u, _ := url.ParseQuery("location=1,2,3,4")
	result := b.Evaluate(u)

	// A valid coordinate list produces a single GEO-WITHIN predicate.
	expect := exp.Predicate{
		Field:    "location",
		Operator: exp.OperatorGeoWithin,
		Value:    geo.NewPolygonFromString("1,2,3,4"),
	}
	require.Equal(t, expect, result)
}

func TestBuilder_Polygon_Invalid(t *testing.T) {

	b := NewBuilder().Polygon("location")

	u, _ := url.ParseQuery("location=not-coordinates")
	require.Equal(t, exp.EmptyExpression{}, b.Evaluate(u))
}

func TestBuilder_Evaluate_MultipleFields(t *testing.T) {

	b := NewBuilder().
		String("firstName").
		Int("age")

	u, _ := url.ParseQuery("firstName=John&age=42")
	result := b.Evaluate(u)

	// Two distinct present fields are combined with AND. Map iteration order is
	// random, so assert on membership rather than position.
	and, ok := result.(exp.AndExpression)
	require.True(t, ok, "expected AndExpression, got %T", result)
	require.Equal(t, 2, len(and))
	require.Contains(t, and, exp.Predicate{Field: "firstName", Operator: "=", Value: "John"})
	require.Contains(t, and, exp.Predicate{Field: "age", Operator: "=", Value: 42})
}

func TestBuilder_Evaluate_SkipsAbsentAndEmpty(t *testing.T) {

	b := NewBuilder().
		String("firstName").
		String("lastName")

	// "lastName" is present but empty, "firstName" is present with a value, and
	// "unknown" is not in the Builder. Only "firstName" contributes a predicate.
	u, _ := url.ParseQuery("firstName=John&lastName=&unknown=x")
	result := b.Evaluate(u)

	expect := exp.Predicate{Field: "firstName", Operator: "=", Value: "John"}
	require.Equal(t, expect, result)
}

func TestBuilder_Evaluate_Empty(t *testing.T) {

	// A Builder with no matching URL params returns the empty expression.
	b := NewBuilder().String("firstName")

	u, _ := url.ParseQuery("other=value")
	require.Equal(t, exp.EmptyExpression{}, b.Evaluate(u))
}

func TestBuilder_EvaluateField_MultipleValues_Or(t *testing.T) {

	// Multiple values for a single Int field combine with OR (each value parsed
	// independently). This exercises the Or-combination path for a non-string type.
	b := NewBuilder().Int("age")

	u, _ := url.ParseQuery("age=10&age=20&age=lt:5")
	result := b.Evaluate(u)

	expect := exp.OrExpression{
		exp.Predicate{Field: "age", Operator: "=", Value: 10},
		exp.Predicate{Field: "age", Operator: "=", Value: 20},
		exp.Predicate{Field: "age", Operator: "<", Value: 5},
	}
	require.Equal(t, expect, result)
}

func TestBuilder_EvaluateField_MixedValidInvalid(t *testing.T) {

	// When one value parses and another does not, only the valid one produces a
	// predicate -- collapsing to a single Predicate rather than an OrExpression.
	b := NewBuilder().Int("age")

	u, _ := url.ParseQuery("age=42&age=not-a-number")
	result := b.Evaluate(u)

	expect := exp.Predicate{Field: "age", Operator: "=", Value: 42}
	require.Equal(t, expect, result)
}

func TestBuilder_Bool_CaseInsensitive(t *testing.T) {

	// Bool parsing lower-cases the input, so mixed/upper case still matches.
	b := NewBuilder().Bool("flag")

	run := func(query string, expected bool) {
		u, _ := url.ParseQuery(query)
		expect := exp.Predicate{Field: "flag", Operator: "=", Value: expected}
		require.Equal(t, expect, b.Evaluate(u), query)
	}

	run("flag=TRUE", true)
	run("flag=True", true)
	run("flag=FALSE", false)
	run("flag=False", false)
}

func TestBuilder_Polygon_StripsOperatorPrefix(t *testing.T) {

	// An "OP:" prefix is stripped before the coordinates are parsed (consistent
	// with every other data type), and Polygon always emits OperatorGeoWithin.
	// So "eq:1,2,3,4" parses the same coordinates as "1,2,3,4".
	b := NewBuilder().Polygon("location")

	u, _ := url.ParseQuery("location=eq:1,2,3,4")
	result := b.Evaluate(u)

	expect := exp.Predicate{
		Field:    "location",
		Operator: exp.OperatorGeoWithin,
		Value:    geo.NewPolygonFromString("1,2,3,4"),
	}
	require.Equal(t, expect, result)
}

func TestBuilder_ObjectID_MultipleValues_Or(t *testing.T) {

	id1, _ := primitive.ObjectIDFromHex("123456781234567812345678")
	id2, _ := primitive.ObjectIDFromHex("aabbccddeeff001122334455")

	b := NewBuilder().ObjectID("parentId")

	u, _ := url.ParseQuery("parentId=123456781234567812345678&parentId=aabbccddeeff001122334455")
	result := b.Evaluate(u)

	expect := exp.OrExpression{
		exp.Predicate{Field: "parentId", Operator: "=", Value: id1},
		exp.Predicate{Field: "parentId", Operator: "=", Value: id2},
	}
	require.Equal(t, expect, result)
}

func TestBuilder_EvaluateAll_MultipleRequiredFields(t *testing.T) {

	b := NewBuilder().
		String("firstName").
		String("lastName")

	// All required fields present: the result ANDs both predicates.
	u, _ := url.ParseQuery("firstName=John&lastName=Connor")
	result, err := b.EvaluateAll(u)
	require.Nil(t, err)

	and, ok := result.(exp.AndExpression)
	require.True(t, ok, "expected AndExpression, got %T", result)
	require.Equal(t, 2, len(and))
	require.Contains(t, and, exp.Predicate{Field: "firstName", Operator: "=", Value: "John"})
	require.Contains(t, and, exp.Predicate{Field: "lastName", Operator: "=", Value: "Connor"})
}

func TestBuilder_EvaluateAll_PresentButEmptyFails(t *testing.T) {

	// A required field that is present but empty fails EvaluateAll, just like an
	// absent one.
	b := NewBuilder().String("firstName")

	u, _ := url.ParseQuery("firstName=")
	result, err := b.EvaluateAll(u)
	require.NotNil(t, err)
	require.Equal(t, exp.EmptyExpression{}, result)
}

// FuzzBuilderEvaluate throws arbitrary query strings at a Builder containing
// every supported data type. The parser must never panic, and Evaluate must
// always return a usable (non-nil) Expression.
func FuzzBuilderEvaluate(f *testing.F) {

	seeds := []string{
		"firstName=John",
		"firstName=a&firstName=b&publishDate=ne:7",
		"publishDate=123",
		"publishDate=MIN",
		"publishDate=lt:5",
		"bignum=MAX",
		"isPublished=true",
		"parentId=123456781234567812345678",
		"timeValue=past-30-days",
		"timeValue=2020-01-01",
		"location=1,2,3,4",
		"",
		"=",
		"&&",
	}
	for _, seed := range seeds {
		f.Add(seed)
	}

	b := NewBuilder().
		String("firstName").
		Int("publishDate").
		Int64("bignum").
		Bool("isPublished").
		ObjectID("parentId").
		Time("timeValue").
		Polygon("location")

	f.Fuzz(func(t *testing.T, query string) {

		values, err := url.ParseQuery(query)
		if err != nil {
			return // Ignore inputs that are not valid query strings.
		}

		// Evaluate must never panic and always returns a usable Expression.
		result := b.Evaluate(values)
		require.NotNil(t, result)

		// EvaluateAll walks the same parsing paths and must also never panic.
		_, _ = b.EvaluateAll(values)
	})
}

// FuzzEvaluateField feeds arbitrary raw values into every data type's parsing
// path. Each parser (int, int64, bool, ObjectID, time, polygon) must tolerate
// hostile input without panicking, and EvaluateField must always return a
// non-nil Expression.
func FuzzEvaluateField(f *testing.F) {

	seeds := []string{
		"", "123", "lt:5", "MIN", "MAX", "true", "FALSE",
		"123456781234567812345678", "not-an-id",
		"2020-01-02", "past-30-days", "garbage",
		"1,2,3,4", "lt:1,2", ":", "EQ:", "ne:",
	}
	for _, seed := range seeds {
		f.Add(seed)
	}

	dataTypes := []string{
		DataTypeString, DataTypeInt, DataTypeInt64, DataTypeBool,
		DataTypeObjectID, DataTypeTime, DataTypePolygon,
	}

	b := NewBuilder()

	f.Fuzz(func(t *testing.T, value string) {

		// Run the same raw value through each data type's parser.
		for _, dataType := range dataTypes {
			field := NewField("field", dataType)

			// Must never panic and must always return a usable Expression.
			result := b.EvaluateField(field, []string{value})
			require.NotNil(t, result, "dataType=%q value=%q", dataType, value)
		}
	})
}
