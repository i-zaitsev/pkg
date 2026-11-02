// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Package assert is a minimal set of test assertions.
//
// Every assertion reports through t.Errorf, so a failing check marks the test as
// failed but lets it continue. Values can be checked with plain functions,
// assert.Eq(t, got, want), or fluently, assert.Val(t, got).NotZero().Eq(want).
//
// The assertion logic lives in check, a pure backend that records failure
// messages without touching testing. The exported functions run a check and
// report its failures to t.
package assert

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// check is the pure assertion backend: it evaluates conditions and records
// failure messages instead of reporting them to a testing.TB.
type check struct {
	fails []string
	dummy bool
}

func (c *check) fail(format string, args ...any) {
	c.fails = append(c.fails, fmt.Sprintf(format, args...))
}

// report forwards recorded failures to t.
func (c *check) report(t testing.TB) {
	t.Helper()
	if !c.dummy {
		for _, msg := range c.fails {
			t.Errorf("%s", msg)
		}
	}
}

func (c *check) err(err error) {
	if err == nil {
		c.fail("expected error, got nil")
	}
}

func (c *check) notErr(err error) {
	if err != nil {
		c.fail("unexpected error: %v", err)
	}
}

func (c *check) errIs(err, target error) {
	if !errors.Is(err, target) {
		c.fail("expected error matching %q, got: %v", target, err)
	}
}

func (c *check) errContains(err error, substr string) {
	if err == nil {
		c.fail("expected error containing %q, got nil", substr)
		return
	}
	if !strings.Contains(err.Error(), substr) {
		c.fail("expected error containing %q, got: %q", substr, err.Error())
	}
}

func (c *check) noDiff(want, got any, opts ...cmp.Option) {
	if diff := cmp.Diff(want, got, opts...); diff != "" {
		c.fail("(-want, +got):\n%s", diff)
	}
}

func (c *check) isTrue(cond bool) {
	if !cond {
		c.fail("expected true, got false")
	}
}

func (c *check) isFalse(cond bool) {
	if cond {
		c.fail("expected false, got true")
	}
}

func (c *check) isNil(val any) {
	if !isNil(val) {
		c.fail("expected nil, got: %+v", val)
	}
}

func (c *check) notNil(val any) {
	if isNil(val) {
		c.fail("expected non-nil value, got nil")
	}
}

func (c *check) contains(s, substr string) {
	if !strings.Contains(s, substr) {
		c.fail("expected %q to contain %q", s, substr)
	}
}

func (c *check) notContains(s, substr string) {
	if strings.Contains(s, substr) {
		c.fail("expected %q not to contain %q", s, substr)
	}
}

func (c *check) pathExists(path string) {
	if _, err := os.Stat(path); err != nil {
		c.fail("expected path %q to exist: %v", path, err)
	}
}

func (c *check) noPathExists(path string) {
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		c.fail("expected path %q not to exist", path)
	}
}

func (c *check) panics(fn func()) {
	defer func() {
		if recover() == nil {
			c.fail("expected panic, function returned normally")
		}
	}()
	fn()
}

func asErr[E error](c *check, err error) E {
	var e E
	if !errors.As(err, &e) {
		c.fail("expected error of type %T, got: %v", e, err)
	}
	return e
}

func eq[T any](c *check, got, want T) {
	if !reflect.DeepEqual(got, want) {
		c.fail("not equal:\n got: %+v\nwant: %+v", got, want)
	}
}

func notEq[T any](c *check, got, other T) {
	if reflect.DeepEqual(got, other) {
		c.fail("expected values to differ, both are: %+v", got)
	}
}

func zero[T any](c *check, val T) {
	if !isZero(val) {
		c.fail("expected zero value, got: %+v", val)
	}
}

func notZero[T any](c *check, val T) {
	if isZero(val) {
		c.fail("expected non-zero value, got: %+v", val)
	}
}

func length[S ~[]E, E any](c *check, s S, n int) {
	if len(s) != n {
		c.fail("expected length %d, got %d: %+v", n, len(s), s)
	}
}

func lengthMap[M ~map[K]V, K comparable, V any](c *check, m M, n int) {
	if len(m) != n {
		c.fail("expected length %d, got %d: %+v", n, len(m), m)
	}
}

func hasKey[M ~map[K]V, K comparable, V any](c *check, m M, key K) {
	if _, ok := m[key]; !ok {
		c.fail("expected map to have key %v", key)
	}
}

func noKey[M ~map[K]V, K comparable, V any](c *check, m M, key K) {
	if _, ok := m[key]; ok {
		c.fail("expected map not to have key %v", key)
	}
}

func isZero[T any](val T) bool {
	return reflect.ValueOf(&val).Elem().IsZero()
}

func isNil(val any) bool {
	if val == nil {
		return true
	}
	rv := reflect.ValueOf(val)
	switch rv.Kind() {
	case reflect.Pointer, reflect.Map, reflect.Slice, reflect.Chan, reflect.Func, reflect.Interface:
		return rv.IsNil()
	default:
		return false
	}
}

var createCheck = func() *check {
	return &check{}
}

// Err checks that err is not nil.
func Err(t testing.TB, err error) {
	t.Helper()
	c := createCheck()
	c.err(err)
	c.report(t)
}

// NotErr checks that err is nil.
func NotErr(t testing.TB, err error) {
	t.Helper()
	c := createCheck()
	c.notErr(err)
	c.report(t)
}

// ErrIs checks that err matches the target, according to errors.Is.
func ErrIs(t testing.TB, err, target error) {
	t.Helper()
	c := createCheck()
	c.errIs(err, target)
	c.report(t)
}

// ErrContains checks that err is not nil and its message contains substr.
func ErrContains(t testing.TB, err error, substr string) {
	t.Helper()
	c := createCheck()
	c.errContains(err, substr)
	c.report(t)
}

// AsErr checks that err matches the error type E according to errors.As,
// and returns the typed error for further checks.
func AsErr[E error](t testing.TB, err error) E {
	t.Helper()
	c := createCheck()
	e := asErr[E](c, err)
	c.report(t)
	return e
}

// Eq checks that got and want are deeply equal.
func Eq[T any](t testing.TB, got, want T) {
	t.Helper()
	c := createCheck()
	eq(c, got, want)
	c.report(t)
}

// NotEq checks that got and other are not deeply equal.
func NotEq[T any](t testing.TB, got, other T) {
	t.Helper()
	c := createCheck()
	notEq(c, got, other)
	c.report(t)
}

// NoDiff checks that want and got have no difference according to cmp.Diff.
func NoDiff(t testing.TB, want, got any, opts ...cmp.Option) {
	t.Helper()
	c := createCheck()
	c.noDiff(want, got, opts...)
	c.report(t)
}

// True checks that cond is true.
func True(t testing.TB, cond bool) {
	t.Helper()
	c := createCheck()
	c.isTrue(cond)
	c.report(t)
}

// False checks that cond is false.
func False(t testing.TB, cond bool) {
	t.Helper()
	c := createCheck()
	c.isFalse(cond)
	c.report(t)
}

// Zero checks that val is the zero value of its type.
func Zero[T any](t testing.TB, val T) {
	t.Helper()
	c := createCheck()
	zero(c, val)
	c.report(t)
}

// NotZero checks that val is not the zero value of its type.
func NotZero[T any](t testing.TB, val T) {
	t.Helper()
	c := createCheck()
	notZero(c, val)
	c.report(t)
}

// Nil checks that val is nil or a nil pointer, map, slice, channel, func, or interface.
func Nil(t testing.TB, val any) {
	t.Helper()
	c := createCheck()
	c.isNil(val)
	c.report(t)
}

// NotNil checks that val is not nil, see Nil.
func NotNil(t testing.TB, val any) {
	t.Helper()
	c := createCheck()
	c.notNil(val)
	c.report(t)
}

// Len checks that the slice s has length n.
func Len[S ~[]E, E any](t testing.TB, s S, n int) {
	t.Helper()
	c := createCheck()
	length(c, s, n)
	c.report(t)
}

// LenMap checks that the map m has n entries.
func LenMap[M ~map[K]V, K comparable, V any](t testing.TB, m M, n int) {
	t.Helper()
	c := createCheck()
	lengthMap(c, m, n)
	c.report(t)
}

// HasKey checks that the map m has the key.
func HasKey[M ~map[K]V, K comparable, V any](t testing.TB, m M, key K) {
	t.Helper()
	c := createCheck()
	hasKey(c, m, key)
	c.report(t)
}

// NoKey checks that the map m does not have the key.
func NoKey[M ~map[K]V, K comparable, V any](t testing.TB, m M, key K) {
	t.Helper()
	c := createCheck()
	noKey(c, m, key)
	c.report(t)
}

// Contains checks that s contains substr.
func Contains(t testing.TB, s, substr string) {
	t.Helper()
	c := createCheck()
	c.contains(s, substr)
	c.report(t)
}

// ContainsAll checks that s contains every one of substrs.
func ContainsAll(t testing.TB, s string, substrs ...string) {
	t.Helper()
	c := createCheck()
	for _, substr := range substrs {
		c.contains(s, substr)
	}
	c.report(t)
}

// ContainsNone checks that s contains none of substrs.
func ContainsNone(t testing.TB, s string, substrs ...string) {
	t.Helper()
	c := createCheck()
	for _, substr := range substrs {
		c.notContains(s, substr)
	}
	c.report(t)
}

// PathExists checks that the file or directory at path exists.
func PathExists(t testing.TB, path string) {
	t.Helper()
	c := createCheck()
	c.pathExists(path)
	c.report(t)
}

// NoPathExists checks that nothing exists at path.
func NoPathExists(t testing.TB, path string) {
	t.Helper()
	c := createCheck()
	c.noPathExists(path)
	c.report(t)
}

// Panics checks that fn panics.
func Panics(t testing.TB, fn func()) {
	t.Helper()
	c := createCheck()
	c.panics(fn)
	c.report(t)
}

// Value wraps a value under test for fluent, chainable assertions.
type Value[T any] struct {
	t   testing.TB
	val T
}

// Val starts a fluent assertion on val.
func Val[T any](t testing.TB, val T) *Value[T] {
	return &Value[T]{t: t, val: val}
}

// Eq checks that the value is deeply equal to want.
func (v *Value[T]) Eq(want T) *Value[T] {
	v.t.Helper()
	Eq(v.t, v.val, want)
	return v
}

// NotEq checks that the value is not deeply equal to other.
func (v *Value[T]) NotEq(other T) *Value[T] {
	v.t.Helper()
	NotEq(v.t, v.val, other)
	return v
}

// Zero checks that the value is the zero value of its type.
func (v *Value[T]) Zero() *Value[T] {
	v.t.Helper()
	Zero(v.t, v.val)
	return v
}

// NotZero checks that the value is not the zero value of its type.
func (v *Value[T]) NotZero() *Value[T] {
	v.t.Helper()
	NotZero(v.t, v.val)
	return v
}
