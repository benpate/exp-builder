// Package builder generates dynamic expressions based on a set of url.Values and a map of allowed fields.
// It supports several different data types, along with a simple list of operators that can be applied to
// each predicate.
package builder

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/benpate/derp"
	"github.com/benpate/exp"
	"github.com/benpate/rosetta/convert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Builder map[string]Field

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

// Int64 adds an 64-bit integer-based parameter to the expression Builder
func (b Builder) Int64(name string, options ...FieldOption) Builder {
	b[name] = NewField(name, DataTypeInt64, options...)
	return b
}

// Location adds a location-based parameter to the expression Builder
func (b Builder) Location(name string, options ...FieldOption) Builder {
	b[name] = NewField(name, DataTypeLocation, options...)
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

// Time adds a stime-based parameter to the expression Builder
func (b Builder) Time(name string, options ...FieldOption) Builder {
	b[name] = NewField(name, DataTypeTime, options...)
	return b
}

// Evaluate returns an Expression based on the specific url.Values provided
func (b Builder) Evaluate(values url.Values) exp.Expression {

	result := exp.Empty()

	for name, field := range b {

		if value, ok := values[name]; ok {
			if sliceNotEmpty(value) {
				result = result.And(b.EvaluateField(field, value))
			}
		}
	}

	return result
}

// EvaluateAll returns an Expression that requires ALL of the values in the builder are present.
func (b Builder) EvaluateAll(values url.Values) (exp.Expression, error) {

	result := exp.Empty()

	for name, field := range b {

		if value, ok := values[name]; ok {
			if sliceNotEmpty(value) {
				result = result.And(b.EvaluateField(field, value))
				continue
			}
		}

		return exp.Empty(), derp.BadRequestError("builder.MissingField", "Missing required field", field)
	}

	return result, nil
}

// HasURLParams returns TRUE if the URL contains any parameters that match the Builder.
// It does not test the validity of those values.
func (b Builder) HasURLParams(values url.Values) bool {

	for field := range b {
		if _, ok := values[field]; ok {
			return true
		}
	}

	return false
}

func (b Builder) EvaluateField(field Field, values []string) exp.Expression {

	result := exp.Empty()

	for _, input := range values {

		operator, stringValue := parseValue(input, field.Operator)
		operator = exp.Operator(operator)

		// Apply filters to input before comparing data types
		for _, filter := range field.Filters {
			stringValue = filter(stringValue)
		}

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

			if value, err := strconv.Atoi(stringValue); err == nil {
				result = result.Or(exp.New(field.Name, operator, value))
			}

		case DataTypeInt64:

			if value, err := strconv.ParseInt(stringValue, 10, 64); err == nil {
				result = result.Or(exp.New(field.Name, operator, value))
			}

		case DataTypeLocation:

			/*/ Try to parse time range statements
			if geoPoint := parseGeoPoint(input); geoPoint.NotZero() {
				result = result.Or(exp.New(field.Name, ">=", beginDate).And(exp.New(field.Name, "<", endDate)))
				continue
			}*/

		case DataTypeObjectID:

			if value, err := primitive.ObjectIDFromHex(stringValue); err == nil {
				result = result.Or(exp.New(field.Name, operator, value))
			}

		case DataTypeTime:

			// Try to parse time range statements
			if beginDate, endDate := parseTimeRange(input); !beginDate.IsZero() {
				result = result.Or(exp.New(field.Name, ">=", beginDate).And(exp.New(field.Name, "<", endDate)))
				continue
			}

			// Otherwise, parse individual time values
			if value := convert.Time(stringValue); !value.IsZero() {
				result = result.Or(exp.New(field.Name, operator, value))
			}
		}
	}

	return result
}
