// Copyright ©2012 The bíogo.kdtree Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package kdtree implements a k-d tree.
package kdtree

import (
	"fmt"
	"math"
)

type Interface interface {
	// Index returns the ith element of the list of points.
	Index(i int) Comparable

	// Len returns the length of the list.
	Len() int

	// Pivot partitions the list based on the dimension specified.
	Pivot(Dim) int

	// Slice returns a slice of the list.
	Slice(start, end int) Interface
}

// An Bounder returns a bounding volume containing the list of points. Bounds may return nil.
type Bounder interface {
	Bounds() *Bounding
}

type bounder interface {
	Interface
	Bounder
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
}

// An Extender can increase a bounding volume to include the point. Extend may return nil.
type Extender interface {
	Extend(*Bounding) *Bounding
}

type extender interface {
	Comparable
	Extender
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

// New returns a k-d tree constructed from the values in p. If p is a Bounder and
// bounding is true, bounds are determined for each node.
func New(p Interface, bounding bool) *Tree {
	if p, ok := p.(bounder); ok && bounding {
		return &Tree{
			Root:  buildBounded(p, 0, bounding),
			Count: p.Len(),
		}
	}
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
		Point:    d,
		Plane:    plane,
		Left:     build(p.Slice(0, piv), np),
		Right:    build(p.Slice(piv+1, p.Len()), np),
		Bounding: nil,
	}
}

func buildBounded(p bounder, plane Dim, bounding bool) *Node {
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
		Left:     buildBounded(p.Slice(0, piv).(bounder), np, bounding),
		Right:    buildBounded(p.Slice(piv+1, p.Len()).(bounder), np, bounding),
		Bounding: b,
	}
}

// Insert adds a point to the tree, updating the bounding volumes if bounding is
// true, and the tree is empty or the tree already has bounding volumes stored,
// and c is an Extender. No rebalancing of the tree is performed.
func (t *Tree) Insert(c Comparable, bounding bool) {
	t.Count++
	if t.Root != nil {
		bounding = t.Root.Bounding != nil
	}
	if c, ok := c.(extender); ok && bounding {
		t.Root = t.Root.insertBounded(c, 0, bounding)
		return
	} else if !ok && t.Root != nil {
		// If we are not rebounding, mark the tree as non-bounded.
		t.Root.Bounding = nil
	}
	t.Root = t.Root.insert(c, 0)
}

func (n *Node) insert(c Comparable, d Dim) *Node {
	if n == nil {
		return &Node{
			Point:    c,
			Plane:    d,
			Bounding: nil,
		}
	}

	d = (n.Plane + 1) % Dim(c.Dims())
	if c.Compare(n.Point, n.Plane) <= 0 {
		n.Left = n.Left.insert(c, d)
	} else {
		n.Right = n.Right.insert(c, d)
	}

	return n
}

func (n *Node) insertBounded(c extender, d Dim, bounding bool) *Node {
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

	if bounding {
		n.Bounding = c.Extend(n.Bounding)
	}
	d = (n.Plane + 1) % Dim(c.Dims())
	if c.Compare(n.Point, n.Plane) <= 0 {
		n.Left = n.Left.insertBounded(c, d, bounding)
	} else {
		n.Right = n.Right.insertBounded(c, d, bounding)
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
