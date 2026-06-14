package builder

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// NOTE: parseGeoPolygon is not currently used by the package itself
// (EvaluateField parses polygons via geo.NewPolygonFromString). These tests
// document and lock in its behavior in case it is wired up later.

func TestGeoPolygon_Success(t *testing.T) {

	// An even number of coordinates is grouped into lon/lat pairs.
	polygon := parseGeoPolygon("1,2,3,4")
	require.Equal(t, [][]float64{{1, 2}, {3, 4}}, polygon)
}

func TestGeoPolygon_Empty(t *testing.T) {

	// An empty string produces an empty (but non-nil) result.
	polygon := parseGeoPolygon("")
	require.Empty(t, polygon)
}

func TestGeoPolygon_Failure(t *testing.T) {

	// This should fail because there are an odd number of values
	values := "1,2,3"

	polygon := parseGeoPolygon(values)
	require.Empty(t, polygon)
}

// FuzzParseGeoPolygon confirms the coordinate parser never panics and that
// every coordinate it returns is a complete lon/lat pair.
func FuzzParseGeoPolygon(f *testing.F) {

	seeds := []string{"1,2,3,4", "1,2,3", "", "-118.5,34.25", "not,a,number", "1,2,"}
	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, data string) {

		result := parseGeoPolygon(data)

		// Every returned coordinate must be a complete pair.
		for _, pair := range result {
			require.Equal(t, 2, len(pair))
		}
	})
}
