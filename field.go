package builder

import "github.com/benpate/exp"

// Field describes a single queryable parameter: the expression field name, its
// data type, the default comparison operator, and any input filters.
type Field struct {
	Name     string
	DataType string
	Operator string
	Filters  []Filter
}

// NewField returns a Field for the given name and data type, applying any
// options in order.
func NewField(name string, dataType string, options ...FieldOption) Field {
	result := Field{
		Name:     name,
		DataType: dataType,
		Operator: exp.OperatorEqual,
		Filters:  make([]Filter, 0),
	}

	for _, option := range options {
		option(&result)
	}

	return result
}
