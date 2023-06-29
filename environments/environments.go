// Tideland Go Audit - Environments
//
// Copyright (C) 2012-2023 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package environments // import "tideland.dev/go/audit/environments"

//--------------------
// IMPORTS
//--------------------

import (
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"

	"tideland.dev/go/audit/asserts"
)

//--------------------
// TEMPDIR
//--------------------

// TempDir represents a temporary directory and possible subdirectories
// for testing purposes. It simply is created with
//
//	assert := asserts.NewTesting(t, asserts.FailContinue)
//	td := environments.NewTempDir(assert)
//	defer td.Restore()
//
//	tdName := td.String()
//	subName:= td.Mkdir("my", "sub", "directory")
//
// The deferred Restore() removes the temporary directory with all
// contents.
type TempDir struct {
	assert *asserts.Asserts
	dir    string
}

// NewTempDir creates a new temporary directory usable for direct
// usage or further subdirectories.
func NewTempDir(assert *asserts.Asserts) *TempDir {
	id := make([]byte, 8)
	td := &TempDir{
		assert: assert,
	}
	for i := 0; i < 256; i++ {
		_, err := rand.Read(id[:])
		td.assert.Nil(err)
		dir := filepath.Join(os.TempDir(), fmt.Sprintf("goaudit-%x", id))
		if err = os.Mkdir(dir, 0700); err == nil {
			td.dir = dir
			break
		}
		if td.dir == "" {
			msg := fmt.Sprintf("cannot create temporary directory %q: %v", td.dir, err)
			td.assert.Fail(msg)
			return nil
		}
	}
	return td
}

// Restore deletes the temporary directory and all contents.
func (td *TempDir) Restore() {
	err := os.RemoveAll(td.dir)
	if err != nil {
		msg := fmt.Sprintf("cannot remove temporary directory %q: %v", td.dir, err)
		td.assert.Fail(msg)
	}
}

// Mkdir creates a potentially nested directory inside the
// temporary directory.
func (td *TempDir) Mkdir(name ...string) string {
	innerName := filepath.Join(name...)
	fullName := filepath.Join(td.dir, innerName)
	if err := os.MkdirAll(fullName, 0700); err != nil {
		msg := fmt.Sprintf("cannot create nested temporary directory %q: %v", fullName, err)
		td.assert.Fail(msg)
	}
	return fullName
}

// String returns the temporary directory.
func (td *TempDir) String() string {
	return td.dir
}

//--------------------
// VARIABLES
//--------------------

// Variables allows to change and restore environment variables. The
// same variable can be set multiple times. Simply do
//
//	assert := asserts.NewTesting(t, asserts.FailContinue)
//	ev := environments.NewVariables(assert)
//	defer ev.Restore()
//
//	ev.Set("MY_VAR", myValue)
//
//	...
//
//	ev.Set("MY_VAR", anotherValue)
//
// The deferred Restore() resets to the original values.
type Variables struct {
	assert *asserts.Asserts
	vars   map[string]string
}

// NewVariables create a new changer for environment variables.
func NewVariables(assert *asserts.Asserts) *Variables {
	v := &Variables{
		assert: assert,
		vars:   make(map[string]string),
	}
	return v
}

// Restore resets all changed environment variables
func (v *Variables) Restore() {
	for key, value := range v.vars {
		if err := os.Setenv(key, value); err != nil {
			msg := fmt.Sprintf("cannot reset environment variable %q: %v", key, err)
			v.assert.Fail(msg)
		}
	}
}

// Set sets an environment variable to a new value.
func (v *Variables) Set(key, value string) {
	ov := os.Getenv(key)
	_, ok := v.vars[key]
	if !ok {
		v.vars[key] = ov
	}
	if err := os.Setenv(key, value); err != nil {
		msg := fmt.Sprintf("cannot set environment variable %q: %v", key, err)
		v.assert.Fail(msg)
	}
}

// Unset unsets an environment variable.
func (v *Variables) Unset(key string) {
	ov := os.Getenv(key)
	_, ok := v.vars[key]
	if !ok {
		v.vars[key] = ov
	}
	if err := os.Unsetenv(key); err != nil {
		msg := fmt.Sprintf("cannot unset environment variable %q: %v", key, err)
		v.assert.Fail(msg)
	}
}

// EOF
