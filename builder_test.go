package builder

import (
	"net/url"
	"testing"

	"github.com/benpate/exp"
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

	{
		u, _ := url.ParseQuery("firstName=John&firstName=Sara&firstName=ne:Kyle")
		expect := exp.OrExpression{
			exp.Predicate{Field: "firstName", Operator: "=", Value: "John"},
			exp.Predicate{Field: "firstName", Operator: "=", Value: "Sara"},
			exp.Predicate{Field: "firstName", Operator: "!=", Value: "Kyle"},
		}
		require.Equal(t, expect, b.Evaluate(u))
	}
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

	{
		u, _ := url.ParseQuery("parentId=123456781234567812345678")
		expect := exp.Predicate{Field: "parentId", Operator: "=", Value: objectID}
		require.Equal(t, expect, b.Evaluate(u))
	}

	{
		u, _ := url.ParseQuery("parentId=Not-An-ObjectID")
		expect := exp.EmptyExpression{}
		require.Equal(t, expect, b.Evaluate(u))
	}
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
		require.Equal(t, "GT", operator)
		require.Equal(t, "7", value)
	}

	{
		operator, value := parseValue("GTE:7", "=")
		require.Equal(t, "GTE", operator)
		require.Equal(t, "7", value)
	}
}
