package bosgo

import (
	"context"
	"encoding/json"
	"net/url"
)

// CredentialsService provides access to credential related API services that also
// require an authenticated developer session.
type CredentialsService struct {
	client *DevClient
}

func NewCredentialsService(c *DevClient) *CredentialsService {
	return &CredentialsService{client: c}
}

// Get returns a request that may be used to get a set of stored credentials.
func (d *CredentialsService) Get(credentialID string) *GetCredentialReq {
	return &GetCredentialReq{
		req: d.client.newReq(apiV1 + "/developers/credentials/" + url.PathEscape(credentialID)),
	}
}

type GetCredentialReq struct {
	req
}

// Context sets the context to be used during this request. If no context is supplied then
// the request will use context.Background.
func (r *GetCredentialReq) Context(ctx context.Context) *GetCredentialReq {
	r.req.ctx = ctx
	return r
}

// ClientID sets a client identifier that will be passed to the Bankrs API in
// the X-Client-Id header.
func (r *GetCredentialReq) ClientID(id string) *GetCredentialReq {
	r.req.clientID = id
	return r
}

func (r *GetCredentialReq) Send() (*Credential, error) {
	res, cleanup, err := r.req.get()
	defer cleanup()
	if err != nil {
		return nil, err
	}

	var cred Credential
	if err := json.NewDecoder(res.Body).Decode(&cred); err != nil {
		return nil, decodeError(err, res)
	}

	return &cred, nil
}

// Delete returns a request that may be used to get a delete a set of stored credentials.
func (d *CredentialsService) Delete(credentialID string) *DeleteCredentialReq {
	return &DeleteCredentialReq{
		req: d.client.newReq(apiV1 + "/developers/credentials/" + url.PathEscape(credentialID)),
	}
}

type DeleteCredentialReq struct {
	req
}

// Context sets the context to be used during this request. If no context is supplied then
// the request will use context.Background.
func (r *DeleteCredentialReq) Context(ctx context.Context) *DeleteCredentialReq {
	r.req.ctx = ctx
	return r
}

// ClientID sets a client identifier that will be passed to the Bankrs API in
// the X-Client-Id header.
func (r *DeleteCredentialReq) ClientID(id string) *DeleteCredentialReq {
	r.req.clientID = id
	return r
}

func (r *DeleteCredentialReq) Send() error {
	_, cleanup, err := r.req.delete(nil)
	defer cleanup()
	if err != nil {
		return err
	}

	return nil
}

// Update returns a request that may be used to update a set of stored credentials.
func (d *CredentialsService) Update(credentialID string, credentials map[string]string) *UpdateCredentialReq {
	return &UpdateCredentialReq{
		req:   d.client.newReq(apiV1 + "/developers/credentials/" + url.PathEscape(credentialID)),
		creds: credentials,
	}
}

type UpdateCredentialReq struct {
	req
	creds map[string]string
}

// Context sets the context to be used during this request. If no context is supplied then
// the request will use context.Background.
func (r *UpdateCredentialReq) Context(ctx context.Context) *UpdateCredentialReq {
	r.req.ctx = ctx
	return r
}

// ClientID sets a client identifier that will be passed to the Bankrs API in
// the X-Client-Id header.
func (r *UpdateCredentialReq) ClientID(id string) *UpdateCredentialReq {
	r.req.clientID = id
	return r
}

func (r *UpdateCredentialReq) Send() error {
	var data = struct {
		Credentials map[string]string `json:"keys"`
	}{
		Credentials: r.creds,
	}

	_, cleanup, err := r.req.putJSON(data)
	defer cleanup()
	if err != nil {
		return err
	}
	return nil
}

// ListProviders returns a request that may be used to get a list of supported providers for
// credential sets.
func (d *CredentialsService) ListProviders() *ListCredentialProvidersReq {
	return &ListCredentialProvidersReq{
		req: d.client.newReq(apiV1 + "/developers/credentials/providers"),
	}
}

type ListCredentialProvidersReq struct {
	req
}

// Context sets the context to be used during this request. If no context is supplied then
// the request will use context.Background.
func (r *ListCredentialProvidersReq) Context(ctx context.Context) *ListCredentialProvidersReq {
	r.req.ctx = ctx
	return r
}

// ClientID sets a client identifier that will be passed to the Bankrs API in
// the X-Client-Id header.
func (r *ListCredentialProvidersReq) ClientID(id string) *ListCredentialProvidersReq {
	r.req.clientID = id
	return r
}

func (r *ListCredentialProvidersReq) Send() (*CredentialProviderPage, error) {
	res, cleanup, err := r.req.get()
	defer cleanup()
	if err != nil {
		return nil, err
	}

	var page CredentialProviderPage
	if err := json.NewDecoder(res.Body).Decode(&page.Providers); err != nil {
		return nil, decodeError(err, res)
	}

	return &page, nil
}
