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
	"fmt"
	"math/rand"
	"sort"
)

// Partition partitions list such that all elements less than the value at pivot prior to the
// call are placed before that element and all elements greater than that value are placed after it.
// The final location of the element at pivot prior to the call is returned.
func Partition(list sort.Interface, pivot int) int {
	var index, last int
	if last = list.Len() - 1; last < 0 {
		return -1
	}
	list.Swap(pivot, last)
	for i := 0; i < last; i++ {
		if !list.Less(last, i) {
			list.Swap(index, i)
			index++
		}
	}
	list.Swap(last, index)
	return index
}

// A SortSlicer satisfies the sort.Interface and is able to slice itself.
type SortSlicer interface {
	sort.Interface
	Slice(start, end int) SortSlicer
}

// Select partitions list such that all elements less than the kth largest element are
// placed placed before k in the resulting list and all elements greater than it are placed
// after the position k.
func Select(list SortSlicer, k int) int {
	var (
		start int
		end   = list.Len()
	)
	if k >= end {
		if k == 0 {
			return 0
		}
		panic(fmt.Sprintf("kdtree: index out of range"))
	}
	if start == end-1 {
		return k
	}

	for {
		if start == end {
			panic("kdtree: internal inconsistency")
		}
		sub := list.Slice(start, end)
		pivot := Partition(sub, rand.Intn(sub.Len()))
		switch {
		case pivot == k:
			return k
		case k < pivot:
			end = pivot + start
		default:
			k -= pivot
			start += pivot
		}
	}

	panic("cannot reach")
}

func min(a, b int) int {
	if a > b {
		return b
	}
	return a
}

// MedianOfMedians returns the index to the median value of the medians of groups of 5 consecutive elements.
func MedianOfMedians(list SortSlicer) int {
	n := list.Len() / 5
	for i := 0; i < n; i++ {
		left := i * 5
		sub := list.Slice(left, min(left+5, list.Len()-1))
		Select(sub, 2)
		list.Swap(i, left+2)
	}
	Select(list.Slice(0, min(n, list.Len()-1)), min(list.Len(), n/2))
	return n / 2
}

// MedianOfRandoms returns the index to the median value of up to n randomly chosen elements in list.
func MedianOfRandoms(list SortSlicer, n int) int {
	if l := list.Len(); n <= l {
		for i := 0; i < n; i++ {
			list.Swap(i, rand.Intn(n))
		}
	} else {
		n = l
	}
	Select(list.Slice(0, n), n/2)
	return n / 2
}
