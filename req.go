// Copyright 2017 Bankrs AG.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package bosgo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
)

type req struct {
	hc      *http.Client
	ctx     context.Context
	addr    string
	path    string
	par     params
	headers headers
}

func (r *req) url() *url.URL {
	u := url.URL{
		Scheme:   "https",
		Host:     r.addr,
		Path:     r.path,
		RawQuery: r.par.Encode(),
	}
	return &u
}

func (r *req) get() (*http.Response, func(), error) {
	req, err := http.NewRequest("GET", r.url().String(), nil)
	if err != nil {
		return nil, func() {}, err
	}
	if r.ctx != nil {
		req = req.WithContext(r.ctx)
	}
	// TODO: remove x-environment header
	req.Header.Set("x-environment", "sandbox")
	for k, v := range r.headers {
		req.Header.Set(k, v)
	}

	res, err := r.hc.Do(req)
	if err != nil {
		return nil, func() {}, err
	}
	if err := responseError(res); err != nil {
		return nil, func() {}, err
	}
	return res, cleanup(res), nil
}

func (r *req) postJSON(data interface{}) (*http.Response, func(), error) {
	var body io.Reader
	if data != nil {
		var encoded bytes.Buffer
		err := json.NewEncoder(&encoded).Encode(data)
		if err != nil {
			return nil, func() {}, err
		}
		body = &encoded
	}

	req, err := http.NewRequest("POST", r.url().String(), body)
	if err != nil {
		return nil, func() {}, err
	}
	if r.ctx != nil {
		req = req.WithContext(r.ctx)
	}
	req.Header.Set("Content-Type", "application/json")
	// TODO: remove x-environment header
	req.Header.Set("x-environment", "sandbox")
	for k, v := range r.headers {
		req.Header.Set(k, v)
	}

	res, err := r.hc.Do(req)
	if err != nil {
		return nil, func() {}, err
	}
	if err := responseError(res); err != nil {
		return nil, func() {}, err
	}
	return res, cleanup(res), nil
}

func (r *req) putJSON(data interface{}) (*http.Response, func(), error) {
	var body io.Reader
	if data != nil {
		var encoded bytes.Buffer
		err := json.NewEncoder(&encoded).Encode(data)
		if err != nil {
			return nil, func() {}, err
		}
		body = &encoded
	}

	req, err := http.NewRequest("PUT", r.url().String(), body)
	if err != nil {
		return nil, func() {}, err
	}
	if r.ctx != nil {
		req = req.WithContext(r.ctx)
	}
	req.Header.Set("Content-Type", "application/json")
	// TODO: remove x-environment header
	req.Header.Set("x-environment", "sandbox")
	for k, v := range r.headers {
		req.Header.Set(k, v)
	}

	res, err := r.hc.Do(req)
	if err != nil {
		return nil, func() {}, err
	}
	if err := responseError(res); err != nil {
		return nil, func() {}, err
	}
	return res, cleanup(res), nil
}

func (r *req) delete() (*http.Response, func(), error) {
	req, err := http.NewRequest("DELETE", r.url().String(), nil)
	if err != nil {
		return nil, func() {}, err
	}
	if r.ctx != nil {
		req = req.WithContext(r.ctx)
	}
	// TODO: remove x-environment header
	req.Header.Set("x-environment", "sandbox")
	for k, v := range r.headers {
		req.Header.Set(k, v)
	}

	res, err := r.hc.Do(req)
	if err != nil {
		return nil, func() {}, err
	}
	if err := responseError(res); err != nil {
		return nil, func() {}, err
	}
	return res, cleanup(res), nil
}

func (r *req) deleteJSON(data interface{}) (*http.Response, func(), error) {
	var body io.Reader
	if data != nil {
		var encoded bytes.Buffer
		err := json.NewEncoder(&encoded).Encode(data)
		if err != nil {
			return nil, func() {}, err
		}
		body = &encoded
	}

	req, err := http.NewRequest("DELETE", r.url().String(), body)
	if err != nil {
		return nil, func() {}, err
	}
	if r.ctx != nil {
		req = req.WithContext(r.ctx)
	}
	req.Header.Set("Content-Type", "application/json")
	// TODO: remove x-environment header
	req.Header.Set("x-environment", "sandbox")
	for k, v := range r.headers {
		req.Header.Set(k, v)
	}

	res, err := r.hc.Do(req)
	if err != nil {
		return nil, func() {}, err
	}
	if err := responseError(res); err != nil {
		return nil, func() {}, err
	}
	return res, cleanup(res), nil
}

func cleanup(res *http.Response) func() {
	return func() {
		if res == nil || res.Body == nil {
			return
		}
		res.Body.Close()
	}
}

type params map[string][]string

func (p params) Get(key string) string {
	vs := p[key]
	if len(vs) == 0 {
		return ""
	}
	return vs[0]
}

func (p params) Set(key, value string) {
	p[key] = []string{value}
}

func (p params) Encode() string {
	return url.Values(p).Encode()
}

type headers map[string]string

// Error contains an error response from a service.
type Error struct {
	Errors     []ErrorItem `json:"errors"` // error messages reported by the service
	StatusCode int         // the HTTP status code from the service response
	Status     string      // the HTTP status line from the service response
	Header     http.Header // the HTTP headers from the service response
	RequestID  string      // the ID of the request that generated the error
}

// ErrorItem is a detailed error code & message.
type ErrorItem struct {
	Code    string `json:"code"`    // standard error code
	Message string `json:"message"` // additional information about the error
}

func (e *Error) Error() string {
	if len(e.Errors) == 1 {
		if e.Errors[0].Message == "" {
			return fmt.Sprintf("%s: %s [request-id: %s]", e.Errors[0].Code, e.Status, e.RequestID)
		}
		return fmt.Sprintf("%s: %s [request-id: %s]", e.Errors[0].Code, e.Errors[0].Message, e.RequestID)
	}
	// TODO: expand on error message
	return fmt.Sprintf("request failed with status %s [request-id: %s]", e.Status, e.RequestID)
}

func responseError(res *http.Response) error {
	if res == nil {
		return &Error{
			Status: "no response found",
		}
	}
	if res.StatusCode/100 == 2 {
		return nil
	}

	rerr := &Error{
		StatusCode: res.StatusCode,
		Status:     res.Status,
		Header:     res.Header,
		RequestID:  res.Header.Get("X-Request-Id"),
	}

	if res.Body == nil {
		return rerr
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		rerr.Errors = append(rerr.Errors, ErrorItem{
			Code:    "unable_to_read_error_response",
			Message: err.Error(),
		})
		return rerr
	}

	var serr Error
	err = json.Unmarshal(body, &serr)
	if err != nil {
		rerr.Errors = append(rerr.Errors, ErrorItem{
			Code:    "unable_to_unmarshal_error_response",
			Message: err.Error(),
		})
		return rerr
	}

	rerr.Errors = append(rerr.Errors, serr.Errors...)
	return rerr
}

func wrap(prefix string, err error) error {
	return fmt.Errorf("%s: %v", prefix, err)
}
