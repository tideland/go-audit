// Tideland Go Audit - Environments - Unit Tests
//
// Copyright (C) 2012-2020 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package environments_test

//--------------------
// IMPORTS
//--------------------

import (
	"net/http"
	"testing"

	"tideland.dev/go/audit/asserts"
	"tideland.dev/go/audit/environments"
)

//--------------------
// TESTS
//--------------------

// TestSimpleRequests tests simple requests to individual handlers.
func TestSimpleRequests(t *testing.T) {
	assert := asserts.NewTesting(t, asserts.FailStop)
	wa := StartWebAsserter(assert)
	defer wa.Close()

	tests := []struct {
		method      string
		path        string
		statusCode  int
		contentType string
		body        string
	}{
		{
			method:      http.MethodGet,
			path:        "/hello/world",
			statusCode:  http.StatusOK,
			contentType: environments.ContentTypePlain,
			body:        "Hello, World!",
		}, {
			method:      http.MethodGet,
			path:        "/hello/tester",
			statusCode:  http.StatusOK,
			contentType: environments.ContentTypePlain,
			body:        "Hello, Tester!",
		}, {
			method:      http.MethodPost,
			path:        "/hello/postman",
			statusCode:  http.StatusOK,
			contentType: environments.ContentTypePlain,
			body:        "Hello, Postman!",
		}, {
			method:     http.MethodOptions,
			path:       "/path/does/not/exist",
			statusCode: http.StatusNotFound,
			body:       "404 page not found",
		},
	}
	for i, test := range tests {
		assert.Logf("test case #%d: %s %s", i, test.method, test.path)
		wreq := wa.CreateRequest(test.method, test.path)
		wresp := wreq.Do()
		wresp.AssertStatusCodeEquals(test.statusCode)
		if test.contentType != "" {
			wresp.Header().AssertKeyValueEquals(environments.HeaderContentType, test.contentType)
		}
		if test.body != "" {
			wresp.AssertBodyMatches(test.body)
		}
	}
}

// TestHeaderCookies tests access to header and cookies.
func TestHeaderCookies(t *testing.T) {
	assert := asserts.NewTesting(t, asserts.FailStop)
	wa := StartWebAsserter(assert)
	defer wa.Close()

	tests := []struct {
		path   string
		header string
		cookie string
	}{
		{
			path:   "/header/cookies",
			header: "foo",
			cookie: "12345",
		}, {
			path:   "/header/cookies",
			header: "bar",
			cookie: "98765",
		},
	}
	for i, test := range tests {
		assert.Logf("test case #%d: GET %s", i, test.path)
		wreq := wa.CreateRequest(http.MethodGet, test.path)
		wreq.Header().Add("Header-In", test.header)
		wreq.Header().Add("Cookie-In", test.cookie)
		wresp := wreq.Do()
		wresp.AssertStatusCodeEquals(http.StatusOK)
		wresp.Header().AssertKeyValueEquals("Header-Out", test.header)
		wresp.Cookies().AssertKeyValueEquals("Cookie-Out", test.cookie)
		wresp.AssertBodyGrep(".*[Dd]one.*")
		wresp.AssertBodyContains("!")
	}
}

//--------------------
// WEB ASSERTER AND HANDLER
//--------------------

// StartTestServer initialises and starts the asserter for the tests.
func StartWebAsserter(assert *asserts.Asserts) *environments.WebAsserter {
	wa := environments.NewWebAsserter(assert)

	wa.Handle("/hello/world/", MakeHelloWorldHandler(assert, "World"))
	wa.Handle("/hello/tester/", MakeHelloWorldHandler(assert, "Tester"))
	wa.Handle("/hello/postman/", MakeHelloWorldHandler(assert, "Postman"))
	wa.Handle("/header/cookies/", MakeHeaderCookiesHandler(assert))
	return wa
}

// MakeHelloWorldHandler creates a "Hello, World" handler.
func MakeHelloWorldHandler(assert *asserts.Asserts, who string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reply := "Hello, " + who + "!"
		w.Header().Add(environments.HeaderContentType, environments.ContentTypePlain)
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(reply)); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

// MakeHeaderCookiesHandler creates a handler for header and cookies.
func MakeHeaderCookiesHandler(assert *asserts.Asserts) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		headerOut := r.Header.Get("Header-In")
		cookieOut := r.Header.Get("Cookie-In")
		http.SetCookie(w, &http.Cookie{
			Name:  "Cookie-Out",
			Value: cookieOut,
		})
		w.WriteHeader(http.StatusOK)
		w.Header().Set(environments.HeaderContentType, environments.ContentTypePlain)
		w.Header().Set("Header-Out", headerOut)
		if _, err := w.Write([]byte("Done!")); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

// EOF
