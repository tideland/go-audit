// Tideland Go Audit - Web
//
// Copyright (C) 2012-2023 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

// Package web helps testing web handlers. A simulator can be started
// with the handler to test. Then standard http requests can be sent
// and the returned http responses can be analyzed.
//
//	h := NewMyHandler()
//	s := web.NewSimulator(h)
//
//	req := s.CreateRequest(http.MethodGet, "http://localhost:8080/", nil)
//
//	resp, err := s.Do(req)
//	assert.NoError(err)
//	assert.Equal(resp.StatusCode, http.StatusOK)
//	body, err := web.BodyToString(resp)
//	assert.NoError(err)
//	assert.Equal(body, test.expected)
//
// Some smaller functions help working with the requests and responses.
package web // import "tideland.dev/go/audit/web"

// EOF
