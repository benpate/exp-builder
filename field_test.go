package builder

import (
	"testing"

	"github.com/benpate/exp"
	"github.com/stretchr/testify/require"
)

func TestNewField_Defaults(t *testing.T) {
	field := NewField("firstName", DataTypeString)

	require.Equal(t, "firstName", field.Name)
	require.Equal(t, DataTypeString, field.DataType)
	require.Equal(t, exp.OperatorEqual, field.Operator)

	// Filters defaults to a non-nil, empty slice.
	require.NotNil(t, field.Filters)
	require.Equal(t, 0, len(field.Filters))
}

func TestNewField_WithOptions(t *testing.T) {
	// Options are applied in order, so a later option overrides an earlier one.
	field := NewField("firstName", DataTypeInt,
		WithAlias("alias"),
		WithDefaultOperator(exp.OperatorGreaterThan),
	)

	require.Equal(t, "alias", field.Name)
	require.Equal(t, DataTypeInt, field.DataType)
	require.Equal(t, exp.OperatorGreaterThan, field.Operator)
}
