package builder

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGeoPolygon_Success(t *testing.T) {

}

func TestGeoPolygon_Failure(t *testing.T) {

	// This should fail because there are an odd number of values
	values := "1,2,3"

	polygon := parseGeoPolygon(values)
	require.Empty(t, polygon)
}
