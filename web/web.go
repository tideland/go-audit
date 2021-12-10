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
	"net/http"
)

//--------------------
// RESPONSE
//--------------------

// Response contains the response of the simulated HTTP request.
type Response struct {
	header     http.Header
	statusCode int
	body       []byte
}

// newResponse creates a new initialized response.
func newResponse() *Response {
	return &Response{
		header:     make(http.Header),
		statusCode: http.StatusOK,
	}
}

// Header returns the header values of the response.
func (r *Response) Header() http.Header {
	return r.header
}

// WriteHeader writes the status code of the response.
func (r *Response) WriteHeader(statusCode int) {
	if len(r.body) == 0 {
		r.statusCode = statusCode
	}
}

// StatusCode returns the status code of the response.
func (r *Response) StatusCode() int {
	return r.statusCode
}

// Write implements the io.Writer interface.
func (r *Response) Write(bs []byte) (int, error) {
	r.body = append(r.body, bs...)
	return len(r.body), nil
}

// Body returns a copy of the body of the response.
func (r *Response) Body() []byte {
	bs := make([]byte, len(r.body))
	copy(bs, r.body)
	return bs
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

// Do executes first all registered preprocessors and then lets
// the handler executes it. The build response is returned.
func (s *Simulator) Do(r *http.Request) (*Response, error) {
	for _, pp := range s.pps {
		if err := pp(r); err != nil {
			return nil, err
		}
	}
	rw := newResponse()
	s.h.ServeHTTP(rw, r)
	return rw, nil
}

// EOF
