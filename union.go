package contour90

import (
	"errors"
	"sort"
)

var (
	// ErrWrongOrientation indicates that an attempt was made to use clockwise polygons
	// (or a mix of counter and clock-wise polygons), while this library's algorithms
	// only works with counter-clockwise polygons
	ErrWrongOrientation = errors.New("polygons should be oriented counterclockwise for this algorithm")
)

var (
	// errDifferentVH indicates that a different number of vertical and horizontal edges
	// was used
	errDifferentVH = errors.New("number of vertical and horizontal edges should be the same")
	// errDisconnectedEdge signals that one horizontal edge is not connected
	// to its (supposedly) preceding vertical edge
	errDisconnectedEdge = errors.New("horizontal edge not connected")
	// errEmptyPolygon signals an empty polygon was encountered where it was expected
	errEmptyPolygon = errors.New("empty polygon")
	// errClosingPolygon signals an error when attempting to close a polygon
	errClosingPolygon = errors.New("closing polygon")
)

// CreateContour returns the boolean union of the polygons
func CreateContour(polygons []Polygon) (Contour, error) {

	if len(polygons) == 0 {
		return Contour{}, nil
	}

	if !areCounterClockwisePolygons(polygons) {
		return Contour{}, ErrWrongOrientation
	}

	// trivial case : only one input polygon, return it
	if len(polygons) == 1 {
		c := append(Contour{}, polygons...)
		return c, nil
	}

	polygonVerticalEdges := getPolygonSliceVerticalEdges(polygons)

	sortVerticalEdges(polygonVerticalEdges)

	// Initialize the segment tree that is used by the sweep() function
	segmentTree, err := createSegmentTree(getPolygonSliceYPositions(polygons))
	if err != nil {
		return nil, err
	}

	// Find the vertical edges of the merged contour. This is the meat of the algorithm...
	contourVerticalEdges := sweep(segmentTree, polygonVerticalEdges)

	// Deduce the horizontal edges from the vertical ones
	contourHorizontalEdges := verticalsToHorizontals(contourVerticalEdges)

	c, err := finalizeContour(contourVerticalEdges, contourHorizontalEdges)
	return c, err
}

// sort vertical edges in ascending x order
// if same x, insure that left edges are before right edges
// within same x, order by increasing bottommost y
// Mind your steps ! This sorting is critical to the contour merging algorithm !
func sortVerticalEdges(edges []verticalEdge) {
	sort.Slice(edges, func(i, j int) bool {
		e1 := edges[i]
		e2 := edges[j]
		x1 := e1.begin().x
		x2 := e2.begin().x
		if EqualFloat(x1, x2) {
			if isLeftEdge(e1) && isRightEdge(e2) {
				return true
			}
			if isRightEdge(e1) && isLeftEdge(e2) {
				return false
			}
			y1 := bottom(e1)
			y2 := bottom(e2)
			return y1 < y2
		} else if x1 < x2 {
			return true
		} else {
			return false
		}
	})
}

// verticalsToHorizontals generates horizontal edges from the vertical ones
// The horizontals are ordered relative to the verticals, i.e. the first horizontal
// should be the edge __following__ the first vertical, etc...
func verticalsToHorizontals(verticals []verticalEdge) []horizontalEdge {
	horizontals := make([]horizontalEdge, len(verticals))

	type vertexWithRef struct {
		first  Vertex
		second int
	}

	vertices := []vertexWithRef{}

	for i, e := range verticals {
		vertices = append(vertices, vertexWithRef{e.begin(), i})
		vertices = append(vertices, vertexWithRef{e.end(), i})
	}

	sort.Slice(vertices, func(i, j int) bool {
		lhs := vertices[i].first
		rhs := vertices[j].first
		if lhs.y < rhs.y {
			return true
		}
		if rhs.y < lhs.y {
			return false
		}
		return lhs.x < rhs.x
	})

	for i := 0; i < len(vertices)/2; i++ {

		p1 := vertices[i*2]
		p2 := vertices[i*2+1]
		refEdge := verticals[p1.second]
		e := p1.first.x
		b := p2.first.x
		if (EqualFloat(p1.first.y, bottom(refEdge)) && isLeftEdge(refEdge)) ||
			(EqualFloat(p1.first.y, top(refEdge)) && isRightEdge(refEdge)) {
			e, b = b, e
		}
		h := horizontalEdge{p1.first.y, b, e}
		// which vertical edge is preceding this horizontal ?
		preceding := p1.second
		next := p2.second
		if b > e {
			next, preceding = preceding, next
		}
		horizontals[preceding] = h
	}
	return horizontals
}

func firstFalse(b []bool) int {
	for i := 0; i < len(b); i++ {
		if b[i] == false {
			return i
		}
	}
	return -1
}

func finalizeContour(v []verticalEdge, h []horizontalEdge) (Contour, error) {
	if len(v) != len(h) {
		return nil, errDifferentVH
	}

	for i := 0; i < len(v); i++ {
		if !EqualVertex(h[i].begin(), v[i].end()) {
			return nil, errDisconnectedEdge
		}
	}

	all := []manhattanEdge{}

	for i := 0; i < len(v); i++ {
		all = append(all, v[i])
		all = append(all, h[i])
	}

	alreadyAdded := make([]bool, len(all))
	for i := range alreadyAdded {
		alreadyAdded[i] = false
	}

	inorder := []int{}

	var nofUsed int
	var iCurrent int

	startSegment := all[iCurrent]

	contour := Contour{}

	for nofUsed < len(all) {
		currentSegment := all[iCurrent]
		inorder = append(inorder, iCurrent)
		alreadyAdded[iCurrent] = true
		nofUsed++
		if EqualVertex(currentSegment.end(), startSegment.begin()) {
			if len(inorder) == 0 {
				return nil, errEmptyPolygon
			}
			vertices := make([]Vertex, len(inorder))
			for i := range inorder {
				vertices[i] = all[inorder[i]].begin()
			}
			p, err := closePolygon(vertices)
			if err != nil {
				return nil, errClosingPolygon
			}
			contour = append(contour, p)
			iCurrent = firstFalse(alreadyAdded)
			inorder = []int{}
			if iCurrent > 0 {
				startSegment = all[iCurrent]
			}
		}
		for i := 0; i < len(alreadyAdded); i++ {
			if i != iCurrent && alreadyAdded[i] == false {
				if EqualVertex(currentSegment.end(), all[i].begin()) {
					iCurrent = i
					break
				}
			}
		}
	}
	return contour, nil
}

func getPolygonVerticalEdges(polygon *Polygon) []verticalEdge {
	edges := []verticalEdge{}
	for i := 0; i < len(*polygon)-1; i++ {
		current := (*polygon)[i]
		next := (*polygon)[i+1]
		if EqualFloat(current.x, next.x) {
			edges = append(edges, verticalEdge{current.x, current.y, next.y})
		}
	}
	return edges
}

func getPolygonSliceVerticalEdges(polygons []Polygon) []verticalEdge {
	edges := []verticalEdge{}
	for _, p := range polygons {
		e := getPolygonVerticalEdges(&p)
		edges = append(edges, e...)
	}
	return edges
}

func getPolygonSliceYPositions(polygons []Polygon) []float64 {
	ypos := []float64{}
	for i := 0; i < len(polygons); i++ {
		for j := 0; j < len(polygons[i]); j++ {
			ypos = append(ypos, polygons[i][j].y)
		}
	}
	sort.Float64s(ypos)
	return removeDuplicates(ypos)
}

func getInterval(v verticalEdge) interval {
	y1 := v.begin().y
	y2 := v.end().y
	if y2 > y1 {
		return interval{y1, y2}
	}
	return interval{y2, y1}
}

func sweep(segmentTree *node, polygonVerticalEdges []verticalEdge) []verticalEdge {

	contourVerticalEdges := []verticalEdge{}

	edgeStack := []interval{}

	for i := 0; i < len(polygonVerticalEdges); i++ {

		edge := polygonVerticalEdges[i]
		ival := getInterval(edge)

		if isLeftEdge(edge) {
			segmentTree.contribution(ival, &edgeStack)
			segmentTree.insertInterval(ival)
		} else {
			segmentTree.deleteInterval(ival)
			segmentTree.contribution(ival, &edgeStack)
		}

		e1 := edge

		if i < len(polygonVerticalEdges)-1 {
			e1 = polygonVerticalEdges[i+1]
		}

		if (isLeftEdge(edge) != isLeftEdge(e1)) ||
			(!EqualFloat(edge.begin().x, e1.begin().x)) ||
			(i == len(polygonVerticalEdges)-1) {
			for _, es := range edgeStack {
				var newEdge verticalEdge
				if isRightEdge(edge) {
					newEdge = verticalEdge{edge.begin().x, es.begin(), es.end()}
				} else {
					newEdge = verticalEdge{edge.begin().x, es.end(), es.begin()}
				}
				contourVerticalEdges = append(contourVerticalEdges, newEdge)
			}
			edgeStack = []interval{}
		}
	}
	return contourVerticalEdges
}
