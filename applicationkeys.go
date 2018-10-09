package bosgo

import (
	"context"
	"net/url"
)

// ApplicationKeysService provides access to application key related API services that also
// require an authenticated developer session.
type ApplicationKeysService struct {
	client *DevClient
}

func NewApplicationKeysService(c *DevClient) *ApplicationKeysService {
	return &ApplicationKeysService{client: c}
}

// Delete returns a request that may be used to remove the specified key from application.
func (d *ApplicationKeysService) Delete(applicationKey string) *DeleteAppKeyReq {
	return &DeleteAppKeyReq{
		req: d.client.newReq(apiV1 + "/developers/application_keys/" + url.PathEscape(applicationKey)),
	}
}

type DeleteAppKeyReq struct {
	req
}

// Context sets the context to be used during this request. If no context is supplied then
// the request will use context.Background.
func (r *DeleteAppKeyReq) Context(ctx context.Context) *DeleteAppKeyReq {
	r.req.ctx = ctx
	return r
}

// ClientID sets a client identifier that will be passed to the Bankrs API in
// the X-Client-Id header.
func (r *DeleteAppKeyReq) ClientID(id string) *DeleteAppKeyReq {
	r.req.clientID = id
	return r
}

func (r *DeleteAppKeyReq) Send() error {
	_, cleanup, err := r.req.delete(nil)
	defer cleanup()
	if err != nil {
		return err
	}

	return nil
}
