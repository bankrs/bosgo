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
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type req struct {
	hc                *http.Client
	ctx               context.Context
	clientID          string
	addr              string
	path              string
	par               params
	headers           headers
	environment       string
	requestsAttempted int
	retryPolicy       RetryPolicy
	allowRetry        bool
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

func (r *req) policyAllowsRetry() bool {
	return r.retryPolicy.canRetry(r.requestsAttempted + 1)
}

func (r *req) nextReq() (*req, time.Duration) {
	r2 := &req{
		hc:                r.hc,
		ctx:               r.ctx,
		clientID:          r.clientID,
		addr:              r.addr,
		path:              r.path,
		par:               r.par,
		headers:           r.headers,
		environment:       r.environment,
		requestsAttempted: r.requestsAttempted + 1,
		retryPolicy:       r.retryPolicy,
		allowRetry:        r.allowRetry,
	}
	return r2, r.retryPolicy.NextWait(r2.requestsAttempted)
}

func (r *req) get() (*http.Response, func(), error) {
	req, err := http.NewRequest("GET", r.url().String(), nil)
	if err != nil {
		return nil, func() {}, err
	}
	if r.ctx != nil {
		req = req.WithContext(r.ctx)
	}
	if r.clientID != "" {
		req.Header.Set("X-Client-Id", r.clientID)
	}
	if r.environment != "" {
		req.Header.Set("X-Environment", r.environment)
	}
	for k, v := range r.headers {
		req.Header.Set(k, v)
	}

	res, err := r.hc.Do(req)
	if err != nil {
		return nil, func() {}, err
	}
	if err, retry := responseError(res); err != nil {
		// By default all GETs are deemed to be retryable
		if retry && r.policyAllowsRetry() {
			nextReq, wait := r.nextReq()
			time.Sleep(wait)
			return nextReq.get()
		}

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
	if r.clientID != "" {
		req.Header.Set("X-Client-Id", r.clientID)
	}
	if r.environment != "" {
		req.Header.Set("X-Environment", r.environment)
	}
	for k, v := range r.headers {
		req.Header.Set(k, v)
	}

	res, err := r.hc.Do(req)
	if err != nil {
		return nil, func() {}, err
	}
	if err, retry := responseError(res); err != nil {
		if retry && r.allowRetry && r.policyAllowsRetry() {
			nextReq, wait := r.nextReq()
			time.Sleep(wait)
			return nextReq.postJSON(data)
		}
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
	if r.clientID != "" {
		req.Header.Set("X-Client-Id", r.clientID)
	}
	if r.environment != "" {
		req.Header.Set("X-Environment", r.environment)
	}
	for k, v := range r.headers {
		req.Header.Set(k, v)
	}

	res, err := r.hc.Do(req)
	if err != nil {
		return nil, func() {}, err
	}
	if err, retry := responseError(res); err != nil {
		if retry && r.allowRetry && r.policyAllowsRetry() {
			nextReq, wait := r.nextReq()
			time.Sleep(wait)
			return nextReq.putJSON(data)
		}
		return nil, func() {}, err
	}
	return res, cleanup(res), nil
}

func (r *req) delete(data interface{}) (*http.Response, func(), error) {
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
	if r.clientID != "" {
		req.Header.Set("X-Client-Id", r.clientID)
	}
	if r.environment != "" {
		req.Header.Set("X-Environment", r.environment)
	}
	for k, v := range r.headers {
		req.Header.Set(k, v)
	}

	res, err := r.hc.Do(req)
	if err != nil {
		return nil, func() {}, err
	}
	if err, retry := responseError(res); err != nil {
		if retry && r.allowRetry && r.policyAllowsRetry() {
			nextReq, wait := r.nextReq()
			time.Sleep(wait)
			return nextReq.delete(data)
		}
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
	if r.clientID != "" {
		req.Header.Set("X-Client-Id", r.clientID)
	}
	if r.environment != "" {
		req.Header.Set("X-Environment", r.environment)
	}
	for k, v := range r.headers {
		req.Header.Set(k, v)
	}

	res, err := r.hc.Do(req)
	if err != nil {
		return nil, func() {}, err
	}
	if err, retry := responseError(res); err != nil {
		if retry && r.allowRetry && r.policyAllowsRetry() {
			nextReq, wait := r.nextReq()
			time.Sleep(wait)
			return nextReq.deleteJSON(data)
		}
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
	URL        string      // the request URL
}

func (e *Error) Error() string {
	if len(e.Errors) == 1 {
		if e.Errors[0].Message == "" {
			return fmt.Sprintf("%s: %s [request-id: %s; URL: %s]", e.Errors[0].Code, e.Status, e.RequestID, e.URL)
		}
		return fmt.Sprintf("%s: %s [request-id: %s; Status: %s; URL: %s]", e.Errors[0].Code, e.Errors[0].Message, e.RequestID, e.Status, e.URL)
	}
	// TODO: expand on error message
	return fmt.Sprintf("request failed with status %s [request-id: %s; URL: %s]", e.Status, e.RequestID, e.URL)
}

// ErrorItem is a detailed error code & message.
type ErrorItem struct {
	Code    string              `json:"code"`    // standard error code
	Message string              `json:"message"` // additional information about the error
	Payload map[string][]string `json:"payload,omitempty"`
}

func (ei *ErrorItem) Description() string {
	var buf bytes.Buffer
	if ei.Message != "" {
		buf.WriteString(ei.Message)
	}

	if len(ei.Payload) > 0 {
		buf.WriteString("(")
		doneFirst := false
		for k, v := range ei.Payload {
			if doneFirst {
				buf.WriteString("; ")
			}
			buf.WriteString(k)
			buf.WriteString("=")
			buf.WriteString(strings.Join(v, ", "))
			doneFirst = true
		}
		buf.WriteString(")")
	}

	return buf.String()
}

func responseError(res *http.Response) (error, bool) {
	if res == nil {
		return &Error{
			Status: "no response found",
		}, false
	}
	if res.StatusCode/100 == 2 {
		return nil, false
	}

	rerr := &Error{
		StatusCode: res.StatusCode,
		Status:     res.Status,
		Header:     res.Header,
		RequestID:  res.Header.Get("X-Request-Id"),
		URL:        res.Request.URL.String(),
	}

	var retryable bool
	if res.StatusCode/100 == 5 {
		retryable = true
	}

	if res.Body == nil {
		return rerr, retryable
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		rerr.Errors = append(rerr.Errors, ErrorItem{
			Code:    "unable_to_read_error_response",
			Message: err.Error(),
		})
		return rerr, retryable
	}

	var serr Error
	err = json.Unmarshal(body, &serr)
	if err != nil {

		n := bytes.IndexByte(body, 0x0)
		if n == -1 {
			n = len(body)
		}
		msg := strings.Replace(strings.Replace(string(body[:n]), "\r", " ", -1), "\n", " ", -1)

		rerr.Errors = append(rerr.Errors, ErrorItem{
			Code:    "unable_to_unmarshal_error_response",
			Message: fmt.Sprintf("received %s", msg),
		})
		return rerr, retryable
	}

	rerr.Errors = append(rerr.Errors, serr.Errors...)
	return rerr, retryable
}

func decodeError(err error, res *http.Response) error {
	rerr := &Error{
		Errors: []ErrorItem{
			{
				Code:    "unable_to_unmarshal_json_response",
				Message: err.Error(),
			},
		},
	}

	if res != nil {
		rerr.StatusCode = res.StatusCode
		rerr.Status = res.Status
		rerr.Header = res.Header
		rerr.RequestID = res.Header.Get("X-Request-Id")
		rerr.URL = res.Request.URL.String()
	}

	return rerr
}

type RetryPolicy struct {
	// MaxRetries is the maximum number of requests that will be made after the original request.
	MaxRetries int

	// Wait is the base time to wait between attempts.
	Wait time.Duration

	// MaxWait is the maximum length of time to wait between retries.
	MaxWait time.Duration

	// Multiplier is the multiplier applied to Wait on each retry after the first. Set to zero
	// to implement a linear backoff.
	Multiplier float64

	// Jitter controls the amount of randomness applied to each wait period. A
	// random amount of time up to +/- Jitter is added to the period.
	Jitter time.Duration
}

func (r RetryPolicy) canRetry(requestsAttempted int) bool {
	return requestsAttempted <= r.MaxRetries
}

func (r RetryPolicy) NextWait(requestsAttempted int) time.Duration {
	if requestsAttempted == 0 {
		return 0
	}

	wait := float64(r.Wait)
	if r.Multiplier > 0 {
		for i := 0; i < requestsAttempted-1; i++ {
			wait *= r.Multiplier
		}
	}

	if r.MaxWait != 0 && wait > float64(r.MaxWait) {
		wait = float64(r.MaxWait)
	}

	if r.Jitter > 0 {
		wait += float64(r.Jitter) * ((rand.Float64() * 2) - 1)
	}

	return time.Duration(wait)
}
