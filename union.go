package contour90

import "errors"

// CreateContour returns the boolean union of the polygons
func CreateContour(polygons []Polygon) (Contour, error) {

	if len(polygons) == 0 {
		return Contour{}, nil
	}

	if !areCounterClockwisePolygons(polygons) {
		return Contour{}, errors.New("polygons should be oriented counterclockwise for this algorithm")
	}

	// trivial case : only one input polygon, return it
	if len(polygons) == 1 {
		c := append(Contour{}, polygons...)
		return c, nil
	}
	return nil, nil
}
