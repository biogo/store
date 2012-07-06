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
	"flag"
	"fmt"
	check "launchpad.net/gocheck"
	"math/rand"
	"os"
	"strings"
	"testing"
	"unsafe"
)

const (
	none = iota
	first
	all
	printTrees = first
)

var genDot = flag.Bool("dot", false, "Generate dot code for TestDeleteRight trees.")

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
				if cn.Left == nil && cn.Right == nil {
					return nil, i
				}
				continue
			}
			if b != ';' {
				cn.Elem = compRune(b)
			}
			return cn, i
		}

		panic("cannot reach")
	}

	n, _ = build([]rune(desc))
	if n.Left == nil && n.Right == nil {
		n = nil
	}

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
	if n == nil {
		s = []rune("()")
	} else {
		follow(n)
	}
	s = append(s, ';')

	return string(s)
}

// Tests
func Test(t *testing.T) { check.TestingT(t) }

type S struct{}

var _ = check.Suite(&S{})

func (s *S) TestMakeAndDescribeTree(c *check.C) {
	c.Check(describeTree((*Node)(nil), true, false), check.DeepEquals, "();")
	for _, desc := range []string{
		"();",
		"((a,c)b,(e,g)f)d;",
	} {
		t := makeTree(desc)
		c.Check(describeTree(t, true, false), check.DeepEquals, desc)
	}
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

func filterDiff(t *Tree, _ int) *Tree { return t }
func (s *S) TestNilOperations(c *check.C) {
	var e *Tree
	for _, t := range []*Tree{nil, {}} {
		c.Check(t.Min(), check.Equals, nil)
		c.Check(t.Max(), check.Equals, nil)
		c.Check(filterDiff(t.DeleteMin()), check.Equals, e)
		c.Check(filterDiff(t.DeleteMax()), check.Equals, e)
	}
}

func (s *S) TestInsertion(c *check.C) {
	var (
		printed  = false
		min, max = compRune(0), compRune(1000)
		d        int
	)
	for _, t := range []*Tree{nil, {}} {
		for i := min; i <= max; i++ {
			t, d = t.Insert(i)
			c.Check(d, check.Equals, 1)
			failed := false
			failed = failed || !c.Check(t.isBST(), check.Equals, true)
			failed = failed || !c.Check(t.is23_234(), check.Equals, true)
			failed = failed || !c.Check(t.isBalanced(), check.Equals, true)
			if failed && (printTrees > none && !printed) || printTrees == all {
				printed = true
				c.Logf("Failing tree: %s\n\n", describeTree((*Node)(t), false, true))
			}
		}
		c.Check(t.Min(), check.Equals, compRune(min))
		c.Check(t.Max(), check.Equals, compRune(max))
	}
}

func (s *S) TestDeletion(c *check.C) {
	var (
		printed  = false
		min, max = compRune(0), compRune(10000)
		d, e     int
	)
	for _, t := range []*Tree{nil, {}} {
		for i := min; i <= max; i++ {
			t, _ = t.Insert(i)
		}
		for i := min; i <= max; i++ {
			if t.Get(i) != nil {
				e = -1
			} else {
				e = 0
			}
			t, d = t.Delete(i)
			c.Check(d, check.Equals, e)
			if i < max {
				failed := false
				failed = failed || !c.Check(t.isBST(), check.Equals, true)
				failed = failed || !c.Check(t.is23_234(), check.Equals, true)
				failed = failed || !c.Check(t.isBalanced(), check.Equals, true)
				if failed && (printTrees > none && !printed) || printTrees == all {
					printed = true
					c.Logf("Failing tree: %s\n\n", describeTree((*Node)(t), false, true))
				}
			}
		}
		c.Check(t, check.Equals, (*Tree)(nil))
	}
}

func (s *S) TestGet(c *check.C) {
	min, max := compRune(0), compRune(100000)
	for _, t := range []*Tree{nil, {}} {
		for i := min; i <= max; i++ {
			if i&1 == 0 {
				t, _ = t.Insert(i)
			}
		}
		for i := min; i <= max; i++ {
			if i&1 == 0 {
				c.Check(t.Get(i), check.DeepEquals, compRune(i)) // Check inserted elements are present.
			} else {
				c.Check(t.Get(i), check.Equals, Comparable(nil)) // Check inserted elements are absent.
			}
		}
	}
}

func (s *S) TestRandomlyInsertedGet(c *check.C) {
	count, max := 100000, 1000
	for _, t := range []*Tree{nil, {}} {
		verify := map[rune]struct{}{}
		for i := 0; i < count; i++ {
			v := compRune(rand.Intn(max))
			t, _ = t.Insert(v)
			verify[rune(v)] = struct{}{}
		}
		// Random fetch order - check only those inserted.
		for v := range verify {
			c.Check(t.Get(compRune(v)), check.DeepEquals, compRune(v)) // Check inserted elements are present.
		}
		// Check all possible insertions.
		for i := compRune(0); i <= compRune(max); i++ {
			if _, ok := verify[rune(i)]; ok {
				c.Check(t.Get(i), check.DeepEquals, compRune(i)) // Check inserted elements are present.
			} else {
				c.Check(t.Get(i), check.Equals, Comparable(nil)) // Check inserted elements are absent.
			}
		}
	}
}

func (s *S) TestRandomInsertion(c *check.C) {
	var (
		printed    bool
		count, max = 100000, 1000
		t          *Tree
	)
	for i := 0; i < count; i++ {
		t, _ = t.Insert(compRune(rand.Intn(max)))
		failed := false
		failed = failed || !c.Check(t.isBST(), check.Equals, true)
		failed = failed || !c.Check(t.is23_234(), check.Equals, true)
		failed = failed || !c.Check(t.isBalanced(), check.Equals, true)
		if failed && (printTrees > none && !printed) || printTrees == all {
			printed = true
			c.Logf("Failing tree: %s\n\n", describeTree((*Node)(t), false, true))
		}
	}
}

func (s *S) TestRandomDeletion(c *check.C) {
	var (
		printed    bool
		count, max = 100000, 1000
		r          = make([]compRune, count)
		t          *Tree
	)
	for i := range r {
		r[i] = compRune(rand.Intn(max))
		t, _ = t.Insert(r[i])
	}
	for _, v := range r {
		t, _ = t.Delete(v)
		if t != nil {
			failed := false
			failed = failed || !c.Check(t.isBST(), check.Equals, true)
			failed = failed || !c.Check(t.is23_234(), check.Equals, true)
			failed = failed || !c.Check(t.isBalanced(), check.Equals, true)
			if failed && (printTrees > none && !printed) || printTrees == all {
				printed = true
				c.Logf("Failing tree: %s\n\n", describeTree((*Node)(t), false, true))
			}
		}
	}
	c.Check(t, check.Equals, (*Tree)(nil))
}

func (s *S) TestDeleteMinMax(c *check.C) {
	var (
		printed  bool
		min, max = compRune(0), compRune(10)
		t        *Tree
		d, dI    int
	)
	for i := min; i <= max; i++ {
		t, d = t.Insert(i)
		dI += d
	}
	c.Check(dI, check.Equals, int(max-min+1))
	for i, m := 0, int(max); i < m/2; i++ {
		failed := false
		t, d = t.DeleteMin()
		c.Check(d, check.Equals, -1)
		min++
		failed = failed || !c.Check(t.Min(), check.Equals, min)
		t, d = t.DeleteMax()
		c.Check(d, check.Equals, -1)
		max--
		failed = failed || !c.Check(t.Max(), check.Equals, max)
		failed = failed || !c.Check(t.isBST(), check.Equals, true)
		failed = failed || !c.Check(t.is23_234(), check.Equals, true)
		failed = failed || !c.Check(t.isBalanced(), check.Equals, true)
		if failed && (printTrees > none && !printed) || printTrees == all {
			printed = true
			c.Logf("Failing tree: %s\n\n", describeTree((*Node)(t), false, true))
		}
	}
}

func (s *S) TestRandomInsertionDeletion(c *check.C) {
	var (
		printed    bool
		count, max = 100000, 1000
		t          *Tree
		dI, dD     int
		verify     = map[int]struct{}{}
	)
	for i := 0; i < count; i++ {
		var d, e int
		if rand.Float64() < 0.5 {
			r := rand.Intn(max)
			if _, ok := verify[r]; ok {
				e = 0
			} else {
				e = 1
			}
			t, d = t.Insert(compRune(r))
			verify[r] = struct{}{}
			c.Check(d, check.Equals, e)
			dI += d
		}
		if rand.Float64() < 0.5 {
			r := rand.Intn(max)
			if _, ok := verify[r]; ok {
				e = -1
			} else {
				e = 0
			}
			t, d = t.Delete(compRune(r))
			delete(verify, r)
			dD += d
		}
		failed := false
		failed = failed || !c.Check(t.isBST(), check.Equals, true)
		failed = failed || !c.Check(t.is23_234(), check.Equals, true)
		failed = failed || !c.Check(t.isBalanced(), check.Equals, true)
		if failed && (printTrees > none && !printed) || printTrees == all {
			printed = true
			c.Logf("Failing tree: %s\n\n", describeTree((*Node)(t), false, true))
		}
	}
	c.Check(dI+dD, check.Equals, len(verify), check.Commentf("Insertions: %d Deletions: %d", dI, -dD))
}

var (
	modeName = []string{TD234: "TD234", BU23: "BU23"}
	arrows   = map[Color]string{Red: "none", Black: "normal"}
)

func (s *S) TestDeleteRight(c *check.C) {
	type target struct {
		min, max, target compRune
	}
	var d int
	for _, r := range []target{
		{0, 14, 14},
		{0, 15, 15},
		{0, 16, 15},
		{0, 16, 16},
		{0, 17, 16},
		{0, 17, 17},
	} {
		var (
			t      *Tree
			format string
		)
		for i := r.min; i <= r.max; i++ {
			t, _ = t.Insert(i)
		}
		before := describeTree((*Node)(t), false, true)
		format = "Before deletion: %#v %s"
		checkTree(t, c, format, r, before)
		if *genDot {
			err := dot(t, fmt.Sprintf("%s_before_del_%d_%d_%d", modeName[Mode], r.min, r.max, r.target))
			if err != nil {
				c.Errorf("Dot file write failed: %v", err)
			}
		}
		t, d = t.Delete(r.target)
		c.Check(d, check.Equals, -1)
		format = "%#v\nBefore deletion: %s\nAfter deletion:  %s"
		checkTree(t, c, format, r, before, describeTree((*Node)(t), false, true))
		if *genDot {
			err := dot(t, fmt.Sprintf("%s_after_del_%d_%d_%d", modeName[Mode], r.min, r.max, r.target))
			if err != nil {
				c.Errorf("Dot file write failed: %v", err)
			}
		}
	}
}

func checkTree(t *Tree, c *check.C, f string, i ...interface{}) {
	comm := check.Commentf(f, i...)
	c.Check(t.isBST(), check.Equals, true, comm)
	c.Check(t.is23_234(), check.Equals, true, comm)
	c.Check(t.isBalanced(), check.Equals, true, comm)
}

func dot(t *Tree, label string) (err error) {
	if t == nil {
		return
	}
	var (
		s      []string
		follow func(*Node)
	)
	follow = func(n *Node) {
		if n == nil {
			return
		}
		id := uintptr(unsafe.Pointer(n))
		c := fmt.Sprintf("%d[label = \"<Left> |<Elem> %d|<Right>\"];", id, n.Elem)
		if n.Left != nil {
			c += fmt.Sprintf("\n\t\tedge [color=%v,arrowhead=%s]; \"%d\":Left -> \"%d\":Elem;",
				n.Left.color(), arrows[n.Left.color()], id, uintptr(unsafe.Pointer(n.Left)))
			follow(n.Left)
		}
		if n.Right != nil {
			c += fmt.Sprintf("\n\t\tedge [color=%v,arrowhead=%s]; \"%d\":Right -> \"%d\":Elem;",
				n.Right.color(), arrows[n.Right.color()], id, uintptr(unsafe.Pointer(n.Right)))
			follow(n.Right)
		}
		s = append(s, c)
	}
	follow((*Node)(t))
	f, err := os.Create(label + ".dot")
	if err != nil {
		return
	}
	defer f.Close()
	fmt.Fprintf(f, "digraph %s {\n\tnode [shape=record,height=0.1];\n\t%s\n}\n",
		label,
		strings.Join(s, "\n\t"),
	)
	return
}

// Benchmarks

type compInt int

func (ci compInt) Compare(i Comparable) int {
	return int(ci) - int(i.(compInt))
}

type compIntNoRep int

func (ci compIntNoRep) Compare(i Comparable) int {
	c := int(ci) - int(i.(compIntNoRep))
	if c == 0 {
		return 1
	}
	return c
}

func BenchmarkInsert(b *testing.B) {
	var t *Tree
	for i := 0; i < b.N; i++ {
		t, _ = t.Insert(compInt(b.N - i))
	}
}

func BenchmarkInsertNoRep(b *testing.B) {
	var t *Tree
	for i := 0; i < b.N; i++ {
		t, _ = t.Insert(compIntNoRep(b.N - i))
	}
}

func BenchmarkGet(b *testing.B) {
	b.StopTimer()
	var t *Tree
	for i := 0; i < b.N; i++ {
		t, _ = t.Insert(compInt(b.N - i))
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		t.Get(compInt(i))
	}
}

func BenchmarkDelete(b *testing.B) {
	b.StopTimer()
	var t *Tree
	for i := 0; i < b.N; i++ {
		t, _ = t.Insert(compInt(b.N - i))
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		t, _ = t.Delete(compInt(i))
	}
}

func BenchmarkDeleteMin(b *testing.B) {
	b.StopTimer()
	var t *Tree
	for i := 0; i < b.N; i++ {
		t, _ = t.Insert(compInt(b.N - i))
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		t, _ = t.DeleteMin()
	}
}
