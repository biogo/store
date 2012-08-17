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
	Sub            []Interval
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

var ivs = []Interval{
	{start: 0, end: 2},
	{start: 2, end: 4},
	{start: 1, end: 6},
	{start: 3, end: 4},
	{start: 1, end: 3},
	{start: 4, end: 6},
	{start: 5, end: 8},
	{start: 6, end: 8},
	{start: 5, end: 7},
	{start: 8, end: 9},
}

func Example_1() {
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

	// Output:
	// Generic interval tree:
	// [[1,6)#2 [2,4)#1 [3,4)#3 [4,6)#5 [5,8)#6 [5,7)#8]
}
