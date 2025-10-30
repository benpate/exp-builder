package builder

import (
	"strconv"
	"strings"
)

type geoPoint struct {
	latitude  float64
	longitude float64
	radius    int
	units     string
}

func (p geoPoint) IsZero() bool {
	return (p.radius == 0) && (p.latitude == 0) && (p.longitude == 0)
}

func (p geoPoint) NotZero() bool {
	return !p.IsZero()
}

func parseGeoPoint(value string) geoPoint {

	// Split the value into parts
	items := strings.Split(value, ",")

	if len(items) < 2 {
		return geoPoint{}
	}

	// Parse longitude
	longitude, err := strconv.ParseFloat(items[0], 64)

	if err != nil {
		return geoPoint{}
	}

	// Parse latitude
	latitude, err := strconv.ParseFloat(items[1], 64)

	if err != nil {
		return geoPoint{}
	}

	result := geoPoint{
		latitude:  latitude,
		longitude: longitude,
	}

	if len(items) > 2 {
		radius, err := strconv.Atoi(items[2])

		if err != nil {
			return geoPoint{}
		}

		result.radius = radius

		if len(items) > 3 {
			result.units = items[3]
		}
	}

	return result
}
