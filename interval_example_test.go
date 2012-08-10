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

type Interval struct {
	Start, End, UID Int
	Payload         interface{}
}

func (i Interval) Overlap(b interval.Range) bool {
	var start, end Int
	switch bc := b.(type) {
	case Interval:
		start, end = bc.Start, bc.End
	case *Mutable:
		start, end = bc.Start, bc.End
	default:
		panic("unknown type")
	}

	// Half-open interval indexing.
	return i.End > start && i.Start < end
}
func (i Interval) ID() interval.Comparable      { return i.UID }
func (i Interval) Min() interval.Comparable     { return i.Start }
func (i Interval) Max() interval.Comparable     { return i.End }
func (i Interval) NewMutable() interval.Mutable { return &Mutable{i.Start, i.End} }
func (i Interval) String() string               { return fmt.Sprintf("[%d,%d)#%d", i.Start, i.End, i.UID) }

type Mutable struct{ Start, End Int }

func (m *Mutable) Min() interval.Comparable     { return m.Start }
func (m *Mutable) Max() interval.Comparable     { return m.End }
func (m *Mutable) SetMin(c interval.Comparable) { m.Start = c.(Int) }
func (m *Mutable) SetMax(c interval.Comparable) { m.End = c.(Int) }

func Example() {
	ivs := []Interval{
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

	t := &interval.Tree{}
	for i, iv := range ivs {
		iv.UID = Int(i)
		err := t.Insert(iv)
		if err != nil {
			fmt.Println(err)
		}
	}

	results, err := t.Get(Interval{Start: 3, End: 6})
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(results)

	// Output:
	// [[1,6)#2 [2,4)#1 [3,4)#3 [4,6)#5 [5,8)#6 [5,9)#8]
}
