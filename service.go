package bosgo

import (
	"context"
	"encoding/json"
	"net/http"
)

const (
	SandboxAddr    = "api.sandbox.bankrs.com"
	ProductionAddr = "api.bankrs.com"
)

const (
	errContextInvalidServiceResponse = "invalid service response"
)

// Client is the base client used for interacting with services that do not
// require authentication. Use Login to initiate a developer session.  It is
// safe for concurrent use by multiple goroutines.
type Client struct {
	// never modified once they have been set
	hc   *http.Client
	addr string
}

// New creates a new client that will use the supplied HTTP client and connect
// via the specified API host address.
func New(client *http.Client, addr string) *Client {
	c := &Client{
		hc:   client,
		addr: addr,
	}
	return c
}

func (c *Client) newReq(path string) req {
	return req{
		hc:      c.hc,
		addr:    c.addr,
		path:    path,
		headers: headers{},
		par:     params{},
	}
}

// Login prepares and returns a request to log a developer into the Bankrs
// API. Sending a successful request will return a new client that allows
// access to services requiring a valid developer session.
func (c *Client) Login(email, password string) *DeveloperLoginReq {
	return &DeveloperLoginReq{
		client: c,
		req:    c.newReq("/v1/developers/login"),
		data: DeveloperCredentials{
			Email:    email,
			Password: password,
		},
	}
}

type DeveloperLoginReq struct {
	req
	client *Client
	data   DeveloperCredentials
}

type DeveloperCredentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Context sets the context to be used during this request. If no context is supplied then
// the request will use context.Background.
func (r *DeveloperLoginReq) Context(ctx context.Context) *DeveloperLoginReq {
	r.req.ctx = ctx
	return r
}

// Send sends the login request and returns a client that can be used to
// access services within the developer's session.
func (r *DeveloperLoginReq) Send() (*DevClient, error) {
	res, cleanup, err := r.req.postJSON(&r.data)
	defer cleanup()
	if err != nil {
		return nil, err
	}

	var t SessionToken
	if err := json.NewDecoder(res.Body).Decode(&t); err != nil {
		return nil, wrap(errContextInvalidServiceResponse, err)
	}

	return NewDevClient(r.client.hc, r.client.addr, t.Token), nil
}

// CreateDeveloper prepares and returns a request to create a developer account for the
// Bankrs API. Sending a successful request will return a new client that
// allows access to services requiring a valid developer session.
func (c *Client) CreateDeveloper(email, password string) *DeveloperCreateReq {
	return &DeveloperCreateReq{
		client: c,
		req:    c.newReq("/v1/developers"),
		data: DeveloperCredentials{
			Email:    email,
			Password: password,
		},
	}
}

type DeveloperCreateReq struct {
	req
	client *Client
	data   DeveloperCredentials
}

// Context sets the context to be used during this request. If no context is supplied then
// the request will use context.Background.
func (r *DeveloperCreateReq) Context(ctx context.Context) *DeveloperCreateReq {
	r.req.ctx = ctx
	return r
}

// Send sends the create request and returns a client that can be used to
// access services within the developer's session.
func (r *DeveloperCreateReq) Send() (*DevClient, error) {
	res, cleanup, err := r.req.postJSON(&r.data)
	defer cleanup()
	if err != nil {
		return nil, err
	}

	var t SessionToken
	if err := json.NewDecoder(res.Body).Decode(&t); err != nil {
		return nil, wrap(errContextInvalidServiceResponse, err)
	}

	return NewDevClient(r.client.hc, r.client.addr, t.Token), nil
}

type SessionToken struct {
	Token string `json:"token"`
}

// CreateDeveloper prepares and returns a request to start the lost password process.
func (c *Client) LostPassword(email string) *LostPasswordReq {
	return &LostPasswordReq{
		req: c.newReq("/v1/developers/lost_password"),
		data: DeveloperEmail{
			Email: email,
		},
	}
}

type DeveloperEmail struct {
	Email string `json:"email"`
}

type LostPasswordReq struct {
	req
	data DeveloperEmail
}

// Context sets the context to be used during this request. If no context is supplied then
// the request will use context.Background.
func (r *LostPasswordReq) Context(ctx context.Context) *LostPasswordReq {
	r.req.ctx = ctx
	return r
}

// Send sends the lost password request.
func (r *LostPasswordReq) Send() error {
	_, cleanup, err := r.req.postJSON(&r.data)
	defer cleanup()
	if err != nil {
		return err
	}

	return nil
}

// CreateDeveloper prepares and returns a request to start the lost password process.
func (c *Client) ResetPassword(password string, token string) *ResetPasswordReq {
	return &ResetPasswordReq{
		req: c.newReq("/v1/developers/reset_password"),
		data: DeveloperPasswordReset{
			Password: password,
			Token:    token,
		},
	}
}

type DeveloperPasswordReset struct {
	Password string `json:"email"`
	Token    string `json:"token"`
}

type ResetPasswordReq struct {
	req
	data DeveloperPasswordReset
}

// Context sets the context to be used during this request. If no context is supplied then
// the request will use context.Background.
func (r *ResetPasswordReq) Context(ctx context.Context) *ResetPasswordReq {
	r.req.ctx = ctx
	return r
}

// Send sends the reset password request.
func (r *ResetPasswordReq) Send() error {
	_, cleanup, err := r.req.postJSON(&r.data)
	defer cleanup()
	if err != nil {
		return err
	}

	return nil
}
