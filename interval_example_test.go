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

package interval_test

import (
	"code.google.com/p/biogo.interval"
	"fmt"
)

// Generic intervals
type Int int

func (c Int) Compare(b interval.Comparable) int {
	return int(c - b.(Int))
}

type Interval struct {
	start, end, id Int
	Payload        interface{}
}

func (i Interval) Overlap(b interval.Range) bool {
	var start, end Int
	switch bc := b.(type) {
	case Interval:
		start, end = bc.start, bc.end
	case *Mutable:
		start, end = bc.start, bc.end
	default:
		panic("unknown type")
	}

	// Half-open interval indexing.
	return i.end > start && i.start < end
}
func (i Interval) ID() interval.Comparable      { return i.id }
func (i Interval) Start() interval.Comparable   { return i.start }
func (i Interval) End() interval.Comparable     { return i.end }
func (i Interval) NewMutable() interval.Mutable { return &Mutable{i.start, i.end} }
func (i Interval) String() string               { return fmt.Sprintf("[%d,%d)#%d", i.start, i.end, i.id) }

type Mutable struct{ start, end Int }

func (m *Mutable) Start() interval.Comparable     { return m.start }
func (m *Mutable) End() interval.Comparable       { return m.end }
func (m *Mutable) SetStart(c interval.Comparable) { m.start = c.(Int) }
func (m *Mutable) SetEnd(c interval.Comparable)   { m.end = c.(Int) }

// Integer-specific intervals
type IntInterval struct {
	Start, End int
	UID        uintptr
	Payload    interface{}
}

func (i IntInterval) Overlap(b interval.IntRange) bool {
	// Half-open interval indexing.
	return i.End > b.Start && i.Start < b.End
}
func (i IntInterval) ID() uintptr              { return i.UID }
func (i IntInterval) Range() interval.IntRange { return interval.IntRange{i.Start, i.End} }
func (i IntInterval) String() string           { return fmt.Sprintf("[%d,%d)#%d", i.Start, i.End, i.UID) }

func Example() {
	// Generic intervals
	{
		ivs := []Interval{
			{start: 0, end: 2},
			{start: 2, end: 4},
			{start: 1, end: 6},
			{start: 3, end: 4},
			{start: 1, end: 3},
			{start: 4, end: 6},
			{start: 5, end: 8},
			{start: 6, end: 8},
			{start: 5, end: 9},
		}

		t := &interval.Tree{}
		for i, iv := range ivs {
			iv.id = Int(i)
			err := t.Insert(iv, false)
			if err != nil {
				fmt.Println(err)
			}
		}

		fmt.Println("Generic interval tree:")
		fmt.Println(t.Get(Interval{start: 3, end: 6}))
	}

	// Integer-specific intervals
	{
		ivs := []IntInterval{
			{Start: 0, End: 2},
			{Start: 2, End: 4},
			{Start: 1, End: 6},
			{Start: 3, End: 4},
			{Start: 1, End: 3},
			{Start: 4, End: 6},
			{Start: 5, End: 8},
			{Start: 6, End: 8},
			{Start: 5, End: 9},
		}

		t := &interval.IntTree{}
		for i, iv := range ivs {
			iv.UID = uintptr(i)
			err := t.Insert(iv, false)
			if err != nil {
				fmt.Println(err)
			}
		}

		fmt.Println("Integer-specific interval tree:")
		fmt.Println(t.Get(IntInterval{Start: 3, End: 6}))
	}

	// Output:
	// Generic interval tree:
	// [[1,6)#2 [2,4)#1 [3,4)#3 [4,6)#5 [5,8)#6 [5,9)#8]
	// Integer-specific interval tree:
	// [[1,6)#2 [2,4)#1 [3,4)#3 [4,6)#5 [5,8)#6 [5,9)#8]
}

func min(a, b interval.Comparable) interval.Comparable {
	if a.Compare(b) < 0 {
		return a
	}
	return b
}

func max(a, b interval.Comparable) interval.Comparable {
	if a.Compare(b) > 0 {
		return a
	}
	return b
}

func ExampleTree_Do() {
	// Flatten all overlapping intervals, storing originals as sub-intervals.

	// Given...
	type Interval struct {
		start, end interval.Comparable
		sub        []*Interval
		interval.Interface
	}
	t := &interval.Tree{}

	var (
		fi = true
		ti []*Interval
	)

	t.Do(
		func(e interval.Interface) (done bool) {
			iv := e.(*Interval)
			if fi || iv.start.Compare(ti[len(ti)-1].end) > 0 {
				ti = append(ti, &Interval{
					start: iv.start,
					end:   iv.end,
				})
				fi = false
			} else {
				ti[len(ti)-1].end = max(ti[len(ti)-1].end, iv.end)
			}
			ti[len(ti)-1].sub = append(ti[len(ti)-1].sub, iv)

			return
		},
	)
	t.Root, t.Count = nil, 0
	for _, iv := range ti {
		t.Insert(iv, true)
	}
	t.AdjustRanges()
}

func ExampleTree_DoMatching() {
	// Merge an interval into the tree, replacing overlapping intervals, but retaining them as sub intervals.

	// Given...
	type Interval struct {
		start, end interval.Comparable
		sub        []*Interval
		interval.Interface
	}
	t := &interval.Tree{}
	ni := &Interval{}

	var (
		fi = true
		qi = &Interval{start: ni.start, end: ni.end}
		r  []interval.Interface
	)

	t.DoMatching(
		func(e interval.Interface) (done bool) {
			iv := e.(*Interval)
			r = append(r, e)
			ni.sub = append(ni.sub, iv)

			// Flatten merge history.
			ni.sub = append(ni.sub, iv.sub...)
			iv.sub = nil

			if fi {
				ni.start = min(iv.start, ni.start)
				fi = false
			}
			ni.end = max(iv.end, ni.end)

			return
		},
		qi,
	)
	for _, d := range r {
		t.Delete(d, false)
	}
	t.Insert(ni, false)
}
