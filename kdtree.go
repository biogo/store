// Copyright ©2012 Dan Kortschak <dan.kortschak@adelaide.edu.au>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

// Package kdtree implements a k-d tree.
package kdtree

import (
	"math"
)

// Randoms is the maximum number of random values to sample for calculation of median of
// random elements.
var Randoms = 100

var (
	_ Interface  = Points{}
	_ Comparable = Point{}
)

type Interface interface {
	Index(int) Comparable
	Len() int
	Pivot(Dim) int
	Slice(start, end int) Interface
}

// A Dim is an index into a point's coordinates.
type Dim int

// A Comparable is the element interface for values stored in a k-d tree.
type Comparable interface {
	// Compare returns a value indicating the sort order relationship between the
	// receiver and the parameter at the dimension specified.
	//
	// Given c = a.Compare(b, d):
	//  c < 0 if a_d < b_d ;
	//  c == 0 if a_d == b_d; and
	//  c > 0 if a_d > b_d.
	//
	Compare(Comparable, Dim) float64

	// Dims returns the number of dimensions described in the Comparable.
	Dims() int

	// Distance resturns the distance between the receiver and the parameter.
	Distance(Comparable) float64
}

// A Point represents a point in a k-d space that satisfies the Comparable interface.
type Point []float64

func (p Point) Compare(c Comparable, d Dim) float64 { q := c.(Point); return p[d] - q[d] }
func (p Point) Dims() int                           { return len(p) }
func (p Point) Distance(c Comparable) float64 {
	q := c.(Point)
	var sum float64
	for dim, c := range p {
		d := c - q[dim]
		sum += d * d
	}
	return sum
}

// A Points is a collection of point values that satisfies the Interface.
type Points []Point

func (p Points) Index(i int) Comparable         { return p[i] }
func (p Points) Len() int                       { return len(p) }
func (p Points) Pivot(d Dim) int                { return Plane{Points: p, Dim: d}.Pivot() }
func (p Points) Slice(start, end int) Interface { return p[start:end] }

// A Plane is a wrapping type that allows a Points type be pivoted on a dimension.
type Plane struct {
	Dim
	Points
}

func (p Plane) Less(i, j int) bool              { return p.Points[i][p.Dim] < p.Points[j][p.Dim] }
func (p Plane) Pivot() int                      { return Partition(p, MedianOfRandoms(p, Randoms)) }
func (p Plane) Slice(start, end int) SortSlicer { p.Points = p.Points[start:end]; return p }
func (p Plane) Swap(i, j int) {
	p.Points[i], p.Points[j] = p.Points[j], p.Points[i]
}

// A Node holds a single point value in a k-d tree.
type Node struct {
	Point       Comparable
	Plane       Dim
	Left, Right *Node
}

// A Tree implements a k-d tree creation and nearest neighbour search.
type Tree struct {
	Root  *Node
	Count int
}

// New returns a k-d tree constructed from the values in p.
func New(p Interface) *Tree {
	return &Tree{
		Root:  build(p, 0),
		Count: p.Len(),
	}
}

func build(p Interface, plane Dim) *Node {
	if p.Len() == 0 {
		return nil
	}

	piv := p.Pivot(plane)
	d := p.Index(piv)
	np := (plane + 1) % Dim(d.Dims())

	return &Node{
		Point: d,
		Plane: plane,
		Left:  build(p.Slice(0, piv), np),
		Right: build(p.Slice(piv+1, p.Len()), np),
	}
}

// Len returns the number of elements in the tree.
func (t *Tree) Len() int { return t.Count }

var inf = math.Inf(1)

// Nearest returns the nearest value to the query and the distance between them.
func (t *Tree) Nearest(q Comparable) (Comparable, float64) {
	if t.Root == nil {
		return nil, inf
	}
	n, dist := t.Root.search(q, 0, inf)
	if n == nil {
		return nil, inf
	}
	return n.Point, dist
}

func (n *Node) search(q Comparable, d Dim, dist float64) (*Node, float64) {
	if n == nil {
		return nil, inf
	}
	c := q.Compare(n.Point, d)
	dist = math.Min(dist, q.Distance(n.Point))
	d = (d + 1) % Dim(q.Dims())
	switch {
	case c < 0:
		rn, rd := n.Left.search(q, d, dist)
		if rd < dist {
			n, dist = rn, rd
		}
		if c*c <= dist {
			rn, rd := n.Right.search(q, d, dist)
			if rd < dist {
				n, dist = rn, rd
			}
		}
	case c > 0:
		rn, rd := n.Right.search(q, d, dist)
		if rd < dist {
			n, dist = rn, rd
		}
		if c*c <= dist {
			rn, rd := n.Left.search(q, d, dist)
			if rd < dist {
				n, dist = rn, rd
			}
		}
	default:
		var (
			rn *Node
			rd float64
		)
		rn, rd = n.Left.search(q, d, dist)
		if rd < dist {
			n, dist = rn, rd
		}
		rn, rd = n.Right.search(q, d, dist)
		if rd < dist {
			n, dist = rn, rd
		}
	}
	return n, dist
}
