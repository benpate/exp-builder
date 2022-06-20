package builder

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/benpate/exp"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestBuilder(t *testing.T) {

	objectID, _ := primitive.ObjectIDFromHex("123456781234567812345678")

	b := NewBuilder().
		ObjectID("parentId").
		String("templateId").
		String("firstName").
		String("lastName").
		Int("publishDate")

	{
		u, _ := url.ParseQuery("publishDate=ge:123")
		expect := exp.Predicate{Field: "publishDate", Operator: ">=", Value: 123}
		require.Equal(t, expect, b.Evaluate(u))
	}

	{
		u, _ := url.ParseQuery("parentId=123456781234567812345678")
		expect := exp.Predicate{Field: "parentId", Operator: "=", Value: objectID}
		require.Equal(t, expect, b.Evaluate(u))
	}

	{
		u, _ := url.ParseQuery("firstName=John&firstName=Sara&firstName=ne:Kyle")
		expect := exp.OrExpression{
			exp.Predicate{Field: "firstName", Operator: "=", Value: "John"},
			exp.Predicate{Field: "firstName", Operator: "=", Value: "Sara"},
			exp.Predicate{Field: "firstName", Operator: "!=", Value: "Kyle"},
		}
		require.Equal(t, expect, b.Evaluate(u))
	}

	{
		u, _ := url.ParseQuery("publishDate=ge:123&parentId=123456781234567812345678")
		fmt.Println(u)
	}

}
