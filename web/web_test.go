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
		assert.Equal(resp.StatusCode, http.StatusOK)
		body, err := web.BodyToString(resp)
		assert.NoError(err)
		assert.Equal(body, test.expected)
	}
}

// TestJSONBody verifies the reading and writing of JSON bodies.
func TestJSONBody(t *testing.T) {
	assert := asserts.NewTesting(t, asserts.FailStop)

	// Correctly marshalling data.
	s := web.NewFuncSimulator(func(w http.ResponseWriter, r *http.Request) {
		b, err := ioutil.ReadAll(r.Body)
		assert.NoError(err)
		_, err = w.Write(b)
		assert.NoError(err)
	})
	req := s.CreateRequest(http.MethodPost, "https://localhost:8080/", nil)
	err := web.JSONToBody(data{"correct", 12345, true}, req)
	assert.NoError(err)
	resp, err := s.Do(req)
	assert.NoError(err)
	assert.Equal(resp.StatusCode, http.StatusOK)
	var obj data
	err = web.BodyToJSON(resp, &obj)
	assert.NoError(err)
	assert.Equal(obj.A, "correct")
	assert.Equal(obj.B, 12345)
	assert.Equal(obj.C, true)

	// Failing marshalling data.
	s = web.NewFuncSimulator(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("{xyz)[]"))
		assert.NoError(err)
	})
	req = s.CreateRequest(http.MethodGet, "https://localhost:8080/", nil)
	err = web.JSONToBody(data{"correct", 12345, true}, req)
	assert.NoError(err)
	resp, err = s.Do(req)
	assert.NoError(err)
	assert.Equal(resp.StatusCode, http.StatusOK)
	err = web.BodyToJSON(resp, &obj)
	assert.ErrorContains(err, "invalid character")
}

// TestConvenience verifies the correct working of the convenient helper
// methods.
func TestConvenience(t *testing.T) {
	assert := asserts.NewTesting(t, asserts.FailStop)

	// Correctly marshalling data.
	s := web.NewFuncSimulator(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			b := []byte(`{"A":"first", "B":54321, "C":false}`)
			_, err := w.Write(b)
			assert.NoError(err)
			return
		}
		b, err := ioutil.ReadAll(r.Body)
		assert.NoError(err)
		_, err = w.Write(b)
		assert.NoError(err)
	})

	// Step 1: GET.
	resp, err := s.Get("https://localhost:8080/")
	assert.NoError(err)
	assert.Equal(resp.StatusCode, http.StatusOK)
	var obj data
	err = web.BodyToJSON(resp, &obj)
	assert.NoError(err)
	assert.Equal(obj.A, "first")
	assert.Equal(obj.B, 54321)
	assert.Equal(obj.C, false)

	// Steo 2: Simple POST.
	resp, err = s.Post("https://localhost:8080/", "text/plain", strings.NewReader("second"))
	assert.NoError(err)
	assert.Equal(resp.StatusCode, http.StatusOK)
	text, err := web.BodyToString(resp)
	assert.NoError(err)
	assert.Equal(text, "second")

	// Steo 3: Simple POST string.
	resp, err = s.PostString("https://localhost:8080/", "text/plain", "third")
	assert.NoError(err)
	assert.Equal(resp.StatusCode, http.StatusOK)
	text, err = web.BodyToString(resp)
	assert.NoError(err)
	assert.Equal(text, "third")

	// Steo 4: Simple POST JSON object.
	resp, err = s.PostJSON("https://localhost:8080/", "content/json", data{"fourth", 10101, true})
	assert.NoError(err)
	assert.Equal(resp.StatusCode, http.StatusOK)
	err = web.BodyToJSON(resp, &obj)
	assert.NoError(err)
	assert.Equal(obj.A, "fourth")
	assert.Equal(obj.B, 10101)
	assert.Equal(obj.C, true)
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
	req := s.CreateRequest(http.MethodGet, "https://localhost:8080/", nil)
	resp, err := s.Do(req)
	assert.NoError(err)
	assert.Equal(resp.StatusCode, http.StatusPartialContent)

	// Illegally set status after body.
	s = web.NewFuncSimulator(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "body")
		w.WriteHeader(http.StatusPartialContent)
	})
	req = s.CreateRequest(http.MethodGet, "https://localhost:8080/", nil)
	assert.NoError(err)
	resp, err = s.Do(req)
	assert.NoError(err)
	assert.Equal(resp.StatusCode, http.StatusOK)
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
		{http.MethodGet, nil, "m(GET) p(/test/) ct(text/plain) a(text/plain) b()"},
		{http.MethodPost, strings.NewReader("posting data"), "m(POST) p(/test/) ct(text/plain) a(text/plain) b(posting data)"},
		{http.MethodPut, strings.NewReader("posting data"), "m(PUT) p(/test/) ct(text/plain) a(text/plain) b(posting data)"},
		{http.MethodDelete, nil, "m(DELETE) p(/test/) ct(text/plain) a(text/plain) b()"},
	}
	for i, test := range tests {
		assert.Logf("no %d: method %q", i, test.method)
		req := s.CreateRequest(test.method, "http://localhost:8080/test/", test.body)

		resp, err := s.Do(req)
		assert.NoError(err)
		assert.Equal(resp.StatusCode, http.StatusOK)
		body, err := web.BodyToString(resp)
		assert.NoError(err)
		assert.Equal(body, test.expected)
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

// data is used when testing the JSON marshalling.
type data struct {
	A string `json:"a"`
	B int    `json:"b"`
	C bool   `json:"c"`
}

// EOF
