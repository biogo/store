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

package kdtree

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

var (
	genDot   = flag.Bool("dot", false, "Generate dot code for failing trees.")
	dotLimit = flag.Int("dotmax", 100, "Maximum size for tree output for dot format.")
)

func Test(t *testing.T) { check.TestingT(t) }

type S struct{}

var _ = check.Suite(&S{})

var (
	// Using example from WP article.
	wpData  = Points{{2, 3}, {5, 4}, {9, 6}, {4, 7}, {8, 1}, {7, 2}}
	wpBound = &Bounding{Point{2, 1}, Point{9, 7}}
	bData   = func(i int) Points {
		p := make(Points, i)
		for i := range p {
			p[i] = Point{rand.Float64(), rand.Float64(), rand.Float64()}
		}
		return p
	}(1e2)
	bTree = New(bData, true)
)

func (s *S) TestNew(c *check.C) {
	var t *Tree
	NewTreePanics := func() (panicked bool) {
		defer func() {
			if r := recover(); r != nil {
				panicked = true
			}
		}()
		t = New(wpData, true)
		return
	}
	c.Check(NewTreePanics(), check.Equals, false)
	c.Check(t.Root.isKDTree(), check.Equals, true)
	for _, p := range wpData {
		c.Check(t.Contains(p), check.Equals, true)
	}
	c.Check(t.Root.Bounding, check.DeepEquals, wpBound)
	if c.Failed() && *genDot && t.Len() <= *dotLimit {
		err := dotFile(t, "TestNew", "")
		if err != nil {
			c.Errorf("Dot file write failed: %v", err)
		}
	}
}

func (s *S) TestInsert(c *check.C) {
	t := New(wpData, true)
	t.Insert(Point{0, 0}, true)
	t.Insert(Point{10, 10}, true)
	c.Check(t.Root.isKDTree(), check.Equals, true)
	c.Check(t.Root.Bounding, check.DeepEquals, &Bounding{Point{0, 0}, Point{10, 10}})
	if c.Failed() && *genDot && t.Len() <= *dotLimit {
		err := dotFile(t, "TestInsert", "")
		if err != nil {
			c.Errorf("Dot file write failed: %v", err)
		}
	}
}

type compFn func(float64) bool

func left(v float64) bool  { return v <= 0 }
func right(v float64) bool { return !left(v) }

func (n *Node) isKDTree() bool {
	if n == nil {
		return true
	}
	if !n.Left.isPartitioned(n.Point, left, n.Plane) {
		return false
	}
	if !n.Right.isPartitioned(n.Point, right, n.Plane) {
		return false
	}
	return n.Left.isKDTree() && n.Right.isKDTree()
}

func (n *Node) isPartitioned(pivot Comparable, fn compFn, plane Dim) bool {
	if n == nil {
		return true
	}
	if n.Left != nil && fn(pivot.Compare(n.Left.Point, plane)) {
		return false
	}
	if n.Right != nil && fn(pivot.Compare(n.Right.Point, plane)) {
		return false
	}
	return n.Left.isPartitioned(pivot, fn, plane) && n.Right.isPartitioned(pivot, fn, plane)
}

func nearest(q Point, p Points) (Point, float64) {
	min := q.Distance(p[0])
	var r int
	for i := 1; i < p.Len(); i++ {
		d := q.Distance(p[i])
		if d < min {
			min = d
			r = i
		}
	}
	return p[r], min
}

func (s *S) TestNearest(c *check.C) {
	t := New(wpData, false)
	for i, q := range append([]Point{
		{4, 6},
		{7, 5},
		{8, 7},
		{6, -5},
		{1e5, 1e5},
		{1e5, -1e5},
		{-1e5, 1e5},
		{-1e5, -1e5},
		{1e5, 0},
		{0, -1e5},
		{0, 1e5},
		{-1e5, 0},
	}, wpData...) {
		p, d := t.Nearest(q)
		ep, ed := nearest(q, wpData)
		c.Check(p, check.DeepEquals, ep, check.Commentf("Test %d: query %.3f expects %.3f", i, q, ep))
		c.Check(d, check.Equals, ed)
	}
}

func BenchmarkNew(b *testing.B) {
	b.StopTimer()
	p := make(Points, 1e5)
	for i := range p {
		p[i] = Point{rand.Float64(), rand.Float64(), rand.Float64()}
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_ = New(p, false)
	}
}

func BenchmarkNewBounds(b *testing.B) {
	b.StopTimer()
	p := make(Points, 1e5)
	for i := range p {
		p[i] = Point{rand.Float64(), rand.Float64(), rand.Float64()}
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_ = New(p, true)
	}
}

func BenchmarkInsert(b *testing.B) {
	rand.Seed(1)
	t := &Tree{}
	for i := 0; i < b.N; i++ {
		t.Insert(Point{rand.Float64(), rand.Float64(), rand.Float64()}, false)
	}
}

func BenchmarkInsertBounds(b *testing.B) {
	rand.Seed(1)
	t := &Tree{}
	for i := 0; i < b.N; i++ {
		t.Insert(Point{rand.Float64(), rand.Float64(), rand.Float64()}, true)
	}
}

func (s *S) TestBenches(c *check.C) {
	c.Check(bTree.Root.isKDTree(), check.Equals, true)
	for i := 0; i < 1e3; i++ {
		q := Point{rand.Float64(), rand.Float64(), rand.Float64()}
		p, d := bTree.Nearest(q)
		ep, ed := nearest(q, bData)
		c.Check(p, check.DeepEquals, ep, check.Commentf("Test %d: query %.3f expects %.3f", i, q, ep))
		c.Check(d, check.Equals, ed)
	}
	if c.Failed() && *genDot && bTree.Len() <= *dotLimit {
		err := dotFile(bTree, "TestBenches", "")
		if err != nil {
			c.Errorf("Dot file write failed: %v", err)
		}
	}
}

func BenchmarkNearest(b *testing.B) {
	var (
		r Comparable
		d float64
	)
	for i := 0; i < b.N; i++ {
		r, d = bTree.Nearest(Point{rand.Float64(), rand.Float64(), rand.Float64()})
	}
	_, _ = r, d
}

func BenchmarkNearBrute(b *testing.B) {
	var (
		r Comparable
		d float64
	)
	for i := 0; i < b.N; i++ {
		r, d = nearest(Point{rand.Float64(), rand.Float64(), rand.Float64()}, bData)
	}
	_, _ = r, d
}

func dot(t *Tree, label string) string {
	if t == nil {
		return ""
	}
	var (
		s      []string
		follow func(*Node)
	)
	follow = func(n *Node) {
		id := uintptr(unsafe.Pointer(n))
		c := fmt.Sprintf("%d[label = \"<Left> |<Elem> %s/%.3f\\n%.3f|<Right>\"];",
			id, n, n.Point.(Point)[n.Plane], *n.Bounding)
		if n.Left != nil {
			c += fmt.Sprintf("\n\t\tedge [arrowhead=normal]; \"%d\":Left -> \"%d\":Elem;",
				id, uintptr(unsafe.Pointer(n.Left)))
			follow(n.Left)
		}
		if n.Right != nil {
			c += fmt.Sprintf("\n\t\tedge [arrowhead=normal]; \"%d\":Right -> \"%d\":Elem;",
				id, uintptr(unsafe.Pointer(n.Right)))
			follow(n.Right)
		}
		s = append(s, c)
	}
	if t.Root != nil {
		follow(t.Root)
	}
	return fmt.Sprintf("digraph %s {\n\tnode [shape=record,height=0.1];\n\t%s\n}\n",
		label,
		strings.Join(s, "\n\t"),
	)
}

func dotFile(t *Tree, label, dotString string) (err error) {
	if t == nil && dotString == "" {
		return
	}
	f, err := os.Create(label + ".dot")
	if err != nil {
		return
	}
	defer f.Close()
	if dotString == "" {
		fmt.Fprintf(f, dot(t, label))
	} else {
		fmt.Fprintf(f, dotString)
	}
	return
}
