// Tideland Go Audit - Asserts
//
// Copyright (C) 2012-2021 Frank Mueller / Tideland / Oldenburg / Germany
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
	"os"
	"reflect"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

//--------------------
// TESTER
//--------------------

// isTrue checks if obtained is true.
func isTrue(obtained bool) bool {
	return obtained
}

// isNil checks if obtained is nil in a safe way.
func isNil(obtained any) bool {
	if obtained == nil {
		// Standard test.
		return true
	}
	// Some types have to be tested via reflection.
	value := reflect.ValueOf(obtained)
	kind := value.Kind()
	switch kind {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return value.IsNil()
	}
	return false
}

// isEqual checks if obtained and expected are equal.
func isEqual(obtained, expected any) bool {
	return reflect.DeepEqual(obtained, expected)
}

// isAbout checks if obtained and expected are to a given extent almost equal.
func isAbout(obtained, expected, extent float64) bool {
	if extent < 0.0 {
		extent = extent * (-1)
	}
	low := expected - extent
	high := expected + extent
	return low <= obtained && obtained <= high
}

// isInRange checks for range assertions
func isInRange(obtained, low, high any) (bool, error) {
	// First standard types.
	switch o := obtained.(type) {
	case byte:
		l, lok := low.(byte)
		h, hok := high.(byte)
		if !lok && !hok {
			return false, errors.New("low and/or high are no byte")
		}
		return l <= o && o <= h, nil
	case int:
		l, lok := low.(int)
		h, hok := high.(int)
		if !lok && !hok {
			return false, errors.New("low and/or high are no int")
		}
		return l <= o && o <= h, nil
	case float64:
		l, lok := low.(float64)
		h, hok := high.(float64)
		if !lok && !hok {
			return false, errors.New("low and/or high are no float64")
		}
		return l <= o && o <= h, nil
	case rune:
		l, lok := low.(rune)
		h, hok := high.(rune)
		if !lok && !hok {
			return false, errors.New("low and/or high are no rune")
		}
		return l <= o && o <= h, nil
	case string:
		l, lok := low.(string)
		h, hok := high.(string)
		if !lok && !hok {
			return false, errors.New("low and/or high are no string")
		}
		return l <= o && o <= h, nil
	case time.Time:
		l, lok := low.(time.Time)
		h, hok := high.(time.Time)
		if !lok && !hok {
			return false, errors.New("low and/or high are no time")
		}
		return (l.Equal(o) || l.Before(o)) &&
			(h.After(o) || h.Equal(o)), nil
	case time.Duration:
		l, lok := low.(time.Duration)
		h, hok := high.(time.Duration)
		if !lok && !hok {
			return false, errors.New("low and/or high are no duration")
		}
		return l <= o && o <= h, nil
	}
	// Now check the collection types.
	_, ol, err := hasLength(obtained, 0)
	if err != nil {
		return false, errors.New("no valid type with a length")
	}
	l, lok := low.(int)
	h, hok := high.(int)
	if !lok && !hok {
		return false, errors.New("low and/or high are no int")
	}
	return l <= ol && ol <= h, nil
}

// contains checks if the part type is matching to the full type and
// if the full data contains the part data.
func contains(part, full any) (bool, error) {
	switch fullValue := full.(type) {
	case string:
		// Content of a string.
		switch partValue := part.(type) {
		case string:
			return strings.Contains(fullValue, partValue), nil
		case []byte:
			return strings.Contains(fullValue, string(partValue)), nil
		default:
			partString := fmt.Sprintf("%v", partValue)
			return strings.Contains(fullValue, partString), nil
		}
	case []byte:
		// Content of a byte slice.
		switch partValue := part.(type) {
		case string:
			return bytes.Contains(fullValue, []byte(partValue)), nil
		case []byte:
			return bytes.Contains(fullValue, partValue), nil
		default:
			partBytes := []byte(fmt.Sprintf("%v", partValue))
			return bytes.Contains(fullValue, partBytes), nil
		}
	default:
		// Content of any array or slice, use reflection.
		value := reflect.ValueOf(full)
		kind := value.Kind()
		if kind == reflect.Array || kind == reflect.Slice {
			length := value.Len()
			for i := 0; i < length; i++ {
				current := value.Index(i)
				if reflect.DeepEqual(part, current.Interface()) {
					return true, nil
				}
			}
			return false, nil
		}
	}
	return false, errors.New("full value is no string, array, or slice")
}

// isSubstring checks if obtained is a substring of the full string.
func isSubstring(obtained, full string) bool {
	return strings.Contains(full, obtained)
}

// isCase checks if the obtained string is uppercase or lowercase.
func isCase(obtained string, upperCase bool) bool {
	if upperCase {
		return obtained == strings.ToUpper(obtained)
	}
	return obtained == strings.ToLower(obtained)
}

// isMatching checks if the obtained string matches a regular expression.
func isMatching(obtained, regex string) (bool, error) {
	return regexp.MatchString("^"+regex+"$", obtained)
}

// isImplementor checks if obtained implements the expected interface variable pointer.
func isImplementor(obtained, expected any) (bool, error) {
	obtainedValue := reflect.ValueOf(obtained)
	expectedValue := reflect.ValueOf(expected)
	if !obtainedValue.IsValid() {
		return false, fmt.Errorf("obtained value is invalid: %v", obtained)
	}
	if !expectedValue.IsValid() || expectedValue.Kind() != reflect.Ptr || expectedValue.Elem().Kind() != reflect.Interface {
		return false, fmt.Errorf("expected value is no interface variable pointer: %v", expected)
	}
	return obtainedValue.Type().Implements(expectedValue.Elem().Type()), nil
}

// isAssignable checks if the types of obtained and expected are assignable.
func isAssignable(obtained, expected any) bool {
	obtainedValue := reflect.ValueOf(obtained)
	expectedValue := reflect.ValueOf(expected)
	return obtainedValue.Type().AssignableTo(expectedValue.Type())
}

// hasLength checks the length of the obtained string, array, slice, map or channel.
func hasLength(obtained any, expected int) (bool, int, error) {
	// Check using the lenable interface.
	if ol, ok := obtained.(lenable); ok {
		l := ol.Len()
		return l == expected, l, nil
	}
	// Check for sting due to UTF-8 rune handling.
	if s, ok := obtained.(string); ok {
		l := utf8.RuneCountInString(s)
		return l == expected, l, nil
	}
	// Check the standard types.
	ov := reflect.ValueOf(obtained)
	ok := ov.Kind()
	switch ok {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice:
		l := ov.Len()
		return l == expected, l, nil
	default:
		descr := ValueDescription(obtained)
		return false, 0, fmt.Errorf("obtained %s is no array, chan, map, slice, string and does not understand Len()", descr)
	}
}

// hasPanic checks if the passed function panics.
func hasPanic(pf func(), reason any) (ok bool) {
	defer func() {
		if r := recover(); r != nil {
			// Panic, so far okay.
			if reason == nil {
				ok = true
			} else {
				if reflect.DeepEqual(r, reason) {
					ok = true
				} else {
					ok = false
				}
			}
		}
	}()
	pf()
	return false
}

// isValidPath checks if the given directory or file path exists.
func isValidPath(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

// EOF
