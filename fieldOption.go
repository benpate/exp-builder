package builder

import "github.com/benpate/exp"

type FieldOption func(*Field)

// WithDefaultOpEqual defines "EQUAL" as the default operator to use when creating an expression.
func WithDefaultOpEqual() FieldOption {
	return WithDefaultOperator(exp.OperatorEqual)
}

// WithDefaultOpEqual defines "CONTAINS" as the default operator to use when creating an expression.
func WithDefaultOpContains() FieldOption {
	return WithDefaultOperator(exp.OperatorContains)
}

// WithDefaultOpBeginsWith defines "BEGINS" as the default operator to use when creating an expression.
func WithDefaultOpBeginsWith() FieldOption {
	return WithDefaultOperator(exp.OperatorBeginsWith)
}

// WithDefaultOperator defines the default operator to use when creating an expression
func WithDefaultOperator(operator string) FieldOption {
	return func(f *Field) {
		f.Operator = operator
	}
}

// WithAlias defines the field name to use when creating an expression,
// which may be different from the field name to read from the URL.
func WithAlias(name string) FieldOption {
	return func(f *Field) {
		f.Name = name
	}
}
