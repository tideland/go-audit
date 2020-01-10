// Tideland Go Audit - Environments
//
// Copyright (C) 2012-2020 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package environments // import "tideland.dev/go/audit/environments"

//--------------------
// IMPORTS
//--------------------

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"html/template"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"regexp"

	"tideland.dev/go/audit/asserts"
)

//--------------------
// CONSTANTS
//--------------------

// Header and content-types.
const (
	HeaderAccept      = "Accept"
	HeaderContentType = "Content-Type"

	ContentTypePlain      = "text/plain"
	ContentTypeHTML       = "text/html"
	ContentTypeXML        = "application/xml"
	ContentTypeJSON       = "application/json"
	ContentTypeURLEncoded = "application/x-www-form-urlencoded"
)

//--------------------
// VALUES
//--------------------

// Values wraps header, cookie, query, and form values.
type Values struct {
	wa   *WebAsserter
	data map[string][]string
}

// newValues creates a new values instance.
func newValues(wa *WebAsserter) *Values {
	vs := &Values{
		wa:   wa,
		data: make(map[string][]string),
	}
	return vs
}

// consumeHeader consumes its values from the HTTP response header.
func consumeHeader(wa *WebAsserter, resp *http.Response) *Values {
	vs := newValues(wa)
	for key, values := range resp.Header {
		for _, value := range values {
			vs.Add(key, value)
		}
	}
	return vs
}

// consumeCookies consumes its values from the HTTP response cookies.
func consumeCookies(wa *WebAsserter, resp *http.Response) *Values {
	vs := newValues(wa)
	for _, cookie := range resp.Cookies() {
		vs.Add(cookie.Name, cookie.Value)
	}
	return vs
}

// Add adds or appends a value to a named field.
func (vs *Values) Add(key, value string) {
	kd := append(vs.data[key], value)
	vs.data[key] = kd
}

// Set sets value of a named field.
func (vs *Values) Set(key, value string) {
	vs.data[key] = []string{value}
}

// Get returns the values for the passed key. May be nil.
func (vs *Values) Get(key string) []string {
	return vs.data[key]
}

// AssertKeyExists tests if the values contain the passed key.
func (vs *Values) AssertKeyExists(key string, msgs ...string) {
	restore := vs.wa.assert.IncrCallstackOffset()
	defer restore()
	_, ok := vs.data[key]
	vs.wa.assert.True(ok, msgs...)
}

// AssertKeyContainsValue tests if the values contain the passed key
// and that the passed value.
func (vs *Values) AssertKeyContainsValue(key, expected string, msgs ...string) {
	restore := vs.wa.assert.IncrCallstackOffset()
	defer restore()
	kd, ok := vs.data[key]
	vs.wa.assert.True(ok, msgs...)
	vs.wa.assert.Contents(expected, kd, msgs...)
}

// AssertKeyValueEquals tests if the first value for a key equals the expected value.
func (vs *Values) AssertKeyValueEquals(key, expected string, msgs ...string) {
	restore := vs.wa.assert.IncrCallstackOffset()
	defer restore()
	values, ok := vs.data[key]
	vs.wa.assert.True(ok, msgs...)
	vs.wa.assert.NotEmpty(values, msgs...)
	vs.wa.assert.Equal(values[0], expected, msgs...)
}

// applyHeader applies its values to the HTTP request header.
func (vs *Values) applyHeader(r *http.Request) {
	for key, values := range vs.data {
		for _, value := range values {
			r.Header.Add(key, value)
		}
	}
}

// applyCookies applies its values to the HTTP request cookies.
func (vs *Values) applyCookies(r *http.Request) {
	restore := vs.wa.assert.IncrCallstackOffset()
	defer restore()
	for key, kd := range vs.data {
		vs.wa.assert.NotEmpty(kd, "cookie must not be empty")
		cookie := &http.Cookie{
			Name:  key,
			Value: kd[0],
		}
		r.AddCookie(cookie)
	}
}

//--------------------
// WEB RESPONSE
//--------------------

// WebResponse provides simplified access to a response in context of
// a web asserter.
type WebResponse struct {
	wa      *WebAsserter
	resp    *http.Response
	header  *Values
	cookies *Values
	body    []byte
}

// Header returns the header values of the response.
func (wresp *WebResponse) Header() *Values {
	return wresp.header
}

// Cookies returns the cookie values of the response.
func (wresp *WebResponse) Cookies() *Values {
	return wresp.cookies
}

// Body returns the body of the response.
func (wresp *WebResponse) Body() []byte {
	return wresp.body
}

// AssertStatusCodeEquals checks if the status is the expected one.
func (wresp *WebResponse) AssertStatusCodeEquals(expected int) {
	restore := wresp.wa.assert.IncrCallstackOffset()
	defer restore()
	wresp.wa.assert.Equal(wresp.resp.StatusCode, expected, "response status differs")
}

// AssertUnmarshalledBody retrieves the body based on the content type
// and unmarshals it accordingly. It asserts that everything works fine.
func (wresp *WebResponse) AssertUnmarshalledBody(data interface{}) {
	restore := wresp.wa.assert.IncrCallstackOffset()
	defer restore()
	contentType := wresp.header.Get(HeaderContentType)
	wresp.wa.assert.NotEmpty(contentType)
	switch contentType[0] {
	case ContentTypeJSON:
		err := json.Unmarshal(wresp.body, data)
		wresp.wa.assert.Nil(err, "cannot unmarshal JSON body")
	case ContentTypeXML:
		err := xml.Unmarshal(wresp.body, data)
		wresp.wa.assert.Nil(err, "cannot unmarshal XML body")
	default:
		wresp.wa.assert.Fail("unmarshalled content type: " + contentType[0])
	}
}

// AssertBodyMatches checks if the body matches a regular expression.
func (wresp *WebResponse) AssertBodyMatches(pattern string) {
	restore := wresp.wa.assert.IncrCallstackOffset()
	defer restore()
	ok, err := regexp.MatchString(pattern, string(wresp.body))
	wresp.wa.assert.Nil(err, "illegal content match pattern")
	wresp.wa.assert.True(ok, "body doesn't match pattern")
}

// AssertBodyGrep greps content out of the body.
func (wresp *WebResponse) AssertBodyGrep(pattern string) []string {
	restore := wresp.wa.assert.IncrCallstackOffset()
	defer restore()
	expr, err := regexp.Compile(pattern)
	wresp.wa.assert.Nil(err, "illegal content grep pattern")
	return expr.FindAllString(string(wresp.body), -1)
}

// AssertBodyContains checks if the body contains a string.
func (wresp *WebResponse) AssertBodyContains(expected string) {
	restore := wresp.wa.assert.IncrCallstackOffset()
	defer restore()
	wresp.wa.assert.Contents(expected, wresp.body, "body doesn't contains expected")
}

//--------------------
// WEB REQUEST
//--------------------

// WebRequest provides simplified access to a request in context of
// a web asserter.
type WebRequest struct {
	wa        *WebAsserter
	method    string
	path      string
	header    *Values
	cookies   *Values
	fieldname string
	filename  string
	body      []byte
}

// Header returns a values instance for request header.
func (wreq *WebRequest) Header() *Values {
	if wreq.header == nil {
		wreq.header = newValues(wreq.wa)
	}
	return wreq.header
}

// Cookies returns a values instance for request cookies.
func (wreq *WebRequest) Cookies() *Values {
	if wreq.cookies == nil {
		wreq.cookies = newValues(wreq.wa)
	}
	return wreq.cookies
}

// SetContentType sets the header Content-Type.
func (wreq *WebRequest) SetContentType(contentType string) {
	wreq.Header().Add(HeaderContentType, contentType)
}

// SetAccept sets the header Accept.
func (wreq *WebRequest) SetAccept(contentType string) {
	wreq.Header().Set(HeaderAccept, contentType)
}

// Upload sets the request as a file upload request.
func (wreq *WebRequest) Upload(fieldname, filename, data string) {
	wreq.fieldname = fieldname
	wreq.filename = filename
	wreq.body = []byte(data)
}

// AssertMarshalBody sets the request body based on the set content type and
// the marshalled data and asserts that everything works fine.
func (wreq *WebRequest) AssertMarshalBody(data interface{}) {
	restore := wreq.wa.assert.IncrCallstackOffset()
	defer restore()
	// Marshal the passed data into the request body.
	contentType := wreq.Header().Get(HeaderContentType)
	wreq.wa.assert.NotEmpty(contentType, "content type must be set for marshalling")
	switch contentType[0] {
	case ContentTypeJSON:
		body, err := json.Marshal(data)
		wreq.wa.assert.Nil(err, "cannot marshal data to JSON")
		wreq.body = body
		wreq.Header().Add(HeaderContentType, ContentTypeJSON)
		wreq.Header().Add(HeaderAccept, ContentTypeJSON)
	case ContentTypeXML:
		body, err := xml.Marshal(data)
		wreq.wa.assert.Nil(err, "cannot marshal data to XML")
		wreq.body = body
		wreq.Header().Add(HeaderContentType, ContentTypeXML)
		wreq.Header().Add(HeaderAccept, ContentTypeXML)
	}
}

// AssertRenderTemplate renders the passed data into the template and
// assigns it to the request body. It asserts that everything works fine.
func (wreq *WebRequest) AssertRenderTemplate(templateSource string, data interface{}) {
	restore := wreq.wa.assert.IncrCallstackOffset()
	defer restore()
	// Render template.
	t, err := template.New(wreq.path).Parse(templateSource)
	wreq.wa.assert.Nil(err, "cannot parse template")
	body := &bytes.Buffer{}
	err = t.Execute(body, data)
	wreq.wa.assert.Nil(err, "cannot render template")
	wreq.body = body.Bytes()
}

// Do performes the web request with the passed method.
func (wreq *WebRequest) Do() *WebResponse {
	restore := wreq.wa.assert.IncrCallstackOffset()
	defer restore()
	// First prepare it.
	var bodyReader io.Reader
	if wreq.filename != "" {
		// Upload file content.
		buffer := &bytes.Buffer{}
		writer := multipart.NewWriter(buffer)
		part, err := writer.CreateFormFile(wreq.fieldname, wreq.filename)
		wreq.wa.assert.Nil(err, "cannot create form file")
		_, err = io.WriteString(part, string(wreq.body))
		wreq.wa.assert.Nil(err, "cannot write data")
		wreq.SetContentType(writer.FormDataContentType())
		err = writer.Close()
		wreq.wa.assert.Nil(err, "cannot close multipart writer")
		wreq.method = http.MethodPost
		bodyReader = ioutil.NopCloser(buffer)
	} else if wreq.body != nil {
		// Upload body content.
		bodyReader = ioutil.NopCloser(bytes.NewBuffer(wreq.body))
	}
	req, err := http.NewRequest(wreq.method, wreq.wa.URL()+wreq.path, bodyReader)
	wreq.wa.assert.Nil(err, "cannot prepare request")
	wreq.Header().applyHeader(req)
	wreq.Cookies().applyCookies(req)
	// Create client and perform request.
	c := http.Client{
		Transport: &http.Transport{},
	}
	resp, err := c.Do(req)
	wreq.wa.assert.Nil(err, "cannot perform test request")
	// Create web response.
	wresp := &WebResponse{
		wa:      wreq.wa,
		resp:    resp,
		header:  consumeHeader(wreq.wa, resp),
		cookies: consumeCookies(wreq.wa, resp),
	}
	body, err := ioutil.ReadAll(resp.Body)
	wreq.wa.assert.Nil(err, "cannot read response")
	defer resp.Body.Close()
	wresp.body = body
	return wresp
}

//--------------------
// WEB ASSERTER
//--------------------

// WebAsserter defines the test server with methods for requests
// and uploads.
type WebAsserter struct {
	assert *asserts.Asserts
	server *httptest.Server
	mux    *http.ServeMux
}

// NewWebAsserter creates a web test server for the tests of own handler
// or the mocking of external systems.
func NewWebAsserter(assert *asserts.Asserts) *WebAsserter {
	wa := &WebAsserter{
		assert: assert,
		mux:    http.NewServeMux(),
	}
	wa.server = httptest.NewServer(wa.mux)
	return wa
}

// Handle registers the handler for the given pattern. If a handler
// already exists for pattern, Handle panics.
func (wa *WebAsserter) Handle(pattern string, handler http.Handler) {
	wa.mux.Handle(pattern, handler)
}

// HandleFunc registers the handler function for the given pattern
func (wa *WebAsserter) HandleFunc(pattern string, handler func(w http.ResponseWriter, r *http.Request)) {
	wa.mux.HandleFunc(pattern, handler)
}

// URL returns the local URL of the internal test server.
func (wa *WebAsserter) URL() string {
	return wa.server.URL
}

// Close shuts down the internal test server and blocks until all
// outstanding requests have completed.
func (wa *WebAsserter) Close() {
	wa.server.Close()
}

// CreateRequest prepares a web request to be performed
// against this web asserter.
func (wa *WebAsserter) CreateRequest(method, path string) *WebRequest {
	return &WebRequest{
		wa:     wa,
		method: method,
		path:   path,
	}
}

// EOF
