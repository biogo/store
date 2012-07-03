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

// Package llrb implements a Left-Leaning Red Black tree as described in
//  http://www.cs.princeton.edu/~rs/talks/LLRB/LLRB.pdf
//  http://www.cs.princeton.edu/~rs/talks/LLRB/Java/RedBlackBST.java
//  http://www.teachsolaisgames.com/articles/balanced_left_leaning.html
package llrb

const (
	TD234 = iota
	BU23
)

// Operation mode of the LLRB tree. Currently only BU23 tests as correct.
const Mode = BU23

// A Comparable is a type that can be inserted into a Tree or used as a range
// or equality query on the tree,
type Comparable interface {
	// Compare returns a value indicating the sort order relationship between the
	// receiver and the parameter.
	//
	// Given c = a.Compare(b):
	//  c < 0 if a r < b;
	//  c == 0 if a == b; and
	//  c > 0 if a > b.
	//
	Compare(Comparable) int
}

// A Color represents the color of a Node.
type Color bool

// String returns a string representation of a Color.
func (c Color) String() string {
	if c {
		return "Black"
	}
	return "Red"
}

const (
	// Red as false give us the defined behaviour that new nodes are red. Although this
	// is incorrect for the root node, that is resolved on the first insertion.
	Red   Color = false
	Black Color = true
)

// A Node represents a node in the LLRB tree.
type Node struct {
	Elem        Comparable
	Left, Right *Node
	Color       Color
}

// A Tree represent the root node of an LLRB tree. Public methods of nodes are exposed
// through this type.
type Tree Node

// Helper methods

// color returns the effect color of a Node. A nil node returns black.
func (self *Node) color() Color {
	if self == nil {
		return Black
	}
	return self.Color
}

// (a,c)b -rotL-> ((a,)b,)c
func (self *Node) rotateLeft() (root *Node) {
	// Assumes: self has two children.
	root = self.Right
	self.Right = root.Left
	root.Left = self
	root.Color = self.Color
	self.Color = Red
	return
}

// (a,c)b -rotR-> (,(,c)b)a
func (self *Node) rotateRight() (root *Node) {
	// Assumes: self has two children.
	root = self.Left
	self.Left = root.Right
	root.Right = self
	root.Color = self.Color
	self.Color = Red
	return
}

// (aR,cR)bB -flipC-> (aB,cB)bR | (aB,cB)bR -flipC-> (aR,cR)bB 
func (self *Node) flipColors() {
	// Assumes: self has two children.
	self.Color = !self.Color
	self.Left.Color = !self.Left.Color
	self.Right.Color = !self.Right.Color
}

// fixUp ensures that black link balance is correct, that red nodes lean left,
// and that 4 nodes are split.
func (self *Node) fixUp() *Node {
	if self.Right.color() == Red {
		self = self.rotateLeft()
	}
	if self.Left.color() == Red && self.Left.Left.color() == Red {
		self = self.rotateRight()
	}
	if self.Left.color() == Red && self.Right.color() == Red {
		self.flipColors()
	}
	return self
}

func (self *Node) moveRedLeft() *Node {
	self.flipColors()
	if self.Right.Left.color() == Red {
		self.Right = self.Right.rotateRight()
		self = self.rotateLeft()
		self.flipColors()
	}
	return self
}

func (self *Node) moveRedRight() *Node {
	self.flipColors()
	if self.Left.Left.color() == Red {
		self = self.rotateRight()
		self.flipColors()
	}
	return self
}

// Get returns the first match of q in the Tree. If insertion without
// replacement is used, this is probably not what you want.
func (self *Tree) Get(q Comparable) Comparable {
	return (*Node)(self).search(q).Elem
}

func (self *Node) search(q Comparable) (n *Node) {
	n = self
	for n != nil {
		switch c := q.Compare(n.Elem); {
		case c == 0:
			return n
		case c < 0:
			n = n.Left
		default:
			n = n.Right
		}
	}

	return
}

// Insert inserts the Comparable e into the Tree at the first match found
// with e or when a nil node is reached. Insertion without replacement can
// specified by ensuring that e.Compare() never returns 0. If insert without
// replacement is performed, a distinct query Comparable must be used that
// does return 0 for a Compare() call.
func (self *Tree) Insert(e Comparable) (root *Tree) {
	root = (*Tree)((*Node)(self).insert(e))
	root.Color = Black
	return
}

func (self *Node) insert(e Comparable) (root *Node) {
	if self == nil {
		return &Node{Elem: e}
	} else if self.Elem == nil {
		self.Elem = e
		return self
	}

	if Mode == TD234 {
		if self.Left.color() == Red && self.Right.color() == Red {
			self.flipColors()
		}
	}

	switch c := e.Compare(self.Elem); {
	case c == 0:
		self.Elem = e
	case c < 0:
		self.Left = self.Left.insert(e)
	default:
		self.Right = self.Right.insert(e)
	}

	if self.Right.color() == Red && self.Left.color() == Black {
		self = self.rotateLeft()
	}
	if self.Left.color() == Red && self.Left.Left.color() == Red {
		self = self.rotateRight()
	}

	if Mode == BU23 {
		if self.Left.color() == Red && self.Right.color() == Red {
			self.flipColors()
		}
	}

	return self
}

// DeleteMin deletes the node with the minimum value in the tree.
func (self *Tree) DeleteMin() (root *Tree) {
	root = (*Tree)((*Node)(self).deleteMin())
	root.Color = Black
	return
}

func (self *Node) deleteMin() *Node {
	if self.Left == nil {
		return nil
	}
	if self.Left.color() == Black && self.Left.Left.color() == Black {
		self = self.moveRedLeft()
	}
	self.Left = self.Left.deleteMin()
	return self.fixUp()
}

// DeleteMax deletes the node with the maximum value in the tree.
func (self *Tree) DeleteMax() (root *Tree) {
	root = (*Tree)((*Node)(self).deleteMax())
	root.Color = Black
	return
}

func (self *Node) deleteMax() *Node {
	if self.Left != nil && self.Left.Color == Red {
		self = self.rotateRight()
	}
	if self.Right == nil {
		return nil
	}
	if self.Right.color() == Black && self.Right.Left.color() == Black {
		self = self.moveRedRight()
	}
	self.Right = self.Right.deleteMax()
	return self.fixUp()
}

// Delete deletes the first node found that matches e according to Compare().
func (self *Tree) Delete(e Comparable) (root *Tree) {
	root = (*Tree)((*Node)(self).delete(e))
	if root == nil {
		return
	}
	root.Color = Black
	return
}

func (self *Node) delete(e Comparable) (root *Node) {
	if self == nil {
		return
	}
	if e.Compare(self.Elem) < 0 {
		if self.Left.color() == Black && self.Left != nil && self.Left.Left.color() == Black {
			self = self.moveRedLeft()
		}
		self.Left = self.Left.delete(e)
	} else {
		if self.Left.color() == Red {
			self = self.rotateRight()
		}
		if e.Compare(self.Elem) == 0 && self.Right == nil {
			return nil
		}
		if self.Right.color() == Black && self.Right != nil && self.Right.Left.color() == Black {
			self = self.moveRedRight()
		}
		if e.Compare(self.Elem) == 0 {
			self.Elem = self.Right.min().Elem
			self.Right = self.Right.deleteMin()
		} else {
			self.Right = self.Right.delete(e)
		}
	}
	return self.fixUp()
}

// Return the minimum value stored in the tree. This will be the left-most value if
// insertion without replacement is allowed.
func (self *Tree) Min() Comparable {
	return (*Node)(self).min().Elem
}

func (self *Node) min() *Node {
	if self.Left == nil {
		return self
	}
	return self.Left.min()
}

// Return the maximum value stored in the tree. This will be the right-most value if
// insertion without replacement is allowed.
func (self *Tree) Max() Comparable {
	return (*Node)(self).max().Elem
}

func (self *Node) max() *Node {
	if self.Right == nil {
		return self
	}
	return self.Right.max()
}

// An Operation is a function that operates on a Comparable.
type Operation func(Comparable)

// Do performs fn on all values stored in the tree. If fn alters the values of
// stored values, future tree operation behaviors are undefined.
func (self *Tree) Do(fn Operation) {
	(*Node)(self).do(fn)
}

func (self *Node) do(fn Operation) {
	if self == nil {
		return
	}
	self.Left.do(fn)
	fn(self.Elem)
	self.Right.do(fn)
}

// DoRange performs fn on all values stored in the tree between from and to. If from is
// less than to, the operations are performed from left to right. If from is greater than
// to then the operations are performed from right to left. If fn alters the values of
// stored values, future tree operation behaviors are undefined.
func (self *Tree) DoRange(fn Operation, from, to Comparable) {
	switch order := from.Compare(to); {
	case order < 0:
		(*Node)(self).doRange(fn, from, to)
	case order > 0:
		(*Node)(self).doRangeReverse(fn, from, to)
	default:
		(*Node)(self).doMatch(fn, from)
	}
}

func (self *Node) doRange(fn Operation, from, to Comparable) {
	if self == nil {
		return
	}
	fc, tc := from.Compare(self.Elem), to.Compare(self.Elem)
	if fc <= 0 {
		self.Left.doRange(fn, from, to)
	}
	if fc <= 0 && tc >= 0 {
		fn(self.Elem)
	}
	if tc >= 0 {
		self.Right.doRange(fn, from, to)
	}
}

func (self *Node) doRangeReverse(fn Operation, from, to Comparable) {
	if self == nil {
		return
	}
	fc, tc := from.Compare(self.Elem), to.Compare(self.Elem)
	if tc >= 0 {
		self.Right.doRangeReverse(fn, from, to)
	}
	if fc <= 0 && tc >= 0 {
		fn(self.Elem)
	}
	if fc <= 0 {
		self.Left.doRangeReverse(fn, from, to)
	}
}

// DoMatch performs fn on all values stored in the tree that match q according to Compare.
// If fn alters the values of stored values, future tree operation behaviors are undefined.
func (self *Tree) DoMatching(fn Operation, q Comparable) {
	(*Node)(self).doMatch(fn, q)
}

func (self *Node) doMatch(fn Operation, q Comparable) {
	if self == nil {
		return
	}
	c := q.Compare(self.Elem)
	if c <= 0 {
		self.Left.doMatch(fn, q)
	}
	if c == 0 {
		fn(self.Elem)
	}
	if c >= 0 {
		self.Right.doMatch(fn, q)
	}
}
