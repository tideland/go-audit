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
	"bytes"
	"errors"
	"fmt"
	"path"
	"runtime"
	"strings"
	"sync"
	"time"
)

//--------------------
// FAILER
//--------------------

// Failer describes a type controlling how an assert
// reacts after a failure.
type Failer interface {
	// SetPrinter sets a new Printer for the Failer and
	// returns the current one, e.g. for restoring.
	SetPrinter(printer Printer) Printer

	// IncrCallstackOffset increases the callstack offset for
	// the assertion output (see Asserts) and returns a function
	// for restoring.
	IncrCallstackOffset() func()

	// Logf can be used to display useful information during testing.
	Logf(format string, args ...interface{})

	// Fail will be called if an assert fails.
	Fail(test Test, obtained, expected interface{}, msgs ...string) bool
}

// FailureDetail contains detailed information of a failure.
type FailureDetail interface {
	// TImestamp tells when the failure has happened.
	Timestamp() time.Time

	// Locations returns file name with line number and
	// function name of the failure.
	Location() (string, string)

	// Test tells which kind of test has failed.
	Test() Test

	// Error returns the failure as error.
	Error() error

	// Message return the optional test message.
	Message() string
}

// failureDetail implements the FailureDetail interface.
type failureDetail struct {
	timestamp time.Time
	location  string
	fun       string
	test      Test
	err       error
	message   string
}

// TImestamp implements the FailureDetail interface.
func (d *failureDetail) Timestamp() time.Time {
	return d.timestamp
}

// Locations implements the FailureDetail interface.
func (d *failureDetail) Location() (string, string) {
	return d.location, d.fun
}

// Test implements the FailureDetail interface.
func (d *failureDetail) Test() Test {
	return d.test
}

// Error implements the FailureDetail interface.
func (d *failureDetail) Error() error {
	return d.err
}

// Message implements the FailureDetail interface.
func (d *failureDetail) Message() string {
	return d.message
}

// Failures collects the collected failures
// of a validation assertion.
type Failures interface {
	// HasErrors returns true, if assertion failures happened.
	HasErrors() bool

	// Details returns the collected details.
	Details() []FailureDetail

	// Errors returns the so far collected errors.
	Errors() []error

	// Error returns the collected errors as one error.
	Error() error
}

//--------------------
// PANIC FAILER
//--------------------

// panicFailer reacts with a panic.
type panicFailer struct {
	printer Printer
}

// SetPrinter implements Failer.
func (f *panicFailer) SetPrinter(printer Printer) Printer {
	old := f.printer
	f.printer = printer
	return old
}

// IncrCallstackOffset implements Failer.
func (f *panicFailer) IncrCallstackOffset() func() {
	return func() {}
}

// Logf implements Failer.
func (f *panicFailer) Logf(format string, args ...interface{}) {
	f.printer.Logf(format+"\n", args...)
}

// Fail implements the Failer interface.
func (f panicFailer) Fail(test Test, obtained, expected interface{}, msgs ...string) bool {
	obex := obexString(test, obtained, expected)
	failStr := failString(test, obex, msgs...)
	f.printer.Errorf(failStr)
	panic(failStr)
}

// NewPanic creates a new Asserts instance which panics if a test fails.
func NewPanic() *Asserts {
	return New(&panicFailer{
		printer: NewStandardPrinter(),
	})
}

//--------------------
// VALIDATION FAILER
//--------------------

// validationFailer collects validation errors, e.g. when
// validating form input data.
type validationFailer struct {
	mu      sync.Mutex
	printer Printer
	offset  int
	details []FailureDetail
	errs    []error
}

// HasErrors implements Failures.
func (f *validationFailer) HasErrors() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.errs) > 0
}

// Details implements Failures.
func (f *validationFailer) Details() []FailureDetail {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.details
}

// Errors implements Failures.
func (f *validationFailer) Errors() []error {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.errs
}

// Error implements Failures.
func (f *validationFailer) Error() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	strs := []string{}
	for i, err := range f.errs {
		strs = append(strs, fmt.Sprintf("[%d] %v", i, err))
	}
	return errors.New(strings.Join(strs, " / "))
}

// SetPrinter implements Failer.
func (f *validationFailer) SetPrinter(printer Printer) Printer {
	f.mu.Lock()
	defer f.mu.Unlock()
	old := f.printer
	f.printer = printer
	return old
}

// IncrCallstackOffset implements Failer.
func (f *validationFailer) IncrCallstackOffset() func() {
	f.mu.Lock()
	defer f.mu.Unlock()
	offset := f.offset
	f.offset++
	return func() {
		f.mu.Lock()
		defer f.mu.Unlock()
		f.offset = offset
	}
}

// Logf implements Failer.
func (f *validationFailer) Logf(format string, args ...interface{}) {
	f.mu.Lock()
	defer f.mu.Unlock()
	location, fun := here(f.offset)
	prefix := fmt.Sprintf("%s %s(): ", location, fun)
	f.printer.Logf(prefix+format+"\n", args...)
}

// Fail implements Failer.
func (f *validationFailer) Fail(test Test, obtained, expected interface{}, msgs ...string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	location, fun := here(f.offset)
	obex := obexString(test, obtained, expected)
	err := errors.New(failString(test, obex, msgs...))
	detail := &failureDetail{
		timestamp: time.Now(),
		location:  location,
		fun:       fun,
		test:      test,
		err:       err,
		message:   strings.Join(msgs, " "),
	}
	f.details = append(f.details, detail)
	f.errs = append(f.errs, err)
	return false
}

// NewValidation creates a new Asserts instance which collections
// validation failures. The returned Failures instance allows to test an access
// them.
func NewValidation() (*Asserts, Failures) {
	vf := &validationFailer{
		printer: NewStandardPrinter(),
		offset:  4,
		details: []FailureDetail{},
		errs:    []error{},
	}
	return New(vf), vf
}

//--------------------
// TESTING FAILER
//--------------------

// Failable allows an assertion to signal a fail to an external instance
// like testing.T or testing.B.
type Failable interface {
	Fail()
	FailNow()
}

// FailMode defines how to react on failing test asserts.
type FailMode int

// Fail modes for test failer.
const (
	NoFailing    FailMode = 0 // NoFailing simply logs a failing.
	FailContinue FailMode = 1 // FailContinue logs a failing and calls Failable.Fail().
	FailStop     FailMode = 2 // FailStop logs a failing and calls Failable.FailNow().
)

// testingFailer works together with the testing package of Go and
// may signal the fail to it.
type testingFailer struct {
	mu       sync.Mutex
	printer  Printer
	failable Failable
	offset   int
	mode     FailMode
}

// SetPrinter implements Failer.
func (f *testingFailer) SetPrinter(printer Printer) Printer {
	f.mu.Lock()
	defer f.mu.Unlock()
	old := f.printer
	f.printer = printer
	return old
}

// IncrCallstackOffset implements Failer.
func (f *testingFailer) IncrCallstackOffset() func() {
	f.mu.Lock()
	defer f.mu.Unlock()
	offset := f.offset
	f.offset++
	return func() {
		f.mu.Lock()
		defer f.mu.Unlock()
		f.offset = offset
	}
}

// Logf implements Failer.
func (f *testingFailer) Logf(format string, args ...interface{}) {
	f.mu.Lock()
	defer f.mu.Unlock()
	location, fun := here(f.offset)
	prefix := fmt.Sprintf("%s %s(): ", location, fun)
	f.printer.Logf(prefix+format+"\n", args...)
}

// Fail implements Failer.
func (f *testingFailer) Fail(test Test, obtained, expected interface{}, msgs ...string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	location, fun := here(f.offset)
	buffer := &bytes.Buffer{}

	if test == Fail {
		fmt.Fprintf(buffer, "%s assert in %s() failed {", location, fun)
	} else {
		fmt.Fprintf(buffer, "%s assert '%s' in %s() failed {", location, test, fun)
	}
	switch test {
	case True, False, Nil, NotNil, NoError, Empty, NotEmpty, Panics:
		fmt.Fprintf(buffer, "got: %v", obtained)
	case Implementor, Assignable, Unassignable:
		fmt.Fprintf(buffer, "got: %v, want: %v", ValueDescription(obtained), ValueDescription(expected))
	case Contains, NotContains:
		switch typedObtained := obtained.(type) {
		case string:
			fmt.Fprintf(buffer, "part: %s, full: %s", typedObtained, expected)
		default:
			fmt.Fprintf(buffer, "part: %v, full: %v", obtained, expected)
		}
	case Fail:
	default:
		fmt.Fprintf(buffer, "got: %v, want: %v", TypedValue(obtained), TypedValue(expected))
	}
	if len(msgs) > 0 {
		if buffer.Bytes()[buffer.Len()-1] != byte('{') {
			fmt.Fprintf(buffer, ", ")
		}
		fmt.Fprintf(buffer, "info: %s", strings.Join(msgs, " "))
	}
	fmt.Fprintf(buffer, "}\n")

	switch f.mode {
	case NoFailing:
		f.printer.Logf(buffer.String())
	case FailContinue:
		f.printer.Errorf(buffer.String())
		f.failable.Fail()
	case FailStop:
		f.printer.Errorf(buffer.String())
		f.failable.FailNow()
	}
	return false
}

// NewTesting creates a new Asserts instance for use with the testing
// package. The *testing.T has to be passed as failable, the argument.
// shallFail controls if a failing assertion also lets fail the Go test.
func NewTesting(f Failable, mode FailMode) *Asserts {
	p, ok := f.(Printer)
	if ok {
		p = NewWrappedPrinter(p)
	} else {
		p = NewStandardPrinter()
	}
	return New(&testingFailer{
		printer:  p,
		failable: f,
		offset:   4,
		mode:     mode,
	})
}

//--------------------
// HELPERS
//--------------------

// here returns the location at the given offset.
func here(offset int) (string, string) {
	// Retrieve program counters.
	pcs := make([]uintptr, 1)
	n := runtime.Callers(offset, pcs)
	if n == 0 {
		return "", ""
	}
	pcs = pcs[:n]
	// Build ID based on program counters.
	frames := runtime.CallersFrames(pcs)
	for {
		frame, more := frames.Next()
		_, fun := path.Split(frame.Function)
		parts := strings.Split(fun, ".")
		fun = strings.Join(parts[1:], ".")
		_, file := path.Split(frame.File)
		location := fmt.Sprintf("%s:%d:0:", file, frame.Line)
		if !more {
			return location, fun
		}
	}
}

// EOF
