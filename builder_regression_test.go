/******************************************
 * Regression & Behavior Tests
 *
 * Most tests here pin a defect that this package
 * actually shipped. The TestBehavior_* ones pin the
 * opposite: behavior that LOOKS like a defect, was
 * "fixed" once, and is deliberately being kept.
 *
 * They share a file because what they have in common
 * is their history, not their subject.
 ******************************************/

package builder

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"testing"

	"github.com/benpate/exp"
	"github.com/stretchr/testify/require"
)

// An empty value IS a legitimate predicate, and must be preserved. A field that
// may be unset is queried as (field == "") OR (field == "VALUE"), so a blank
// value carries real meaning and cannot be discarded as noise.
//
// A guard that dropped these was written and then removed; see BUG-131 for the
// unresolved case against it. This test exists so that the guard cannot come
// back without the decision being revisited.
func TestBehavior_EmptyValueIsAPredicate(t *testing.T) {

	b := NewBuilder().String("firstName")

	// "Unset, or John" -- both halves survive.
	result := b.Evaluate(url.Values{"firstName": {"", "John"}})
	require.Equal(t, exp.Or(exp.Equal("firstName", ""), exp.Equal("firstName", "John")), result)

	// A wholly blank parameter is still skipped, by Evaluate's sliceNotEmpty
	// check rather than by anything in EvaluateField.
	require.Equal(t, exp.Empty(), b.Evaluate(url.Values{"firstName": {""}}))
	require.Equal(t, exp.Empty(), b.Evaluate(url.Values{"firstName": {"", "", ""}}))

	// EvaluateField has no such check, so it emits the blank comparison on its own.
	require.Equal(t, exp.Equal("firstName", ""), b.EvaluateField(b["firstName"], []string{""}))
}

// A filter that empties a value produces an empty comparison too. With CONTAINS
// that matches every record, which is the open question in BUG-131 -- this test
// pins what the package does TODAY so the answer is a visible change.
func TestBehavior_FilterEmptiedValueIsAPredicate(t *testing.T) {

	// Strips everything that is not a lowercase letter, which is the shape of a
	// real token filter -- "???" reduces to "".
	lettersOnly := func(value string) string {
		return strings.Map(func(r rune) rune {
			if (r >= 'a') && (r <= 'z') {
				return r
			}
			return -1
		}, value)
	}

	b := NewBuilder().String("tags", WithFilter(lettersOnly), WithDefaultOpContains())

	require.Equal(t, exp.New("tags", exp.OperatorContains, ""), b.Evaluate(url.Values{"tags": {"???"}}))
	require.Equal(t, exp.New("tags", exp.OperatorContains, "abc"), b.Evaluate(url.Values{"tags": {"abc"}}))
}

// A named time range must be read from the operator-stripped value. Parsing the
// raw input meant that "gt:today" matched no range, produced no predicate, and
// silently returned an UNFILTERED result set.
func TestRegression_TimeRangeStripsOperatorPrefix(t *testing.T) {

	b := NewBuilder().Time("date")

	bare := b.Evaluate(url.Values{"date": {"today"}})
	prefixed := b.Evaluate(url.Values{"date": {"gt:today"}})

	require.NotEqual(t, exp.Empty(), bare)
	require.Equal(t, bare, prefixed, "an operator prefix must not discard the range")
}

// Filters must reach the time-range parser, too. They were applied to the
// stripped value while the range parser read the raw input, so a filter that
// normalized a range name had no effect.
func TestRegression_TimeRangeSeesFilteredValue(t *testing.T) {

	b := NewBuilder().Time("date", WithFilter(strings.ToLower))

	require.NotEqual(t, exp.Empty(), b.Evaluate(url.Values{"date": {"TODAY"}}))
	require.Equal(t,
		b.Evaluate(url.Values{"date": {"today"}}),
		b.Evaluate(url.Values{"date": {"TODAY"}}),
	)
}

// The same URL must always produce the same expression. Ranging the Builder map
// directly shuffled the predicates on every call, changing the query shape that
// the database was handed.
func TestRegression_EvaluateIsDeterministic(t *testing.T) {

	b := NewBuilder().String("a").String("b").String("c").String("d").String("e")
	values := url.Values{"a": {"1"}, "b": {"2"}, "c": {"3"}, "d": {"4"}, "e": {"5"}}

	expected := fmt.Sprintf("%v", b.Evaluate(values))

	for range 100 {
		require.Equal(t, expected, fmt.Sprintf("%v", b.Evaluate(values)))
	}
}

// A request missing several required fields must always name the same one,
// rather than picking at random from a map. The name travels in the error's
// DETAIL, not in its message, so the assertion has to read the marshaled error.
func TestRegression_EvaluateAllReportsAStableField(t *testing.T) {

	b := NewBuilder().String("alpha").String("beta").String("gamma")

	for range 100 {

		_, err := b.EvaluateAll(url.Values{})
		require.Error(t, err)

		marshaled, marshalErr := json.Marshal(err)
		require.NoError(t, marshalErr)
		require.Contains(t, string(marshaled), `"details":["alpha"]`)
	}
}

// The EvaluateAll error must survive being marshaled into a JSON response.
// It carried the whole Field as a detail, and a Field holds filter FUNCTIONS,
// which encoding/json refuses outright.
func TestRegression_EvaluateAllErrorIsMarshalable(t *testing.T) {

	b := NewBuilder().
		String("alpha", WithFilter(strings.ToLower)).
		String("beta", WithFilter(strings.ToUpper))

	_, err := b.EvaluateAll(url.Values{})
	require.Error(t, err)

	marshaled, marshalErr := json.Marshal(err)
	require.NoError(t, marshalErr)
	require.Contains(t, string(marshaled), "alpha")
}

// Nothing up my sleeve
