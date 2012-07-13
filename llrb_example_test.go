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

package llrb_test

import (
	"code.google.com/p/biogo.llrb"
	"fmt"
)

type (
	CompInt     int
	CompIntNoEq int
)

func (c CompInt) Compare(b llrb.Comparable) int {
	switch i := b.(type) {
	case CompInt:
		return int(c) - int(i)
	case CompIntNoEq:
		return int(c) - int(i)
	}
	panic("unknown type")
}

func (c CompIntNoEq) Compare(b llrb.Comparable) int {
	var d int
	switch i := b.(type) {
	case CompInt:
		d = int(c) - int(i)
	case CompIntNoEq:
		d = int(c) - int(i)
	}
	if d == 0 {
		return 1
	}
	return d
}

func Example() {
	values := []int{0, 1, 2, 3, 4, 2, 3, 5, 5, 65, 32, 3, 23}

	// Insert using a type that reports equality:
	{
		t := &llrb.Tree{}
		for _, v := range values {
			t.Insert(CompInt(v)) // Insert with replacement.
		}

		results := []int(nil)
		// More efficiently retrieved using Get(Compint(3))...
		t.DoMatching(func(c llrb.Comparable) (done bool) {
			results = append(results, int(c.(CompInt)))
			return
		}, CompInt(3))

		fmt.Println("With replacement:   ", results)
	}

	// Insert using a type that does not report equality:
	{
		t := &llrb.Tree{}
		for _, v := range values {
			t.Insert(CompIntNoEq(v)) // Insert without replacement.
		}

		results := []int(nil)
		t.DoMatching(func(c llrb.Comparable) (done bool) {
			results = append(results, int(c.(CompIntNoEq)))
			return
		}, CompInt(3))

		fmt.Println("Without replacement:", results)
	}

	// Output:
	// With replacement:    [3]
	// Without replacement: [3 3 3]
}

/*func Example_2() {
	values := []int{0, 1, 2, 3, 4, 2, 3, 5, 5, 65, 32, 3, 23}
	t := &Tree{}
	for _, v := range values {
		t.Insert(CompInt(v)) // Insert with replacement.
	}

	results := []int(nil)
	t.DoMatching(func(c Comparable) (done bool) {
		results = append(results, int(c.(CompInt)))
		return
	}, CompInt(3))

	fmt.Println(results)
	// Output:
	// [3]
}
*/
