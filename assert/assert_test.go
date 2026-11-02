// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package assert

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp/cmpopts"
)

// backend collects failures of all assertions made since the last
// noFails or hasFails call, so a failure early in a chain is not lost.
var backend = check{dummy: true}

func createTestCheck() *check {
	return &backend
}

func (c *check) noFails(t *testing.T, msg string, args ...any) {
	t.Helper()
	if len(c.fails) != 0 {
		t.Errorf(msg, args...)
	}
	c.fails = nil
}

func (c *check) hasFails(t *testing.T, msg string, args ...any) {
	t.Helper()
	if len(c.fails) == 0 {
		t.Errorf(msg, args...)
	}
	c.fails = nil
}

func TestMain(m *testing.M) {
	createCheck = createTestCheck
	os.Exit(m.Run())
}

type recorder struct {
	testing.TB
	msgs []string
}

func (r *recorder) Helper() {}

func (r *recorder) Errorf(format string, args ...any) {
	r.msgs = append(r.msgs, fmt.Sprintf(format, args...))
}

func TestReport_realCheck_forwardsFailures(t *testing.T) {
	rec := &recorder{TB: t}
	c := &check{}

	c.fail("first")
	c.fail("second")
	c.report(rec)

	if len(rec.msgs) != 2 || rec.msgs[0] != "first" || rec.msgs[1] != "second" {
		t.Errorf("reported %q, want [first second]", rec.msgs)
	}
}

func TestErr(t *testing.T) {
	Err(t, errors.New("expected"))
	backend.noFails(t, "non-nil error failed")

	Err(t, nil)
	backend.hasFails(t, "nil error did not fail")
}

func TestNotErr(t *testing.T) {
	NotErr(t, nil)
	backend.noFails(t, "nil error failed")

	NotErr(t, errors.New("unexpected"))
	backend.hasFails(t, "non-nil error did not fail")
}

func TestErrIs(t *testing.T) {
	tests := []struct {
		err, target error
		match       bool
	}{
		{os.ErrExist, os.ErrExist, true},
		{os.ErrExist, os.ErrNotExist, false},
	}
	for _, tt := range tests {
		ErrIs(t, tt.err, tt.target)
		if tt.match {
			backend.noFails(t, "errs match but ErrIs failed")
		} else {
			backend.hasFails(t, "errs do not match but ErrIs succeeded")
		}
	}
}

func TestErrContains(t *testing.T) {
	tests := []struct {
		err    error
		substr string
		match  bool
	}{
		{errors.New("file not found"), "not found", true},
		{errors.New("file not found"), "denied", false},
		{nil, "not found", false},
	}
	for _, tt := range tests {
		ErrContains(t, tt.err, tt.substr)
		if tt.match {
			backend.noFails(t, "err %v contains %q but ErrContains failed", tt.err, tt.substr)
		} else {
			backend.hasFails(t, "err %v does not contain %q but ErrContains succeeded", tt.err, tt.substr)
		}
	}
}

type errA struct{}

func (e *errA) Error() string { return "a" }

type errB struct{}

func (e *errB) Error() string { return "b" }

func TestAsErr(t *testing.T) {
	a := &errA{}

	got := AsErr[*errA](t, a)
	backend.noFails(t, "error cast expected to succeed but failed")
	if got != a {
		t.Errorf("AsErr returned %v, want %v", got, a)
	}

	_ = AsErr[*errB](t, a)
	backend.hasFails(t, "error cast expected to fail but succeeded")
}

func TestEquality(t *testing.T) {
	tests := []struct {
		left, right any
		equal       bool
	}{
		{1, 1, true},
		{1, 2, false},
		{1.0, 1.0, true},
		{1.000001, 1.0, false},
		{"a", "a", true},
		{"aa", "a", false},
		{[]int{1, 2}, []int{1, 2}, true},
		{[]int{1, 2}, []int{2, 1}, false},
	}
	for _, tt := range tests {
		t.Run("Eq", func(t *testing.T) {
			Eq(t, tt.left, tt.right)
			if tt.equal {
				backend.noFails(t, "%v == %v but Eq failed", tt.left, tt.right)
			} else {
				backend.hasFails(t, "%v != %v but Eq succeeded", tt.left, tt.right)
			}
		})
		t.Run("NotEq", func(t *testing.T) {
			NotEq(t, tt.left, tt.right)
			if tt.equal {
				backend.hasFails(t, "%v == %v but NotEq succeeded", tt.left, tt.right)
			} else {
				backend.noFails(t, "%v != %v but NotEq failed", tt.left, tt.right)
			}
		})
	}
}

func TestNoDiff(t *testing.T) {
	type item struct {
		Name string
		tags []string
	}

	NoDiff(t, item{Name: "a", tags: []string{"x"}}, item{Name: "a"}, cmpopts.IgnoreUnexported(item{}))
	backend.noFails(t, "values equal under options failed")

	NoDiff(t, item{Name: "a"}, item{Name: "b"}, cmpopts.IgnoreUnexported(item{}))
	backend.hasFails(t, "different values did not fail")
}

func TestTrue(t *testing.T) {
	True(t, true)
	backend.noFails(t, "true value failed")

	True(t, false)
	backend.hasFails(t, "false value did not fail")
}

func TestFalse(t *testing.T) {
	False(t, false)
	backend.noFails(t, "false value failed")

	False(t, true)
	backend.hasFails(t, "true value did not fail")
}

func TestZero(t *testing.T) {
	testZero(t, 0, true)
	testZero(t, 1, false)
	testZero(t, "", true)
	testZero(t, "a", false)
	testZero(t, []int(nil), true)
	testZero(t, []int{}, false)
}

// testZero runs Zero and NotZero on val.
// It is generic rather than table-driven over any,
// because Zero[any] would test the interface, not the value inside.
func testZero[T any](t *testing.T, val T, zero bool) {
	t.Helper()
	t.Run("Zero", func(t *testing.T) {
		Zero(t, val)
		if zero {
			backend.noFails(t, "%#v is zero but Zero failed", val)
		} else {
			backend.hasFails(t, "%#v is not zero but Zero succeeded", val)
		}
	})
	t.Run("NotZero", func(t *testing.T) {
		NotZero(t, val)
		if zero {
			backend.hasFails(t, "%#v is zero but NotZero succeeded", val)
		} else {
			backend.noFails(t, "%#v is not zero but NotZero failed", val)
		}
	})
}

func TestNil(t *testing.T) {
	var nilPtr *int
	var nilMap map[string]int
	tests := []struct {
		name  string
		val   any
		isNil bool
	}{
		{"untyped_nil", nil, true},
		{"nil_pointer", nilPtr, true},
		{"nil_map", nilMap, true},
		{"pointer", new(int), false},
		{"value", 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Nil(t, tt.val)
			if tt.isNil {
				backend.noFails(t, "%#v is nil but Nil failed", tt.val)
			} else {
				backend.hasFails(t, "%#v is not nil but Nil succeeded", tt.val)
			}

			NotNil(t, tt.val)
			if tt.isNil {
				backend.hasFails(t, "%#v is nil but NotNil succeeded", tt.val)
			} else {
				backend.noFails(t, "%#v is not nil but NotNil failed", tt.val)
			}
		})
	}
}

func TestLen(t *testing.T) {
	Len(t, []int{1, 2}, 2)
	backend.noFails(t, "slice of matching length failed")

	Len(t, []int{1}, 2)
	backend.hasFails(t, "slice of wrong length did not fail")
}

func TestLenMap(t *testing.T) {
	LenMap(t, map[string]int{"a": 1}, 1)
	backend.noFails(t, "map of matching length failed")

	LenMap(t, map[string]int{}, 1)
	backend.hasFails(t, "map of wrong length did not fail")
}

func TestHasKey(t *testing.T) {
	m := map[string]int{"a": 1}

	HasKey(t, m, "a")
	backend.noFails(t, "present key failed")

	HasKey(t, m, "b")
	backend.hasFails(t, "missing key did not fail")

	NoKey(t, m, "b")
	backend.noFails(t, "missing key failed")

	NoKey(t, m, "a")
	backend.hasFails(t, "present key did not fail")
}

func TestContains(t *testing.T) {
	tests := []struct {
		s, substr string
		match     bool
	}{
		{"hello", "ell", true},
		{"hello", "", true},
		{"hello", "bye", false},
		{"", "a", false},
	}
	for _, tt := range tests {
		Contains(t, tt.s, tt.substr)
		if tt.match {
			backend.noFails(t, "%q contains %q but Contains failed", tt.s, tt.substr)
		} else {
			backend.hasFails(t, "%q does not contain %q but Contains succeeded", tt.s, tt.substr)
		}
	}
}

func TestContainsAll(t *testing.T) {
	ContainsAll(t, "hello world", "hello", "world")
	backend.noFails(t, "all substrings present but ContainsAll failed")

	ContainsAll(t, "hello world", "hello", "bye")
	backend.hasFails(t, "missing substring did not fail")
}

func TestContainsNone(t *testing.T) {
	ContainsNone(t, "hello world", "bye", "moon")
	backend.noFails(t, "no substrings present but ContainsNone failed")

	ContainsNone(t, "hello world", "bye", "world")
	backend.hasFails(t, "present substring did not fail")
}

func TestPathExists(t *testing.T) {
	dir := t.TempDir()
	missing := filepath.Join(dir, "missing")

	PathExists(t, dir)
	backend.noFails(t, "existing path failed")

	PathExists(t, missing)
	backend.hasFails(t, "missing path did not fail")

	NoPathExists(t, missing)
	backend.noFails(t, "missing path failed")

	NoPathExists(t, dir)
	backend.hasFails(t, "existing path did not fail")
}

func TestPanics(t *testing.T) {
	Panics(t, func() { panic("boom") })
	backend.noFails(t, "panicking func failed")

	Panics(t, func() {})
	backend.hasFails(t, "returning func did not fail")
}

func TestValue(t *testing.T) {
	tests := []struct {
		name  string
		chain func(t testing.TB)
		fails bool
	}{
		{"non_zero", func(t testing.TB) { Val(t, 1).Eq(1).NotEq(2).NotZero() }, false},
		{"zero", func(t testing.TB) { Val(t, 0).Eq(0).Zero() }, false},
		{"eq_fails", func(t testing.TB) { Val(t, 1).Eq(2) }, true},
		{"not_eq_fails", func(t testing.TB) { Val(t, "a").NotEq("a") }, true},
		{"zero_fails", func(t testing.TB) { Val(t, []int{}).Zero() }, true},
		{"not_zero_fails", func(t testing.TB) { Val(t, "").NotZero() }, true},
		{"mid_chain_fails", func(t testing.TB) { Val(t, 1).Eq(2).NotZero() }, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.chain(t)
			if tt.fails {
				backend.hasFails(t, "chain should fail but succeeded")
			} else {
				backend.noFails(t, "chain should succeed but failed")
			}
		})
	}
}
