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

// Operation mode of the LLRB tree.
const Mode = BU23

func init() {
	if Mode != TD234 && Mode != BU23 {
		panic("llrb: unknown mode")
	}
}

// A Comparable is a type that can be inserted into a Tree or used as a range
// or equality query on the tree,
type Comparable interface {
	// Compare returns a value indicating the sort order relationship between the
	// receiver and the parameter.
	//
	// Given c = a.Compare(b):
	//  c < 0 if a < b;
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
// and that 4 nodes are split in the case of BU23 and properly balanced in TD234.
func (self *Node) fixUp() *Node {
	if self.Right.color() == Red {
		if Mode == TD234 && self.Right.Left.color() == Red {
			self.Right = self.Right.rotateRight()
		}
		self = self.rotateLeft()
	}
	if self.Left.color() == Red && self.Left.Left.color() == Red {
		self = self.rotateRight()
	}
	if Mode == BU23 && self.Left.color() == Red && self.Right.color() == Red {
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
		if Mode == TD234 && self.Right.Right.color() == Red {
			self.Right = self.Right.rotateLeft()
		}
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

// Len returns the number of elements stored in the Tree.
func (self *Tree) Len() int {
	return self.Count
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

	root = self

	return
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

	root = self.fixUp()

	return
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

	root = self.fixUp()

	return
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

	root = self.fixUp()

	return
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

// Floor returns the greatest value equal to or less than the query q according to q.Compare().
func (self *Tree) Floor(q Comparable) Comparable {
	if self.Root == nil {
		return nil
	}
	n := self.Root.floor(q)
	if n == nil {
		return nil
	}
	return n.Elem
}

func (self *Node) floor(q Comparable) *Node {
	if self == nil {
		return nil
	}
	switch c := q.Compare(self.Elem); {
	case c == 0:
		return self
	case c < 0:
		return self.Left.floor(q)
	default:
		if r := self.Right.floor(q); r != nil {
			return r
		}
	}
	return self
}

// Ceil returns the smallest value equal to or greater than the query q according to q.Compare().
func (self *Tree) Ceil(q Comparable) Comparable {
	if self.Root == nil {
		return nil
	}
	n := self.Root.ceil(q)
	if n == nil {
		return nil
	}
	return n.Elem
}

func (self *Node) ceil(q Comparable) *Node {
	if self == nil {
		return nil
	}
	switch c := q.Compare(self.Elem); {
	case c == 0:
		return self
	case c > 0:
		return self.Right.ceil(q)
	default:
		if l := self.Left.ceil(q); l != nil {
			return l
		}
	}
	return self
}

// An Operation is a function that operates on a Comparable. If done is returned true, the
// Operation is indicating that no further work needs to be done and so the Do function should
// traverse no further.
type Operation func(Comparable) (done bool)

// Do performs fn on all values stored in the tree. A boolean is returned indicating whether the
// Do traversal was interupted by an Operation returning true. If fn alters stored values' sort
// relationships, future tree operation behaviors are undefined.
func (self *Tree) Do(fn Operation) bool {
	if self.Root == nil {
		return false
	}
	return self.Root.do(fn)
}

func (self *Node) do(fn Operation) (done bool) {
	if self.Left != nil {
		done = self.Left.do(fn)
		if done {
			return
		}
	}
	done = fn(self.Elem)
	if done {
		return
	}
	if self.Right != nil {
		done = self.Right.do(fn)
	}
	return
}

// DoReverse performs fn on all values stored in the tree, but in reverse of sort order. A boolean
// is returned indicating whether the Do traversal was interupted by an Operation returning true.
// If fn alters stored values' sort relationships, future tree operation behaviors are undefined.
func (self *Tree) DoReverse(fn Operation) bool {
	if self.Root == nil {
		return false
	}
	return self.Root.doReverse(fn)
}

func (self *Node) doReverse(fn Operation) (done bool) {
	if self.Right != nil {
		done = self.Right.doReverse(fn)
		if done {
			return
		}
	}
	done = fn(self.Elem)
	if done {
		return
	}
	if self.Left != nil {
		done = self.Left.doReverse(fn)
	}
	return
}

// DoRange performs fn on all values stored in the tree over the interval [from, to) from left
// to right. If to equals from the call is a no-op, and if to is less than from DoRange will
// panic. A boolean is returned indicating whether the Do traversal was interupted by an
// Operation returning true. If fn alters stored values' sort relationships future tree
// operation behaviors are undefined.
func (self *Tree) DoRange(fn Operation, from, to Comparable) bool {
	if self.Root == nil {
		return false
	}
	switch order := from.Compare(to); {
	case order < 0:
		return self.Root.doRange(fn, from, to)
	case order > 0:
		panic("llrb: inverted range")
	}
	return false
}

func (self *Node) doRange(fn Operation, lo, hi Comparable) (done bool) {
	lc, hc := lo.Compare(self.Elem), hi.Compare(self.Elem)
	if lc <= 0 && self.Left != nil {
		done = self.Left.doRange(fn, lo, hi)
		if done {
			return
		}
	}
	if lc <= 0 && hc > 0 {
		done = fn(self.Elem)
		if done {
			return
		}
	}
	if hc > 0 && self.Right != nil {
		done = self.Right.doRange(fn, lo, hi)
	}
	return
}

// DoRangeReverse performs fn on all values stored in the tree over the interval [to, from) from
// right to left. If from equals to the call is a no-op, and if from is less than to DoRange will
// panic. A boolean is returned indicating whether the Do traversal was interupted by an Operation
// returning true. If fn alters stored values' sort relationships future tree operation behaviors
// are undefined.
func (self *Tree) DoRangeReverse(fn Operation, from, to Comparable) bool {
	if self.Root == nil {
		return false
	}
	switch order := from.Compare(to); {
	case order > 0:
		return self.Root.doRangeReverse(fn, from, to)
	case order < 0:
		panic("llrb: inverted range")
	}
	return false
}

func (self *Node) doRangeReverse(fn Operation, hi, lo Comparable) (done bool) {
	lc, hc := lo.Compare(self.Elem), hi.Compare(self.Elem)
	if hc > 0 && self.Right != nil {
		done = self.Right.doRangeReverse(fn, hi, lo)
		if done {
			return
		}
	}
	if lc <= 0 && hc > 0 {
		done = fn(self.Elem)
		if done {
			return
		}
	}
	if lc <= 0 && self.Left != nil {
		done = self.Left.doRangeReverse(fn, hi, lo)
	}
	return
}

// DoMatch performs fn on all values stored in the tree that match q according to Compare, with
// q.Compare() used to guide tree traversal, so DoMatching() will out perform Do() with a called
// conditional function if the condition is based on sort order, but can not be reliably used if
// the condition is independent of sort order. A boolean is returned indicating whether the Do
// traversal was interupted by an Operation returning true.If fn alters stored values' sort
// relationships, future tree operation behaviors are undefined.
func (self *Tree) DoMatching(fn Operation, q Comparable) bool {
	if self.Root == nil {
		return false
	}
	return self.Root.doMatch(fn, q)
}

func (self *Node) doMatch(fn Operation, q Comparable) (done bool) {
	c := q.Compare(self.Elem)
	if c <= 0 && self.Left != nil {
		done = self.Left.doMatch(fn, q)
		if done {
			return
		}
	}
	if c == 0 {
		done = fn(self.Elem)
		if done {
			return
		}
	}
	if c >= 0 && self.Right != nil {
		done = self.Right.doMatch(fn, q)
	}
	return
}
