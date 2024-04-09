package builderer

import "github.com/benpate/exp"

type Field struct {
	Name     string
	DataType string
	Operator string
}

func NewField(name string, dataType string, options ...FieldOption) Field {
	result := Field{
		Name:     name,
		DataType: dataType,
		Operator: exp.OperatorEqual,
	}

	for _, option := range options {
		option(&result)
	}

	return result
}
