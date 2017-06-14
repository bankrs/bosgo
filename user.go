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

func (u *UserClient) newReq() req {
	return req{
		hc: u.hc,
	}
}

// Logout returns a request that may be used to log a user out of the Bankrs
// API. Once this request has been sent the user client is no longer valid and
// should not be used.
func (u *UserClient) Logout() *UserLogoutReq {
	return &UserLogoutReq{
		req: u.newReq(),
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
	_, cleanup, err := r.req.get()
	defer cleanup()
	if err != nil {
		return err
	}
	return nil
}
