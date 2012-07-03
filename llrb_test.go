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

package llrb

import (
	"fmt"
	check "launchpad.net/gocheck"
	"math/rand"
	"testing"
)

// Integrity checks - translated from http://www.cs.princeton.edu/~rs/talks/LLRB/Java/RedBlackBST.java

// Is this tree a BST?
func (t *Tree) isBST() bool {
	return (*Node)(t).isBST(t.Min(), t.Max())
}

// Are all the values in the BST rooted at x between min and max,
// and does the same property hold for both subtrees?
func (n *Node) isBST(min, max Comparable) bool {
	if n == nil {
		return true
	}
	if n.Elem.Compare(min) < 0 || n.Elem.Compare(max) > 0 {
		return false
	}
	return n.Left.isBST(min, n.Elem) || n.Right.isBST(n.Elem, max)
}

// Test BU and TD234 invariants.
func (t *Tree) is23_234() bool { return (*Node)(t).is23_234() }
func (n *Node) is23_234() bool {
	if n == nil {
		return true
	}
	if Mode == BU23 {
		// If the node has two children, only one of them may be red.
		// The other must be black...
		if (n.Left != nil) && (n.Right != nil) {
			if n.Left.color() == Red && n.Right.color() == Red {
				return false
			}
		}
		// and the red node should really should be the left one.
		if n.Right.color() == Red {
			return false
		}
	} else if Mode == TD234 {
		// This test is altered from that shown in the java since the trees
		// shown in the paper do not conform to the test as it existed and the
		// current situation does not break the 2-3-4 definition of the LLRB.
		if n.Right.color() == Red && n.Left.color() == Black && n.Left != nil {
			return false
		}
	} else {
		panic("unknown mode")
	}
	if n.color() == Red && n.Left.color() == Red && n.Left.Left.color() == Red {
		return false
	}
	return n.Left.is23_234() && n.Right.is23_234()
}

// Do all paths from root to leaf have same number of black edges?
func (t *Tree) isBalanced() bool {
	var black int // number of black links on path from root to min
	for x := (*Node)(t); x != nil; x = x.Left {
		if x.color() == Black {
			black++
		}
	}
	return (*Node)(t).isBalanced(black)
}

// Does every path from the root to a leaf have the given number 
// of black links?
func (n *Node) isBalanced(black int) bool {
	if n == nil && black == 0 {
		return true
	} else if n == nil && black != 0 {
		return false
	}
	if n.color() == Black {
		black--
	}
	return n.Left.isBalanced(black) && n.Right.isBalanced(black)
}

// Test helpers

type compRune rune

func (cr compRune) Compare(r Comparable) int {
	return int(cr) - int(r.(compRune))
}

// Build a tree from a simplified Newick format returning the root node.
// Single letter node names only, no error checking and all nodes are full or leaf.
func makeTree(desc string) (n *Node) {
	var build func([]rune) (*Node, int)
	build = func(desc []rune) (cn *Node, i int) {
		if len(desc) == 0 || desc[0] == ';' {
			return nil, 0
		}

		var c int
		cn = &Node{}
		for {
			b := desc[i]
			i++
			if b == '(' {
				cn.Left, c = build(desc[i:])
				i += c
				continue
			}
			if b == ',' {
				cn.Right, c = build(desc[i:])
				i += c
				continue
			}
			if b == ')' {
				continue
			}
			cn.Elem = compRune(b)
			return cn, i
		}

		panic("cannot reach")
	}

	n, _ = build([]rune(desc))

	return
}

// Return a Newick format description of a tree defined by a node
func describeTree(n *Node, char, color bool) string {
	s := []rune(nil)

	var follow func(*Node)
	follow = func(n *Node) {
		children := n.Left != nil || n.Right != nil
		if children {
			s = append(s, '(')
		}
		if n.Left != nil {
			follow(n.Left)
		}
		if children {
			s = append(s, ',')
		}
		if n.Right != nil {
			follow(n.Right)
		}
		if children {
			s = append(s, ')')
		}
		if n.Elem != nil {
			if char {
				s = append(s, rune(n.Elem.(compRune)))
			} else {
				s = append(s, []rune(fmt.Sprintf("%d", n.Elem))...)
			}
			if color {
				s = append(s, []rune(fmt.Sprintf(" %v", n.color()))...)
			}
		}
	}
	follow(n)
	s = append(s, ';')

	return string(s)
}

// Tests
func Test(t *testing.T) { check.TestingT(t) }

type S struct{}

var _ = check.Suite(&S{})

func (s *S) TestMakeAndDescribeTree(c *check.C) {
	desc := "((a,c)b,(e,g)f)d;"
	tree := makeTree(desc)
	c.Check(describeTree(tree, true, false), check.DeepEquals, desc)
}

// ((a,c)b,(e,g)f)d -rotL-> (((a,c)b,e)d,g)f
func (s *S) TestRotateLeft(c *check.C) {
	orig := "((a,c)b,(e,g)f)d;"
	rot := "(((a,c)b,e)d,g)f;"

	tree := makeTree(orig)

	tree = tree.rotateLeft()
	c.Check(describeTree(tree, true, false), check.DeepEquals, rot)

	rotTree := makeTree(rot)
	c.Check(tree, check.DeepEquals, rotTree)
}

// ((a,c)b,(e,g)f)d -rotR-> (a,(c,(e,g)f)d)b
func (s *S) TestRotateRight(c *check.C) {
	orig := "((a,c)b,(e,g)f)d;"
	rot := "(a,(c,(e,g)f)d)b;"

	tree := makeTree(orig)

	tree = tree.rotateRight()
	c.Check(describeTree(tree, true, false), check.DeepEquals, rot)

	rotTree := makeTree(rot)
	c.Check(tree, check.DeepEquals, rotTree)
}

func (s *S) TestInsertion(c *check.C) {
	min, max := compRune(0), compRune(1000)
	t := &Tree{}
	for i := min; i <= max; i++ {
		t = t.Insert(i)
		c.Check(t.isBST(), check.Equals, true)
		c.Check(t.is23_234(), check.Equals, true)
		c.Check(t.isBalanced(), check.Equals, true)
	}
	if c.Failed() {
		c.Log(describeTree((*Node)(t), false, true))
	}
	c.Check(t.Min(), check.Equals, compRune(min))
	c.Check(t.Max(), check.Equals, compRune(max))
}

func (s *S) TestDeletion(c *check.C) {
	min, max := compRune(0), compRune(10000)
	t := &Tree{}
	for i := min; i <= max; i++ {
		t = t.Insert(i)
	}
	for i := min; i <= max; i++ {
		t = t.Delete(i)
		if i < max {
			c.Check(t.isBST(), check.Equals, true)
			c.Check(t.is23_234(), check.Equals, true)
			c.Check(t.isBalanced(), check.Equals, true)
			if c.Failed() {
				c.Log(describeTree((*Node)(t), false, true))
			}
		}
	}
	c.Check(t, check.Equals, (*Tree)(nil))
}

func (s *S) TestRandomInsertion(c *check.C) {
	count, max := 100000, 1000
	t := &Tree{}
	for i := 0; i < count; i++ {
		t = t.Insert(compRune(rand.Intn(max)))
		c.Check(t.isBST(), check.Equals, true)
		c.Check(t.is23_234(), check.Equals, true)
		c.Check(t.isBalanced(), check.Equals, true)
	}
	if c.Failed() {
		c.Log(describeTree((*Node)(t), false, true))
	}
}

func (s *S) TestRandomDeletion(c *check.C) {
	count, max := 100000, 1000
	r := make([]compRune, count)
	t := &Tree{}
	for i := range r {
		r[i] = compRune(rand.Intn(max))
		t = t.Insert(r[i])
	}
	for _, v := range r {
		t = t.Delete(v)
		if t != nil {
			c.Check(t.isBST(), check.Equals, true)
			c.Check(t.is23_234(), check.Equals, true)
			c.Check(t.isBalanced(), check.Equals, true)
			if c.Failed() {
				c.Log(describeTree((*Node)(t), false, true))
			}
		}
	}
	c.Check(t, check.Equals, (*Tree)(nil))
}

func (s *S) TestDeleteMinMax(c *check.C) {
	min, max := compRune(0), compRune(10000)
	t := &Tree{}
	for i := min; i <= max; i++ {
		t = t.Insert(i)
	}
	for i, m := 0, int(max); i < m/2; i++ {
		t = t.DeleteMin()
		min++
		c.Check(t.Min(), check.Equals, min)
		t = t.DeleteMax()
		max--
		c.Check(t.Max(), check.Equals, max)
		c.Check(t.isBST(), check.Equals, true)
		c.Check(t.is23_234(), check.Equals, true)
		c.Check(t.isBalanced(), check.Equals, true)
		if c.Failed() {
			c.Log(describeTree((*Node)(t), false, true))
		}
	}
}

func (s *S) TestRandomInsertionDeletion(c *check.C) {
	count, max := 100000, 1000
	t := &Tree{}
	for i := 0; i < count; i++ {
		if rand.Float64() < 0.5 {
			t = t.Insert(compRune(rand.Intn(max)))
		}
		if rand.Float64() < 0.5 {
			t = t.Delete(compRune(rand.Intn(max)))
		}
		c.Check(t.isBST(), check.Equals, true)
		c.Check(t.is23_234(), check.Equals, true)
		c.Check(t.isBalanced(), check.Equals, true)
	}
	if c.Failed() {
		c.Log(describeTree((*Node)(t), false, true))
	}
}
