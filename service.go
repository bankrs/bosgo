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

const (
	SandboxAddr    = "api.sandbox.bankrs.com"
	ProductionAddr = "api.bankrs.com"
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
		return nil, err
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
		return nil, err
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
	fmt.Printf("%+v\n", req)
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

	fmt.Printf("%+v\n", req)

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
	fmt.Printf("%+v\n", req)
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
	fmt.Printf("%+v\n", req)
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

type Error struct {
	Errors     []ErrorItem `json:"errors"`
	Header     http.Header
	StatusCode int
	Status     string
}

type ErrorItem struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *Error) Error() string {
	if len(e.Errors) == 1 {
		if e.Errors[0].Message == "" {
			return fmt.Sprintf("%s (%s)", e.Errors[0].Code, e.Status)
		}
		return fmt.Sprintf("%s: %s (%s)", e.Errors[0].Code, e.Errors[0].Message, e.Status)
	}
	// TODO: expand on error message
	return fmt.Sprintf("request failed with status %s (%+v)", e.Status, *e)
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

	fmt.Printf("body: %s\n", body)

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
