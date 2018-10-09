package bosgo

import (
	"context"
	"encoding/json"
	"net/url"
)

// WebhooksService provides access to webhook related API services.
type WebhooksService struct {
	client *DevClient
}

func NewWebhooksService(c *DevClient) *WebhooksService { return &WebhooksService{client: c} }

// Create prepares and returns a request to create a new webhook.
func (d *WebhooksService) Create(apiVersion int, url string, events []string) *CreateWebhookReq {
	return &CreateWebhookReq{
		req: d.client.newReq(apiV1 + "/webhooks"),
		data: createWebhookParams{
			URL:        url,
			Events:     events,
			APIVersion: apiVersion,
		},
	}
}

type createWebhookParams struct {
	URL        string   `json:"url"`
	Events     []string `json:"events"`
	APIVersion int      `json:"api_version"`
}

type CreateWebhookReq struct {
	req
	data createWebhookParams
}

// Context sets the context to be used during this request. If no context is supplied then
// the request will use context.Background.
func (r *CreateWebhookReq) Context(ctx context.Context) *CreateWebhookReq {
	r.req.ctx = ctx
	return r
}

// ClientID sets a client identifier that will be passed to the Bankrs API in
// the X-Client-Id header.
func (r *CreateWebhookReq) ClientID(id string) *CreateWebhookReq {
	r.req.clientID = id
	return r
}

func (r *CreateWebhookReq) Send() (string, error) {
	res, cleanup, err := r.req.postJSON(r.data)
	defer cleanup()
	if err != nil {
		return "", err
	}

	var id struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(res.Body).Decode(&id); err != nil {
		return "", decodeError(err, res)
	}

	return id.ID, nil
}

// Get prepares and returns a request to get details of an existing webhook.
func (d *WebhooksService) Get(id string) *GetWebhookReq {
	return &GetWebhookReq{
		req: d.client.newReq(apiV1 + "/webhooks/" + url.PathEscape(id)),
	}
}

type GetWebhookReq struct {
	req
}

// Context sets the context to be used during this request. If no context is supplied then
// the request will use context.Background.
func (r *GetWebhookReq) Context(ctx context.Context) *GetWebhookReq {
	r.req.ctx = ctx
	return r
}

// ClientID sets a client identifier that will be passed to the Bankrs API in
// the X-Client-Id header.
func (r *GetWebhookReq) ClientID(id string) *GetWebhookReq {
	r.req.clientID = id
	return r
}

func (r *GetWebhookReq) Send() (*Webhook, error) {
	res, cleanup, err := r.req.get()
	defer cleanup()
	if err != nil {
		return nil, err
	}

	var wh Webhook
	if err := json.NewDecoder(res.Body).Decode(&wh); err != nil {
		return nil, decodeError(err, res)
	}

	return &wh, nil
}

// List prepares and returns a request to list details of all webhooks.
func (d *WebhooksService) List() *ListWebhookReq {
	return &ListWebhookReq{
		req: d.client.newReq(apiV1 + "/webhooks"),
	}
}

type ListWebhookReq struct {
	req
}

// Context sets the context to be used during this request. If no context is supplied then
// the request will use context.Background.
func (r *ListWebhookReq) Context(ctx context.Context) *ListWebhookReq {
	r.req.ctx = ctx
	return r
}

// ClientID sets a client identifier that will be passed to the Bankrs API in
// the X-Client-Id header.
func (r *ListWebhookReq) ClientID(id string) *ListWebhookReq {
	r.req.clientID = id
	return r
}

func (r *ListWebhookReq) Send() (*WebhookPage, error) {
	res, cleanup, err := r.req.get()
	defer cleanup()
	if err != nil {
		return nil, err
	}

	var page WebhookPage
	if err := json.NewDecoder(res.Body).Decode(&page.Webhooks); err != nil {
		return nil, decodeError(err, res)
	}

	return &page, nil
}

// Update prepares and returns a request to update an existing webhook.
func (d *WebhooksService) Update(id string, apiVersion int, u string, events []string) *UpdateWebhookReq {
	return &UpdateWebhookReq{
		req: d.client.newReq(apiV1 + "/webhooks/" + url.PathEscape(id)),
		data: UpdateWebhookParams{
			URL:        u,
			Events:     events,
			APIVersion: apiVersion,
		},
	}
}

type UpdateWebhookParams struct {
	URL        string   `json:"url"`
	Events     []string `json:"events"`
	APIVersion int      `json:"api_version"`
}

type UpdateWebhookReq struct {
	req
	data UpdateWebhookParams
}

// Context sets the context to be used during this request. If no context is supplied then
// the request will use context.Background.
func (r *UpdateWebhookReq) Context(ctx context.Context) *UpdateWebhookReq {
	r.req.ctx = ctx
	return r
}

// ClientID sets a client identifier that will be passed to the Bankrs API in
// the X-Client-Id header.
func (r *UpdateWebhookReq) ClientID(id string) *UpdateWebhookReq {
	r.req.clientID = id
	return r
}

func (r *UpdateWebhookReq) Send() error {
	_, cleanup, err := r.req.putJSON(r.data)
	defer cleanup()
	if err != nil {
		return err
	}

	return nil
}

// Delete prepares and returns a request to delete an existing webhook.
func (d *WebhooksService) Delete(id string) *DeleteWebhookReq {
	return &DeleteWebhookReq{
		req: d.client.newReq(apiV1 + "/webhooks/" + url.PathEscape(id)),
	}
}

type DeleteWebhookReq struct {
	req
}

// Context sets the context to be used during this request. If no context is supplied then
// the request will use context.Background.
func (r *DeleteWebhookReq) Context(ctx context.Context) *DeleteWebhookReq {
	r.req.ctx = ctx
	return r
}

// ClientID sets a client identifier that will be passed to the Bankrs API in
// the X-Client-Id header.
func (r *DeleteWebhookReq) ClientID(id string) *DeleteWebhookReq {
	r.req.clientID = id
	return r
}

func (r *DeleteWebhookReq) Send() error {
	_, cleanup, err := r.req.delete(nil)
	defer cleanup()
	if err != nil {
		return err
	}

	return nil
}

// Test prepares and returns a request to test a webhook.
func (d *WebhooksService) Test(id string, event string) *TestWebhookReq {
	return &TestWebhookReq{
		req: d.client.newReq(apiV1 + "/webhooks/" + url.PathEscape(id)),
		data: testWebhookParams{
			Event: event,
		},
	}
}

type testWebhookParams struct {
	Event string `json:"event"`
}

type TestWebhookReq struct {
	req
	data testWebhookParams
}

// Context sets the context to be used during this request. If no context is supplied then
// the request will use context.Background.
func (r *TestWebhookReq) Context(ctx context.Context) *TestWebhookReq {
	r.req.ctx = ctx
	return r
}

// ClientID sets a client identifier that will be passed to the Bankrs API in
// the X-Client-Id header.
func (r *TestWebhookReq) ClientID(id string) *TestWebhookReq {
	r.req.clientID = id
	return r
}

func (r *TestWebhookReq) Send() (*WebhookTestResult, error) {
	res, cleanup, err := r.req.postJSON(r.data)
	defer cleanup()
	if err != nil {
		return nil, err
	}

	var testResponse WebhookTestResult
	if err := json.NewDecoder(res.Body).Decode(&testResponse); err != nil {
		return nil, decodeError(err, res)
	}

	return &testResponse, nil
}
