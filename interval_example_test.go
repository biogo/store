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

type Int int

func (c Int) Compare(b interval.Comparable) int {
	return int(c - b.(Int))
}

type IntOverlap struct{ Start, End Int }

func (o *IntOverlap) Overlap(b interval.Overlapper) int {
	var start, end Int
	switch bc := b.(type) {
	case *IntOverlap:
		start, end = bc.Start, bc.End
	case *IntRange:
		start, end = bc.Start, bc.End
	default:
		panic("unknown type")
	}

	// Half-open interval indexing.
	if o.End > start && o.Start < end {
		return 0
	}

	if o.End <= start {
		return -1
	}
	if o.Start >= end {
		return 1
	}
	panic("cannot reach")
}
func (o *IntOverlap) Min() interval.Comparable  { return o.Start }
func (o *IntOverlap) Max() interval.Comparable  { return o.End }
func (o *IntOverlap) Mutable() interval.Mutable { return &IntRange{*o} }
func (o *IntOverlap) String() string            { return fmt.Sprintf("[%d,%d)", o.Start, o.End) }

type IntRange struct{ IntOverlap }

func (r *IntRange) SetMin(m interval.Comparable) { r.IntOverlap.Start = m.(Int) }
func (r *IntRange) SetMax(m interval.Comparable) { r.IntOverlap.End = m.(Int) }

func Example() {
	ivs := []*IntOverlap{
		{0, 2},
		{2, 4},
		{1, 6},
		{3, 4},
		{1, 3},
		{4, 6},
		{5, 8},
		{6, 8},
		{5, 9},
	}

	t := &interval.Tree{}
	for _, iv := range ivs {
		err := t.Insert(iv)
		if err != nil {
			fmt.Println(err)
		}
	}

	results, err := t.Get(&IntOverlap{3, 6})
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(results)

	// Output:
	// [[1,6) [2,4) [3,4) [4,6) [5,8) [5,9)]
}
