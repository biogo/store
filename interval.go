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

// Package interval implements an interval tree based on an augmented
// Left-Leaning Red Black tree.
package interval

import (
	"code.google.com/p/biogo.llrb"
	"errors"
)

const (
	TD234 = iota
	BU23
)

// Operation mode of the underlying LLRB tree.
const Mode = BU23

func init() {
	if Mode != TD234 && Mode != BU23 {
		panic("interval: unknown mode")
	}
}

// ErrInvertedRange is returned if an Range is used where the minimum value is
// greater than the maximum value according to Compare().
var ErrInvertedRange = errors.New("interval: inverted range")

// An Overlapper can determine whether it overlaps a range.
type Overlapper interface {
	// Overlap returns a boolean indicating whether the receiver overlaps the parameter.
	Overlap(Range) bool
}

// A Range is a type that describes the basic characteristics of an interval.
type Range interface {
	// Return a Comparable equal to the Minimum value of the Overlapper.
	Min() Comparable
	// Return a Comparable equal to the Maximum value of the Overlapper.
	Max() Comparable
}

// An Interface is a type that can be inserted into a Tree.
type Interface interface {
	Overlapper
	Range
	ID() Comparable      // Returns a unique ID for the element.
	NewMutable() Mutable // Returns an mutable copy of the Interface's range.
}

// A Mutable is a Range that can have its range altered.
type Mutable interface {
	Range
	SetMin(Comparable) // Set the minimum value.
	SetMax(Comparable) // Set the maximum value.
}

// A Comparable is a type that describes the ends of an Overlapper.
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

// A Node represents a node in the LLRB tree.
type Node struct {
	Elem        Interface
	Range       Mutable
	Left, Right *Node
	Color       llrb.Color
}

// A Tree manages the root node of an interval tree. Public methods are exposed through this type.
type Tree struct {
	Root  *Node // Root node of the tree.
	Count int   // Number of elements stored.
}

// Helper methods

// color returns the effect color of a Node. A nil node returns black.
func (self *Node) color() llrb.Color {
	if self == nil {
		return llrb.Black
	}
	return self.Color
}

// (a,c)b -rotL-> ((a,)b,)c
func (self *Node) rotateLeft() (root *Node) {
	// Assumes: self has two children.
	root = self.Right
	if root.Left != nil {
		self.Range.SetMax(max(self.Elem.Max(), root.Left.Range.Max()))
	} else {
		self.Range.SetMax(self.Elem.Max())
	}
	root.Range.SetMin(min(root.Elem.Min(), self.Range.Min()))
	self.Right = root.Left
	root.Left = self
	root.Color = self.Color
	self.Color = llrb.Red
	return
}

// (a,c)b -rotR-> (,(,c)b)a
func (self *Node) rotateRight() (root *Node) {
	// Assumes: self has two children.
	root = self.Left
	if root.Right != nil {
		self.Range.SetMin(min(self.Elem.Min(), root.Right.Range.Min()))
	} else {
		self.Range.SetMin(self.Elem.Min())
	}
	root.Range.SetMax(max(root.Elem.Max(), self.Range.Max()))
	self.Left = root.Right
	root.Right = self
	root.Color = self.Color
	self.Color = llrb.Red
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
func (self *Node) fixUp(fast bool) *Node {
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
func (self *Node) adjustRange() {
	if self.Left != nil {
		self.Range.SetMin(min(self.Elem.Min(), self.Left.Range.Min()))
		self.Range.SetMax(max(self.Elem.Max(), self.Left.Range.Max()))
	}
	if self.Right != nil {
		self.Range.SetMax(max(self.Elem.Max(), self.Right.Range.Max()))
	}
}

func min(a, b Comparable) Comparable {
	if a.Compare(b) < 0 {
		return a
	}
	return b
}

func max(a, b Comparable) Comparable {
	if a.Compare(b) > 0 {
		return a
	}
	return b
}

func (self *Node) moveRedLeft() *Node {
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

func (self *Node) moveRedRight() *Node {
	self.flipColors()
	if self.Left.Left.color() == llrb.Red {
		self = self.rotateRight()
		self.flipColors()
	}
	return self
}

// Len returns the number of intervals stored in the Tree.
func (self *Tree) Len() int {
	return self.Count
}

// Get returns a slice of Interfaces that overlap q in the Tree according
// to Overlap.
func (self *Tree) Get(q Overlapper) (o []Interface) {
	if self.Root != nil && q.Overlap(self.Root.Range) {
		self.Root.doMatch(func(e Interface) (done bool) { o = append(o, e); return }, q)
	}
	return
}

// AdjustRanges fixes range fields for all Nodes in the Tree. This must be called
// before Get or DoMatching* is used if fast insertion or deletion has been performed.
func (self *Tree) AdjustRanges() {
	if self.Root == nil {
		return
	}
	self.Root.adjustRanges()
}

func (self *Node) adjustRanges() {
	if self.Left != nil {
		self.Left.adjustRanges()
	}
	if self.Right != nil {
		self.Right.adjustRanges()
	}
	self.adjustRange()
}

// Insert inserts the Interface e into the Tree. Insertions may replace
// existing stored intervals.
func (self *Tree) Insert(e Interface, fast bool) (err error) {
	if e.Min().Compare(e.Max()) > 0 {
		return ErrInvertedRange
	}
	var d int
	self.Root, d = self.Root.insert(e, e.Min(), e.ID(), fast)
	self.Count += d
	self.Root.Color = llrb.Black
	return
}

func (self *Node) insert(e Interface, min, id Comparable, fast bool) (root *Node, d int) {
	if self == nil {
		return &Node{Elem: e, Range: e.NewMutable()}, 1
	} else if self.Elem == nil {
		self.Elem = e
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

	switch c := min.Compare(self.Elem.Min()); {
	case c == 0:
		switch cid := id.Compare(self.Elem.ID()); {
		case cid == 0:
			self.Elem = e
			if !fast {
				self.Range.SetMax(e.Max())
			}
		case cid < 0:
			self.Left, d = self.Left.insert(e, min, id, fast)
		default:
			self.Right, d = self.Right.insert(e, min, id, fast)
		}
	case c < 0:
		self.Left, d = self.Left.insert(e, min, id, fast)
	default:
		self.Right, d = self.Right.insert(e, min, id, fast)
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
func (self *Tree) DeleteMin(fast bool) {
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

func (self *Node) deleteMin(fast bool) (root *Node, d int) {
	if self.Left == nil {
		return nil, -1
	}
	if self.Left.color() == llrb.Black && self.Left.Left.color() == llrb.Black {
		self = self.moveRedLeft()
	}
	self.Left, d = self.Left.deleteMin(fast)
	if self.Left == nil {
		self.Range.SetMin(self.Elem.Min())
	}

	root = self.fixUp(fast)

	return
}

// DeleteMax deletes the right-most interval.
func (self *Tree) DeleteMax(fast bool) {
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

func (self *Node) deleteMax(fast bool) (root *Node, d int) {
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
		self.Range.SetMax(self.Elem.Max())
	}

	root = self.fixUp(fast)

	return
}

// Delete deletes the element e if it exists in the Tree.
func (self *Tree) Delete(e Interface, fast bool) (err error) {
	if e.Min().Compare(e.Max()) > 0 {
		return ErrInvertedRange
	}
	if self.Root == nil || !e.Overlap(self.Root.Range) {
		return
	}
	var d int
	self.Root, d = self.Root.delete(e, e.Min(), e.ID(), fast)
	self.Count += d
	if self.Root == nil {
		return
	}
	self.Root.Color = llrb.Black
	return
}

func (self *Node) delete(e Interface, min, id Comparable, fast bool) (root *Node, d int) {
	if p := min.Compare(self.Elem.Min()); p < 0 || (p == 0 && id.Compare(self.Elem.ID()) < 0) {
		if self.Left != nil && e.Overlap(self.Left.Range) {
			if self.Left.color() == llrb.Black && self.Left.Left.color() == llrb.Black {
				self = self.moveRedLeft()
			}
			self.Left, d = self.Left.delete(e, min, id, fast)
			if self.Left == nil {
				self.Range.SetMin(self.Elem.Min())
			}
		}
	} else {
		if self.Left.color() == llrb.Red {
			self = self.rotateRight()
		}
		if self.Right == nil && id.Compare(self.Elem.ID()) == 0 {
			return nil, -1
		}
		if self.Right != nil {
			if self.Right.color() == llrb.Black && self.Right.Left.color() == llrb.Black {
				self = self.moveRedRight()
			}
			if id.Compare(self.Elem.ID()) == 0 {
				self.Elem = self.Right.min().Elem
				self.Right, d = self.Right.deleteMin(fast)
			} else {
				self.Right, d = self.Right.delete(e, min, id, fast)
			}
			if self.Right == nil {
				self.Range.SetMax(self.Elem.Max())
			}
		}
	}

	root = self.fixUp(fast)

	return
}

// Return the left-most interval stored in the tree.
func (self *Tree) Min() Interface {
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

// Return the right-most interval stored in the tree.
func (self *Tree) Max() Interface {
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

// Floor returns the largest value equal to or less than the query q according to
// q.Min().Compare(), with ties broken by q.ID().Compare().
func (self *Tree) Floor(q Interface) (o Interface, err error) {
	if self.Root == nil {
		return
	}
	n := self.Root.floor(q.Min(), q.ID())
	if n == nil {
		return
	}
	return n.Elem, nil
}

func (self *Node) floor(m, id Comparable) *Node {
	if self == nil {
		return nil
	}
	switch c := m.Compare(self.Elem.Min()); {
	case c == 0:
		switch cid := id.Compare(self.Elem.ID()); {
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
func (self *Tree) Ceil(q Interface) (o Interface, err error) {
	if self.Root == nil {
		return
	}
	n := self.Root.ceil(q.Min(), q.ID())
	if n == nil {
		return
	}
	return n.Elem, nil
}

func (self *Node) ceil(m, id Comparable) *Node {
	if self == nil {
		return nil
	}
	switch c := m.Compare(self.Elem.Min()); {
	case c == 0:
		switch cid := id.Compare(self.Elem.ID()); {
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

// An Operation is a function that operates on an Interface. If done is returned true, the
// Operation is indicating that no further work needs to be done and so the Do function should
// traverse no further.
type Operation func(Interface) (done bool)

// Do performs fn on all intervals stored in the tree. A boolean is returned indicating whether the
// Do traversal was interrupted by an Operation returning true. If fn alters stored intervals' sort
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

// DoReverse performs fn on all intervals stored in the tree, but in reverse of sort order. A boolean
// is returned indicating whether the Do traversal was interrupted by an Operation returning true.
// If fn alters stored intervals' sort relationships, future tree operation behaviors are undefined.
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

// DoMatch performs fn on all intervals stored in the tree that match q according to Overlap, with
// q.Overlap() used to guide tree traversal, so DoMatching() will out perform Do() with a called
// conditional function if the condition is based on sort order, but can not be reliably used if
// the condition is independent of sort order. A boolean is returned indicating whether the Do
// traversal was interrupted by an Operation returning true. If fn alters stored intervals' sort
// relationships, future tree operation behaviors are undefined.
func (self *Tree) DoMatching(fn Operation, q Overlapper) (t bool) {
	if self.Root != nil && q.Overlap(self.Root.Range) {
		return self.Root.doMatch(fn, q)
	}
	return
}

func (self *Node) doMatch(fn Operation, q Overlapper) (done bool) {
	if self.Left != nil && q.Overlap(self.Left.Range) {
		done = self.Left.doMatch(fn, q)
		if done {
			return
		}
	}
	if q.Overlap(self.Elem) {
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
// traversal was interrupted by an Operation returning true. If fn alters stored intervals' sort
// relationships, future tree operation behaviors are undefined.
func (self *Tree) DoMatchingReverse(fn Operation, q Overlapper) (t bool) {
	if self.Root != nil && q.Overlap(self.Root.Range) {
		return self.Root.doMatch(fn, q)
	}
	return
}

func (self *Node) doMatchReverse(fn Operation, q Overlapper) (done bool) {
	if self.Right != nil && q.Overlap(self.Right.Range) {
		done = self.Right.doMatchReverse(fn, q)
		if done {
			return
		}
	}
	if q.Overlap(self.Elem) {
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
