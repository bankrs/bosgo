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
	"context"
	"encoding/json"
	"net/http"
	"net/url"
)

// AppClient is a client used for interacting with services in the context of
// a registered application without a valid user or developer session. It is safe
// for concurrent use by multiple goroutines.
type AppClient struct {
	// never modified once they have been set
	hc            *http.Client
	addr          string
	applicationID string
	ua            string

	Categories *CategoriesService
	Providers  *ProvidersService
	Users      *AppUsersService
	IBAN       *IBANService
}

// NewAppClient creates a new client that may be used to interact with
// services that require a specific application context.
func NewAppClient(client *http.Client, addr string, applicationID string) *AppClient {
	ac := &AppClient{
		hc:            client,
		addr:          addr,
		applicationID: applicationID,
	}

	ac.Categories = NewCategoriesService(ac)
	ac.Providers = NewProvidersService(ac)
	ac.Users = NewAppUsersService(ac)
	ac.IBAN = NewIBANService(ac)
	return ac
}

func (a *AppClient) newReq(path string) req {
	return req{
		hc:   a.hc,
		addr: a.addr,
		path: path,
		headers: headers{
			"User-Agent":       a.userAgent(),
			"x-application-id": a.applicationID,
		},
		par: params{},
	}
}

func (a *AppClient) userAgent() string {
	if a.ua == "" {
		return DefaultUserAgent
	}

	return DefaultUserAgent + " " + a.ua
}

// CategoriesService provides access to category related API services.
type CategoriesService struct {
	client *AppClient
}

func NewCategoriesService(a *AppClient) *CategoriesService { return &CategoriesService{client: a} }

// List returns a request that may be used to request a list of classification categories.
func (c *CategoriesService) List() *CategoriesReq {
	return &CategoriesReq{
		req: c.client.newReq(apiV1 + "/categories"),
	}
}

type CategoriesReq struct {
	req
}

// Context sets the context to be used during this request. If no context is supplied then
// the request will use context.Background.
func (r *CategoriesReq) Context(ctx context.Context) *CategoriesReq {
	r.req.ctx = ctx
	return r
}

// ClientID sets a client identifier that will be passed to the Bankrs API in
// the X-Client-Id header.
func (r *CategoriesReq) ClientID(id string) *CategoriesReq {
	r.req.clientID = id
	return r
}

// Send sends the request to list categories.
func (r *CategoriesReq) Send() (*CategoryList, error) {
	res, cleanup, err := r.req.get()
	defer cleanup()
	if err != nil {
		return nil, err
	}

	var list CategoryList
	if err := json.NewDecoder(res.Body).Decode(&list); err != nil {
		return nil, decodeError(err, res)
	}

	return &list, nil
}

// ProvidersService provides access to financial provider related API services.
type ProvidersService struct {
	client *AppClient
}

func NewProvidersService(a *AppClient) *ProvidersService { return &ProvidersService{client: a} }

// Search returns a request that may be used to search the list of financial providers.
func (c *ProvidersService) Search(query string) *ProvidersSearchReq {
	r := c.client.newReq(apiV1 + "/providers")
	r.par.Set("q", query)
	return &ProvidersSearchReq{
		req: r,
	}
}

type ProvidersSearchReq struct {
	req
}

// Context sets the context to be used during this request. If no context is supplied then
// the request will use context.Background.
func (r *ProvidersSearchReq) Context(ctx context.Context) *ProvidersSearchReq {
	r.req.ctx = ctx
	return r
}

// ClientID sets a client identifier that will be passed to the Bankrs API in
// the X-Client-Id header.
func (r *ProvidersSearchReq) ClientID(id string) *ProvidersSearchReq {
	r.req.clientID = id
	return r
}

// Send sends the request to search providers.
func (r *ProvidersSearchReq) Send() (*ProviderSearchResults, error) {
	res, cleanup, err := r.req.get()
	defer cleanup()
	if err != nil {
		return nil, err
	}

	var srch ProviderSearchResults
	if err := json.NewDecoder(res.Body).Decode(&srch); err != nil {
		return nil, decodeError(err, res)
	}

	return &srch, nil
}

// Get returns a request that may be used to get the details of a single financial provider.
func (c *ProvidersService) Get(id string) *ProvidersGetReq {
	return &ProvidersGetReq{
		req: c.client.newReq(apiV1 + "/providers/" + url.PathEscape(id)),
	}
}

type ProvidersGetReq struct {
	req
}

// Context sets the context to be used during this request. If no context is supplied then
// the request will use context.Background.
func (r *ProvidersGetReq) Context(ctx context.Context) *ProvidersGetReq {
	r.req.ctx = ctx
	return r
}

// ClientID sets a client identifier that will be passed to the Bankrs API in
// the X-Client-Id header.
func (r *ProvidersGetReq) ClientID(id string) *ProvidersGetReq {
	r.req.clientID = id
	return r
}

// Send sends the request to get a single financial provider.
func (r *ProvidersGetReq) Send() (*Provider, error) {
	res, cleanup, err := r.req.get()
	defer cleanup()
	if err != nil {
		return nil, err
	}

	var p Provider
	if err := json.NewDecoder(res.Body).Decode(&p); err != nil {
		return nil, decodeError(err, res)
	}

	return &p, nil
}

// AppUsersService provides access to application user related API services.
type AppUsersService struct {
	client *AppClient
}

func NewAppUsersService(c *AppClient) *AppUsersService { return &AppUsersService{client: c} }

// Create returns a request that may be used to create a user with the given username and password.
func (a *AppUsersService) Create(username, password string) *UserCreateReq {
	return &UserCreateReq{
		req:    a.client.newReq(apiV1 + "/users"),
		client: a.client,
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

// Context sets the context to be used during this request. If no context is supplied then
// the request will use context.Background.
func (r *UserCreateReq) Context(ctx context.Context) *UserCreateReq {
	r.req.ctx = ctx
	return r
}

// ClientID sets a client identifier that will be passed to the Bankrs API in
// the X-Client-Id header.
func (r *UserCreateReq) ClientID(id string) *UserCreateReq {
	r.req.clientID = id
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
		return nil, decodeError(err, res)
	}

	uc := NewUserClient(r.client.hc, r.client.addr, t.Token, r.client.applicationID)
	uc.ua = r.client.ua
	return uc, nil
}

// Login returns a request that may be used to login a user with the given username and password.
func (a *AppUsersService) Login(username, password string) *UserLoginReq {
	return &UserLoginReq{
		req:    a.client.newReq(apiV1 + "/users/login"),
		client: a.client,
		data: UserCredentials{
			Username: username,
			Password: password,
		},
	}
}

type UserLoginReq struct {
	req
	client *AppClient
	data   UserCredentials
}

// Context sets the context to be used during this request. If no context is supplied then
// the request will use context.Background.
func (r *UserLoginReq) Context(ctx context.Context) *UserLoginReq {
	r.req.ctx = ctx
	return r
}

// ClientID sets a client identifier that will be passed to the Bankrs API in
// the X-Client-Id header.
func (r *UserLoginReq) ClientID(id string) *UserLoginReq {
	r.req.clientID = id
	return r
}

// Send sends the request to login the user and returns a client that can be
// used to access services within the new users's session.
func (r *UserLoginReq) Send() (*UserClient, error) {
	res, cleanup, err := r.req.postJSON(r.data)
	defer cleanup()
	if err != nil {
		return nil, err
	}

	var t UserToken
	if err := json.NewDecoder(res.Body).Decode(&t); err != nil {
		return nil, decodeError(err, res)
	}

	uc := NewUserClient(r.client.hc, r.client.addr, t.Token, r.client.applicationID)
	uc.ua = r.client.ua
	return uc, nil
}

// ResetPassword prepares and returns a request to reset a user's password.
func (a *AppUsersService) ResetPassword(username, password string) *ResetUserPasswordReq {
	return &ResetUserPasswordReq{
		req: a.client.newReq(apiV1 + "/users/reset_password"),
		data: UserCredentials{
			Username: username,
			Password: password,
		},
	}
}

type ResetUserPasswordReq struct {
	req
	data UserCredentials
}

// Context sets the context to be used during this request. If no context is supplied then
// the request will use context.Background.
func (r *ResetUserPasswordReq) Context(ctx context.Context) *ResetUserPasswordReq {
	r.req.ctx = ctx
	return r
}

// ClientID sets a client identifier that will be passed to the Bankrs API in
// the X-Client-Id header.
func (r *ResetUserPasswordReq) ClientID(id string) *ResetUserPasswordReq {
	r.req.clientID = id
	return r
}

// Send sends the request to reset a user's password.
func (r *ResetUserPasswordReq) Send() error {
	_, cleanup, err := r.req.postJSON(r.data)
	defer cleanup()
	if err != nil {
		return err
	}
	return nil
}

// IBANService provides access to IBAN related API services.
type IBANService struct {
	client *AppClient
}

func NewIBANService(c *AppClient) *IBANService { return &IBANService{client: c} }

// Validate returns a request that may be used to validate an IBAN.
func (a *IBANService) Validate(iban string) *ValidateIBANReq {
	return &ValidateIBANReq{
		req:    a.client.newReq(apiV1 + "/iban/" + url.PathEscape(iban)),
		client: a.client,
	}
}

// ValidateIBANReq is a request that may be used to validate an IBAN.
type ValidateIBANReq struct {
	req
	client *AppClient
}

// Context sets the context to be used during this request. If no context is supplied then
// the request will use context.Background.
func (r *ValidateIBANReq) Context(ctx context.Context) *ValidateIBANReq {
	r.req.ctx = ctx
	return r
}

// ClientID sets a client identifier that will be passed to the Bankrs API in
// the X-Client-Id header.
func (r *ValidateIBANReq) ClientID(id string) *ValidateIBANReq {
	r.req.clientID = id
	return r
}

// Send sends the request to validate the IBAN and returns details about the IBAN.
func (r *ValidateIBANReq) Send() (*IBANDetails, error) {
	res, cleanup, err := r.req.get()
	defer cleanup()
	if err != nil {
		return nil, err
	}

	var id IBANDetails
	if err := json.NewDecoder(res.Body).Decode(&id); err != nil {
		return nil, decodeError(err, res)
	}

	return &id, nil
}
