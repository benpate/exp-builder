package builder

import (
	"maps"
	"math"
	"net/url"
	"slices"
	"strconv"
	"strings"

	"github.com/benpate/derp"
	"github.com/benpate/exp"
	"github.com/benpate/geo"
	"github.com/benpate/rosetta/convert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Builder maps URL parameter names to the Fields that describe how to turn
// their values into an expression.
type Builder map[string]Field

// NewBuilder returns an empty Builder.
func NewBuilder() Builder {
	return make(Builder)
}

// Bool adds a boolean-based parameter to the expression Builder
func (b Builder) Bool(name string, options ...FieldOption) Builder {
	b[name] = NewField(name, DataTypeBool, options...)
	return b
}

// Int adds an integer-based parameter to the expression Builder
func (b Builder) Int(name string, options ...FieldOption) Builder {
	b[name] = NewField(name, DataTypeInt, options...)
	return b
}

// Int64 adds a 64-bit integer-based parameter to the expression Builder
func (b Builder) Int64(name string, options ...FieldOption) Builder {
	b[name] = NewField(name, DataTypeInt64, options...)
	return b
}

// ObjectID adds a mongodb ObjectID-based parameter to the expression Builder
func (b Builder) ObjectID(name string, options ...FieldOption) Builder {
	b[name] = NewField(name, DataTypeObjectID, options...)
	return b
}

// String adds a string-based parameter to the expression Builder
func (b Builder) String(name string, options ...FieldOption) Builder {
	b[name] = NewField(name, DataTypeString, options...)
	return b
}

// Time adds a time-based parameter to the expression Builder
func (b Builder) Time(name string, options ...FieldOption) Builder {
	b[name] = NewField(name, DataTypeTime, options...)
	return b
}

// Polygon adds a GeoJSON polygon-based parameter to the expression Builder
func (b Builder) Polygon(name string, options ...FieldOption) Builder {
	b[name] = NewField(name, DataTypePolygon, options...)
	return b
}

// Evaluate returns an Expression built from the url.Values provided, using only
// the parameters that this Builder declares. Everything else is ignored.
func (b Builder) Evaluate(values url.Values) exp.Expression {

	var result exp.Expression = exp.Empty()

	// Fields are visited in name order so that the same URL always produces the
	// same expression.  Ranging a map directly would shuffle the predicates on
	// every call, which changes the query shape the database sees.
	for _, name := range slices.Sorted(maps.Keys(b)) {

		if value, ok := values[name]; ok {
			if sliceNotEmpty(value) {
				result = result.And(b.EvaluateField(b[name], value))
			}
		}
	}

	// Nothing but the parameters we asked for
	return result
}

// EvaluateAll is like Evaluate, but every field defined in the Builder must be
// present and non-empty in the URL values. It returns an error if any is missing.
func (b Builder) EvaluateAll(values url.Values) (exp.Expression, error) {

	const location = "builder.Builder.EvaluateAll"

	var result exp.Expression = exp.Empty()

	// Sorted for the same reason as Evaluate, and so that a request missing
	// several fields always reports the same one instead of a random pick.
	for _, name := range slices.Sorted(maps.Keys(b)) {

		if value, ok := values[name]; ok {
			if sliceNotEmpty(value) {
				result = result.And(b.EvaluateField(b[name], value))
				continue
			}
		}

		// The detail is the field NAME, not the Field: a Field carries filter
		// functions, which cannot be marshaled into a JSON error response.
		return exp.Empty(), derp.BadRequest(location, "Missing required field", name)
	}

	// All present and accounted for
	return result, nil
}

// HasURLParams returns TRUE if the URL contains any parameters that match the
// Builder. It tests only for the parameter's presence, so a blank or unparseable
// value still counts here while Evaluate would discard it.
func (b Builder) HasURLParams(values url.Values) bool {

	for field := range b {
		if _, ok := values[field]; ok {
			return true
		}
	}

	// Nobody here but us chickens
	return false
}

// EvaluateField converts the raw URL values for a single field into an
// expression, parsing each value according to the field's data type.
func (b Builder) EvaluateField(field Field, values []string) exp.Expression {

	var result exp.Expression = exp.Empty()

	for _, input := range values {

		// Split an "OP:" prefix off the front of the value
		operator, stringValue := parseValue(input, field.Operator)
		operator = exp.Operator(operator)

		// Apply filters to input before comparing data types
		for _, filter := range field.Filters {
			stringValue = filter(stringValue)
		}

		// An empty value is a legitimate comparison: a field that may be unset is
		// queried as (field == "") OR (field == "VALUE").  See BUG-131 for the
		// case AGAINST this -- a filter that empties a value turns "CONTAINS" into
		// a match on every record -- which is unresolved, and deliberately not
		// guarded against here.
		//
		// Convert the value into the field's data type, and add it to the result.
		// A value that will not convert is dropped without complaint.
		switch field.DataType {

		case DataTypeString:

			result = result.Or(exp.New(field.Name, operator, stringValue))

		case DataTypeBool:

			switch strings.ToLower(stringValue) {

			case "true":
				result = result.Or(exp.New(field.Name, operator, true))
				continue

			case "false":
				result = result.Or(exp.New(field.Name, operator, false))
				continue
			}

		case DataTypeInt:

			// Handle magic values
			switch stringValue {

			case "MIN":
				result = result.Or(exp.New(field.Name, operator, math.MinInt))
				continue
			case "MAX":
				result = result.Or(exp.New(field.Name, operator, math.MaxInt))
				continue
			}

			// Otherwise, try to parse the value as an integer
			if value, err := strconv.Atoi(stringValue); err == nil {
				result = result.Or(exp.New(field.Name, operator, value))
			}

		case DataTypeInt64:

			// Handle magic values. The explicit int64 conversions keep these
			// consistent with the strconv.ParseInt path below (and avoid an
			// int overflow on 32-bit platforms).
			switch stringValue {

			case "MIN":
				result = result.Or(exp.New(field.Name, operator, int64(math.MinInt64)))
				continue
			case "MAX":
				result = result.Or(exp.New(field.Name, operator, int64(math.MaxInt64)))
				continue
			}

			// Otherwise, try to parse the value as an int64
			if value, err := strconv.ParseInt(stringValue, 10, 64); err == nil {
				result = result.Or(exp.New(field.Name, operator, value))
				continue
			}

		case DataTypePolygon:

			// Parse the operator-stripped stringValue (consistent with every other
			// data type) so an "OP:" prefix never leaks into the coordinate parser.
			if polygon := geo.NewPolygonFromString(stringValue); polygon.NotZero() {
				result = result.Or(exp.New(field.Name, exp.OperatorGeoWithin, polygon))
				continue
			}

		case DataTypeObjectID:

			if value, err := primitive.ObjectIDFromHex(stringValue); err == nil {
				result = result.Or(exp.New(field.Name, operator, value))
				continue
			}

		case DataTypeTime:

			// Try to parse time range statements.  A named range supplies its own
			// begin/end comparisons, so -- exactly like a Polygon -- any operator
			// prefix is stripped first and then discarded.
			if beginDate, endDate := parseTimeRange(stringValue); !beginDate.IsZero() {
				result = result.Or(exp.New(field.Name, exp.OperatorGreaterOrEqual, beginDate).And(exp.New(field.Name, exp.OperatorLessThan, endDate)))
				continue
			}

			// Otherwise, parse individual time values
			if value := convert.Time(stringValue); !value.IsZero() {
				result = result.Or(exp.New(field.Name, operator, value))
				continue
			}
		}
	}

	// One field, however many values it brought along
	return result
}
