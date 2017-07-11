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
	"testing"
)

var (
	userTokenHandler = func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"token":"usertoken", "id":"userid"}`)
	}
)

func TestUserLogin(t *testing.T) {
	routes := routeMap{
		"/v1/users/login": {
			http.MethodPost: userTokenHandler,
		},
	}

	hc, cleanup := startTestServer(t, routes)
	defer cleanup()

	appClient := NewAppClient(hc, SandboxAddr, "appid")
	userClient, err := appClient.Users.Login("name", "password").Send()
	if err != nil {
		t.Fatalf("failed to send logout request: %v", err)
	}

	if userClient == nil {
		t.Fatal("got nil user client, wanted non-nil")
	}

	if userClient.SessionToken() != "usertoken" {
		t.Errorf("got session token %q, wanted %q", userClient.SessionToken(), "usertoken")
	}
}

func TestUserLoginUnknown(t *testing.T) {
	routes := routeMap{
		"/v1/users/login": {
			http.MethodPost: unauthorizedHandler,
		},
	}

	hc, cleanup := startTestServer(t, routes)
	defer cleanup()

	appClient := NewAppClient(hc, SandboxAddr, "appid")
	_, err := appClient.Users.Login("name", "password").Send()

	if err == nil {
		t.Fatal("got nil error, wanted non-nil")
	}

	serr := err.(*Error)
	if serr.StatusCode != http.StatusUnauthorized {
		t.Errorf("got status code %d, wanted %d", serr.StatusCode, http.StatusUnauthorized)
	}
}

func TestUserCreate(t *testing.T) {
	routes := routeMap{
		"/v1/users": {
			http.MethodPost: userTokenHandler,
		},
	}

	hc, cleanup := startTestServer(t, routes)
	defer cleanup()

	appClient := NewAppClient(hc, SandboxAddr, "appid")
	userClient, err := appClient.Users.Create("name", "password").Send()
	if err != nil {
		t.Fatalf("failed to send logout request: %v", err)
	}

	if userClient == nil {
		t.Fatal("got nil user client, wanted non-nil")
	}

	if userClient.SessionToken() != "usertoken" {
		t.Errorf("got session token %q, wanted %q", userClient.SessionToken(), "usertoken")
	}

	if userClient.addr != appClient.addr {
		t.Errorf("got addr %q, wanted %q", userClient.addr, appClient.addr)
	}

	if userClient.ua != appClient.ua {
		t.Errorf("got ua %q, wanted %q", userClient.ua, appClient.ua)
	}

}
