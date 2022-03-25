// Tideland Go Audit - Web
//
// Copyright (C) 2012-2021 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package web // import "tideland.dev/go/audit/web"

//--------------------
// IMPORTS
//--------------------

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
)

//--------------------
// BODY HELPER
//--------------------

// StringToBody sets the request body to a given string.
func StringToBody(s string, r *http.Request) {
	r.Body = ioutil.NopCloser(bytes.NewBufferString(s))
}

// JSONToBody sets the request body to the JSON representation of
// the given object.
func JSONToBody(obj any, r *http.Request) error {
	b := bytes.NewBuffer(nil)
	if err := json.NewEncoder(b).Encode(obj); err != nil {
		return err
	}
	r.Body = ioutil.NopCloser(b)
	return nil
}

// BodyToString reads the whole body and simply interprets it as string.
func BodyToString(r *http.Response) (string, error) {
	bs, err := ioutil.ReadAll(r.Body)
	return string(bs), err
}

// BodyToJSON reads the whole body and decodes the JSON content into the
// given object.
func BodyToJSON(r *http.Response, obj any) error {
	return json.NewDecoder(r.Body).Decode(obj)
}

//--------------------
// SIMULATOR
//--------------------

// Preprocessor will be executed before a request is passed to the
// handler.
type Preprocessor func(r *http.Request) error

// Simulator locally simulates HTTP requests to handler.
type Simulator struct {
	h   http.Handler
	pps []Preprocessor
}

// NewSimulator creates a new local HTTP request simulator.
func NewSimulator(h http.Handler, pps ...Preprocessor) *Simulator {
	return &Simulator{
		h:   h,
		pps: pps,
	}
}

// NewFuncSimulator is a convenient variant of NewSimulator just for
// a http.HandlerFunc.
func NewFuncSimulator(f http.HandlerFunc, pps ...Preprocessor) *Simulator {
	return NewSimulator(f, pps...)
}

// CreateRequest creates a request for the simulator.
func (s *Simulator) CreateRequest(method, target string, body io.Reader) *http.Request {
	return httptest.NewRequest(method, target, body)
}

// Do executes first all registered preprocessors and then lets
// the handler executes it. The build response is returned.
func (s *Simulator) Do(r *http.Request) (*http.Response, error) {
	for _, pp := range s.pps {
		if err := pp(r); err != nil {
			return nil, err
		}
	}
	w := httptest.NewRecorder()
	s.h.ServeHTTP(w, r)
	return w.Result(), nil
}

// Get conveniently executes a simple GET request.
func (s *Simulator) Get(target string) (*http.Response, error) {
	req := s.CreateRequest(http.MethodGet, target, nil)

	return s.Do(req)
}

// Post conveniently executes a simple POST request.
func (s *Simulator) Post(target, contentType string, body io.Reader) (*http.Response, error) {
	req := s.CreateRequest(http.MethodPost, target, body)

	return s.Do(req)
}

// PostText conveniently executes a simple POST request with the given string body.
func (s *Simulator) PostText(target, body string) (*http.Response, error) {
	req := s.CreateRequest(http.MethodPost, target, nil)
	req.Header.Set("Content-Type", "text/plain")

	StringToBody(body, req)

	return s.Do(req)
}

// PostJSON conveniently executes a simple POST request with the given interface  body.
func (s *Simulator) PostJSON(target string, body any) (*http.Response, error) {
	req := s.CreateRequest(http.MethodPost, target, nil)
	req.Header.Set("Content-Type", "application/json")

	if err := JSONToBody(body, req); err != nil {
		return nil, err
	}

	return s.Do(req)
}

// EOF
