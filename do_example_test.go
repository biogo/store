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

func min(a, b Int) Int {
	if a < b {
		return a
	}
	return b
}

func max(a, b Int) Int {
	if a > b {
		return a
	}
	return b
}

// Flatten all overlapping intervals, storing originals as sub-intervals.
func Flatten(t *interval.Tree) {
	var (
		fi  = true
		ti  []Interval
		mid Int
	)

	t.Do(
		func(e interval.Interface) (done bool) {
			iv := e.(Interval)
			if fi || iv.start >= ti[len(ti)-1].end {
				ti = append(ti, Interval{
					start: iv.start,
					end:   iv.end,
				})
				fi = false
			} else {
				ti[len(ti)-1].end = max(ti[len(ti)-1].end, iv.end)
			}
			ti[len(ti)-1].Sub = append(ti[len(ti)-1].Sub, iv)
			if iv.id > mid {
				mid = iv.id
			}

			return
		},
	)

	mid++
	t.Root, t.Count = nil, 0
	for i, iv := range ti {
		iv.id = Int(i) + mid
		t.Insert(iv, true)
	}
	t.AdjustRanges()
}

func ExampleTree_Do() {
	t := &interval.Tree{}
	for i, iv := range ivs {
		iv.id = Int(i)
		err := t.Insert(iv, false)
		if err != nil {
			fmt.Println(err)
		}
	}

	Flatten(t)
	t.Do(func(e interval.Interface) (done bool) { fmt.Printf("%s: %v\n", e, e.(Interval).Sub); return })

	// Output:
	// [0,8)#10: [[0,2)#0 [1,6)#2 [1,3)#4 [2,4)#1 [3,4)#3 [4,6)#5 [5,8)#6 [5,7)#8 [6,8)#7]
	// [8,9)#11: [[8,9)#9]
}
