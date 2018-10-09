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
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path"
	"strings"
	"testing"
	"time"
)

var (
	noContentHandler = func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}

	errorHandler = func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, `{"errors":[{"code":"general"}]}`)
	}

	devTokenHandler = func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"token":"devtoken"}`)
	}

	unauthorizedHandler = func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, `{"errors":[{"code":"authentication_failed"}]}`)
	}
)

type routeMap map[string]map[string]http.HandlerFunc

func startTestServer(t *testing.T, routes routeMap) (*http.Client, func()) {
	mux := http.NewServeMux()
	for path, methodHandlers := range routes {
		for method, handler := range methodHandlers {
			mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
				if r.Method != method {
					w.WriteHeader(http.StatusNotFound)
					return
				}
				handler(w, r)
			})
		}
	}

	ts := httptest.NewServer(mux)

	u, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatalf("failed to parse httptest.Server URL: %v", err)
	}

	client := &http.Client{
		Transport: urlRewriteTransport{URL: u},
	}

	return client, func() {
		ts.Close()
	}
}

type urlRewriteTransport struct {
	Transport    http.RoundTripper
	URL          *url.URL
	StripVersion bool
}

func (t urlRewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.URL.Scheme = t.URL.Scheme
	req.URL.Host = t.URL.Host
	if t.StripVersion {
		req.URL.Path = req.URL.Path[strings.IndexByte(req.URL.Path[1:], '/')+1:]
	}
	req.URL.Path = path.Join(t.URL.Path, req.URL.Path)
	rt := t.Transport
	if rt == nil {
		rt = http.DefaultTransport
	}
	return rt.RoundTrip(req)
}

func TestDeveloperLoginSuccess(t *testing.T) {
	routes := routeMap{
		"/v1/developers/login": {
			http.MethodPost: devTokenHandler,
		},
	}

	hc, cleanup := startTestServer(t, routes)
	defer cleanup()

	client := New(hc, SandboxAddr)
	devClient, err := client.Login("dev@example.com", "pwd").Send()

	if err != nil {
		t.Fatalf("failed to send request: %v", err)
	}

	if devClient == nil {
		t.Fatal("got nil developer client, wanted non-nil")
	}

	if devClient.SessionToken() != "devtoken" {
		t.Errorf("got session token %q, wanted %q", devClient.SessionToken(), "devtoken")
	}
}

func TestDeveloperLoginUnknown(t *testing.T) {
	routes := routeMap{
		"/v1/developers/login": {
			http.MethodPost: unauthorizedHandler,
		},
	}

	hc, cleanup := startTestServer(t, routes)
	defer cleanup()

	client := New(hc, SandboxAddr)
	_, err := client.Login("dev@example.com", "pwd").Send()

	if err == nil {
		t.Fatal("got nil error, wanted non-nil")
	}

	serr := err.(*Error)
	if serr.StatusCode != http.StatusUnauthorized {
		t.Errorf("got status code %d, wanted %d", serr.StatusCode, http.StatusUnauthorized)
	}
}

func TestCreateDeveloper(t *testing.T) {
	routes := routeMap{
		"/v1/developers": {
			http.MethodPost: devTokenHandler,
		},
	}

	hc, cleanup := startTestServer(t, routes)
	defer cleanup()

	client := New(hc, SandboxAddr)
	devClient, err := client.CreateDeveloper("dev@example.com", "pwd").Send()

	if err != nil {
		t.Fatalf("failed to send request: %v", err)
	}

	if devClient == nil {
		t.Fatal("got nil developer client, wanted non-nil")
	}

	if devClient.SessionToken() != "devtoken" {
		t.Errorf("got session token %q, wanted %q", devClient.SessionToken(), "devtoken")
	}

	if devClient.addr != client.addr {
		t.Errorf("got addr %q, wanted %q", devClient.addr, client.addr)
	}

	if devClient.ua != client.ua {
		t.Errorf("got ua %q, wanted %q", devClient.ua, client.ua)
	}

}

type transientErrorHandler struct {
	retriesNeeded   int
	successResponse string
}

func (t *transientErrorHandler) Handle(w http.ResponseWriter, r *http.Request) {
	t.retriesNeeded--
	if t.retriesNeeded > 0 {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, `{"errors":[{"code":"retry_test_failure"}]}`)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, t.successResponse)

}

func TestRetryGet(t *testing.T) {

	handler := &transientErrorHandler{
		retriesNeeded:   5,
		successResponse: `[{"score":1, "provider":{"id":"DE-BIN-10001000"}}]`,
	}

	routes := routeMap{
		"/v1/providers": {
			http.MethodGet: handler.Handle,
		},
	}

	hc, cleanup := startTestServer(t, routes)
	defer cleanup()

	// Request fails without retry policy
	clientNoRetry := New(hc, SandboxAddr)
	appClientNoRetry := clientNoRetry.WithApplicationID("applicationid")

	_, err := appClientNoRetry.Providers.Search("foo").Send()
	if err == nil {
		t.Fatalf("expected error but did not get one")
	}

	policy := RetryPolicy{
		MaxRetries: 10,
		Wait:       100 * time.Microsecond,
		MaxWait:    500 * time.Microsecond,
	}

	// Request succeeds with retry policy
	clientWithRetry := New(hc, SandboxAddr, WithRetryPolicy(policy))
	appClientWithRetry := clientWithRetry.WithApplicationID("applicationid")

	_, err = appClientWithRetry.Providers.Search("foo").Send()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

}

func TestRetryGetReturnsLastError(t *testing.T) {
	handler := &transientErrorHandler{
		retriesNeeded:   20,
		successResponse: `[{"score":1, "provider":{"id":"DE-BIN-10001000"}}]`,
	}

	routes := routeMap{
		"/v1/providers": {
			http.MethodGet: handler.Handle,
		},
	}

	hc, cleanup := startTestServer(t, routes)
	defer cleanup()

	policy := RetryPolicy{
		MaxRetries: 10, // Fewer retries than are needed by service
		Wait:       100 * time.Microsecond,
		MaxWait:    500 * time.Microsecond,
	}

	// Request succeeds with retry policy
	clientWithRetry := New(hc, SandboxAddr, WithRetryPolicy(policy))
	appClientWithRetry := clientWithRetry.WithApplicationID("applicationid")

	_, err := appClientWithRetry.Providers.Search("foo").Send()
	if err == nil {
		t.Fatalf("expected error but did not get one")
	}

	rerr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected response error of type *Error but got %T", err)
	}

	if rerr.StatusCode != 500 {
		t.Errorf("got status %d, wanted 500", rerr.StatusCode)
	}

	if len(rerr.Errors) != 1 {
		t.Errorf("got %d error messages, wanted 1", len(rerr.Errors))
	}

	if rerr.Errors[0].Code != "retry_test_failure" {
		t.Errorf("got error code %s, wanted retry_test_failure", rerr.Errors[0].Code)
	}
}

func TestRetryPost(t *testing.T) {
	handler1 := &transientErrorHandler{
		retriesNeeded:   5,
		successResponse: `{"id":"foo"}`,
	}
	handler2 := &transientErrorHandler{
		retriesNeeded:   5,
		successResponse: `{"users":[]}`,
	}

	routes := routeMap{
		"/v1/transfers": {
			http.MethodPost: handler1.Handle,
		},
		"/v1/developers/users": {
			http.MethodPost: handler2.Handle,
		},
	}

	hc, cleanup := startTestServer(t, routes)
	defer cleanup()

	policy := RetryPolicy{
		MaxRetries: 10,
		Wait:       100 * time.Microsecond,
		MaxWait:    500 * time.Microsecond,
	}

	// Request fails since retries are not allowed when creating a transfer
	userClient := NewUserClient(hc, SandboxAddr, "usertoken", "applicationid")
	userClient.retryPolicy = policy

	_, err := userClient.Transfers.Create(1, TransferAddress{Name: "test"}, MoneyAmount{Currency: "EUR", Value: "40.15"}).Send()
	if err == nil {
		t.Fatalf("expected error but did not get one")
	}

	// Request succeeds since retries are allowed when searching for users
	devClient := NewDevClient(hc, SandboxAddr, "devtoken")
	devClient.retryPolicy = policy

	_, err = devClient.Applications.ListUsers("applicationid").Limit(40).Send()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

}
