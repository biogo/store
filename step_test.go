// Copyright ©2012 The bíogo.step Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package step

import (
	"fmt"
	check "launchpad.net/gocheck"
	"math"
	"math/rand"
	"reflect"
	"testing"
)

// Tests
func Test(t *testing.T) { check.TestingT(t) }

type S struct{}

var _ = check.Suite(&S{})

type nilable int

func (n *nilable) Equal(e Equaler) bool {
	return n == e.(*nilable)
}

func (s *S) TestCreate(c *check.C) {
	_, err := New(0, 0, nil)
	c.Check(err, check.ErrorMatches, ErrZeroLength.Error())
	for _, vec := range []struct {
		start, end int
		zero       Equaler
	}{
		{1, 10, (*nilable)(nil)},
		{0, 10, (*nilable)(nil)},
		{-1, 100, (*nilable)(nil)},
		{-100, -10, (*nilable)(nil)},
		{1, 10, Int(0)},
		{0, 10, Int(0)},
		{-1, 100, Int(0)},
		{-100, -10, Int(0)},
	} {
		sv, err := New(vec.start, vec.end, vec.zero)
		c.Assert(err, check.Equals, nil)
		c.Check(sv.Start(), check.Equals, vec.start)
		c.Check(sv.End(), check.Equals, vec.end)
		c.Check(sv.Len(), check.Equals, vec.end-vec.start)
		c.Check(sv.Zero, check.DeepEquals, vec.zero)
		var at Equaler
		for i := vec.start; i < vec.end; i++ {
			at, err = sv.At(i)
			c.Check(at, check.DeepEquals, vec.zero)
			c.Check(err, check.Equals, nil)
		}
		_, err = sv.At(vec.start - 1)
		c.Check(err, check.ErrorMatches, ErrOutOfRange.Error())
		_, err = sv.At(vec.start - 1)
		c.Check(err, check.ErrorMatches, ErrOutOfRange.Error())
	}
}

func (s *S) TestSet_1(c *check.C) {
	for i, t := range []struct {
		start, end int
		zero       Equaler
		sets       []position
		expect     string
	}{
		{1, 10, Int(0),
			[]position{
				{1, Int(2)},
				{2, Int(3)},
				{3, Int(3)},
				{4, Int(3)},
				{5, Int(2)},
			},
			"[1:2 2:3 5:2 6:0 10:<nil>]",
		},
		{1, 10, Int(0),
			[]position{
				{3, Int(3)},
				{4, Int(3)},
				{1, Int(2)},
				{2, Int(3)},
				{5, Int(2)},
			},
			"[1:2 2:3 5:2 6:0 10:<nil>]",
		},
		{1, 10, Int(0),
			[]position{
				{3, Int(3)},
				{4, Int(3)},
				{5, Int(2)},
				{1, Int(2)},
				{2, Int(3)},
				{9, Int(2)},
			},
			"[1:2 2:3 5:2 6:0 9:2 10:<nil>]",
		},
		{1, 10, Float(0),
			[]position{
				{3, Float(math.NaN())},
				{4, Float(math.NaN())},
				{5, Float(2)},
				{1, Float(2)},
				{2, Float(math.NaN())},
				{9, Float(2)},
			},
			"[1:2 2:NaN 5:2 6:0 9:2 10:<nil>]",
		},
		{1, 10, Float(math.NaN()),
			[]position{
				{3, Float(3)},
				{4, Float(3)},
				{5, Float(2)},
				{1, Float(2)},
				{2, Float(3)},
				{9, Float(2)},
			},
			"[1:2 2:3 5:2 6:NaN 9:2 10:<nil>]",
		},
	} {
		sv, err := New(t.start, t.end, t.zero)
		c.Assert(err, check.Equals, nil)
		c.Check(func() { sv.Set(t.start-1, nil) }, check.Panics, ErrOutOfRange)
		c.Check(func() { sv.Set(t.end, nil) }, check.Panics, ErrOutOfRange)
		for _, v := range t.sets {
			sv.Set(v.pos, v.val)
			c.Check(sv.min.pos, check.Equals, t.start)
			c.Check(sv.max.pos, check.Equals, t.end)
			c.Check(sv.Len(), check.Equals, t.end-t.start)
		}
		c.Check(sv.String(), check.Equals, t.expect, check.Commentf("subtest %d", i))
		sv.Relaxed = true
		sv.Set(t.start-1, sv.Zero)
		sv.Set(t.end, sv.Zero)
		c.Check(sv.Len(), check.Equals, t.end-t.start+2)
		for _, v := range t.sets {
			sv.Set(v.pos, t.zero)
		}
		sv.Set(t.start-1, t.zero)
		sv.Set(t.end, t.zero)
		c.Check(sv.t.Len(), check.Equals, 2)
		c.Check(sv.String(), check.Equals, fmt.Sprintf("[%d:%v %d:%v]", t.start-1, t.zero, t.end+1, nil))
	}
}

func (s *S) TestSet_2(c *check.C) {
	for i, t := range []struct {
		start, end int
		zero       Int
		sets       []position
		expect     string
		count      int
	}{
		{1, 2, 0,
			[]position{
				{1, Int(2)},
				{2, Int(3)},
				{3, Int(3)},
				{4, Int(3)},
				{5, Int(2)},
				{-1, Int(5)},
				{10, Int(23)},
			},
			"[-1:5 0:0 1:2 2:3 5:2 6:0 10:23 11:<nil>]",
			7,
		},
		{1, 10, 0,
			[]position{
				{0, Int(0)},
			},
			"[0:0 10:<nil>]",
			1,
		},
		{1, 10, 0,
			[]position{
				{-1, Int(0)},
			},
			"[-1:0 10:<nil>]",
			1,
		},
		{1, 10, 0,
			[]position{
				{11, Int(0)},
			},
			"[1:0 12:<nil>]",
			1,
		},
		{1, 10, 0,
			[]position{
				{2, Int(1)},
				{3, Int(1)},
				{4, Int(1)},
				{5, Int(1)},
				{6, Int(1)},
				{7, Int(1)},
				{8, Int(1)},
				{5, Int(1)},
			},
			"[1:0 2:1 9:0 10:<nil>]",
			3,
		},
		{1, 10, 0,
			[]position{
				{3, Int(1)},
				{2, Int(1)},
			},
			"[1:0 2:1 4:0 10:<nil>]",
			3,
		},
	} {
		sv, err := New(t.start, t.end, t.zero)
		c.Assert(err, check.Equals, nil)
		sv.Relaxed = true
		for _, v := range t.sets {
			sv.Set(v.pos, v.val)
		}
		c.Check(sv.String(), check.Equals, t.expect, check.Commentf("subtest %d", i))
		c.Check(sv.Count(), check.Equals, t.count)
	}
}

func (s *S) TestSetRange_1(c *check.C) {
	type posRange struct {
		start, end int
		val        Int
	}
	for i, t := range []struct {
		start, end int
		zero       Int
		sets       []posRange
		expect     string
		count      int
	}{
		{1, 10, 0,
			[]posRange{
				{1, 2, 2},
				{2, 3, 3},
				{3, 4, 3},
				{4, 5, 3},
				{5, 6, 2},
			},
			"[1:2 2:3 5:2 6:0 10:<nil>]",
			4,
		},
		{1, 10, 0,
			[]posRange{
				{3, 4, 3},
				{4, 5, 3},
				{1, 2, 2},
				{2, 3, 3},
				{5, 6, 2},
			},
			"[1:2 2:3 5:2 6:0 10:<nil>]",
			4,
		},
		{1, 10, 0,
			[]posRange{
				{3, 4, 3},
				{4, 5, 3},
				{5, 6, 2},
				{1, 2, 2},
				{2, 3, 3},
				{9, 10, 2},
			},
			"[1:2 2:3 5:2 6:0 9:2 10:<nil>]",
			5,
		},
		{1, 10, 0,
			[]posRange{
				{3, 6, 3},
				{4, 5, 1},
				{5, 7, 2},
				{1, 3, 2},
				{9, 10, 2},
			},
			"[1:2 3:3 4:1 5:2 7:0 9:2 10:<nil>]",
			6,
		},
		{1, 10, 0,
			[]posRange{
				{1, 3, 3},
				{4, 5, 1},
				{7, 8, 2},
				{9, 10, 4},
			},
			"[1:3 3:0 4:1 5:0 7:2 8:0 9:4 10:<nil>]",
			7,
		},
	} {
		sv, err := New(t.start, t.end, t.zero)
		c.Assert(err, check.Equals, nil)
		c.Check(func() { sv.SetRange(t.start-2, t.start, nil) }, check.Panics, ErrOutOfRange)
		c.Check(func() { sv.SetRange(t.end, t.end+2, nil) }, check.Panics, ErrOutOfRange)
		for _, v := range t.sets {
			sv.SetRange(v.start, v.end, v.val)
			c.Check(sv.min.pos, check.Equals, t.start)
			c.Check(sv.max.pos, check.Equals, t.end)
			c.Check(sv.Len(), check.Equals, t.end-t.start)
		}
		c.Check(sv.String(), check.Equals, t.expect, check.Commentf("subtest %d", i))
		c.Check(sv.Count(), check.Equals, t.count)
		sv.Relaxed = true
		sv.SetRange(t.start-1, t.start, sv.Zero)
		sv.SetRange(t.end, t.end+1, sv.Zero)
		c.Check(sv.Len(), check.Equals, t.end-t.start+2)
		sv.SetRange(t.start-1, t.end+1, t.zero)
		c.Check(sv.t.Len(), check.Equals, 2)
		c.Check(sv.String(), check.Equals, fmt.Sprintf("[%d:%v %d:%v]", t.start-1, t.zero, t.end+1, nil))
	}
}

func (s *S) TestSetRange_2(c *check.C) {
	sv, _ := New(0, 1, nil)
	c.Check(func() { sv.SetRange(1, 0, nil) }, check.Panics, ErrInvertedRange)
	type posRange struct {
		start, end int
		val        Int
	}
	for i, t := range []struct {
		start, end int
		zero       Int
		sets       []posRange
		expect     string
	}{
		{1, 10, 0,
			[]posRange{
				{1, 2, 2},
				{2, 3, 3},
				{3, 4, 3},
				{4, 5, 3},
				{5, 6, 2},
				{-10, -1, 4},
				{23, 35, 10},
			},
			"[-10:4 -1:0 1:2 2:3 5:2 6:0 23:10 35:<nil>]",
		},
		{1, 2, 0,
			[]posRange{
				{1, 1, 2},
			},
			"[1:0 2:<nil>]",
		},
		{1, 10, 0,
			[]posRange{
				{-10, 1, 0},
			},
			"[-10:0 10:<nil>]",
		},
		{1, 10, 0,
			[]posRange{
				{-10, 1, 1},
			},
			"[-10:1 1:0 10:<nil>]",
		},
		{1, 10, 0,
			[]posRange{
				{-10, 0, 1},
			},
			"[-10:1 0:0 10:<nil>]",
		},
		{1, 10, 0,
			[]posRange{
				{-10, 0, 0},
			},
			"[-10:0 10:<nil>]",
		},
		{1, 10, 0,
			[]posRange{
				{10, 20, 0},
			},
			"[1:0 20:<nil>]",
		},
		{1, 10, 0,
			[]posRange{
				{10, 20, 1},
			},
			"[1:0 10:1 20:<nil>]",
		},
		{1, 10, 0,
			[]posRange{
				{11, 20, 0},
			},
			"[1:0 20:<nil>]",
		},
		{1, 10, 0,
			[]posRange{
				{11, 20, 1},
			},
			"[1:0 11:1 20:<nil>]",
		},
		{1, 10, 0,
			[]posRange{
				{1, 10, 1},
				{11, 20, 1},
			},
			"[1:1 10:0 11:1 20:<nil>]",
		},
		{1, 10, 0,
			[]posRange{
				{2, 5, 1},
				{2, 5, 0},
			},
			"[1:0 10:<nil>]",
		},
		{1, 10, 0,
			[]posRange{
				{2, 6, 1},
				{2, 5, 0},
			},
			"[1:0 5:1 6:0 10:<nil>]",
		},
		{1, 10, 0,
			[]posRange{
				{1, 3, 1},
				{5, 7, 2},
				{3, 5, 1},
			},
			"[1:1 5:2 7:0 10:<nil>]",
		},
		{1, 10, 0,
			[]posRange{
				{1, 3, 1},
				{5, 7, 2},
				{3, 5, 2},
			},
			"[1:1 3:2 7:0 10:<nil>]",
		},
		{1, 10, 0,
			[]posRange{
				{2, 5, 1},
				{2, 6, 0},
			},
			"[1:0 10:<nil>]",
		},
		{1, 10, 0,
			[]posRange{
				{2, 6, 1},
				{2, 5, 0},
			},
			"[1:0 5:1 6:0 10:<nil>]",
		},
		{1, 10, 0,
			[]posRange{
				{2, 5, 1},
				{2, 5, 2},
			},
			"[1:0 2:2 5:0 10:<nil>]",
		},
		{1, 10, 0,
			[]posRange{
				{2, 5, 1},
				{3, 5, 2},
			},
			"[1:0 2:1 3:2 5:0 10:<nil>]",
		},
		{1, 10, 0,
			[]posRange{
				{2, 5, 1},
				{3, 5, 0},
			},
			"[1:0 2:1 3:0 10:<nil>]",
		},
	} {
		sv, err := New(t.start, t.end, t.zero)
		c.Assert(err, check.Equals, nil)
		sv.Relaxed = true
		for _, v := range t.sets {
			sv.SetRange(v.start, v.end, v.val)
		}
		c.Check(sv.String(), check.Equals, t.expect, check.Commentf("subtest %d", i))
	}
}

func (s *S) TestStepAt(c *check.C) {
	type posRange struct {
		start, end int
		val        Int
	}
	t := struct {
		start, end int
		zero       Int
		sets       []posRange
		expect     string
	}{1, 10, 0,
		[]posRange{
			{1, 3, 3},
			{4, 5, 1},
			{7, 8, 2},
			{9, 10, 4},
		},
		"[1:3 3:0 4:1 5:0 7:2 8:0 9:4 10:<nil>]",
	}

	sv, err := New(t.start, t.end, t.zero)
	c.Assert(err, check.Equals, nil)
	for _, v := range t.sets {
		sv.SetRange(v.start, v.end, v.val)
	}
	c.Check(sv.String(), check.Equals, t.expect)
	for i, v := range t.sets {
		for j := v.start; j < v.end; j++ {
			st, en, at, err := sv.StepAt(v.start)
			c.Check(err, check.Equals, nil)
			c.Check(at, check.DeepEquals, v.val)
			c.Check(st, check.Equals, v.start)
			c.Check(en, check.Equals, v.end)
		}
		st, en, at, err := sv.StepAt(v.end)
		if v.end < sv.End() {
			c.Check(err, check.Equals, nil)
			c.Check(at, check.DeepEquals, sv.Zero)
			c.Check(st, check.Equals, v.end)
			c.Check(en, check.Equals, t.sets[i+1].start)
		} else {
			c.Check(err, check.ErrorMatches, ErrOutOfRange.Error())
		}
	}
	_, _, _, err = sv.StepAt(t.start - 1)
	c.Check(err, check.ErrorMatches, ErrOutOfRange.Error())
}

func (s *S) TestDo(c *check.C) {
	var data interface{}
	type posRange struct {
		start, end int
		val        Int
	}
	for i, t := range []struct {
		start, end int
		zero       Int
		sets       []posRange
		setup      func()
		fn         Operation
		expect     interface{}
	}{
		{1, 10, 0,
			[]posRange{
				{1, 3, 3},
				{4, 5, 1},
				{7, 8, 2},
				{9, 10, 4},
			},
			func() { data = []Int(nil) },
			func(start, end int, vi Equaler) {
				sl := data.([]Int)
				v := vi.(Int)
				for i := start; i < end; i++ {
					sl = append(sl, v)
				}
				data = sl
			},
			[]Int{3, 3, 0, 1, 0, 0, 2, 0, 4},
		},
	} {
		t.setup()
		sv, err := New(t.start, t.end, t.zero)
		c.Assert(err, check.Equals, nil)
		for _, v := range t.sets {
			sv.SetRange(v.start, v.end, v.val)
		}
		sv.Do(t.fn)
		c.Check(data, check.DeepEquals, t.expect, check.Commentf("subtest %d", i))
		c.Check(reflect.ValueOf(data).Len(), check.Equals, sv.Len())
	}
}

func (s *S) TestDoRange(c *check.C) {
	var data interface{}
	type posRange struct {
		start, end int
		val        Int
	}
	for i, t := range []struct {
		start, end int
		zero       Int
		sets       []posRange
		setup      func()
		fn         Operation
		from, to   int
		expect     interface{}
		err        error
	}{
		{1, 10, 0,
			[]posRange{
				{1, 3, 3},
				{4, 5, 1},
				{7, 8, 2},
				{9, 10, 4},
			},
			func() { data = []Int(nil) },
			func(start, end int, vi Equaler) {
				sl := data.([]Int)
				v := vi.(Int)
				for i := start; i < end; i++ {
					sl = append(sl, v)
				}
				data = sl
			},
			2, 8,
			[]Int{3, 0, 1, 0, 0, 2},
			nil,
		},
		{1, 10, 0,
			[]posRange{
				{1, 3, 3},
				{4, 5, 1},
				{7, 8, 2},
				{9, 10, 4},
			},
			func() { data = []Int(nil) },
			func(_, _ int, _ Equaler) {},
			-2, -1,
			[]Int(nil),
			ErrOutOfRange,
		},
		{1, 10, 0,
			[]posRange{
				{1, 3, 3},
				{4, 5, 1},
				{7, 8, 2},
				{9, 10, 4},
			},
			func() { data = []Int(nil) },
			func(_, _ int, _ Equaler) {},
			10, 1,
			[]Int(nil),
			ErrInvertedRange,
		},
	} {
		t.setup()
		sv, err := New(t.start, t.end, t.zero)
		c.Assert(err, check.Equals, nil)
		for _, v := range t.sets {
			sv.SetRange(v.start, v.end, v.val)
		}
		c.Check(sv.DoRange(t.from, t.to, t.fn), check.DeepEquals, t.err)
		c.Check(data, check.DeepEquals, t.expect, check.Commentf("subtest %d", i))
		if t.from <= t.to && t.from < sv.End() && t.to > sv.Start() {
			c.Check(reflect.ValueOf(data).Len(), check.Equals, t.to-t.from)
		}
	}
}

func (s *S) TestApply(c *check.C) {
	type posRange struct {
		start, end int
		val        Equaler
	}
	for i, t := range []struct {
		start, end int
		zero       Equaler
		sets       []posRange
		mutate     Mutator
		expect     string
	}{
		{1, 10, Int(0),
			[]posRange{
				{1, 3, Int(3)},
				{4, 5, Int(1)},
				{7, 8, Int(2)},
				{9, 10, Int(4)},
			},
			IncInt,
			"[1:4 3:1 4:2 5:1 7:3 8:1 9:5 10:<nil>]",
		},
		{1, 10, Int(0),
			[]posRange{
				{1, 3, Int(3)},
				{4, 5, Int(1)},
				{7, 8, Int(2)},
				{9, 10, Int(4)},
			},
			DecInt,
			"[1:2 3:-1 4:0 5:-1 7:1 8:-1 9:3 10:<nil>]",
		},
		{1, 10, Float(0),
			[]posRange{
				{1, 3, Float(3)},
				{4, 5, Float(1)},
				{7, 8, Float(2)},
				{9, 10, Float(4)},
			},
			IncFloat,
			"[1:4 3:1 4:2 5:1 7:3 8:1 9:5 10:<nil>]",
		},
		{1, 10, Float(0),
			[]posRange{
				{1, 3, Float(3)},
				{4, 5, Float(1)},
				{7, 8, Float(2)},
				{9, 10, Float(4)},
			},
			DecFloat,
			"[1:2 3:-1 4:0 5:-1 7:1 8:-1 9:3 10:<nil>]",
		},
		{1, 10, Int(0),
			[]posRange{
				{1, 3, Int(3)},
				{4, 5, Int(1)},
				{7, 8, Int(2)},
				{9, 10, Int(4)},
			},
			func(_ Equaler) Equaler { return Int(0) },
			"[1:0 10:<nil>]",
		},
	} {
		sv, err := New(t.start, t.end, t.zero)
		c.Assert(err, check.Equals, nil)
		for _, v := range t.sets {
			sv.SetRange(v.start, v.end, v.val)
		}
		sv.Apply(t.mutate)
		c.Check(sv.String(), check.Equals, t.expect, check.Commentf("subtest %d", i))
	}
}

func (s *S) TestMutateRange(c *check.C) {
	type posRange struct {
		start, end int
		val        Int
	}
	for i, t := range []struct {
		start, end int
		zero       Int
		sets       []posRange
		mutate     Mutator
		from, to   int
		expect     string
		err        error
	}{
		{1, 10, 0,
			[]posRange{
				{1, 3, 3},
				{4, 5, 1},
				{7, 8, 2},
				{9, 10, 4},
			},
			IncInt,
			2, 8,
			"[1:3 2:4 3:1 4:2 5:1 7:3 8:0 9:4 10:<nil>]",
			nil,
		},
		{1, 10, 0,
			[]posRange{
				{1, 3, 3},
				{7, 8, 2},
				{9, 10, 4},
			},
			IncInt,
			4, 6,
			"[1:3 3:0 4:1 6:0 7:2 8:0 9:4 10:<nil>]",
			nil,
		},
		{1, 10, 0,
			[]posRange{
				{1, 3, 3},
				{7, 8, 1},
				{9, 10, 4},
			},
			IncInt,
			4, 7,
			"[1:3 3:0 4:1 8:0 9:4 10:<nil>]",
			nil,
		},
		{1, 10, 0,
			[]posRange{
				{1, 3, 3},
				{4, 5, 1},
				{7, 8, 2},
				{9, 10, 4},
			},
			func(_ Equaler) Equaler { return Int(0) },
			2, 8,
			"[1:3 2:0 9:4 10:<nil>]",
			nil,
		},
		{1, 10, 0,
			[]posRange{
				{1, 3, 3},
				{7, 8, 1},
				{9, 10, 4},
			},
			IncInt,
			4, 8,
			"[1:3 3:0 4:1 7:2 8:0 9:4 10:<nil>]",
			nil,
		},
		{1, 20, 0,
			[]posRange{
				{5, 10, 1},
				{10, 15, 2},
				{15, 20, 3},
			},
			func(v Equaler) Equaler {
				if v.Equal(Int(3)) {
					return Int(1)
				}
				return v
			},
			8, 18,
			"[1:0 5:1 10:2 15:1 18:3 20:<nil>]",
			nil,
		},
		{1, 10, 0,
			[]posRange{
				{1, 3, 3},
				{4, 5, 1},
				{7, 8, 2},
				{9, 10, 4},
			},
			IncInt,
			-1, 0,
			"[1:3 3:0 4:1 5:0 7:2 8:0 9:4 10:<nil>]",
			ErrOutOfRange,
		},
		{1, 10, 0,
			[]posRange{
				{1, 3, 3},
				{4, 5, 1},
				{7, 8, 2},
				{9, 10, 4},
			},
			IncInt,
			10, 1,
			"[1:3 3:0 4:1 5:0 7:2 8:0 9:4 10:<nil>]",
			ErrInvertedRange,
		},
	} {
		sv, err := New(t.start, t.end, t.zero)
		c.Assert(err, check.Equals, nil)
		for _, v := range t.sets {
			sv.SetRange(v.start, v.end, v.val)
		}
		c.Check(sv.ApplyRange(t.from, t.to, t.mutate), check.DeepEquals, t.err)
		c.Check(sv.String(), check.Equals, t.expect, check.Commentf("subtest %d", i))
	}
}

// Benchmarks

func applyRange(b *testing.B, coverage float64) {
	b.StopTimer()
	var (
		length = 100
		start  = 0
		end    = int(float64(b.N)/coverage) / length
		zero   = Int(0)
		pool   = make([]int, b.N)
	)
	if end == 0 {
		return
	}
	sv, _ := New(start, end, zero)
	for i := 0; i < b.N; i++ {
		pool[i] = rand.Intn(end)
	}
	b.StartTimer()
	for _, r := range pool {
		sv.ApplyRange(r, r+length, IncInt)
	}
}

func BenchmarkApplyRangeXDense(b *testing.B) {
	applyRange(b, 1000)
}
func BenchmarkApplyRangeVDense(b *testing.B) {
	applyRange(b, 100)
}
func BenchmarkApplyRangeDense(b *testing.B) {
	applyRange(b, 10)
}
func BenchmarkApplyRangeUnity(b *testing.B) {
	applyRange(b, 1)
}
func BenchmarkApplyRangeSparse(b *testing.B) {
	applyRange(b, 0.1)
}
func BenchmarkApplyRangeVSparse(b *testing.B) {
	applyRange(b, 0.01)
}
func BenchmarkApplyRangeXSparse(b *testing.B) {
	applyRange(b, 0.001)
}

func atFunc(b *testing.B, coverage float64) {
	b.StopTimer()
	var (
		length = 100
		start  = 0
		end    = int(float64(b.N)/coverage) / length
		zero   = Int(0)
	)
	if end == 0 {
		return
	}
	sv, _ := New(start, end, zero)
	for i := 0; i < b.N; i++ {
		r := rand.Intn(end)
		sv.ApplyRange(r, r+length, IncInt)
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_, err := sv.At(rand.Intn(end))
		if err != nil {
			panic("cannot reach")
		}
	}
}

func BenchmarkAtXDense(b *testing.B) {
	atFunc(b, 1000)
}
func BenchmarkAtVDense(b *testing.B) {
	atFunc(b, 100)
}
func BenchmarkAtDense(b *testing.B) {
	atFunc(b, 10)
}
func BenchmarkAtUnity(b *testing.B) {
	atFunc(b, 1)
}
func BenchmarkAtSparse(b *testing.B) {
	atFunc(b, 0.1)
}
func BenchmarkAtVSparse(b *testing.B) {
	atFunc(b, 0.01)
}
func BenchmarkAtXSparse(b *testing.B) {
	atFunc(b, 0.001)
}
