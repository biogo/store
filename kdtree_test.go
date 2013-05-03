// Copyright ©2012 The bíogo.kdtree Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package kdtree

import (
	"container/heap"
	"flag"
	"fmt"
	check "launchpad.net/gocheck"
	"math/rand"
	"os"
	"reflect"
	"sort"
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
	wpData   = Points{{2, 3}, {5, 4}, {9, 6}, {4, 7}, {8, 1}, {7, 2}}
	nbWpData = nbPoints{{2, 3}, {5, 4}, {9, 6}, {4, 7}, {8, 1}, {7, 2}}
	wpBound  = &Bounding{Point{2, 1}, Point{9, 7}}
	bData    = func(i int) Points {
		p := make(Points, i)
		for i := range p {
			p[i] = Point{rand.Float64(), rand.Float64(), rand.Float64()}
		}
		return p
	}(1e2)
	bTree = New(bData, true)
)

func (s *S) TestNew(c *check.C) {
	for i, test := range []struct {
		data     Interface
		bounding bool
		bounds   *Bounding
	}{
		{wpData, false, nil},
		{nbWpData, false, nil},
		{wpData, true, wpBound},
		{nbWpData, true, nil},
	} {
		var t *Tree
		NewTreePanics := func() (panicked bool) {
			defer func() {
				if r := recover(); r != nil {
					panicked = true
				}
			}()
			t = New(test.data, test.bounding)
			return
		}
		c.Check(NewTreePanics(), check.Equals, false)
		c.Check(t.Root.isKDTree(), check.Equals, true)
		switch data := test.data.(type) {
		case Points:
			for _, p := range data {
				c.Check(t.Contains(p), check.Equals, true)
			}
		case nbPoints:
			for _, p := range data {
				c.Check(t.Contains(p), check.Equals, true)
			}
		}
		c.Check(t.Root.Bounding, check.DeepEquals, test.bounds,
			check.Commentf("Test %d. %T %v", i, test.data, test.bounding))
		if c.Failed() && *genDot && t.Len() <= *dotLimit {
			err := dotFile(t, fmt.Sprintf("TestNew%T", test.data), "")
			if err != nil {
				c.Errorf("Dot file write failed: %v", err)
			}
		}
	}
}

func (s *S) TestInsert(c *check.C) {
	for i, test := range []struct {
		data   Interface
		insert []Comparable
		bounds *Bounding
	}{
		{
			wpData,
			[]Comparable{Point{0, 0}, Point{10, 10}},
			&Bounding{Point{0, 0}, Point{10, 10}},
		},
		{
			nbWpData,
			[]Comparable{nbPoint{0, 0}, nbPoint{10, 10}},
			nil,
		},
	} {
		t := New(test.data, true)
		for _, v := range test.insert {
			t.Insert(v, true)
		}
		c.Check(t.Root.isKDTree(), check.Equals, true)
		c.Check(t.Root.Bounding, check.DeepEquals, test.bounds,
			check.Commentf("Test %d. %T", i, test.data))
		if c.Failed() && *genDot && t.Len() <= *dotLimit {
			err := dotFile(t, fmt.Sprintf("TestInsert%T", test.data), "")
			if err != nil {
				c.Errorf("Dot file write failed: %v", err)
			}
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
	d := n.Point.Dims()
	// Together these define the property of minimal orthogonal bounding.
	if !(n.isContainedBy(n.Bounding) && n.Bounding.planesHaveCoincidentPointsIn(n, [2][]bool{make([]bool, d), make([]bool, d)})) {
		return false
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

func (n *Node) isContainedBy(b *Bounding) bool {
	if n == nil {
		return true
	}
	if !b.Contains(n.Point) {
		return false
	}
	return n.Left.isContainedBy(b) && n.Right.isContainedBy(b)
}

func (b *Bounding) planesHaveCoincidentPointsIn(n *Node, tight [2][]bool) bool {
	if b == nil {
		return true
	}
	if n == nil {
		return true
	}

	b.planesHaveCoincidentPointsIn(n.Left, tight)
	b.planesHaveCoincidentPointsIn(n.Right, tight)

	var ok = true
	for i := range tight {
		for d := 0; d < n.Point.Dims(); d++ {
			if c := n.Point.Compare(b[0], Dim(d)); c == 0 {
				tight[i][d] = true
			}
			ok = ok && tight[i][d]
		}
	}
	return ok
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

type pDist struct {
	Point
	dist float64
}

type pDists []pDist

func newPDists(n int) pDists {
	pd := make(pDists, 1, n)
	pd[0].dist = inf
	return pd
}

func (pd *pDists) Keep(p pDist) {
	if p.dist < (*pd)[0].dist {
		if len(*pd) == cap(*pd) {
			heap.Pop(pd)
		}
		heap.Push(pd, p)
	}
}
func (pd pDists) Len() int              { return len(pd) }
func (pd pDists) Less(i, j int) bool    { return pd[i].dist > pd[j].dist }
func (pd pDists) Swap(i, j int)         { pd[i], pd[j] = pd[j], pd[i] }
func (pd *pDists) Push(x interface{})   { (*pd) = append(*pd, x.(pDist)) }
func (pd *pDists) Pop() (i interface{}) { i, *pd = (*pd)[len(*pd)-1], (*pd)[:len(*pd)-1]; return i }

func nearestN(n int, q Point, p Points) ([]Comparable, []float64) {
	pd := newPDists(n)
	for i := 0; i < p.Len(); i++ {
		pd.Keep(pDist{Point: p[i], dist: q.Distance(p[i])})
	}
	if len(pd) == 1 {
		return []Comparable{pd[0].Point}, []float64{pd[0].dist}
	}
	if pd[0].dist == inf {
		pd = pd[1:]
	}
	sort.Sort(pd)
	for i, j := 0, len(pd)-1; i < j; i, j = i+1, j-1 {
		pd[i], pd[j] = pd[j], pd[i]
	}
	ns := make([]Comparable, len(pd))
	d := make([]float64, len(pd))
	for i, n := range pd {
		ns[i] = n.Point
		d[i] = n.dist
	}
	return ns, d
}

func (s *S) TestNearestN(c *check.C) {
	t := New(wpData, false)
	in := append([]Point{
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
		{-1e5, 0}}, wpData[:len(wpData)-1]...) // The point (9,6) is excluded as it has two pairs of equidistant points.
	for k := 1; k <= len(wpData); k++ {
		for i, q := range in {
			p, d := t.NearestN(k, q)
			ep, ed := nearestN(k, q, wpData)
			c.Check(p, check.DeepEquals, ep, check.Commentf("Test k=%d %d: query %.3f expects %.3f", k, i, q, ep))
			c.Check(d, check.DeepEquals, ed)
		}
	}
}

func (s *S) TestNearestSet(c *check.C) {
	t := New(wpData, false)
	in := append([]Point{
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
		{-1e5, 0}}, wpData...)
	for k := 1; k <= len(wpData); k++ {
		for i, q := range in {
			ep, ed := t.NearestN(k, q)
			keep := newNDists(k)
			t.NearestSet(&keep, q)
			p := make(map[float64][]Comparable)
			d := make([]float64, len(keep))
			for i, n := range keep {
				p[n.Dist] = append(p[n.Dist], n.Point)
				d[i] = n.Dist
			}
			c.Check(d, check.DeepEquals, ed, check.Commentf("Test k=%d %d: query %.3f expects %.3f", k, i, q, ed))
			// Sort order is not the same between the two methods, so we do this...
			for j, d := range ed {
				// Find a point value pv in p[d] that matches the current point in ed.
				var ok bool
				for _, pv := range p[d] {
					if reflect.DeepEqual(pv, ep[j]) {
						ok = true
						break
					}
				}
				c.Check(ok, check.Equals, true)
			}
		}
	}
}

func (s *S) TestDo(c *check.C) {
	var result Points
	t := New(wpData, false)
	f := func(c Comparable, _ *Bounding, _ int) (done bool) {
		result = append(result, c.(Point))
		return
	}
	killed := t.Do(f)
	c.Check(result, check.DeepEquals, wpData)
	c.Check(killed, check.Equals, false)
}

func (s *S) TestDoBounded(c *check.C) {
	for _, test := range []struct {
		bounds *Bounding
		result Points
	}{
		{
			nil,
			wpData,
		},
		{
			&Bounding{Point{0, 0}, Point{10, 10}},
			wpData,
		},
		{
			&Bounding{Point{3, 4}, Point{10, 10}},
			Points{Point{5, 4}, Point{4, 7}, Point{9, 6}},
		},
		{
			&Bounding{Point{3, 3}, Point{10, 10}},
			Points{Point{5, 4}, Point{4, 7}, Point{9, 6}},
		},
		{
			&Bounding{Point{0, 0}, Point{6, 5}},
			Points{Point{2, 3}, Point{5, 4}},
		},
		{
			&Bounding{Point{5, 2}, Point{7, 4}},
			Points{Point{5, 4}, Point{7, 2}},
		},
		{
			&Bounding{Point{2, 2}, Point{7, 4}},
			Points{Point{2, 3}, Point{5, 4}, Point{7, 2}},
		},
		{
			&Bounding{Point{2, 3}, Point{9, 6}},
			Points{Point{2, 3}, Point{5, 4}, Point{9, 6}},
		},
		{
			&Bounding{Point{7, 2}, Point{7, 2}},
			Points{Point{7, 2}},
		},
	} {
		var result Points
		t := New(wpData, false)
		f := func(c Comparable, _ *Bounding, _ int) (done bool) {
			result = append(result, c.(Point))
			return
		}
		killed := t.DoBounded(f, test.bounds)
		c.Check(result, check.DeepEquals, test.result)
		c.Check(killed, check.Equals, false)
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

func BenchmarkNearestN10(b *testing.B) {
	var (
		r []Comparable
		d []float64
	)
	for i := 0; i < b.N; i++ {
		r, d = bTree.NearestN(10, Point{rand.Float64(), rand.Float64(), rand.Float64()})
	}
	_, _ = r, d
}

func BenchmarkNearestSetN10(b *testing.B) {
	var keep = newNDists(10)
	for i := 0; i < b.N; i++ {
		bTree.NearestSet(&keep, Point{rand.Float64(), rand.Float64(), rand.Float64()})
		keep = keep[:1]
		keep[0] = NodeDist{nil, inf}
	}
}

func BenchmarkNearBruteN10(b *testing.B) {
	var (
		r []Comparable
		d []float64
	)
	for i := 0; i < b.N; i++ {
		r, d = nearestN(10, Point{rand.Float64(), rand.Float64(), rand.Float64()}, bData)
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
