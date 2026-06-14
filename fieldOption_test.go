package builder

import (
	"net/url"
	"strings"
	"testing"

	"github.com/benpate/exp"
	"github.com/stretchr/testify/require"
)

func TestWithDefaultOpEqual(t *testing.T) {
	field := NewField("name", DataTypeString, WithDefaultOpEqual())
	require.Equal(t, exp.OperatorEqual, field.Operator)
}

func TestWithDefaultOpContains(t *testing.T) {
	field := NewField("name", DataTypeString, WithDefaultOpContains())
	require.Equal(t, exp.OperatorContains, field.Operator)
}

func TestWithDefaultOpBeginsWith(t *testing.T) {
	field := NewField("name", DataTypeString, WithDefaultOpBeginsWith())
	require.Equal(t, exp.OperatorBeginsWith, field.Operator)
}

func TestWithDefaultOperator(t *testing.T) {
	field := NewField("name", DataTypeString, WithDefaultOperator(exp.OperatorLessThan))
	require.Equal(t, exp.OperatorLessThan, field.Operator)
}

func TestWithAlias(t *testing.T) {
	// The alias replaces the field name used in the generated expression,
	// while the URL parameter is still read under the original key.
	field := NewField("urlName", DataTypeString, WithAlias("databaseName"))
	require.Equal(t, "databaseName", field.Name)
}

func TestWithFilter(t *testing.T) {
	field := NewField("name", DataTypeString, WithFilter(strings.ToUpper))

	require.Equal(t, 1, len(field.Filters))
	require.Equal(t, "HELLO", field.Filters[0]("hello"))
}

func TestWithFilter_Multiple(t *testing.T) {
	// Filters accumulate in the order they are supplied.
	field := NewField("name", DataTypeString,
		WithFilter(strings.ToUpper),
		WithFilter(strings.TrimSpace),
	)
	require.Equal(t, 2, len(field.Filters))
}

func TestWithFilter_AppliedDuringEvaluate(t *testing.T) {
	// A filter transforms the input value before the expression is built.
	b := NewBuilder().String("name", WithFilter(strings.ToUpper))

	u, _ := url.ParseQuery("name=john")
	expect := exp.Predicate{Field: "name", Operator: "=", Value: "JOHN"}
	require.Equal(t, expect, b.Evaluate(u))
}

func TestWithAlias_AppliedDuringEvaluate(t *testing.T) {
	// The URL key is "urlName" but the expression field is the alias.
	b := NewBuilder().String("urlName", WithAlias("databaseName"))

	u, _ := url.ParseQuery("urlName=John")
	expect := exp.Predicate{Field: "databaseName", Operator: "=", Value: "John"}
	require.Equal(t, expect, b.Evaluate(u))
}
