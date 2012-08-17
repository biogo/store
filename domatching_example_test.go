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

// Merge an interval into the tree, replacing overlapping intervals, but retaining them as sub intervals.
func Merge(t *interval.Tree, ni Interval) {
	var (
		fi = true
		qi = &Interval{start: ni.start, end: ni.end}
		r  []interval.Interface
	)

	t.DoMatching(
		func(e interval.Interface) (done bool) {
			iv := e.(Interval)
			r = append(r, e)
			ni.Sub = append(ni.Sub, iv)

			// Flatten merge history.
			ni.Sub = append(ni.Sub, iv.Sub...)
			iv.Sub = nil

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

func ExampleTree_DoMatching() {
	t := &interval.Tree{}

	var (
		i  int
		iv Interval
	)

	for i, iv = range ivs {
		iv.id = uintptr(i)
		err := t.Insert(iv, false)
		if err != nil {
			fmt.Println(err)
		}
	}
	i++

	Merge(t, Interval{start: -1, end: 4, id: uintptr(i)})
	t.Do(func(e interval.Interface) (done bool) {
		fmt.Printf("%s: %v\n", e, e.(Interval).Sub)
		return
	})

	// Output:
	// [-1,6)#10: [[0,2)#0 [1,6)#2 [1,3)#4 [2,4)#1 [3,4)#3]
	// [4,6)#5: []
	// [5,8)#6: []
	// [5,7)#8: []
	// [6,8)#7: []
	// [8,9)#9: []
}
