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

var intIvs = []IntInterval{
	{Start: 0, End: 2},
	{Start: 2, End: 4},
	{Start: 1, End: 6},
	{Start: 3, End: 4},
	{Start: 1, End: 3},
	{Start: 4, End: 6},
	{Start: 5, End: 8},
	{Start: 6, End: 8},
	{Start: 5, End: 7},
	{Start: 8, End: 9},
}

func Example_2() {
	t := &interval.IntTree{}
	for i, iv := range intIvs {
		iv.UID = uintptr(i)
		err := t.Insert(iv, false)
		if err != nil {
			fmt.Println(err)
		}
	}

	fmt.Println("Integer-specific interval tree:")
	fmt.Println(t.Get(IntInterval{Start: 3, End: 6}))

	// Output:
	// Integer-specific interval tree:
	// [[1,6)#2 [2,4)#1 [3,4)#3 [4,6)#5 [5,8)#6 [5,7)#8]
}
