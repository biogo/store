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
	check "launchpad.net/gocheck"
	"math/rand"
	"testing"
)

func Test(t *testing.T) { check.TestingT(t) }

type S struct{}

var _ = check.Suite(&S{})

var (
	// Using example from WP article.
	wpData = Points{{2, 3}, {5, 4}, {9, 6}, {4, 7}, {8, 1}, {7, 2}}
	bData  = func(i int) Points {
		p := make(Points, i)
		for i := range p {
			p[i] = Point{rand.Float64(), rand.Float64(), rand.Float64()}
		}
		return p
	}(1e2)
	bTree = New(bData)
)

func (s *S) TestNew(c *check.C) {
	New(wpData)
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
	t := New(wpData)
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
		_ = New(p)
	}
}

func (s *S) TestBenches(c *check.C) {
	for i := 0; i < 20; i++ {
		q := Point{rand.Float64(), rand.Float64(), rand.Float64()}
		p, d := bTree.Nearest(q)
		ep, ed := nearest(q, bData)
		c.Check(p, check.DeepEquals, ep, check.Commentf("Test %d: query %.3f expects %.3f", i, q, ep))
		c.Check(d, check.Equals, ed)
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
