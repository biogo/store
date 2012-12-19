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
	"fmt"
	"math"
)

type Interface interface {
	// Bounds returns a bounding on the list of point.
	Bounds() *Bounding

	// Index returns the ith element of the list of points.
	Index(i int) Comparable

	// Len returns the length of the list.
	Len() int

	// Pivot partitions the list based on the dimension specified.
	Pivot(Dim) int

	// Slice returns a slice of the list.
	Slice(start, end int) Interface
}

// A Dim is an index into a point's coordinates.
type Dim int

// A Comparable is the element interface for values stored in a k-d tree.
type Comparable interface {
	// Clone returns a copy of the Comparable.
	Clone() Comparable

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

	// Extend increases the bounding volume to include the point. Extend may return nil.
	Extend(*Bounding) *Bounding
}

// A Bounding represents a volume bounding box.
type Bounding [2]Comparable

// Contains returns whether c is within the volume of the Bounding. A nil Bounding
// returns true.
func (b *Bounding) Contains(c Comparable) bool {
	if b == nil {
		return true
	}
	for d := Dim(0); d < Dim(c.Dims()); d++ {
		if c.Compare(b[0], d) < 0 || c.Compare(b[1], d) > 0 {
			return false
		}
	}
	return true
}

// A Node holds a single point value in a k-d tree.
type Node struct {
	Point       Comparable
	Plane       Dim
	Left, Right *Node
	*Bounding
}

func (n *Node) String() string {
	if n == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%.3f %d", n.Point, n.Plane)
}

// A Tree implements a k-d tree creation and nearest neighbour search.
type Tree struct {
	Root  *Node
	Count int
}

// New returns a k-d tree constructed from the values in p. If bounding is true, bounds
// are determined for each node.
func New(p Interface, bounding bool) *Tree {
	return &Tree{
		Root:  build(p, 0, bounding),
		Count: p.Len(),
	}
}

func build(p Interface, plane Dim, bounding bool) *Node {
	if p.Len() == 0 {
		return nil
	}

	piv := p.Pivot(plane)
	d := p.Index(piv)
	np := (plane + 1) % Dim(d.Dims())

	var b *Bounding
	if bounding {
		b = p.Bounds()
	}
	return &Node{
		Point:    d,
		Plane:    plane,
		Left:     build(p.Slice(0, piv), np, bounding),
		Right:    build(p.Slice(piv+1, p.Len()), np, bounding),
		Bounding: b,
	}
}

// Insert add a point to the tree, updating the bounding volumes if bounding is true and
// the tree is empty or the tree already has bounding volumes stored. No rebalancing of the
// tree is performed.
func (t *Tree) Insert(c Comparable, bounding bool) {
	t.Count++
	if t.Root != nil {
		bounding = t.Root.Bounding != nil
	}
	t.Root = t.Root.insert(c, 0, bounding)
}

func (n *Node) insert(c Comparable, d Dim, bounding bool) *Node {
	if n == nil {
		var b *Bounding
		if bounding {
			b = &Bounding{c.Clone(), c.Clone()}
		}
		return &Node{
			Point:    c,
			Plane:    d,
			Bounding: b,
		}
	}

	n.Bounding = c.Extend(n.Bounding)
	d = (n.Plane + 1) % Dim(c.Dims())
	if c.Compare(n.Point, n.Plane) <= 0 {
		n.Left = n.Left.insert(c, d, bounding)
	} else {
		n.Right = n.Right.insert(c, d, bounding)
	}

	return n
}

// Len returns the number of elements in the tree.
func (t *Tree) Len() int { return t.Count }

// Contains returns whether a Comparable is in the bounds of the tree. If no bounding has
// been contructed Contains returns true.
func (t *Tree) Contains(c Comparable) bool {
	if t.Root.Bounding == nil {
		return true
	}
	return t.Root.Contains(c)
}

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

	bn := n
	if c <= 0 {
		ln, ld := n.Left.search(q, d, dist)
		if ld < dist {
			dist = ld
			bn = ln
		}
		if c*c <= dist {
			rn, rd := n.Right.search(q, d, dist)
			if rd < dist {
				bn, dist = rn, rd
			}
		}
		return bn, dist
	}
	rn, rd := n.Right.search(q, d, dist)
	if rd < dist {
		dist = rd
		bn = rn
	}
	if c*c <= dist {
		ln, ld := n.Left.search(q, d, dist)
		if ld < dist {
			bn, dist = ln, ld
		}
	}
	return bn, dist
}
