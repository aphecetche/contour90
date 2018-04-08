package contour90

import "math"

// EqualFloat checks if two float64 are equal within 1E-4 (on difference)
func EqualFloat(a, b float64) bool {
	return (math.Abs(a-b) < 1E-4)
}

// IsStrictlyBelowFloat returns true if a < b
func IsStrictlyBelowFloat(a, b float64) bool {
	return (a < b) && !EqualFloat(a, b)
}
