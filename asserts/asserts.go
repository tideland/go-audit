// Tideland Go Audit - Asserts
//
// Copyright (C) 2012-2020 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package asserts // import "tideland.dev/go/audit/asserts"

//--------------------
// IMPORTS
//--------------------

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

//--------------------
// ASSERTS
//--------------------

// Asserts provides a number of convenient test methods.
type Asserts struct {
	tester Tester
	failer Failer
}

// New creates a new Asserts instance.
func New(f Failer) *Asserts {
	return &Asserts{
		tester: Tester{},
		failer: f,
	}
}

// SetPrinter sets a new Printer used for the output of failing
// tests or logging. The current one is returned, e.g. for a
// later restoring.
func (a *Asserts) SetPrinter(printer Printer) Printer {
	return a.failer.SetPrinter(printer)
}

// SetFailable allows to change the failable possibly used inside
// a failer. This way a testing.T of a sub-test can be injected. A
// restore function is returned.
//
//     t.Run(name, func(t *testing.T)) {
//         defer assert.SetFailable(t)()
//         ...
//     })
//
// So the returned restorer function will be called when
// leaving the sub-test.
func (a *Asserts) SetFailable(f Failable) func() {
	tf, ok := a.failer.(*testingFailer)
	if !ok {
		// Nothing to do.
		return func() {}
	}
	// It's a test assertion.
	old := tf.failable
	tf.failable = f
	return func() {
		tf.failable = old
	}
}

// IncrCallstackOffset allows test libraries using the audit
// package internally to adjust the callstack offset. This
// way test output shows the correct location. Deferring
// the returned function restores the former offset.
func (a *Asserts) IncrCallstackOffset() func() {
	return a.failer.IncrCallstackOffset()
}

// Logf can be used to display helpful information during testing.
func (a *Asserts) Logf(format string, args ...interface{}) {
	a.failer.Logf(format, args...)
}

// OK is a convenient metatest depending in the obtained tyoe. In case
// of a bool it has to be true, an int has to be 0, a string has to
// be empty, and a func() error has to return no error. Any else value
// has to be nil or in case of an ErrorProne its Err() has to return nil.
func (a *Asserts) OK(obtained interface{}, msgs ...string) bool {
	switch o := obtained.(type) {
	case bool:
		return a.True(o, msgs...)
	case int:
		return a.Equal(o, 0, msgs...)
	case string:
		return a.Equal(o, "", msgs...)
	case func() error:
		return a.NoError(o(), msgs...)
	default:
		return a.NoError(obtained, msgs...)
	}
}

// True tests if obtained is true.
func (a *Asserts) True(obtained bool, msgs ...string) bool {
	if !a.tester.IsTrue(obtained) {
		return a.failer.Fail(True, obtained, true, msgs...)
	}
	return true
}

// False tests if obtained is false.
func (a *Asserts) False(obtained bool, msgs ...string) bool {
	if a.tester.IsTrue(obtained) {
		return a.failer.Fail(False, obtained, false, msgs...)
	}
	return true
}

// Nil tests if obtained is nil.
func (a *Asserts) Nil(obtained interface{}, msgs ...string) bool {
	if !a.tester.IsNil(obtained) {
		return a.failer.Fail(Nil, obtained, nil, msgs...)
	}
	return true
}

// NotNil tests if obtained is not nil.
func (a *Asserts) NotNil(obtained interface{}, msgs ...string) bool {
	if a.tester.IsNil(obtained) {
		return a.failer.Fail(NotNil, obtained, nil, msgs...)
	}
	return true
}

// Equal tests if obtained and expected are equal.
func (a *Asserts) Equal(obtained, expected interface{}, msgs ...string) bool {
	if !a.tester.IsEqual(obtained, expected) {
		return a.failer.Fail(Equal, obtained, expected, msgs...)
	}
	return true
}

// Different tests if obtained and expected are different.
func (a *Asserts) Different(obtained, expected interface{}, msgs ...string) bool {
	if a.tester.IsEqual(obtained, expected) {
		return a.failer.Fail(Different, obtained, expected, msgs...)
	}
	return true
}

// NoError tests if the obtained error or ErrorProne.Err() is nil.
func (a *Asserts) NoError(obtained interface{}, msgs ...string) bool {
	err := ifaceToError(obtained)
	if !a.tester.IsNil(err) {
		return a.failer.Fail(NoError, err, nil, msgs...)
	}
	return true
}

// ErrorMatch tests if the obtained error as string matches a
// regular expression.
func (a *Asserts) ErrorMatch(obtained interface{}, regex string, msgs ...string) bool {
	if obtained == nil {
		return a.failer.Fail(ErrorMatch, nil, regex, "error is nil")
	}
	err := ifaceToError(obtained)
	matches, err := a.tester.IsMatching(err.Error(), regex)
	if err != nil {
		return a.failer.Fail(ErrorMatch, err, regex, "can't compile regex: "+err.Error())
	}
	if !matches {
		return a.failer.Fail(ErrorMatch, err, regex, msgs...)
	}
	return true
}

// ErrorContains tests if the obtained error contains a given string.
func (a *Asserts) ErrorContains(obtained interface{}, part string, msgs ...string) bool {
	err := ifaceToError(obtained)
	if !a.tester.IsSubstring(part, err.Error()) {
		return a.failer.Fail(ErrorContains, obtained, part, msgs...)
	}
	return true
}

// Contents tests if the obtained data is part of the expected
// string, array, or slice.
func (a *Asserts) Contents(part, full interface{}, msgs ...string) bool {
	contains, err := a.tester.Contains(part, full)
	if err != nil {
		return a.failer.Fail(Contents, part, full, "type missmatch: "+err.Error())
	}
	if !contains {
		return a.failer.Fail(Contents, part, full, msgs...)
	}
	return true
}

// NotContents tests if the obtained data is not part of the expected
// string, array, or slice.
func (a *Asserts) NotContents(part, full interface{}, msgs ...string) bool {
	contains, err := a.tester.Contains(part, full)
	if err != nil {
		return a.failer.Fail(Contents, part, full, "type missmatch: "+err.Error())
	}
	if contains {
		return a.failer.Fail(Contents, part, full, msgs...)
	}
	return true
}

// About tests if obtained and expected are near to each other
// (within the given extent).
func (a *Asserts) About(obtained, expected, extent float64, msgs ...string) bool {
	if !a.tester.IsAbout(obtained, expected, extent) {
		return a.failer.Fail(About, obtained, expected, msgs...)
	}
	return true
}

// Range tests if obtained is larger or equal low and lower or
// equal high. Allowed are byte, int and float64 for numbers, runes,
// strings, times, and duration. In case of obtained arrays,
// slices, and maps low and high have to be ints for testing
// the length.
func (a *Asserts) Range(obtained, low, high interface{}, msgs ...string) bool {
	expected := &lowHigh{low, high}
	inRange, err := a.tester.IsInRange(obtained, low, high)
	if err != nil {
		return a.failer.Fail(Range, obtained, expected, "type missmatch: "+err.Error())
	}
	if !inRange {
		return a.failer.Fail(Range, obtained, expected, msgs...)
	}
	return true
}

// Substring tests if obtained is a substring of the full string.
func (a *Asserts) Substring(obtained, full string, msgs ...string) bool {
	if !a.tester.IsSubstring(obtained, full) {
		return a.failer.Fail(Substring, obtained, full, msgs...)
	}
	return true
}

// Case tests if obtained string is uppercase or lowercase.
func (a *Asserts) Case(obtained string, upperCase bool, msgs ...string) bool {
	if !a.tester.IsCase(obtained, upperCase) {
		if upperCase {
			return a.failer.Fail(Case, obtained, strings.ToUpper(obtained), msgs...)
		}
		return a.failer.Fail(Case, obtained, strings.ToLower(obtained), msgs...)
	}
	return true
}

// Match tests if the obtained string matches a regular expression.
func (a *Asserts) Match(obtained, regex string, msgs ...string) bool {
	matches, err := a.tester.IsMatching(obtained, regex)
	if err != nil {
		return a.failer.Fail(Match, obtained, regex, "can't compile regex: "+err.Error())
	}
	if !matches {
		return a.failer.Fail(Match, obtained, regex, msgs...)
	}
	return true
}

// Implementor tests if obtained implements the expected
// interface variable pointer.
func (a *Asserts) Implementor(obtained, expected interface{}, msgs ...string) bool {
	implements, err := a.tester.IsImplementor(obtained, expected)
	if err != nil {
		return a.failer.Fail(Implementor, obtained, expected, err.Error())
	}
	if !implements {
		return a.failer.Fail(Implementor, obtained, expected, msgs...)
	}
	return implements
}

// Assignable tests if the types of expected and obtained are assignable.
func (a *Asserts) Assignable(obtained, expected interface{}, msgs ...string) bool {
	if !a.tester.IsAssignable(obtained, expected) {
		return a.failer.Fail(Assignable, obtained, expected, msgs...)
	}
	return true
}

// Unassignable tests if the types of expected and obtained are
// not assignable.
func (a *Asserts) Unassignable(obtained, expected interface{}, msgs ...string) bool {
	if a.tester.IsAssignable(obtained, expected) {
		return a.failer.Fail(Unassignable, obtained, expected, msgs...)
	}
	return true
}

// Empty tests if the len of the obtained string, array, slice
// map, or channel is 0.
func (a *Asserts) Empty(obtained interface{}, msgs ...string) bool {
	length, err := a.tester.Len(obtained)
	if err != nil {
		return a.failer.Fail(Empty, ValueDescription(obtained), 0, err.Error())
	}
	if length > 0 {
		return a.failer.Fail(Empty, length, 0, msgs...)

	}
	return true
}

// NotEmpty tests if the len of the obtained string, array, slice
// map, or channel is greater than 0.
func (a *Asserts) NotEmpty(obtained interface{}, msgs ...string) bool {
	length, err := a.tester.Len(obtained)
	if err != nil {
		return a.failer.Fail(NotEmpty, ValueDescription(obtained), 0, err.Error())
	}
	if length == 0 {
		return a.failer.Fail(NotEmpty, length, 0, msgs...)

	}
	return true
}

// Length tests if the len of the obtained string, array, slice
// map, or channel is equal to the expected one.
func (a *Asserts) Length(obtained interface{}, expected int, msgs ...string) bool {
	length, err := a.tester.Len(obtained)
	if err != nil {
		return a.failer.Fail(Length, ValueDescription(obtained), expected, err.Error())
	}
	if length != expected {
		return a.failer.Fail(Length, length, expected, msgs...)
	}
	return true
}

// Panics checks if the passed function panics.
func (a *Asserts) Panics(pf func(), msgs ...string) bool {
	if !a.tester.HasPanic(pf, nil) {
		return a.failer.Fail(Panics, ValueDescription(pf), nil, msgs...)
	}
	return true
}

// NotPanics checks if the passed function does not panic.
func (a *Asserts) NotPanics(pf func(), msgs ...string) bool {
	if a.tester.HasPanic(pf, nil) {
		return a.failer.Fail(NotPanics, ValueDescription(pf), nil, msgs...)
	}
	return true
}

// PanicsWith checks if the passed function panics with the passed reason.
func (a *Asserts) PanicsWith(pf func(), reason interface{}, msgs ...string) bool {
	if !a.tester.HasPanic(pf, reason) {
		return a.failer.Fail(PanicsWith, ValueDescription(pf), reason, msgs...)
	}
	return true
}

// PathExists checks if the passed path or file exists.
func (a *Asserts) PathExists(obtained string, msgs ...string) bool {
	valid, err := a.tester.IsValidPath(obtained)
	if err != nil {
		return a.failer.Fail(PathExists, obtained, true, err.Error())
	}
	if !valid {
		return a.failer.Fail(PathExists, obtained, true, msgs...)
	}
	return true
}

// Wait receives a signal from a channel and compares it to the
// expired value. Assert also fails on timeout.
func (a *Asserts) Wait(
	sigc <-chan interface{},
	expected interface{},
	timeout time.Duration,
	msgs ...string,
) bool {
	select {
	case obtained := <-sigc:
		if !a.tester.IsEqual(obtained, expected) {
			return a.failer.Fail(Wait, obtained, expected, msgs...)
		}
		return true
	case <-time.After(timeout):
		return a.failer.Fail(Wait, "timeout "+timeout.String(), "signal true", msgs...)
	}
}

// WaitClosed waits until a channel closing, the assert fails on a timeout.
func (a *Asserts) WaitClosed(
	sigc <-chan interface{},
	timeout time.Duration,
	msgs ...string,
) bool {
	done := time.NewTimer(timeout)
	defer done.Stop()
	for {
		select {
		case _, ok := <-sigc:
			if !ok {
				// Only return true if channel has been closed.
				return true
			}
		case <-done.C:
			return a.failer.Fail(WaitClosed, "timeout "+timeout.String(), "closed", msgs...)
		}
	}
}

// WaitGroup waits until a wait group instance is done, the assert fails on a timeout.
func (a *Asserts) WaitGroup(
	wg *sync.WaitGroup,
	timeout time.Duration,
	msgs ...string,
) bool {
	stopc := make(chan struct{}, 1)
	done := time.NewTimer(timeout)
	defer done.Stop()
	go func() {
		wg.Wait()
		stopc <- struct{}{}
	}()
	for {
		select {
		case <-stopc:
			return true
		case <-done.C:
			return a.failer.Fail(WaitGroup, "timeout "+timeout.String(), "done", msgs...)
		}
	}
}

// WaitTested receives a signal from a channel and runs the passed tester
// function on it. That has to return nil for a signal assert. In case of
// a timeout the assert fails.
func (a *Asserts) WaitTested(
	sigc <-chan interface{},
	test func(interface{}) error,
	timeout time.Duration,
	msgs ...string,
) bool {
	select {
	case obtained := <-sigc:
		err := test(obtained)
		return a.Nil(err, msgs...)
	case <-time.After(timeout):
		return a.failer.Fail(WaitTested, "timeout "+timeout.String(), "signal tested", msgs...)
	}
}

// Retry calls the passed function and expects it to return true. Otherwise
// it pauses for the given duration and retries the call the defined number.
func (a *Asserts) Retry(rf func() bool, retries int, pause time.Duration, msgs ...string) bool {
	start := time.Now()
	for r := 0; r < retries; r++ {
		if rf() {
			return true
		}
		time.Sleep(pause)
	}
	needed := time.Since(start)
	info := fmt.Sprintf("timeout after %v and %d retries", needed, retries)
	return a.failer.Fail(Retry, info, "successful call", msgs...)
}

// Fail always fails.
func (a *Asserts) Fail(msgs ...string) bool {
	return a.failer.Fail(Fail, nil, nil, msgs...)
}

// MakeWaitChan is a simple one-liner to create the buffered signal channel
// for the wait assertion.
func MakeWaitChan() chan interface{} {
	return make(chan interface{}, 1)
}

// MakeMultiWaitChan is a simple one-liner to create a sized buffered signal
// channel for the wait assertion.
func MakeMultiWaitChan(size int) chan interface{} {
	if size < 1 {
		size = 1
	}
	return make(chan interface{}, size)
}

//--------------------
// HELPER
//--------------------

// lowHigh transports the expected borders of a range test.
type lowHigh struct {
	low  interface{}
	high interface{}
}

// errable describes a type able to return an error state
// with the method Err().
type errable interface {
	Err() error
}

// ifaceToError converts an interface{} into an error.
func ifaceToError(obtained interface{}) error {
	err, ok := obtained.(error)
	if ok {
		return err
	}
	able, ok := obtained.(errable)
	if ok {
		if able == nil {
			return nil
		}
		return able.Err()
	}
	return err
}

// lenable describes a type able to return its length
// with the method Len().
type lenable interface {
	Len() int
}

// obexString constructs a descriptive sting matching
// to test, obtained, and expected value.
func obexString(test Test, obtained, expected interface{}) string {
	switch test {
	case True, False, Nil, NotNil, Empty, NotEmpty:
		return fmt.Sprintf("'%v'", obtained)
	case Implementor, Assignable, Unassignable:
		return fmt.Sprintf("'%v' <> '%v'", ValueDescription(obtained), ValueDescription(expected))
	case Range:
		lh := expected.(*lowHigh)
		return fmt.Sprintf("not '%v' <= '%v' <= '%v'", lh.low, obtained, lh.high)
	case Fail:
		return "fail intended"
	default:
		return fmt.Sprintf("'%v' <> '%v'", obtained, expected)
	}
}

// failString constructs a fail string for panics or
// validition errors.
func failString(test Test, obex string, msgs ...string) string {
	var out string
	if test == Fail {
		out = fmt.Sprintf("assert failed: %s", obex)
	} else {
		out = fmt.Sprintf("assert '%s' failed: %s", test, obex)
	}
	jmsgs := strings.Join(msgs, " ")
	if len(jmsgs) > 0 {
		out += " (" + jmsgs + ")"
	}
	return out
}

// EOF
