// Tideland Go Audit - Asserts
//
// Copyright (C) 2012-2023 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

// Package asserts helps writing convenient and powerful unit tests. One part of
// those are assertions to compare expected and obtained values. Additional text output
// for failing tests can be added.
//
// In the beginning of a test function a new assertion instance is created with:
//
//	assert := asserts.NewTesting(t, shallFail)
//
// Inside the test an assert looks like:
//
//	assert.Equal(obtained, expected, "obtained value has to be like expected")
//
// If shallFail is set to true a failing assert also lets fail the Go test.
// Otherwise the failing is printed but the tests continue.
package asserts // import "tideland.dev/go/audit/asserts"

// EOF
