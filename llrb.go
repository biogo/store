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

// A Tree manages the root node of an LLRB tree. Public methods are exposed through this type.
type Tree struct {
	Root  *Node // Root node of the tree.
	Count int   // Number of elements stored.
}

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
	if self.Root == nil {
		return nil
	}
	n := self.Root.search(q)
	if n == nil {
		return nil
	}
	return n.Elem
}

// Len returns the number of elements stored in the Tree.
func (self *Tree) Len() int {
	return self.Count
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
// can return 0 with a Compare() call.
func (self *Tree) Insert(e Comparable) {
	var d int
	self.Root, d = self.Root.insert(e)
	self.Count += d
	self.Root.Color = Black
}

func (self *Node) insert(e Comparable) (root *Node, d int) {
	if self == nil {
		return &Node{Elem: e}, 1
	} else if self.Elem == nil {
		self.Elem = e
		return self, 1
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
		self.Left, d = self.Left.insert(e)
	default:
		self.Right, d = self.Right.insert(e)
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

	return self, d
}

// DeleteMin deletes the node with the minimum value in the tree. If insertion without
// replacement has been used the the left-most minimum will be deleted.
func (self *Tree) DeleteMin() {
	if self.Root == nil {
		return
	}
	var d int
	self.Root, d = self.Root.deleteMin()
	self.Count += d
	if self.Root == nil {
		return
	}
	self.Root.Color = Black
}

func (self *Node) deleteMin() (root *Node, d int) {
	if self.Left == nil {
		return nil, -1
	}
	if self.Left.color() == Black && self.Left.Left.color() == Black {
		self = self.moveRedLeft()
	}
	self.Left, d = self.Left.deleteMin()
	return self.fixUp(), d
}

// DeleteMax deletes the node with the maximum value in the tree. If insertion without
// replacement has been used the the right-most maximum will be deleted.
func (self *Tree) DeleteMax() {
	if self.Root == nil {
		return
	}
	var d int
	self.Root, d = self.Root.deleteMax()
	self.Count += d
	if self.Root == nil {
		return
	}
	self.Root.Color = Black
}

func (self *Node) deleteMax() (root *Node, d int) {
	if self.Left != nil && self.Left.color() == Red {
		self = self.rotateRight()
	}
	if self.Right == nil {
		return nil, -1
	}
	if self.Right.color() == Black && self.Right.Left.color() == Black {
		self = self.moveRedRight()
	}
	self.Right, d = self.Right.deleteMax()
	return self.fixUp(), d
}

// Delete deletes the first node found that matches e according to Compare().
func (self *Tree) Delete(e Comparable) {
	if self.Root == nil {
		return
	}
	var d int
	self.Root, d = self.Root.delete(e)
	self.Count += d
	if self.Root == nil {
		return
	}
	self.Root.Color = Black
}

func (self *Node) delete(e Comparable) (root *Node, d int) {
	if e.Compare(self.Elem) < 0 {
		if self.Left != nil {
			if self.Left.color() == Black && self.Left.Left.color() == Black {
				self = self.moveRedLeft()
			}
			self.Left, d = self.Left.delete(e)
		}
	} else {
		if self.Left.color() == Red {
			self = self.rotateRight()
		}
		if e.Compare(self.Elem) == 0 && self.Right == nil {
			return nil, -1
		}
		if self.Right != nil {
			if self.Right.color() == Black && self.Right.Left.color() == Black {
				self = self.moveRedRight()
			}
			if e.Compare(self.Elem) == 0 {
				self.Elem = self.Right.min().Elem
				self.Right, d = self.Right.deleteMin()
			} else {
				self.Right, d = self.Right.delete(e)
			}
		}
	}
	return self.fixUp(), d
}

// Return the minimum value stored in the tree. This will be the left-most minimum value if
// insertion without replacement has been used.
func (self *Tree) Min() Comparable {
	if self.Root == nil {
		return nil
	}
	return self.Root.min().Elem
}

func (self *Node) min() (n *Node) {
	for n = self; n.Left != nil; n = n.Left {
	}
	return
}

// Return the maximum value stored in the tree. This will be the right-most maximum value if
// insertion without replacement has been used.
func (self *Tree) Max() Comparable {
	if self.Root == nil {
		return nil
	}
	return self.Root.max().Elem
}

func (self *Node) max() (n *Node) {
	for n = self; n.Right != nil; n = n.Right {
	}
	return
}

// An Operation is a function that operates on a Comparable.
type Operation func(Comparable)

// Do performs fn on all values stored in the tree. If fn alters stored values, future tree
// operation behaviors are undefined.
func (self *Tree) Do(fn Operation) {
	if self.Root == nil {
		return
	}
	self.Root.do(fn)
}

func (self *Node) do(fn Operation) {
	if self.Left != nil {
		self.Left.do(fn)
	}
	fn(self.Elem)
	if self.Right != nil {
		self.Right.do(fn)
	}
}

// DoRange performs fn on all values stored in the tree between from and to. If from is
// less than to, the operations are performed from left to right. If from is greater than
// to then the operations are performed from right to left. If fn alters stored values,
// future tree operation behaviors are undefined.
func (self *Tree) DoRange(fn Operation, from, to Comparable) {
	if self.Root == nil {
		return
	}
	switch order := from.Compare(to); {
	case order < 0:
		self.Root.doRange(fn, from, to)
	case order > 0:
		self.Root.doRangeReverse(fn, from, to)
	default:
		self.Root.doMatch(fn, from)
	}
}

func (self *Node) doRange(fn Operation, from, to Comparable) {
	fc, tc := from.Compare(self.Elem), to.Compare(self.Elem)
	if fc <= 0 && self.Left != nil {
		self.Left.doRange(fn, from, to)
	}
	if fc <= 0 && tc >= 0 {
		fn(self.Elem)
	}
	if tc >= 0 && self.Right != nil {
		self.Right.doRange(fn, from, to)
	}
}

func (self *Node) doRangeReverse(fn Operation, from, to Comparable) {
	fc, tc := from.Compare(self.Elem), to.Compare(self.Elem)
	if tc >= 0 && self.Right != nil {
		self.Right.doRangeReverse(fn, from, to)
	}
	if fc <= 0 && tc >= 0 {
		fn(self.Elem)
	}
	if fc <= 0 && self.Left != nil {
		self.Left.doRangeReverse(fn, from, to)
	}
}

// DoMatch performs fn on all values stored in the tree that match q according to Compare.
// If fn alters stored values, future tree operation behaviors are undefined.
func (self *Tree) DoMatching(fn Operation, q Comparable) {
	if self.Root == nil {
		return
	}
	self.Root.doMatch(fn, q)
}

func (self *Node) doMatch(fn Operation, q Comparable) {
	c := q.Compare(self.Elem)
	if c <= 0 && self.Left != nil {
		self.Left.doMatch(fn, q)
	}
	if c == 0 {
		fn(self.Elem)
	}
	if c >= 0 && self.Right != nil {
		self.Right.doMatch(fn, q)
	}
}
