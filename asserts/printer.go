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
	"os"
	"reflect"
)

//--------------------
// TEST
//--------------------

// Test represents the test inside an assert.
type Test int

// Tests provided by the assertion.
const (
	Invalid Test = iota + 1
	True
	False
	Nil
	NotNil
	NoError
	Equal
	Different
	Contents
	About
	Range
	Substring
	Case
	Match
	ErrorMatch
	ErrorContains
	Implementor
	Assignable
	Unassignable
	Empty
	NotEmpty
	Length
	Panics
	NotPanics
	PanicsWith
	PathExists
	Wait
	WaitClosed
	WaitGroup
	WaitTested
	Retry
	Fail
	OK
)

// testNames maps the tests to their descriptive names.
var testNames = []string{
	Invalid:      "invalid",
	True:         "true",
	False:        "false",
	Nil:          "nil",
	NotNil:       "not nil",
	NoError:      "no error",
	Equal:        "equal",
	Different:    "different",
	Contents:     "contents",
	About:        "about",
	Range:        "range",
	Substring:    "substring",
	Case:         "case",
	Match:        "match",
	ErrorMatch:   "error match",
	Implementor:  "implementor",
	Assignable:   "assignable",
	Unassignable: "unassignable",
	Empty:        "empty",
	NotEmpty:     "not empty",
	Length:       "length",
	Panics:       "panics",
	NotPanics:    "not panics",
	PanicsWith:   "panics with",
	PathExists:   "path exists",
	Wait:         "wait",
	WaitClosed:   "wait closed",
	WaitGroup:    "wait group",
	WaitTested:   "wait tested",
	Retry:        "retry",
	Fail:         "fail",
}

// String implements fmt.Stringer.
func (t Test) String() string {
	if int(t) < len(testNames) {
		return testNames[t]
	}
	return "invalid"
}

//--------------------
// PRINTER
//--------------------

// Printer allows to switch between different outputs of
// the tests.
type Printer interface {
	// Logf prints a formatted logging information.
	Logf(format string, args ...interface{})

	// Errorf prints a formatted error.
	Errorf(format string, args ...interface{})
}

// wrappedPrinter wraps a type implementing the Printer
// interface..
type wrappedPrinter struct {
	printer Printer
}

// NewWrappedPrinter returns a printer using the passed Printer.
func NewWrappedPrinter(p Printer) Printer {
	return &wrappedPrinter{
		printer: p,
	}
}

// Logf implements Printer.
func (p *wrappedPrinter) Logf(format string, args ...interface{}) {
	p.printer.Logf(format, args...)
}

// Errorf implements Printer.
func (p *wrappedPrinter) Errorf(format string, args ...interface{}) {
	p.printer.Errorf(format, args...)
}

// standardPrinter uses the standard fmt package for printing.
type standardPrinter struct{}

// NewStandardPrinter creates a printer writing its output to
// stdout and stderr.
func NewStandardPrinter() Printer {
	return &standardPrinter{}
}

// Logf implements Printer.
func (p *standardPrinter) Logf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stdout, format, args...)
}

// Errorf implements Printer.
func (p *standardPrinter) Errorf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
}

// BufferedPrinter collects prints to be retrieved later via Flush().
type BufferedPrinter interface {
	Printer

	// Flush returns and resets the buffered prints.
	Flush() []string
}

// bufferedPrinter collects the prints which can be retrieved later.
type bufferedPrinter struct {
	buffer []string
}

// NewBufferedPrinter returns the buffered printer for collecting
// assertion output.
func NewBufferedPrinter() BufferedPrinter {
	return &bufferedPrinter{}
}

// Logf implements Printer.
func (p *bufferedPrinter) Logf(format string, args ...interface{}) {
	s := fmt.Sprintf("[LOG] "+format, args...)
	p.buffer = append(p.buffer, s)
}

// Errorf implements Printer.
func (p *bufferedPrinter) Errorf(format string, args ...interface{}) {
	s := fmt.Sprintf("[ERR] "+format, args...)
	p.buffer = append(p.buffer, s)
}

// Flush implements BufferedPrinter.
func (p *bufferedPrinter) Flush() []string {
	b := p.buffer
	p.buffer = nil
	return b
}

//--------------------
// HELPER
//--------------------

// ValueDescription returns a description of a value as string.
func ValueDescription(value interface{}) string {
	rvalue := reflect.ValueOf(value)
	kind := rvalue.Kind()
	switch kind {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice:
		return kind.String() + " of " + rvalue.Type().Elem().String()
	case reflect.Func:
		return kind.String() + " " + rvalue.Type().Name() + "()"
	case reflect.Interface, reflect.Struct:
		return kind.String() + " " + rvalue.Type().Name()
	case reflect.Ptr:
		return kind.String() + " to " + rvalue.Type().Elem().String()
	default:
		return kind.String()
	}
}

// TypedValue returns a value including its type.
func TypedValue(value interface{}) string {
	kind := reflect.ValueOf(value).Kind()
	switch kind {
	case reflect.String:
		return fmt.Sprintf("%q (string)", value)
	default:
		return fmt.Sprintf("%v (%s)", value, kind.String())
	}
}

// EOF
