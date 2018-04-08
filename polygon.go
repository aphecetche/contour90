package contour90

import (
	"errors"
	"fmt"
	"log"
	"math"
	"sort"
)

// Polygon describes a simple, rectilinear, closed set of vertices
// with a specific orientation
type Polygon []Vertex

func (p *Polygon) isManhattan() bool {
	for i := 0; i < len(*p)-1; i++ {
		if !isVerticalSegment((*p)[i], (*p)[i+1]) &&
			!isHorizontalSegment((*p)[i], (*p)[i+1]) {
			return false
		}
	}
	return true
}

func (p *Polygon) isCounterClockwiseOriented() bool {
	return p.signedArea() > 0
}

func (p *Polygon) signedArea() float64 {
	/// Compute the signed area of this polygon
	/// Algorithm from F. Feito, J.C. Torres and A. Urena,
	/// Comput. & Graphics, Vol. 19, pp. 595-600, 1995
	area := 0.0
	for i := 0; i < len(*p)-1; i++ {
		current := (*p)[i]
		next := (*p)[i+1]
		area += current.x*next.y - next.x*current.y
	}
	return area * 0.5
}

func (p *Polygon) isClosed() bool {
	return (*p)[0] == (*p)[len(*p)-1]
}

// EqualPolygon checks if two polygon are the same
// (same vertices, whetever the order)
func EqualPolygon(a, b Polygon) bool {
	if len(a) != len(b) {
		return false
	}

	if (a == nil) != (b == nil) {
		return false
	}

	sa := a.getSortedVertices()
	sb := b.getSortedVertices()

	for i, v := range sa {
		if !EqualVertex(v, sb[i]) {
			return false
		}
	}
	return true
}

func closePolygon(p Polygon) (Polygon, error) {
	if p.isClosed() {
		return p, nil
	}
	np := Polygon{}
	np = append(np, p...)
	np = append(np, p[0])
	if !np.isManhattan() {
		return nil, errors.New("closing resulted in non Manhattan polygon")
	}
	return np, nil
}

func (p *Polygon) String() string {
	s := fmt.Sprintf("POLYGON (")
	for i := 0; i < len(*p); i++ {
		s += fmt.Sprintf("%f %f", (*p)[i].x, (*p)[i].y)
		if i < len(*p)-1 {
			s += ","
		}
	}
	s += ")"
	return s
}

func (p *Polygon) getSortedVertices() []Vertex {
	size := len(*p)
	if p.isClosed() {
		size--
	}
	c := []Vertex{}
	for i := 0; i < size; i++ {
		c = append(c, (*p)[i])
	}
	sort.Slice(c, func(i, j int) bool {
		if EqualFloat(c[i].x, c[j].x) {
			return c[i].y < c[j].y
		}
		return c[i].x < c[j].x
	})
	return c
}

// Contains returns true if the point (xp,yp) is inside the polygon
//
// Note that this algorithm yields unpredicatable result if the point xp,yp
// is on one edge of the polygon. Should not generally matters, except when comparing
// two different implementations maybe.
//
// TODO : look e.g. to http://alienryderflex.com/polygon/ for some possible optimizations
// (e.g. pre-computation)
func (p *Polygon) Contains(xp, yp float64) (bool, error) {
	if !p.isClosed() {
		return false, errors.New("Contains can only work with closed polygons")
	}

	j := len(*p) - 1
	oddNodes := false
	for i := 0; i < len(*p); i++ {
		if ((*p)[i].y < yp && (*p)[j].y >= yp) || ((*p)[j].y < yp && (*p)[i].y >= yp) {
			if (*p)[i].x+
				(yp-(*p)[i].y)/((*p)[j].y-(*p)[i].y)*((*p)[j].x-(*p)[i].x) <
				xp {
				oddNodes = !oddNodes
			}
		}
		j = i
	}
	return oddNodes, nil
}

// BBox returns the bounding box of the polygon.
func (p *Polygon) BBox() BBox {
	xmin := math.MaxFloat64
	xmax := -xmin
	ymin := xmin
	ymax := -ymin

	for _, v := range *p {
		xmin = math.Min(xmin, v.x)
		ymin = math.Min(ymin, v.y)
		xmax = math.Max(xmax, v.x)
		ymax = math.Max(ymax, v.y)
	}
	bbox, err := NewBBox(xmin, ymin, xmax, ymax)
	if err != nil {
		log.Fatal("got a very unexpected invalid bbox here")
	}
	return bbox
}

// SquaredDistancePointToPolygon return the square of the distance
// between a point and a polygon
func SquaredDistancePointToPolygon(point Vertex, p Polygon) float64 {
	d := math.MaxFloat64
	for i := 0; i < len(p)-1; i++ {
		s0 := p[i]
		s1 := p[i+1]
		d2 := SquaredDistanceOfPointToSegment(point, s0, s1)
		d = math.Min(d, d2)
	}
	return d
}