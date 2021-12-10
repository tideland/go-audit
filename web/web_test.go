// Tideland Go Audit - Web - Unit Tests
//
// Copyright (C) 2012-2021 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package web_test

//--------------------
// IMPORTS
//--------------------

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"tideland.dev/go/audit/asserts"
	"tideland.dev/go/audit/web"
)

//--------------------
// TESTS
//--------------------

// TestSimpleRequests tests handling of requests without preprocessors.
func TestSimpleRequests(t *testing.T) {
	assert := asserts.NewTesting(t, asserts.FailStop)
	h := &echoHandler{assert}
	s := web.NewSimulator(h)

	tests := []struct {
		method   string
		body     io.Reader
		expected string
	}{
		{http.MethodGet, nil, "m(GET) p(/test/) ct() a() b()"},
		{http.MethodPost, strings.NewReader("posting data"), "m(POST) p(/test/) ct() a() b(posting data)"},
		{http.MethodPut, strings.NewReader("posting data"), "m(PUT) p(/test/) ct() a() b(posting data)"},
		{http.MethodDelete, nil, "m(DELETE) p(/test/) ct() a() b()"},
	}
	for i, test := range tests {
		assert.Logf("no %d: method %q", i, test.method)
		req, err := http.NewRequest(test.method, "http://localhost:8080/test/", test.body)
		assert.NoError(err)

		resp, err := s.Do(req)
		assert.NoError(err)
		assert.Equal(resp.StatusCode(), http.StatusOK)

		body := resp.Body()

		assert.Equal(string(body), test.expected)
	}
}

// TestResponseCode verifies that the status code cannot be changed after
// writing to the response body.
func TestResponseCode(t *testing.T) {
	assert := asserts.NewTesting(t, asserts.FailStop)

	// Correctly set status before body.
	s := web.NewFuncSimulator(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusPartialContent)
		fmt.Fprint(w, "body")
	})
	req, err := http.NewRequest(http.MethodGet, "https://localhost:8080/", nil)
	assert.NoError(err)
	resp, err := s.Do(req)
	assert.NoError(err)
	assert.Equal(resp.StatusCode(), http.StatusPartialContent)

	// Illegally set status after body.
	s = web.NewFuncSimulator(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "body")
		w.WriteHeader(http.StatusPartialContent)
	})
	req, err = http.NewRequest(http.MethodGet, "https://localhost:8080/", nil)
	assert.NoError(err)
	resp, err = s.Do(req)
	assert.NoError(err)
	assert.Equal(resp.StatusCode(), http.StatusOK)
}

// TestPreprocessors verifies that preprocessors are correctly modifying
// a request as wanted.
func TestPreprocessors(t *testing.T) {
	assert := asserts.NewTesting(t, asserts.FailStop)
	ppContentType := func(r *http.Request) error {
		if r.Body != nil {
			r.Header.Add("Content-Type", "text/plain")
		}
		return nil
	}
	ppAccept := func(r *http.Request) error {
		r.Header.Add("Accept", "text/plain")
		return nil
	}
	h := &echoHandler{assert}
	s := web.NewSimulator(h, ppContentType, ppAccept)

	tests := []struct {
		method   string
		body     io.Reader
		expected string
	}{
		{http.MethodGet, nil, "m(GET) p(/test/) ct() a(text/plain) b()"},
		{http.MethodPost, strings.NewReader("posting data"), "m(POST) p(/test/) ct(text/plain) a(text/plain) b(posting data)"},
		{http.MethodPut, strings.NewReader("posting data"), "m(PUT) p(/test/) ct(text/plain) a(text/plain) b(posting data)"},
		{http.MethodDelete, nil, "m(DELETE) p(/test/) ct() a(text/plain) b()"},
	}
	for i, test := range tests {
		assert.Logf("no %d: method %q", i, test.method)
		req, err := http.NewRequest(test.method, "http://localhost:8080/test/", test.body)
		assert.NoError(err)

		resp, err := s.Do(req)
		assert.NoError(err)
		assert.Equal(resp.StatusCode(), http.StatusOK)

		body := resp.Body()

		assert.Equal(string(body), test.expected)
	}
}

//--------------------
// HELPER
//--------------------

// echoHandler simply echos some data of the request into the response for testing.
type echoHandler struct {
	assert *asserts.Asserts
}

func (h *echoHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Read the interesting request parts.
	m := r.Method
	p := r.URL.Path
	ct := r.Header.Get("Content-Type")
	a := r.Header.Get("Accept")

	var bs []byte
	var err error

	if r.Body != nil {
		bs, err = ioutil.ReadAll(r.Body)
		h.assert.NoError(err)
	}

	// Echo them.
	w.Header().Add("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "m(%s) p(%s) ct(%s) a(%s) b(%s)", m, p, ct, a, string(bs))
}

// EOF
