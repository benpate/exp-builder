// Package builder generates dynamic expressions based on a set of url.Values and a map of allowed fields.
// It supports several different data types, along with a simple list of operators that can be applied to
// each predicate.
package builder

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/benpate/exp"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Builder map[string]string

func NewBuilder() Builder {
	return make(Builder)
}

func (b Builder) String(name string) Builder {
	b[name] = DataTypeString
	return b
}

func (b Builder) Int(name string) Builder {
	b[name] = DataTypeInt
	return b
}

func (b Builder) Bool(name string) Builder {
	b[name] = DataTypeBool
	return b
}

func (b Builder) ObjectID(name string) Builder {
	b[name] = DataTypeObjectID
	return b
}

func (b Builder) Evaluate(values url.Values) exp.Expression {

	result := exp.Empty()

	for field, dataType := range b {

		if value, ok := values[field]; ok {
			result = result.And(b.evaluateField(field, dataType, value))
		}
	}

	return result
}

func (b Builder) evaluateField(field string, dataType string, values []string) exp.Expression {

	result := exp.Empty()

	for _, input := range values {

		operator, stringValue := parseValue(input)
		operator = exp.Operator(operator)

		var err error
		var value interface{}

		switch dataType {
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

		result = result.Or(exp.New(field, operator, value))

	}

	return result
}

func parseValue(input string) (string, string) {

	var value string
	var operator string

	if len(input) > 0 {
		inputSlice := strings.Split(input, ":")

		switch len(inputSlice) {
		case 0:
		case 1:
			operator = exp.OperatorEqual
			value = inputSlice[0]
		default:
			operator = inputSlice[0]
			value = inputSlice[1]
		}
	}

	return operator, value
}
