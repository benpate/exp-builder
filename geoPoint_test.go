package builder

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGeo(t *testing.T) {

	result := parseGeoPoint("123,456,7,miles")
	require.True(t, result.NotZero())
	require.False(t, result.IsZero())
	require.Equal(t, result.longitude, 123.0)
	require.Equal(t, result.latitude, 456.0)
	require.Equal(t, result.radius, 7)
	require.Equal(t, result.units, "miles")
}

func TestGeo_Failure(t *testing.T) {

	result := parseGeoPoint("bad bad location")
	require.True(t, result.IsZero())
	require.False(t, result.NotZero())
	require.Equal(t, result.longitude, 0.0)
	require.Equal(t, result.latitude, 0.0)
	require.Equal(t, result.radius, 0)
	require.Equal(t, result.units, "")
}
