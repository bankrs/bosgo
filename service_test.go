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
	"testing"
)

var (
	devLoginSuccessHandler = func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"token":"session-token"}`)
	}

	devLoginUnknownHandler = func(w http.ResponseWriter, r *http.Request) {
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
		Transport: RewriteTransport{URL: u},
	}

	return client, func() {
		ts.Close()
	}
}

type RewriteTransport struct {
	Transport http.RoundTripper
	URL       *url.URL
}

func (t RewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.URL.Scheme = t.URL.Scheme
	req.URL.Host = t.URL.Host
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
			http.MethodPost: devLoginSuccessHandler,
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

	if devClient.SessionToken() != "session-token" {
		t.Errorf("got session token %q, wanted %q", devClient.SessionToken(), "session-token")
	}
}

func TestDeveloperLoginUnknown(t *testing.T) {
	routes := routeMap{
		"/v1/developers/login": {
			http.MethodPost: devLoginUnknownHandler,
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
