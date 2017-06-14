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

	Categories *CategoriesService
	Providers  *ProvidersService
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

	ac.Categories = NewCategoriesService(ac)
	ac.Providers = NewProvidersService(ac)
	return ac
}

func (a *AppClient) newReq(path string) req {
	return req{
		hc:   a.hc,
		addr: a.addr,
		path: path,
		headers: headers{
			"x-token":          a.token,
			"x-application-id": a.applicationID,
		},
		par: params{},
	}
}

// CreateUser returns a request that may be used to create a user with the given username and password.
func (a *AppClient) CreateUser(username, password string) *UserCreateReq {
	return &UserCreateReq{
		req:    a.newReq("/v1/users"),
		client: a,
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

// UserLogin returns a request that may be used to login a user with the given username and password.
func (a *AppClient) UserLogin(username, password string) *UserLoginReq {
	return &UserLoginReq{
		req:    a.newReq("/v1/users/login"),
		client: a,
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
		return nil, err
	}

	return NewUserClient(r.client.hc, r.client.addr, t.Token, r.client.applicationID), nil
}

type CategoriesService struct {
	client *AppClient
}

func NewCategoriesService(a *AppClient) *CategoriesService { return &CategoriesService{client: a} }

// Categories returns a request that may be used to request a list of classification categories.
func (c *CategoriesService) List() *CategoriesReq {
	return &CategoriesReq{
		req: c.client.newReq("/v1/categories"),
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

// Send sends the request to list categories.
func (r *CategoriesReq) Send() (*CategoryList, error) {
	res, cleanup, err := r.req.get()
	defer cleanup()
	if err != nil {
		return nil, err
	}

	var list CategoryList
	if err := json.NewDecoder(res.Body).Decode(&list); err != nil {
		return nil, err
	}

	return &list, nil
}

type CategoryList []Category

type Category struct {
	ID    int64             `json:"id"`
	Names map[string]string `json:"names"`
	Group string            `json:"group"` // spending or income, e.g.
}

type ProvidersService struct {
	client *AppClient
}

func NewProvidersService(a *AppClient) *ProvidersService { return &ProvidersService{client: a} }

// Categories returns a request that may be used to search the list of financial providers.
func (c *ProvidersService) Search(query string) *ProvidersSearchReq {
	req := c.client.newReq("/v1/providers")
	req.par.Set("q", query)
	return &ProvidersSearchReq{
		req: req,
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

// Send sends the request to search providers.
func (r *ProvidersSearchReq) Send() (*ProviderSearchResults, error) {
	res, cleanup, err := r.req.get()
	defer cleanup()
	if err != nil {
		return nil, err
	}

	var srch ProviderSearchResults
	if err := json.NewDecoder(res.Body).Decode(&srch); err != nil {
		return nil, err
	}

	return &srch, nil
}

// Get returns a request that may be used to get the details of a single financial provider.
func (c *ProvidersService) Get(id string) *ProvidersGetReq {
	return &ProvidersGetReq{
		req: c.client.newReq("/v1/providers/" + id),
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

// Send sends the request to get a single financial provider.
func (r *ProvidersGetReq) Send() (*Provider, error) {
	res, cleanup, err := r.req.get()
	defer cleanup()
	if err != nil {
		return nil, err
	}

	var p Provider
	if err := json.NewDecoder(res.Body).Decode(&p); err != nil {
		return nil, err
	}

	return &p, nil
}

type ProviderSearchResults []ProviderSearchResult

type ProviderSearchResult struct {
	Score    float64  `json:"score"`
	Provider Provider `json:"provider"`
}

type Provider struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Country     string      `json:"country"`
	URL         string      `json:"url"`
	Address     string      `json:"address"`
	PostalCode  string      `json:"postal_code"`
	Challenges  []Challenge `json:"challenges"`
}

type Challenge struct {
	ID          string            `json:"id"`
	Description string            `json:"desc"`
	Type        string            `json:"type"`
	Secure      bool              `json:"secure"`
	UnStoreable bool              `json:"unstoreable"`
	Options     map[string]string `json:"options,omitempty"`
}
