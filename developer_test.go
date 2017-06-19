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
	"net/http"
	"testing"
)

func TestDeveloperLogout(t *testing.T) {
	routes := routeMap{
		"/v1/developers/logout": {
			http.MethodPost: noContentHandler,
		},
	}

	hc, cleanup := startTestServer(t, routes)
	defer cleanup()

	devClient := NewDevClient(hc, SandboxAddr, "devtoken")
	err := devClient.Logout().Send()
	if err != nil {
		t.Fatalf("failed to send logout request: %v", err)
	}
}

func TestWithApplication(t *testing.T) {
	hc, cleanup := startTestServer(t, nil)
	defer cleanup()

	devClient := NewDevClient(hc, SandboxAddr, "devtoken")
	appClient := devClient.WithApplication("appid")

	if appClient.token != devClient.token {
		t.Errorf("got token %q, wanted %q", appClient.token, devClient.token)
	}

	if appClient.addr != devClient.addr {
		t.Errorf("got addr %q, wanted %q", appClient.addr, devClient.addr)
	}

	if appClient.ua != devClient.ua {
		t.Errorf("got ua %q, wanted %q", appClient.ua, devClient.ua)
	}

}
