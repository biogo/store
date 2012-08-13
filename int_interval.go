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

package interval

import (
	"code.google.com/p/biogo.llrb"
)

// An IntOverlapper can determine whether it overlaps an integer range.
type IntOverlapper interface {
	// Overlap returns a boolean indicating whether the receiver overlaps a range.
	Overlap(IntRange) bool
}

// An IntRange is a type that describes the basic characteristics of an interval over the
// integer number line.
type IntRange struct {
	Min, Max int
}

// An IntInterface is a type that can be inserted into a IntTree.
type IntInterface interface {
	IntOverlapper
	Range() IntRange
	ID() uintptr // Returns a unique ID for the element.
}

// A IntNode represents a node in an IntTree.
type IntNode struct {
	Elem        IntInterface
	Interval    IntRange
	Range       IntRange
	Left, Right *IntNode
	Color       llrb.Color
}

// A IntTree manages the root node of an integer line interval tree.
// Public methods are exposed through this type.
type IntTree struct {
	Root  *IntNode // Root node of the tree.
	Count int      // Number of elements stored.
}

// Helper methods

// color returns the effect color of a IntNode. A nil node returns black.
func (self *IntNode) color() llrb.Color {
	if self == nil {
		return llrb.Black
	}
	return self.Color
}

// (a,c)b -rotL-> ((a,)b,)c
func (self *IntNode) rotateLeft() (root *IntNode) {
	// Assumes: self has two children.
	root = self.Right
	if root.Left != nil {
		self.Range.Max = intMax(self.Interval.Max, root.Left.Range.Max)
	} else {
		self.Range.Max = self.Interval.Max
	}
	root.Range.Min = intMin(root.Interval.Min, self.Range.Min)
	self.Right = root.Left
	root.Left = self
	root.Color = self.Color
	self.Color = llrb.Red
	return
}

// (a,c)b -rotR-> (,(,c)b)a
func (self *IntNode) rotateRight() (root *IntNode) {
	// Assumes: self has two children.
	root = self.Left
	if root.Right != nil {
		self.Range.Min = intMin(self.Interval.Min, root.Right.Range.Min)
	} else {
		self.Range.Min = self.Interval.Min
	}
	root.Range.Max = intMax(root.Interval.Max, self.Range.Max)
	self.Left = root.Right
	root.Right = self
	root.Color = self.Color
	self.Color = llrb.Red
	return
}

// (aR,cR)bB -flipC-> (aB,cB)bR | (aB,cB)bR -flipC-> (aR,cR)bB 
func (self *IntNode) flipColors() {
	// Assumes: self has two children.
	self.Color = !self.Color
	self.Left.Color = !self.Left.Color
	self.Right.Color = !self.Right.Color
}

// fixUp ensures that black link balance is correct, that red nodes lean left,
// and that 4 nodes are split in the case of BU23 and properly balanced in TD234.
func (self *IntNode) fixUp(fast bool) *IntNode {
	if !fast {
		self.adjustRange()
	}
	if self.Right.color() == llrb.Red {
		if Mode == TD234 && self.Right.Left.color() == llrb.Red {
			self.Right = self.Right.rotateRight()
		}
		self = self.rotateLeft()
	}
	if self.Left.color() == llrb.Red && self.Left.Left.color() == llrb.Red {
		self = self.rotateRight()
	}
	if Mode == BU23 && self.Left.color() == llrb.Red && self.Right.color() == llrb.Red {
		self.flipColors()
	}

	return self
}

// adjustRange sets the Range to the maximum extent of the childrens' Range
// spans and the node's Elem span.
func (self *IntNode) adjustRange() {
	if self.Left != nil {
		self.Range.Min = intMin(self.Interval.Min, self.Left.Range.Min)
		self.Range.Max = intMax(self.Interval.Max, self.Left.Range.Max)
	}
	if self.Right != nil {
		self.Range.Max = intMax(self.Interval.Max, self.Right.Range.Max)
	}
}

func intMin(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func intMax(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (self *IntNode) moveRedLeft() *IntNode {
	self.flipColors()
	if self.Right.Left.color() == llrb.Red {
		self.Right = self.Right.rotateRight()
		self = self.rotateLeft()
		self.flipColors()
		if Mode == TD234 && self.Right.Right.color() == llrb.Red {
			self.Right = self.Right.rotateLeft()
		}
	}
	return self
}

func (self *IntNode) moveRedRight() *IntNode {
	self.flipColors()
	if self.Left.Left.color() == llrb.Red {
		self = self.rotateRight()
		self.flipColors()
	}
	return self
}

// Len returns the number of intervals stored in the IntTree.
func (self *IntTree) Len() int {
	return self.Count
}

// Get returns a slice of IntInterfaces that overlap q in the IntTree according
// to Overlap.
func (self *IntTree) Get(q IntOverlapper) (o []IntInterface) {
	if self.Root != nil && q.Overlap(self.Root.Range) {
		self.Root.doMatch(func(e IntInterface) (done bool) { o = append(o, e); return }, q)
	}
	return
}

// AdjustRanges fixes range fields for all IntNodes in the IntTree. This must be called
// before Get or DoMatching* is used if fast insertion or deletion has been performed.
func (self *IntTree) AdjustRanges() {
	if self.Root == nil {
		return
	}
	self.Root.adjustRanges()
}

func (self *IntNode) adjustRanges() {
	if self.Left != nil {
		self.Left.adjustRanges()
	}
	if self.Right != nil {
		self.Right.adjustRanges()
	}
	self.adjustRange()
}

// Insert inserts the IntInterface e into the IntTree. Insertions may replace
// existing stored intervals.
func (self *IntTree) Insert(e IntInterface, fast bool) (err error) {
	if r := e.Range(); r.Min > r.Max {
		return ErrInvertedRange
	}
	var d int
	self.Root, d = self.Root.insert(e, e.Range(), e.ID(), fast)
	self.Count += d
	self.Root.Color = llrb.Black
	return
}

func (self *IntNode) insert(e IntInterface, r IntRange, id uintptr, fast bool) (root *IntNode, d int) {
	if self == nil {
		return &IntNode{Elem: e, Interval: r, Range: r}, 1
	} else if self.Elem == nil {
		self.Elem = e
		self.Interval = r
		if !fast {
			self.adjustRange()
		}
		return self, 1
	}

	if Mode == TD234 {
		if self.Left.color() == llrb.Red && self.Right.color() == llrb.Red {
			self.flipColors()
		}
	}

	switch c := r.Min - self.Interval.Min; {
	case c == 0:
		switch cid := id - self.Elem.ID(); {
		case cid == 0:
			self.Elem = e
			self.Interval = r
			if !fast {
				self.Range.Max = r.Max
			}
		case cid < 0:
			self.Left, d = self.Left.insert(e, r, id, fast)
		default:
			self.Right, d = self.Right.insert(e, r, id, fast)
		}
	case c < 0:
		self.Left, d = self.Left.insert(e, r, id, fast)
	default:
		self.Right, d = self.Right.insert(e, r, id, fast)
	}

	if self.Right.color() == llrb.Red && self.Left.color() == llrb.Black {
		self = self.rotateLeft()
	}
	if self.Left.color() == llrb.Red && self.Left.Left.color() == llrb.Red {
		self = self.rotateRight()
	}

	if Mode == BU23 {
		if self.Left.color() == llrb.Red && self.Right.color() == llrb.Red {
			self.flipColors()
		}
	}

	if !fast {
		self.adjustRange()
	}
	root = self

	return
}

// DeleteMin deletes the left-most interval.
func (self *IntTree) DeleteMin(fast bool) {
	if self.Root == nil {
		return
	}
	var d int
	self.Root, d = self.Root.deleteMin(fast)
	self.Count += d
	if self.Root == nil {
		return
	}
	self.Root.Color = llrb.Black
}

func (self *IntNode) deleteMin(fast bool) (root *IntNode, d int) {
	if self.Left == nil {
		return nil, -1
	}
	if self.Left.color() == llrb.Black && self.Left.Left.color() == llrb.Black {
		self = self.moveRedLeft()
	}
	self.Left, d = self.Left.deleteMin(fast)
	if self.Left == nil {
		self.Range.Min = self.Elem.Range().Min
	}

	root = self.fixUp(fast)

	return
}

// DeleteMax deletes the right-most interval.
func (self *IntTree) DeleteMax(fast bool) {
	if self.Root == nil {
		return
	}
	var d int
	self.Root, d = self.Root.deleteMax(fast)
	self.Count += d
	if self.Root == nil {
		return
	}
	self.Root.Color = llrb.Black
}

func (self *IntNode) deleteMax(fast bool) (root *IntNode, d int) {
	if self.Left != nil && self.Left.color() == llrb.Red {
		self = self.rotateRight()
	}
	if self.Right == nil {
		return nil, -1
	}
	if self.Right.color() == llrb.Black && self.Right.Left.color() == llrb.Black {
		self = self.moveRedRight()
	}
	self.Right, d = self.Right.deleteMax(fast)
	if self.Right == nil {
		self.Range.Max = self.Elem.Range().Max
	}

	root = self.fixUp(fast)

	return
}

// Delete deletes the element e if it exists in the IntTree.
func (self *IntTree) Delete(e IntInterface, fast bool) (err error) {
	if r := e.Range(); r.Min > r.Max {
		return ErrInvertedRange
	}
	if self.Root == nil || !e.Overlap(self.Root.Range) {
		return
	}
	var d int
	self.Root, d = self.Root.delete(e.Range().Min, e.ID(), fast)
	self.Count += d
	if self.Root == nil {
		return
	}
	self.Root.Color = llrb.Black
	return
}

func (self *IntNode) delete(m int, id uintptr, fast bool) (root *IntNode, d int) {
	if p := m - self.Interval.Min; p < 0 || (p == 0 && id < self.Elem.ID()) {
		if self.Left != nil {
			if self.Left.color() == llrb.Black && self.Left.Left.color() == llrb.Black {
				self = self.moveRedLeft()
			}
			self.Left, d = self.Left.delete(m, id, fast)
			if self.Left == nil {
				self.Range.Min = self.Interval.Min
			}
		}
	} else {
		if self.Left.color() == llrb.Red {
			self = self.rotateRight()
		}
		if self.Right == nil && id == self.Elem.ID() {
			return nil, -1
		}
		if self.Right != nil {
			if self.Right.color() == llrb.Black && self.Right.Left.color() == llrb.Black {
				self = self.moveRedRight()
			}
			if id == self.Elem.ID() {
				m := self.Right.min()
				self.Elem = m.Elem
				self.Interval = m.Interval
				self.Right, d = self.Right.deleteMin(fast)
			} else {
				self.Right, d = self.Right.delete(m, id, fast)
			}
			if self.Right == nil {
				self.Range.Max = self.Interval.Max
			}
		}
	}

	root = self.fixUp(fast)

	return
}

// Return the left-most interval stored in the tree.
func (self *IntTree) Min() IntInterface {
	if self.Root == nil {
		return nil
	}
	return self.Root.min().Elem
}

func (self *IntNode) min() (n *IntNode) {
	for n = self; n.Left != nil; n = n.Left {
	}
	return
}

// Return the right-most interval stored in the tree.
func (self *IntTree) Max() IntInterface {
	if self.Root == nil {
		return nil
	}
	return self.Root.max().Elem
}

func (self *IntNode) max() (n *IntNode) {
	for n = self; n.Right != nil; n = n.Right {
	}
	return
}

// Floor returns the largest value equal to or less than the query q according to
// q.Min().Compare(), with ties broken by q.ID().Compare().
func (self *IntTree) Floor(q IntInterface) (o IntInterface, err error) {
	if self.Root == nil {
		return
	}
	n := self.Root.floor(q.Range().Min, q.ID())
	if n == nil {
		return
	}
	return n.Elem, nil
}

func (self *IntNode) floor(m int, id uintptr) *IntNode {
	if self == nil {
		return nil
	}
	switch c := m - self.Interval.Min; {
	case c == 0:
		switch cid := id - self.Elem.ID(); {
		case cid == 0:
			return self
		case cid < 0:
			return self.Left.floor(m, id)
		default:
			if r := self.Right.floor(m, id); r != nil {
				return r
			}
		}
	case c < 0:
		return self.Left.floor(m, id)
	default:
		if r := self.Right.floor(m, id); r != nil {
			return r
		}
	}
	return self
}

// Ceil returns the smallest value equal to or greater than the query q according to
// q.Min().Compare(), with ties broken by q.ID().Compare().
func (self *IntTree) Ceil(q IntInterface) (o IntInterface, err error) {
	if self.Root == nil {
		return
	}
	n := self.Root.ceil(q.Range().Min, q.ID())
	if n == nil {
		return
	}
	return n.Elem, nil
}

func (self *IntNode) ceil(m int, id uintptr) *IntNode {
	if self == nil {
		return nil
	}
	switch c := m - self.Interval.Min; {
	case c == 0:
		switch cid := id - self.Elem.ID(); {
		case cid == 0:
			return self
		case cid > 0:
			return self.Right.ceil(m, id)
		default:
			if l := self.Left.ceil(m, id); l != nil {
				return l
			}
		}
	case c > 0:
		return self.Right.ceil(m, id)
	default:
		if l := self.Left.ceil(m, id); l != nil {
			return l
		}
	}
	return self
}

// An IntOperation is a function that operates on an IntInterface. If done is returned true, the
// IntOperation is indicating that no further work needs to be done and so the Do function should
// traverse no further.
type IntOperation func(IntInterface) (done bool)

// Do performs fn on all intervals stored in the tree. A boolean is returned indicating whether the
// Do traversal was interrupted by an IntOperation returning true. If fn alters stored intervals'
// end points, future tree operation behaviors are undefined.
func (self *IntTree) Do(fn IntOperation) bool {
	if self.Root == nil {
		return false
	}
	return self.Root.do(fn)
}

func (self *IntNode) do(fn IntOperation) (done bool) {
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

// DoReverse performs fn on all intervals stored in the tree, but in reverse of sort order. A boolean
// is returned indicating whether the Do traversal was interrupted by an IntOperation returning true.
// If fn alters stored intervals' end points, future tree operation behaviors are undefined.
func (self *IntTree) DoReverse(fn IntOperation) bool {
	if self.Root == nil {
		return false
	}
	return self.Root.doReverse(fn)
}

func (self *IntNode) doReverse(fn IntOperation) (done bool) {
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

// DoMatch performs fn on all intervals stored in the tree that match q according to Overlap, with
// q.Overlap() used to guide tree traversal, so DoMatching() will out perform Do() with a called
// conditional function if the condition is based on sort order, but can not be reliably used if
// the condition is independent of sort order. A boolean is returned indicating whether the Do
// traversal was interrupted by an IntOperation returning true. If fn alters stored intervals' end
// points, future tree operation behaviors are undefined.
func (self *IntTree) DoMatching(fn IntOperation, q IntOverlapper) (t bool) {
	if self.Root != nil && q.Overlap(self.Root.Range) {
		return self.Root.doMatch(fn, q)
	}
	return
}

func (self *IntNode) doMatch(fn IntOperation, q IntOverlapper) (done bool) {
	if self.Left != nil && q.Overlap(self.Left.Range) {
		done = self.Left.doMatch(fn, q)
		if done {
			return
		}
	}
	if q.Overlap(self.Interval) {
		done = fn(self.Elem)
		if done {
			return
		}
	}
	if self.Right != nil && q.Overlap(self.Right.Range) {
		done = self.Right.doMatch(fn, q)
	}
	return
}

// DoMatchReverse performs fn on all intervals stored in the tree that match q according to Overlap,
// with q.Overlap() used to guide tree traversal, so DoMatching() will out perform Do() with a called
// conditional function if the condition is based on sort order, but can not be reliably used if
// the condition is independent of sort order. A boolean is returned indicating whether the Do
// traversal was interrupted by an IntOperation returning true. If fn alters stored intervals' end
// points, future tree operation behaviors are undefined.
func (self *IntTree) DoMatchingReverse(fn IntOperation, q IntOverlapper) (t bool) {
	if self.Root != nil && q.Overlap(self.Root.Range) {
		return self.Root.doMatch(fn, q)
	}
	return
}

func (self *IntNode) doMatchReverse(fn IntOperation, q IntOverlapper) (done bool) {
	if self.Right != nil && q.Overlap(self.Right.Range) {
		done = self.Right.doMatchReverse(fn, q)
		if done {
			return
		}
	}
	if q.Overlap(self.Interval) {
		done = fn(self.Elem)
		if done {
			return
		}
	}
	if self.Left != nil && q.Overlap(self.Left.Range) {
		done = self.Left.doMatchReverse(fn, q)
	}
	return
}
