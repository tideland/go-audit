// Tideland Go Audit - Capture - Unit Tests
//
// Copyright (C) 2017-2020 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package capture_test

//--------------------
// IMPORTS
//--------------------

import (
	"fmt"
	"os"
	"testing"

	"tideland.dev/go/audit/asserts"
	"tideland.dev/go/audit/capture"
)

//--------------------
// TESTS
//--------------------

// TestStdout tests the capturing of writings to stdout.
func TestStdout(t *testing.T) {
	assert := asserts.NewTesting(t, asserts.FailStop)
	hello := "Hello, World!"
	cptrd := capture.Stdout(func() {
		fmt.Print(hello)
	})
	assert.Equal(cptrd.String(), hello)
	assert.Length(cptrd, len(hello))
}

// TestStderr tests the capturing of writings to stderr.
func TestStderr(t *testing.T) {
	assert := asserts.NewTesting(t, asserts.FailStop)
	ouch := "ouch"
	cptrd := capture.Stderr(func() {
		fmt.Fprint(os.Stderr, ouch)
	})
	assert.Equal(cptrd.String(), ouch)
	assert.Length(cptrd, len(ouch))
}

// TestBoth tests the capturing of writings to stdout
// and stderr.
func TestBoth(t *testing.T) {
	assert := asserts.NewTesting(t, asserts.FailStop)
	hello := "Hello, World!"
	ouch := "ouch"
	cout, cerr := capture.Both(func() {
		fmt.Fprint(os.Stdout, hello)
		fmt.Fprint(os.Stderr, ouch)
	})
	assert.Equal(cout.String(), hello)
	assert.Length(cout, len(hello))
	assert.Equal(cerr.String(), ouch)
	assert.Length(cerr, len(ouch))
}

// TestBytes tests the retrieving of captures as bytes.
func TestBytes(t *testing.T) {
	assert := asserts.NewTesting(t, asserts.FailStop)
	foo := "foo"
	boo := []byte(foo)
	cout, cerr := capture.Both(func() {
		fmt.Fprint(os.Stdout, foo)
		fmt.Fprint(os.Stderr, foo)
	})
	assert.Equal(cout.Bytes(), boo)
	assert.Equal(cerr.Bytes(), boo)
}

// TestRestore tests the restoring of os.Stdout
// and os.Stderr after capturing.
func TestRestore(t *testing.T) {
	assert := asserts.NewTesting(t, asserts.FailStop)
	foo := "foo"
	oldOut := os.Stdout
	oldErr := os.Stderr
	cout, cerr := capture.Both(func() {
		fmt.Fprint(os.Stdout, foo)
		fmt.Fprint(os.Stderr, foo)
	})
	assert.Equal(cout.String(), foo)
	assert.Length(cout, len(foo))
	assert.Equal(cerr.String(), foo)
	assert.Length(cerr, len(foo))
	assert.Equal(os.Stdout, oldOut)
	assert.Equal(os.Stderr, oldErr)
}

// EOF
