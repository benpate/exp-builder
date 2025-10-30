package builder

import "github.com/benpate/rosetta/convert"

func parseGeoPolygon(value string) [][]float64 {

	coords := convert.SliceOfFloat(value)

	// If the coordinates slice is not even, then this is not a valid polygon
	if len(coords)%2 > 0 {
		return [][]float64{}
	}

	// Allocate a result that's half the length of the coordinate pairs
	result := make([][]float64, 0, len(coords)/2)

	// Combiine coordinates into pairs
	for len(coords) > 0 {
		point := []float64{coords[0], coords[1]}
		result = append(result, point)
		coords = coords[2:]
	}

	// UwU
	return result
}
