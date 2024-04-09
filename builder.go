// Package builder generates dynamic expressions based on a set of url.Values and a map of allowed fields.
// It supports several different data types, along with a simple list of operators that can be applied to
// each predicate.
package builderer

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/benpate/derp"
	"github.com/benpate/exp"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Builder map[string]Field

func NewBuilder() Builder {
	return make(Builder)
}

// String adds a string-based parameter to the expression Builder
func (b Builder) String(name string, options ...FieldOption) Builder {
	b[name] = NewField(name, DataTypeString, options...)
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

// Bool adds a boolean-based parameter to the expression Builder
func (b Builder) Bool(name string, options ...FieldOption) Builder {
	b[name] = NewField(name, DataTypeBool, options...)
	return b
}

// ObjectID adds a mongodb ObjectID-based parameter to the expression Builder
func (b Builder) ObjectID(name string, options ...FieldOption) Builder {
	b[name] = NewField(name, DataTypeObjectID, options...)
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

		return exp.Empty(), derp.NewBadRequestError("builder.MissingField", "Missing required field", field)
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

		var err error
		var value any

		switch field.DataType {

		case DataTypeString:
			value = stringValue

		case DataTypeBool:

			switch strings.ToLower(stringValue) {
			case "true":
				value = true
			case "false":
				value = false
			default:
				// Unrecognized values are skipped.
				continue
			}

		case DataTypeInt:
			value, err = strconv.Atoi(stringValue)

			// If this is not a valid Integer, then skip this parameter
			if err != nil {
				continue
			}

		case DataTypeInt64:
			value, err = strconv.ParseInt(stringValue, 10, 64)

			// If this is not a valid Integer, then skip this parameter
			if err != nil {
				continue
			}

		case DataTypeObjectID:
			value, err = primitive.ObjectIDFromHex(stringValue)

			// If this is not a valid ObjectID, then skip this parameter
			if err != nil {
				continue
			}

		default:
			// Unrecognized Types are skipped.  How did you even do this?
			continue
		}

		result = result.Or(exp.New(field.Name, operator, value))
	}

	return result
}

// parseValue parses a single string into an operator and a value, using the form
// "OP:Value" -- so the string "EQ:John" will return ("=", "John").
// If a valid operator is defined in the input, then it is returned along
// with the remaining string as the criteria value.
// Otherwise, the defaultOperator is used.
func parseValue(input string, defaultOperator string) (string, string) {

	if len(input) > 0 {

		// If the input contains a colon, then split it into OPERATOR and VALUE
		if operator, value, found := strings.Cut(input, ":"); found {

			if operator, ok := exp.OperatorOk(operator); ok {
				return operator, value
			}
		}

		// Otherwise, use the default operator argument
		return defaultOperator, input
	}

	return "", ""
}

func sliceNotEmpty(slice []string) bool {

	for _, value := range slice {
		if len(value) > 0 {
			return true
		}
	}

	return false
}
