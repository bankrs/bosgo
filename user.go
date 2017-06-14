package bosgo

import (
	"context"
	"net/http"
)

// UserClient is a client used for interacting with services in the context of
// a registered application and a valid user session. It is safe for
// concurrent use by multiple goroutines.
type UserClient struct {
	// never modified once they have been set
	hc            *http.Client
	addr          string
	token         string // session token
	applicationID string
}

// NewUserClient creates a new user client, ready to use.
func NewUserClient(client *http.Client, addr string, token string, applicationID string) *UserClient {
	uc := &UserClient{
		hc:            client,
		addr:          addr,
		token:         token,
		applicationID: applicationID,
	}

	return uc
}

func (u *UserClient) newReq(path string) req {
	return req{
		hc:   u.hc,
		addr: u.addr,
		path: path,
		headers: headers{
			"x-token":          u.token,
			"x-application-id": u.applicationID,
		},
		par: params{},
	}
}

// Logout returns a request that may be used to log a user out of the Bankrs
// API. Once this request has been sent the user client is no longer valid and
// should not be used.
func (u *UserClient) Logout() *UserLogoutReq {
	return &UserLogoutReq{
		req: u.newReq("/v1/users/logout"),
	}
}

type UserLogoutReq struct {
	req
}

func (r *UserLogoutReq) Context(ctx context.Context) *UserLogoutReq {
	r.req.ctx = ctx
	return r
}

func (r *UserLogoutReq) Send() error {
	_, cleanup, err := r.req.postJSON(nil)
	defer cleanup()
	if err != nil {
		return err
	}
	return nil
}
