// Tideland Go Audit - Asserts - Unit Tests
//
// Copyright (C) 2012-2020 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package asserts_test

//--------------------
// IMPORTS
//--------------------

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"tideland.dev/go/audit/asserts"
)

//--------------------
// TESTS
//--------------------

// withErr helps testing error returning types.
type withErr struct {
	err error
}

// Err implements the needed interface for returning errors.
func (we withErr) Err() error {
	return we.err
}

// TestAssertOK tests the OK() assertion.
func TestAssertOK(t *testing.T) {
	successfulAssert := successfulAsserts(t)
	failingAssert := failingAsserts(t)

	var errA error
	var errB withErr = withErr{errA}
	var errC = errors.New("ouch")
	var errD withErr = withErr{errC}

	successfulAssert.OK(true, "OK true should not fail")
	successfulAssert.OK(errA, "OK nil error should not fail")
	successfulAssert.OK(errB, "OK nil Err() should not fail")
	successfulAssert.OK(0, "OK 0 should not fail")
	successfulAssert.OK("", "OK '' should not fail")
	failingAssert.OK(false, "OK false should fail and be logged")
	failingAssert.OK(errC, "OK ouch error should fail and be logged")
	failingAssert.OK(errD, "OK ouch Err() should fail and be logged")
	failingAssert.OK(1, "OK 1 should fail and be logged")
	failingAssert.OK("ouch", "OK 'ouch' should fail and be logged")
}

// TestAssertTrue tests the True() assertion.
func TestAssertTrue(t *testing.T) {
	successfulAssert := successfulAsserts(t)
	failingAssert := failingAsserts(t)

	successfulAssert.True(true, "should not fail")
	failingAssert.True(false, "should fail and be logged")
}

// TestAssertFalse tests the False() assertion.
func TestAssertFalse(t *testing.T) {
	successfulAssert := successfulAsserts(t)
	failingAssert := failingAsserts(t)

	successfulAssert.False(false, "should not fail")
	failingAssert.False(true, "should fail and be logged")
}

// TestAssertNil tests the Nil() assertion.
func TestAssertNil(t *testing.T) {
	successfulAssert := successfulAsserts(t)
	failingAssert := failingAsserts(t)

	successfulAssert.Nil(nil, "should not fail")
	failingAssert.Nil("not nil", "should fail and be logged")
}

// TestAssertNotNil tests the NotNil() assertion.
func TestAssertNotNil(t *testing.T) {
	successfulAssert := successfulAsserts(t)
	failingAssert := failingAsserts(t)

	successfulAssert.NotNil("not nil", "should not fail")
	failingAssert.NotNil(nil, "should fail and be logged")
}

// TestAssertNoError tests the NoError() assertion.
func TestAssertNoError(t *testing.T) {
	successfulAssert := successfulAsserts(t)
	failingAssert := failingAsserts(t)

	var errA error
	var errB withErr = withErr{errA}
	var errC = errors.New("ouch")
	var errD withErr = withErr{errC}

	successfulAssert.NoError(errA, "should not fail")
	successfulAssert.NoError(errB, "should not fail")
	failingAssert.NoError(errC, "should fail and be logged")
	failingAssert.NoError(errD, "should fail and be logged")
}

// TestAssertEqual tests the Equal() assertion.
func TestAssertEqual(t *testing.T) {
	successfulAssert := successfulAsserts(t)
	failingAssert := failingAsserts(t)

	m := map[string]int{"one": 1, "two": 2, "three": 3}
	now := time.Now()
	nowStr := now.Format(time.RFC3339Nano)
	nowParsedA, errA := time.Parse(time.RFC3339Nano, nowStr)
	nowParsedB, errB := time.Parse(time.RFC3339Nano, nowStr)

	successfulAssert.Nil(errA, "should not fail")
	successfulAssert.Nil(errB, "should not fail")
	successfulAssert.Equal(nowParsedA, nowParsedB, "should not fail")
	successfulAssert.Equal(nil, nil, "should not fail")
	successfulAssert.Equal(true, true, "should not fail")
	successfulAssert.Equal(1, 1, "should not fail")
	successfulAssert.Equal("foo", "foo", "should not fail")
	successfulAssert.Equal(map[string]int{"one": 1, "three": 3, "two": 2}, m, "should not fail")
	failingAssert.Equal("one", 1, "should fail and be logged")
	failingAssert.Equal("two", "2", "should fail and be logged")
}

// TestAssertDifferent tests the Different() assertion.
func TestAssertDifferent(t *testing.T) {
	successfulAssert := successfulAsserts(t)
	failingAssert := failingAsserts(t)

	m := map[string]int{"one": 1, "two": 2, "three": 3}

	successfulAssert.Different(nil, "nil", "should not fail")
	successfulAssert.Different("true", true, "should not fail")
	successfulAssert.Different(1, 2, "should not fail")
	successfulAssert.Different("foo", "bar", "should not fail")
	successfulAssert.Different(map[string]int{"three": 3, "two": 2}, m, "should not fail")
	failingAssert.Different("one", "one", "should fail and be logged")
	failingAssert.Different(2, 2, "should fail and be logged")
}

// TestAssertAbout tests the About() assertion.
func TestAssertAbout(t *testing.T) {
	successfulAssert := successfulAsserts(t)
	failingAssert := failingAsserts(t)

	successfulAssert.About(1.0, 1.0, 0.0, "equal, no extend")
	successfulAssert.About(1.0, 1.0, 0.1, "equal, little extend")
	successfulAssert.About(0.9, 1.0, 0.1, "different, within bounds of extent")
	successfulAssert.About(1.1, 1.0, 0.1, "different, within bounds of extent")
	failingAssert.About(0.8, 1.0, 0.1, "different, out of bounds of extent")
	failingAssert.About(1.2, 1.0, 0.1, "different, out of bounds of extent")
}

// TestAssertRange tests the Range() assertion.
func TestAssertRange(t *testing.T) {
	successfulAssert := successfulAsserts(t)
	failingAssert := failingAsserts(t)
	now := time.Now()

	successfulAssert.Range(byte(9), byte(1), byte(22), "byte in range")
	successfulAssert.Range(9, 1, 22, "int in range")
	successfulAssert.Range(9.0, 1.0, 22.0, "float64 in range")
	successfulAssert.Range('f', 'a', 'z', "rune in range")
	successfulAssert.Range("foo", "a", "zzzzz", "string in range")
	successfulAssert.Range(now, now.Add(-time.Hour), now.Add(time.Hour), "time in range")
	successfulAssert.Range(time.Minute, time.Second, time.Hour, "duration in range")
	successfulAssert.Range([]int{1, 2, 3}, 1, 10, "slice length in range")
	successfulAssert.Range([3]int{1, 2, 3}, 1, 10, "array length in range")
	successfulAssert.Range(map[int]int{3: 1, 2: 2, 1: 3}, 1, 10, "map length in range")
	failingAssert.Range(byte(1), byte(10), byte(20), "byte out of range")
	failingAssert.Range(1, 10, 20, "int out of range")
	failingAssert.Range(1.0, 10.0, 20.0, "float64 out of range")
	failingAssert.Range('a', 'x', 'z', "rune out of range")
	failingAssert.Range("aaa", "uuuuu", "zzzzz", "string out of range")
	failingAssert.Range(now, now.Add(time.Minute), now.Add(time.Hour), "time out of range")
	failingAssert.Range(time.Second, time.Minute, time.Hour, "duration in range")
	failingAssert.Range([]int{1, 2, 3}, 5, 10, "slice length out of range")
	failingAssert.Range([3]int{1, 2, 3}, 5, 10, "array length out of range")
	failingAssert.Range(map[int]int{3: 1, 2: 2, 1: 3}, 5, 10, "map length out of range")
}

// TestAssertContents tests the Contents() assertion.
func TestAssertContents(t *testing.T) {
	successfulAssert := successfulAsserts(t)
	failingAssert := failingAsserts(t)

	successfulAssert.Contents("bar", "foobarbaz")
	successfulAssert.Contents(4711, []int{1, 2, 3, 4711, 5, 6, 7, 8, 9})
	failingAssert.Contents(4711, "12345-4711-67890")
	failingAssert.Contents(4711, "foo")
	failingAssert.Contents(4711, []interface{}{1, "2", 3, "4711", 5, 6, 7, 8, 9})
	successfulAssert.Contents("4711", []interface{}{1, "2", 3, "4711", 5, 6, 7, 8, 9})
	failingAssert.Contents("foobar", []byte("the quick brown fox jumps over the lazy dog"))

	successfulAssert.NotContents("yadda", "foobarbaz")
	successfulAssert.NotContents(123, []int{1, 2, 3, 4711, 5, 6, 7, 8, 9})
	failingAssert.NotContents("4711", "12345-4711-67890")
	failingAssert.NotContents("oba", "foobar")
	failingAssert.NotContents("4711", []interface{}{1, "2", 3, "4711", 5, 6, 7, 8, 9})
	successfulAssert.NotContents(4711, []interface{}{1, "2", 3, "4711", 5, 6, 7, 8, 9})
	failingAssert.NotContents("fox", []byte("the quick brown fox jumps over the lazy dog"))
}

// TestAssertContentsPrint test the visualization of failing content tests.
func TestAssertContentsPrint(t *testing.T) {
	assert := asserts.NewTesting(t, asserts.NoFailing)

	assert.Logf("printing of failing content tests")
	assert.Contents("foobar", []byte("the quick brown fox jumps over the lazy dog"), "test fails but passes, just visualization")
	assert.Contents([]byte("foobar"), []byte("the quick brown ..."), "test fails but passes, just visualization")
}

// TestOffsetPrint test the correct visualization when printing
// with offset.
func TestOffsetPrint(t *testing.T) {
	assert := asserts.NewTesting(t, asserts.NoFailing)

	// Log should reference line below (174).
	failWithOffset(assert, "174")
}

// TestAssertSubstring tests the Substring() assertion.
func TestAssertSubstring(t *testing.T) {
	successfulAssert := successfulAsserts(t)
	failingAssert := failingAsserts(t)

	successfulAssert.Substring("is assert", "this is assert test", "should not fail")
	successfulAssert.Substring("test", "this is 1 test", "should not fail")
	failingAssert.Substring("foo", "this is assert test", "should fail and be logged")
	failingAssert.Substring("this  is  assert  test", "this is assert test", "should fail and be logged")
}

// TestAssertCase tests the Case() assertion.
func TestAssertCase(t *testing.T) {
	successfulAssert := successfulAsserts(t)
	failingAssert := failingAsserts(t)

	successfulAssert.Case("FOO", true, "is all uppercase")
	successfulAssert.Case("foo", false, "is all lowercase")
	failingAssert.Case("Foo", true, "is mixed case")
	failingAssert.Case("Foo", false, "is mixed case")
}

// TestAssertMatch tests the Match() assertion.
func TestAssertMatch(t *testing.T) {
	successfulAssert := successfulAsserts(t)
	failingAssert := failingAsserts(t)

	successfulAssert.Match("this is assert test", "this.*test", "should not fail")
	successfulAssert.Match("this is 1 test", "this is [0-9] test", "should not fail")
	failingAssert.Match("this is assert test", "foo", "should fail and be logged")
	failingAssert.Match("this is assert test", "this*test", "should fail and be logged")
}

// TestAssertErrorMatch tests the ErrorMatch() assertion.
func TestAssertErrorMatch(t *testing.T) {
	successfulAssert := successfulAsserts(t)
	failingAssert := failingAsserts(t)

	errA := errors.New("oops, an error")
	errB := withErr{errA}

	successfulAssert.ErrorMatch(errA, "oops, an error", "should not fail")
	successfulAssert.ErrorMatch(errA, "oops,.*", "should not fail")
	successfulAssert.ErrorMatch(errB, "oops,.*", "should not fail")
	failingAssert.ErrorMatch(errA, "foo", "should fail and be logged")
}

// TestAssertErrorContains tests the ErrorContains() assertion.
func TestAssertErrorContains(t *testing.T) {
	successfulAssert := successfulAsserts(t)
	failingAssert := failingAsserts(t)

	errA := errors.New("oops, an error")
	errB := withErr{errA}

	successfulAssert.ErrorContains(errA, "an error", "should not fail")
	successfulAssert.ErrorContains(errB, "an error", "should not fail")
	failingAssert.ErrorContains(errA, "foo", "should fail and be logged")
}

// TestAssertImplementor tests the Implementor() assertion.
func TestAssertImplementor(t *testing.T) {
	successfulAssert := successfulAsserts(t)
	failingAssert := failingAsserts(t)

	var err error
	var w io.Writer

	successfulAssert.Implementor(errors.New("error test"), &err, "should not fail")
	failingAssert.Implementor("string test", &err, "should fail and be logged")
	failingAssert.Implementor(errors.New("error test"), &w, "should fail and be logged")
}

// TestAssertAssignable tests the Assignable() assertion.
func TestAssertAssignable(t *testing.T) {
	successfulAssert := successfulAsserts(t)
	failingAssert := failingAsserts(t)

	successfulAssert.Assignable(1, 5, "should not fail")
	failingAssert.Assignable("one", 5, "should fail and be logged")
}

// TestAssertUnassignable tests the Unassignable() assertion.
func TestAssertUnassignable(t *testing.T) {
	successfulAssert := successfulAsserts(t)
	failingAssert := failingAsserts(t)

	successfulAssert.Unassignable("one", 5, "should not fail")
	failingAssert.Unassignable(1, 5, "should fail and be logged")
}

// TestAssertEmpty tests the Empty() assertion.
func TestAssertEmpty(t *testing.T) {
	successfulAssert := successfulAsserts(t)
	failingAssert := failingAsserts(t)

	successfulAssert.Empty("", "should not fail")
	successfulAssert.Empty([]bool{}, "should also not fail")
	failingAssert.Empty("not empty", "should fail and be logged")
	failingAssert.Empty([3]int{1, 2, 3}, "should also fail and be logged")
	failingAssert.Empty(true, "illegal type has to fail")
}

// TestAssertNotEmpty tests the NotEmpty() assertion.
func TestAssertNotEmpty(t *testing.T) {
	successfulAssert := successfulAsserts(t)
	failingAssert := failingAsserts(t)

	successfulAssert.NotEmpty("not empty", "should not fail")
	successfulAssert.NotEmpty([3]int{1, 2, 3}, "should also not fail")
	failingAssert.NotEmpty("", "should fail and be logged")
	failingAssert.NotEmpty([]int{}, "should also fail and be logged")
	failingAssert.NotEmpty(true, "illegal type has to fail")
}

// TestAssertLength tests the Length() assertion.
func TestAssertLength(t *testing.T) {
	successfulAssert := successfulAsserts(t)
	failingAssert := failingAsserts(t)

	successfulAssert.Length("", 0, "should not fail")
	successfulAssert.Length([]bool{true, false}, 2, "should also not fail")
	failingAssert.Length("not empty", 0, "should fail and be logged")
	failingAssert.Length([3]int{1, 2, 3}, 10, "should also fail and be logged")
	failingAssert.Length(true, 1, "illegal type has to fail")
}

// TestAssertPanics tests the Panics() assertion.
func TestAssertPanics(t *testing.T) {
	successfulAssert := successfulAsserts(t)
	failingAssert := failingAsserts(t)

	successfulAssert.Panics(func() { panic("ouch") }, "should panic")
	failingAssert.Panics(func() { _ = 1 + 1 }, "should not panic")
}

// TestAssertWait tests the wait testing.
func TestAssertWait(t *testing.T) {
	successfulAssert := successfulAsserts(t)
	failingAssert := failingAsserts(t)

	waitc := asserts.MakeWaitChan()
	go func() {
		time.Sleep(50 * time.Millisecond)
		waitc <- true
	}()
	successfulAssert.Wait(waitc, true, 100*time.Millisecond, "should be true")

	go func() {
		time.Sleep(50 * time.Millisecond)
		waitc <- false
	}()
	failingAssert.Wait(waitc, true, 100*time.Millisecond, "should be false")

	go func() {
		time.Sleep(200 * time.Millisecond)
		waitc <- true
	}()
	failingAssert.Wait(waitc, true, 100*time.Millisecond, "should timeout")
}

// TestAssertWaitClosed tests the wait closed testing.
func TestAssertWaitClosed(t *testing.T) {
	successfulAssert := successfulAsserts(t)
	failingAssert := failingAsserts(t)

	waitc := asserts.MakeWaitChan()
	go func() {
		time.Sleep(50 * time.Millisecond)
		close(waitc)
	}()
	successfulAssert.WaitClosed(waitc, 100*time.Millisecond, "should be true")

	waitc = asserts.MakeWaitChan()
	go func() {
		time.Sleep(500 * time.Millisecond)
		close(waitc)
	}()
	failingAssert.WaitClosed(waitc, 100*time.Millisecond, "should timeout")
}

// TestAssertWaitGroup tests the wait group testing.
func TestAssertWaitGroup(t *testing.T) {
	successfulAssert := successfulAsserts(t)
	failingAssert := failingAsserts(t)

	var wg sync.WaitGroup

	wg.Add(5)
	go func() {
		for i := 0; i < 5; i++ {
			time.Sleep(50 * time.Millisecond)
			wg.Done()
		}
	}()
	successfulAssert.WaitGroup(&wg, 500*time.Millisecond, "should be done")

	wg.Add(5)
	go func() {
		for i := 0; i < 5; i++ {
			time.Sleep(50 * time.Millisecond)
			wg.Done()
		}
	}()
	failingAssert.WaitGroup(&wg, 200*time.Millisecond, "should timeout")
}

// TestAssertWaitTested tests the wait tested testing.
func TestAssertWaitTested(t *testing.T) {
	successfulAssert := successfulAsserts(t)
	failingAssert := failingAsserts(t)
	tester := func(v interface{}) error {
		b, ok := v.(bool)
		if !ok || b == false {
			return errors.New("illegal value")
		}
		return nil
	}

	waitc := asserts.MakeWaitChan()
	go func() {
		time.Sleep(50 * time.Millisecond)
		waitc <- true
	}()
	successfulAssert.WaitTested(waitc, tester, 100*time.Millisecond, "should be true")

	go func() {
		time.Sleep(50 * time.Millisecond)
		waitc <- false
	}()
	failingAssert.WaitTested(waitc, tester, 100*time.Millisecond, "should be false")

	go func() {
		time.Sleep(200 * time.Millisecond)
		waitc <- true
	}()
	failingAssert.WaitTested(waitc, tester, 100*time.Millisecond, "should timeout")
}

// TestAssertRetry tests the retry testing.
func TestAssertRetry(t *testing.T) {
	successfulAssert := successfulAsserts(t)
	failingAssert := failingAsserts(t)

	i := 0
	successfulAssert.Retry(func() bool {
		i++
		return i == 5
	}, 10, 10*time.Millisecond, "should succeed")

	failingAssert.Retry(func() bool { return false }, 10, 10*time.Millisecond, "should fail")
}

// TestAssertPathExists tests the PathExists() assertion.
func TestAssertPathExists(t *testing.T) {
	successfulAssert := successfulAsserts(t)
	failingAssert := failingAsserts(t)

	dir := filepath.Join(os.TempDir(), "assert-path-exists")
	err := os.Mkdir(dir, 0700)
	successfulAssert.Nil(err)
	defer func() {
		err = os.RemoveAll(dir)
		successfulAssert.Nil(err)
	}()

	successfulAssert.PathExists(dir, "temporary directory exists")
	failingAssert.PathExists("/this/path/will/hopefully/not/exist", "illegal path")
}

// TestAssertFail tests the fail testing.
func TestAssertFail(t *testing.T) {
	failingAssert := failingAsserts(t)

	failingAssert.Fail("this should fail")
}

// TestTestingAssertion tests the testing assertion.
func TestTestingAssertion(t *testing.T) {
	assert := asserts.NewTesting(t, asserts.NoFailing)
	foo := func() {}
	bar := 4711

	assert.Assignable(47, 11, "should not fail")
	assert.Assignable(foo, bar, "should fail (but not the test)")
	assert.Assignable(foo, bar)
	assert.Assignable(foo, bar, "this", "should", "fail", "too")
}

// TestPanicAssertion tests if the panic assertions panic when they fail.
func TestPanicAssert(t *testing.T) {
	defer func() {
		if err := recover(); err != nil {
			t.Logf("panic worked: '%v'", err)
		}
	}()

	assert := asserts.NewPanic()
	foo := func() {}

	assert.Assignable(47, 11, "should not fail")
	assert.Assignable(47, foo, "should fail")

	t.Errorf("should not be reached")
}

// TestValidationAssertion test the validation of data.
func TestValidationAssertion(t *testing.T) {
	assert, failures := asserts.NewValidation()

	assert.True(true, "should not fail")
	assert.True(false, "should fail")
	assert.Equal(1, 2, "should fail")

	if !failures.HasErrors() {
		t.Errorf("should have errors")
	}
	if len(failures.Errors()) != 2 {
		t.Errorf("wrong number of errors: %v", failures.Error())
	}

	if len(failures.Details()) != 2 {
		t.Errorf("wrong number of details")
	}
	details := failures.Details()
	location, fun := details[0].Location()
	tt := details[0].Test()
	if location != "asserts_test.go:546:0:" || fun != "TestValidationAssertion" {
		t.Errorf("wrong location %q or function %q of first detail", location, fun)
	}
	if tt != asserts.True {
		t.Errorf("wrong test type of first detail: %v", tt)
	}
	location, fun = details[1].Location()
	tt = details[1].Test()
	if location != "asserts_test.go:547:0:" || fun != "TestValidationAssertion" {
		t.Errorf("wrong location %q or function %q of second detail", location, fun)
	}
	if tt != asserts.Equal {
		t.Errorf("wrong test type of second detail: %v", tt)
	}
}

// TestSetFailable tests the setting of the failable
// to the one of a sub-test.
func TestSetFailable(t *testing.T) {
	successfulAssert := successfulAsserts(t)
	failingAssert := failingAsserts(t)

	t.Run("success", func(t *testing.T) {
		defer successfulAssert.SetFailable(t)()
		successfulAssert.True(true)
	})

	t.Run("fail", func(t *testing.T) {
		defer failingAssert.SetFailable(t)()
		failingAssert.True(false)
	})
}

// TestSetPrinter tests the chaning of the printer.
func TestSetPrinter(t *testing.T) {
	assert := asserts.NewTesting(t, asserts.NoFailing)

	// Must not fail.
	assert.Logf("first %d %s", 1, "(a)")
	assert.Logf("second")

	// Collect in buffer.
	bp := asserts.NewBufferedPrinter()
	assert.SetPrinter(bp)
	assert.Logf("first")
	assert.Logf("second")
	assert.Logf("third")

	b := bp.Flush()
	assert.Length(b, 3)
	assert.Contents(b[0], "first")
	assert.Contents(b[1], "second")
	assert.Contents(b[2], "third")
	b = bp.Flush()
	assert.Length(b, 0)
}

//--------------------
// META FAILER
//--------------------

type metaFailer struct {
	t    *testing.T
	fail bool
}

func (f *metaFailer) SetPrinter(printer asserts.Printer) asserts.Printer {
	return printer
}

func (f *metaFailer) IncrCallstackOffset() func() {
	return func() {}
}

func (f *metaFailer) Logf(format string, args ...interface{}) {
	f.t.Logf(format, args...)
}

func (f *metaFailer) Fail(test asserts.Test, obtained, expected interface{}, msgs ...string) bool {
	msg := strings.Join(msgs, " ")
	if msg != "" {
		msg = " [" + msg + "]"
	}
	format := "testing assert %q failed: '%v' (%v) <> '%v' (%v)" + msg
	obtainedVD := asserts.ValueDescription(obtained)
	expectedVD := asserts.ValueDescription(expected)
	f.Logf(format, test, obtained, obtainedVD, expected, expectedVD)
	if f.fail {
		f.t.FailNow()
	}
	return f.fail
}

//--------------------
// HELPER
//--------------------

// failWithOffset checks the offset increment.
func failWithOffset(assert *asserts.Asserts, line string) {
	restore := assert.IncrCallstackOffset()
	defer restore()

	assert.Fail("should fail referencing line " + line)
}

// successfulAsserts returns an Asserts insrance which doesn't expect a failing.
func successfulAsserts(t *testing.T) *asserts.Asserts {
	return asserts.New(&metaFailer{t, true})
}

// failingAsserts returns an Asserts instance which only logs a failing but doesn't fail.
func failingAsserts(t *testing.T) *asserts.Asserts {
	return asserts.New(&metaFailer{t, false})
}

// EOF
