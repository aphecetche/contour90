package contour90

import "testing"

func TestContourCreationGeneratesEmptyContourForEmptyInput(t *testing.T) {
	polygons := []Polygon{}
	c, err := CreateContour(polygons)
	if err != nil {
		t.Error("should not trigger an error here")
	}
	if len(c) != 0 {
		t.Error("should get an empty contour here")
	}
}
