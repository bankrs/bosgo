package bosgo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// DevClient is a client used for interacting with services that require a
// valid developer session. It is safe for concurrent use by multiple goroutines.
type DevClient struct {
	// never modified once they have been set
	hc    *http.Client
	addr  string
	token string // session token

	Applications *ApplicationsService
	Users        *DevUsersService
}

// NewDevClient creates a new developer client, ready to use.
func NewDevClient(client *http.Client, addr string, token string) *DevClient {
	dc := &DevClient{
		hc:    client,
		addr:  addr,
		token: token,
	}
	dc.Applications = NewApplicationsService(dc)
	dc.Users = NewDevUsersService(dc)

	return dc
}

// WithApplication returns a new client that may be used to interact with
// services that require a specific application context.
func (d *DevClient) WithApplication(applicationID string) *AppClient {
	return NewAppClient(d.hc, d.addr, d.token, applicationID)
}

func (d *DevClient) newReq(path string) req {
	return req{
		hc:   d.hc,
		addr: d.addr,
		path: path,
		headers: headers{
			"x-token": d.token,
		},
	}
}

// Logout prepares and returns a request to log a developer out of the Bankrs
// API. Once this request has been sent the client is no longer valid and
// should not be used.
func (d *DevClient) Logout() *DeveloperLogoutReq {
	return &DeveloperLogoutReq{
		req: d.newReq("/v1/developers/logout"),
	}
}

type DeveloperLogoutReq struct {
	req
}

func (r *DeveloperLogoutReq) Context(ctx context.Context) *DeveloperLogoutReq {
	r.req.ctx = ctx
	return r
}

// Send sends the request to log the developer out and end the session. Once
// this request has been sent the developer client should not be used again.
func (r *DeveloperLogoutReq) Send() error {
	_, cleanup, err := r.req.getJSON()
	defer cleanup()
	if err != nil {
		return err
	}
	return nil
}

type ApplicationsService struct {
	client *DevClient
}

func NewApplicationsService(c *DevClient) *ApplicationsService { return &ApplicationsService{client: c} }

func (d *ApplicationsService) List() *ListApplicationsReq {
	return &ListApplicationsReq{
		req: d.client.newReq("/v1/developers/applications"),
	}
}

type ListApplicationsReq struct {
	req
}

func (r *ListApplicationsReq) Context(ctx context.Context) *ListApplicationsReq {
	r.req.ctx = ctx
	return r
}

func (r *ListApplicationsReq) Send() (*ApplicationListPage, error) {
	res, cleanup, err := r.req.getJSON()
	defer cleanup()
	if err != nil {
		return nil, err
	}

	var page ApplicationListPage
	if err := json.NewDecoder(res.Body).Decode(&page.Applications); err != nil {
		return nil, err
	}
	fmt.Printf("%+v\n", page)

	return &page, nil
}

type ApplicationListPage struct {
	Applications []ApplicationMetadata
}

type ApplicationMetadata struct {
	ApplicationID string    `json:"application_id,omitempty"`
	Label         string    `json:"label,omitempty"`
	CreatedAt     time.Time `json:"created_at,omitempty"`
}

func (d *ApplicationsService) Create(label string) *CreateApplicationsReq {
	return &CreateApplicationsReq{
		req: d.client.newReq("/v1/developers/applications"),
		data: ApplicationMetadata{
			Label: label,
		},
	}
}

type CreateApplicationsReq struct {
	req
	data ApplicationMetadata
}

func (r *CreateApplicationsReq) Context(ctx context.Context) *CreateApplicationsReq {
	r.req.ctx = ctx
	return r
}

func (r *CreateApplicationsReq) Send() (string, error) {
	res, cleanup, err := r.req.postJSON(r.data)
	defer cleanup()
	if err != nil {
		return "", err
	}

	var car CreateApplicationsResponse
	if err := json.NewDecoder(res.Body).Decode(&car); err != nil {
		return "", err
	}

	return car.ApplicationID, nil
}

type CreateApplicationsResponse struct {
	ApplicationID string `json:"application_id"`
}

func (d *ApplicationsService) Delete(applicationID string) *DeleteApplicationsReq {
	return &DeleteApplicationsReq{
		req:           d.client.newReq("/v1/developers/applications/" + applicationID),
		applicationID: applicationID,
	}
}

type DeleteApplicationsReq struct {
	req
	applicationID string
}

func (r *DeleteApplicationsReq) Context(ctx context.Context) *DeleteApplicationsReq {
	r.req.ctx = ctx
	return r
}

func (r *DeleteApplicationsReq) Send() (string, error) {
	_, cleanup, err := r.req.postJSON(nil)
	defer cleanup()
	if err != nil {
		return "", err
	}

	// TODO: parse the applications id
	return "", nil
}

type DevUsersService struct {
	client *DevClient
}

func NewDevUsersService(c *DevClient) *DevUsersService { return &DevUsersService{client: c} }

func (d *DevUsersService) List() *ListDevUsersReq {
	return &ListDevUsersReq{
		req: d.client.newReq("/v1/developers/users"),
	}
}

type ListDevUsersReq struct {
	req
}

func (r *ListDevUsersReq) Context(ctx context.Context) *ListDevUsersReq {
	r.req.ctx = ctx
	return r
}

func (r *ListDevUsersReq) Send() (*UserListPage, error) {
	_, cleanup, err := r.req.getJSON()
	defer cleanup()
	if err != nil {
		return nil, err
	}

	// TODO: parse the user list
	return nil, nil
}

type UserListPage struct {
}
