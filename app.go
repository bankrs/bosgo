package bosgo

import (
	"context"
	"encoding/json"
	"net/http"
)

// AppClient is a client used for interacting with services in the context of
// a registered application and a valid user or developer session. It is safe
// for concurrent use by multiple goroutines.
type AppClient struct {
	// never modified once they have been set
	hc            *http.Client
	addr          string
	token         string // session token
	applicationID string
}

// NewAppClient creates a new client that may be used to interact with
// services that require a specific application context.
func NewAppClient(client *http.Client, addr string, token string, applicationID string) *AppClient {
	ac := &AppClient{
		hc:            client,
		addr:          addr,
		token:         token,
		applicationID: applicationID,
	}

	return ac
}

// CreateUser returns a request that may be used to create a user with the given username and password.
func (a *AppClient) CreateUser(username, password string) *UserCreateReq {
	return &UserCreateReq{
		client: a,
		req: req{
			hc:   a.hc,
			addr: a.addr,
			path: "/v1/users",
			headers: headers{
				"x-token":          a.token,
				"x-application-id": a.applicationID,
				"Content-Type":     "application/json",
			},
		},
		data: UserCredentials{
			Username: username,
			Password: password,
		},
	}
}

// UserCreateReq is a request that may be used to create a user.
type UserCreateReq struct {
	req
	client *AppClient
	data   UserCredentials
}

type UserCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Context sets the context to be used during this request. If no context is supplied then
// the request will use context.Background.
func (r *UserCreateReq) Context(ctx context.Context) *UserCreateReq {
	r.req.ctx = ctx
	return r
}

// Send sends the request to create the user and returns a client that can be
// used to access services within the new users's session.
func (r *UserCreateReq) Send() (*UserClient, error) {
	res, cleanup, err := r.req.postJSON(r.data)
	defer cleanup()
	if err != nil {
		return nil, err
	}

	var t UserToken
	if err := json.NewDecoder(res.Body).Decode(&t); err != nil {
		return nil, err
	}

	return NewUserClient(r.client.hc, r.client.addr, t.Token, r.client.applicationID), nil
}

type UserToken struct {
	ID    string `json:"id"`    // globally unique identifier for a user
	Token string `json:"token"` // session token
}
