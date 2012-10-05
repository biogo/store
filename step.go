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

// Package step implements a step vector type.
//
// A step vector can be used to represent high volume data that would be
// efficiently stored by run-length encoding.
package step

import (
	"code.google.com/p/biogo.llrb"
	"errors"
	"fmt"
)

var (
	ErrOutOfRange    = errors.New("step: index out of range")
	ErrInvertedRange = errors.New("step: inverted range")
	ErrZeroLength    = errors.New("step: attempt to create zero length vector")
)

type (
	position struct {
		pos int
		val Equaler
	}
	query int
	upper int
)

func (p *position) Compare(c llrb.Comparable) int {
	return p.pos - c.(*position).pos
}
func (q query) Compare(c llrb.Comparable) (d int) {
	switch c := c.(type) {
	case *position:
		d = int(q) - c.pos
	case query:
		d = int(q) - int(c)
	}
	return
}
func (q upper) Compare(c llrb.Comparable) (d int) {
	d = int(q) - c.(*position).pos
	if d == 0 {
		d = 1
	}
	return
}

// An Equaler is a type that can return whether it equals another Equaler.
type Equaler interface {
	Equal(Equaler) bool
}

// An Int is an int type satisfying the Equaler interface.
type Int int

// Equal returns whether i equals e. Equal assumes the underlying type of e is Int.
func (i Int) Equal(e Equaler) bool {
	return i == e.(Int)
}

// A Float is a float64 type satisfying the Equaler interface.
type Float float64

// Equal returns whether f equals e. For the purposes of the step package here, NaN == NaN
// evaluates to true. Equal assumes the underlying type of e is Float.
func (f Float) Equal(e Equaler) bool {
	ef := e.(Float)
	if f != f && ef != ef { // For our purposes NaN == NaN.
		return true
	}
	return f == ef
}

// A Vector is type that support the storage of array type data in a run-length
// encoding format.
type Vector struct {
	Zero     Equaler // Ground state for the step vector.
	Relaxed  bool    // If true, dynamic vector resize is allowed.
	t        *llrb.Tree
	min, max *position
}

// New returns a new Vector with the extent defined by start and end,
// and the ground state defined by zero. The Vector's extent is mutable
// if the Relaxed field is set to true. If a zero length vector is requested
// an error is returned.
func New(start, end int, zero Equaler) (*Vector, error) {
	if start >= end {
		return nil, ErrZeroLength
	}
	v := &Vector{
		Zero: zero,
		t:    &llrb.Tree{},
		min: &position{
			pos: start,
			val: zero,
		},
		max: &position{
			pos: end,
			val: nil,
		},
	}
	v.t.Insert(v.min)
	v.t.Insert(v.max)

	return v, nil
}

// Start returns the index of minimum position of the Vector.
func (v *Vector) Start() int { return v.min.pos }

// End returns the index of lowest position beyond the end of the Vector.
func (v *Vector) End() int { return v.max.pos }

// Len returns the length of the represented data array, that is the distance
// between the start and end of the vector.
func (v *Vector) Len() int { return v.End() - v.Start() }

// Count returns the number of steps represented in the vector.
func (v *Vector) Count() int { return v.t.Len() - 1 }

// At returns the value of the vector at position i. If i is outside the extent
// of the vector an error is returned.
func (v *Vector) At(i int) (e Equaler, err error) {
	if i < v.Start() || i >= v.End() {
		return nil, ErrOutOfRange
	}
	st := v.t.Floor(query(i)).(*position)
	return st.val, nil
}

// StepAt returns the value and range of the step at i, where start <= i < end.
// If i is outside the extent of the vector, an error is returned.
func (v *Vector) StepAt(i int) (start, end int, e Equaler, err error) {
	if i < v.Start() || i >= v.End() {
		return 0, 0, nil, ErrOutOfRange
	}
	lo := v.t.Floor(query(i)).(*position)
	hi := v.t.Ceil(upper(i)).(*position)
	return lo.pos, hi.pos, lo.val, nil
}

// Set sets the value of position i to e.
func (v *Vector) Set(i int, e Equaler) {
	if i < v.min.pos || v.max.pos <= i {
		if !v.Relaxed {
			panic(ErrOutOfRange)
		}

		if i < v.min.pos {
			if i == v.min.pos-1 {
				if e.Equal(v.min.val) {
					v.min.pos--
				} else {
					v.min = &position{pos: i, val: e}
					v.t.Insert(v.min)
				}
			} else {
				if v.min.val.Equal(v.Zero) {
					v.min.pos = i + 1
				} else {
					v.min = &position{pos: i + 1, val: v.Zero}
					v.t.Insert(v.min)
				}
				if e.Equal(v.Zero) {
					v.min.pos--
				} else {
					v.min = &position{pos: i, val: e}
					v.t.Insert(v.min)
				}
			}
		} else if i >= v.max.pos {
			if i == v.max.pos {
				v.max.pos++
				prev := v.t.Floor(query(i)).(*position)
				if !e.Equal(prev.val) {
					v.t.Insert(&position{pos: i, val: e})
				}
			} else {
				mpos := v.max.pos
				v.max.pos = i + 1
				prev := v.t.Floor(query(i)).(*position)
				if !prev.val.Equal(v.Zero) {
					v.t.Insert(&position{pos: mpos, val: v.Zero})
				}
				if !e.Equal(v.Zero) {
					v.t.Insert(&position{pos: i, val: e})
				}
			}
		}
		return
	}

	lo := v.t.Floor(query(i)).(*position)
	if e.Equal(lo.val) {
		return
	}
	hi := v.t.Ceil(upper(i)).(*position)

	if lo.pos == i {
		if hi.pos == i+1 {
			if hi != v.max && e.Equal(hi.val) {
				v.t.Delete(query(i))
				hi.pos--
			} else {
				lo.val = e
			}
			if i > v.min.pos {
				prev := v.t.Floor(query(i - 1)).(*position)
				if e.Equal(prev.val) {
					v.t.Delete(query(i))
				}
			}
		} else {
			lo.pos = i + 1
			prev := v.t.Floor(query(i))
			if prev == nil {
				v.min = &position{pos: i, val: e}
				v.t.Insert(v.min)
			} else if !e.Equal(prev.(*position).val) {
				v.t.Insert(&position{pos: i, val: e})
			}
		}
	} else {
		if hi.pos == i+1 {
			if hi != v.max && e.Equal(hi.val) {
				hi.pos--
			} else {
				v.t.Insert(&position{pos: i, val: e})
			}
		} else {
			v.t.Insert(&position{pos: i, val: e})
			v.t.Insert(&position{pos: i + 1, val: lo.val})
		}
	}
}

// SetRange sets the value of positions [start, end) to e.
// The underlying type of e must be comparable by reflect.DeepEqual.
func (v *Vector) SetRange(start, end int, e Equaler) {
	l := end - start
	switch {
	case l == 0:
		return
	case l == 1:
		v.Set(start, e)
		return
	case l < 0:
		panic(ErrInvertedRange)
	}

	if end <= v.min.pos || v.max.pos <= start {
		if !v.Relaxed {
			panic(ErrOutOfRange)
		}

		if end <= v.min.pos {
			if end == v.min.pos {
				if e.Equal(v.min.val) {
					v.min.pos -= l
				} else {
					v.min = &position{pos: start, val: e}
					v.t.Insert(v.min)
				}
			} else {
				if v.min.val.Equal(v.Zero) {
					v.min.pos = end
				} else {
					v.min = &position{pos: end, val: v.Zero}
					v.t.Insert(v.min)
				}
				if e.Equal(v.Zero) {
					v.min.pos -= l
				} else {
					v.min = &position{pos: start, val: e}
					v.t.Insert(v.min)
				}
			}
		} else if start >= v.max.pos {
			if start == v.max.pos {
				v.max.pos += l
				prev := v.t.Floor(query(start)).(*position)
				if !e.Equal(prev.val) {
					v.t.Insert(&position{pos: start, val: e})
				}
			} else {
				mpos := v.max.pos
				v.max.pos = end
				prev := v.t.Floor(query(start)).(*position)
				if !prev.val.Equal(v.Zero) {
					v.t.Insert(&position{pos: mpos, val: v.Zero})
				}
				if !e.Equal(v.Zero) {
					v.t.Insert(&position{pos: start, val: e})
				}
			}
		}
		return
	}

	delQ := []llrb.Comparable{}
	v.t.DoRange(func(c llrb.Comparable) (done bool) {
		delQ = append(delQ, c)
		return
	}, query(start), query(end))
	for _, p := range delQ {
		v.t.Delete(p)
	}

	var la, lo *position
	if len(delQ) > 0 {
		lo = delQ[0].(*position)
		la = delQ[len(delQ)-1].(*position)
	} else {
		lo = v.t.Floor(query(start)).(*position)
		la = &position{}
		*la = *lo
	}

	hi := v.t.Ceil(query(end)).(*position)
	if start == lo.pos {
		var prevSame bool
		prev := v.t.Floor(query(start - 1))
		if prev != nil {
			prevSame = e.Equal(prev.(*position).val)
		}
		hiSame := hi != v.max && e.Equal(hi.val)
		if hi.pos == end {
			switch {
			case hiSame && prevSame:
				v.t.Delete(hi)
			case prevSame:
				return
			case hiSame:
				hi.pos = start
			default:
				if prev == nil {
					v.min = &position{pos: start, val: e}
					v.t.Insert(v.min)
				} else {
					v.t.Insert(&position{pos: start, val: e})
				}
			}
		} else {
			la.pos = end
			if !e.Equal(la.val) {
				v.t.Insert(la)
			}
			if prev == nil {
				v.min = &position{pos: start, val: e}
				v.t.Insert(v.min)
			} else if !prevSame {
				v.t.Insert(&position{pos: start, val: e})
			}
		}
	} else {
		if hi.pos == end {
			if hi != v.max && e.Equal(hi.val) {
				hi.pos = start
			} else {
				v.t.Insert(&position{pos: start, val: e})
			}
		} else {
			v.t.Insert(&position{pos: start, val: e})
			la.pos = end
			if !e.Equal(la.val) {
				v.t.Insert(la)
			}
		}
	}
}

// An Operation is a non-mutating function that can be applied to a vector using Do
// and DoRange.
type Operation func(start, end int, e Equaler)

// Do performs the function fn on steps stored in the Vector in ascending sort order
// of start position. fn is passed the start, end and value of the step.
func (v *Vector) Do(fn Operation) {
	var (
		la  *position
		min = v.min.pos
	)

	v.t.Do(func(c llrb.Comparable) (done bool) {
		p := c.(*position)
		if p.pos != min {
			fn(la.pos, p.pos, la.val)
		}
		la = p
		return
	})
}

// Do performs the function fn on steps stored in the Vector over the range [from, to)
// in ascending sort order of start position. fn is passed the start, end and value of
// the step.
func (v *Vector) DoRange(from, to int, fn Operation) (err error) {
	if to < from {
		return ErrInvertedRange
	}
	var (
		la  *position
		min = v.min.pos
		max = v.max.pos
	)
	if to <= min || from >= max {
		return ErrOutOfRange
	}

	_, end, e, _ := v.StepAt(from)
	fn(from, end, e)
	v.t.DoRange(func(c llrb.Comparable) (done bool) {
		p := c.(*position)
		if p.pos != end {
			fn(la.pos, p.pos, la.val)
		}
		la = p
		return
	}, query(end), query(to))
	if to > la.pos {
		fn(la.pos, to, la.val)
	}

	return
}

// A Mutator is a function that is used by Apply and ApplyRange to alter values within
// a Vector.
type Mutator func(Equaler) Equaler

// Convenience mutator functions. Mutator functions are used by Apply and ApplyRange
// to alter step values in a value-dependent manner. These mutators assume the stored
// type matches the function and will panic is this is not true.
var (
	IncInt   Mutator = incInt   // Increment an int value.
	DecInt   Mutator = decInt   // Decrement an int value.
	IncFloat Mutator = incFloat // Increment a float64 value.
	DecFloat Mutator = decFloat // Decrement a float64 value.
)

func incInt(e Equaler) Equaler   { return e.(Int) + 1 }
func decInt(e Equaler) Equaler   { return e.(Int) - 1 }
func incFloat(e Equaler) Equaler { return e.(Float) + 1 }
func decFloat(e Equaler) Equaler { return e.(Float) - 1 }

// Apply applies the mutator function m to steps stored in the Vector in ascending sort order
// of start position. Redundant steps resulting from changes in step values are erased.
func (v *Vector) Apply(m Mutator) {
	var (
		la   Equaler
		min  = v.min.pos
		max  = v.max.pos
		delQ []query
	)

	v.t.Do(func(c llrb.Comparable) (done bool) {
		p := c.(*position)
		if p.pos == max {
			return true
		}
		p.val = m(p.val)
		if p.pos != min && p.pos != max && p.val.Equal(la) {
			delQ = append(delQ, query(p.pos))
		}
		la = p.val
		return
	})

	for _, d := range delQ {
		v.t.Delete(d)
	}
}

// Apply applies the mutator function m to steps stored in the Vector in over the range
// [from, to) in ascending sort order of start position. Redundant steps resulting from
// changes in step values are erased.
func (v *Vector) ApplyRange(from, to int, m Mutator) (err error) {
	if to < from {
		return ErrInvertedRange
	}
	var (
		la   Equaler
		old  position
		min  = v.min.pos
		max  = v.max.pos
		delQ []query
	)
	if to <= min || from >= max {
		return ErrOutOfRange
	}

	var end int
	old.pos, end, old.val, _ = v.StepAt(from)
	la = old.val
	la = m(la)
	if to <= end {
		v.SetRange(from, to, la)
		return
	}
	if !la.Equal(old.val) {
		v.t.Insert(&position{from, la})
	}
	v.t.DoRange(func(c llrb.Comparable) (done bool) {
		p := c.(*position)
		if p.pos == max {
			return true
		}
		old = *p // Needed for fix-up of last step if to is not at a step boundary.
		p.val = m(p.val)
		if p.pos != min && p.val.Equal(la) {
			delQ = append(delQ, query(p.pos))
		}
		la = p.val
		return
	}, query(end), query(to))

	if to < max {
		p := v.t.Ceil(query(to)).(*position)
		if p.pos > to && (p == v.max || !p.val.Equal(old.val)) {
			v.t.Insert(&position{pos: to, val: old.val})
		} else if p.val.Equal(la) {
			delQ = append(delQ, query(p.pos))
		}
	}

	for _, d := range delQ {
		v.t.Delete(d)
	}

	return
}

// String returns a string representation a Vector, displaying step start
// positions and values. The last step indicates the end of the vector and
// always has an associated value of nil.
func (v *Vector) String() string {
	sb := []string(nil)
	v.t.Do(func(c llrb.Comparable) (done bool) {
		p := c.(*position)
		sb = append(sb, fmt.Sprintf("%d:%v", p.pos, p.val))
		return
	})
	return fmt.Sprintf("%v", sb)
}
