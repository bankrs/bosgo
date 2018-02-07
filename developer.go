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
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// DevClient is a client used for interacting with services that require a
// valid developer session. It is safe for concurrent use by multiple goroutines.
type DevClient struct {
	// never modified once they have been set
	hc    *http.Client
	addr  string
	token string // session token
	ua    string

	Applications *ApplicationsService
	Stats        *StatsService
}

// NewDevClient creates a new developer client, ready to use.
func NewDevClient(client *http.Client, addr string, token string) *DevClient {
	dc := &DevClient{
		hc:    client,
		addr:  addr,
		token: token,
	}
	dc.Applications = NewApplicationsService(dc)
	dc.Stats = NewStatsService(dc)

	return dc
}

func (d *DevClient) userAgent() string {
	if d.ua == "" {
		return DefaultUserAgent
	}

	return DefaultUserAgent + " " + d.ua
}

// SessionToken returns the current session token.
func (d *DevClient) SessionToken() string {
	return d.token
}

func (d *DevClient) newReq(path string) req {
	return req{
		hc:   d.hc,
		addr: d.addr,
		path: path,
		headers: headers{
			"User-Agent": d.userAgent(),
			"x-token":    d.token,
		},
		par: params{},
	}
}

// Logout prepares and returns a request to log a developer out of the Bankrs
// API. Once this request has been sent the client is no longer valid and
// should not be used.
func (d *DevClient) Logout() *DeveloperLogoutReq {
	return &DeveloperLogoutReq{
		req: d.newReq(apiV1 + "/developers/logout"),
	}
}

type DeveloperLogoutReq struct {
	req
}

// Context sets the context to be used during this request. If no context is supplied then
// the request will use context.Background.
func (r *DeveloperLogoutReq) Context(ctx context.Context) *DeveloperLogoutReq {
	r.req.ctx = ctx
	return r
}

// ClientID sets a client identifier that will be passed to the Bankrs API in
// the X-Client-Id header.
func (r *DeveloperLogoutReq) ClientID(id string) *DeveloperLogoutReq {
	r.req.clientID = id
	return r
}

// Send sends the request to log the developer out and end the session. Once
// this request has been sent the developer client should not be used again.
func (r *DeveloperLogoutReq) Send() error {
	_, cleanup, err := r.req.postJSON(nil)
	defer cleanup()
	if err != nil {
		return err
	}
	return nil
}

// Delete prepares and returns a request to delete the developer account and
// all it's associated data in all environments. Once this request has been
// sent the client is no longer valid and should not be used.
func (d *DevClient) Delete() *DeveloperDeleteReq {
	return &DeveloperDeleteReq{
		req: d.newReq(apiV1 + "/developers"),
	}
}

type DeveloperDeleteReq struct {
	req
}

// Context sets the context to be used during this request. If no context is supplied then
// the request will use context.Background.
func (r *DeveloperDeleteReq) Context(ctx context.Context) *DeveloperDeleteReq {
	r.req.ctx = ctx
	return r
}

// ClientID sets a client identifier that will be passed to the Bankrs API in
// the X-Client-Id header.
func (r *DeveloperDeleteReq) ClientID(id string) *DeveloperDeleteReq {
	r.req.clientID = id
	return r
}

// Send sends the request to delete developer. Once this request has been sent
// the developer client should not be used again.
func (r *DeveloperDeleteReq) Send() error {
	_, cleanup, err := r.req.delete(nil)
	defer cleanup()
	if err != nil {
		return err
	}
	return nil
}

// ChangePassword prepares and returns a request to change a developer's
// password.
func (d *DevClient) ChangePassword(old, new string) *DeveloperChangePasswordReq {
	return &DeveloperChangePasswordReq{
		req: d.newReq(apiV1 + "/developers/password"),
		data: developerChangePasswordData{
			OldPassword: old,
			NewPassword: new,
		},
	}
}

type developerChangePasswordData struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

type DeveloperChangePasswordReq struct {
	req
	data developerChangePasswordData
}

// Context sets the context to be used during this request. If no context is supplied then
// the request will use context.Background.
func (r *DeveloperChangePasswordReq) Context(ctx context.Context) *DeveloperChangePasswordReq {
	r.req.ctx = ctx
	return r
}

// ClientID sets a client identifier that will be passed to the Bankrs API in
// the X-Client-Id header.
func (r *DeveloperChangePasswordReq) ClientID(id string) *DeveloperChangePasswordReq {
	r.req.clientID = id
	return r
}

// Send sends the request to change the developer's password.
func (r *DeveloperChangePasswordReq) Send() error {
	_, cleanup, err := r.req.postJSON(r.data)
	defer cleanup()
	if err != nil {
		return err
	}
	return nil
}

// Profile retrieves the developer's profile.
func (d *DevClient) Profile() *DeveloperProfileReq {
	return &DeveloperProfileReq{
		req: d.newReq(apiV1 + "/developers/profile"),
	}
}

type DeveloperProfileReq struct {
	req
}

// Context sets the context to be used during this request. If no context is supplied then
// the request will use context.Background.
func (r *DeveloperProfileReq) Context(ctx context.Context) *DeveloperProfileReq {
	r.req.ctx = ctx
	return r
}

// ClientID sets a client identifier that will be passed to the Bankrs API in
// the X-Client-Id header.
func (r *DeveloperProfileReq) ClientID(id string) *DeveloperProfileReq {
	r.req.clientID = id
	return r
}

// Send sends the request to retrieve the developer's profile.
func (r *DeveloperProfileReq) Send() (*DeveloperProfile, error) {
	res, cleanup, err := r.req.get()
	defer cleanup()
	if err != nil {
		return nil, err
	}
	var profile DeveloperProfile
	if err := json.NewDecoder(res.Body).Decode(&profile); err != nil {
		return nil, decodeError(err, res)
	}

	return &profile, nil
}

// SetProfile sets the developer's profile.
func (d *DevClient) SetProfile(profile *DeveloperProfile) *DeveloperSetProfileReq {
	return &DeveloperSetProfileReq{
		req:  d.newReq(apiV1 + "/developers/profile"),
		data: *profile,
	}
}

type DeveloperSetProfileReq struct {
	req
	data DeveloperProfile
}

// Context sets the context to be used during this request. If no context is supplied then
// the request will use context.Background.
func (r *DeveloperSetProfileReq) Context(ctx context.Context) *DeveloperSetProfileReq {
	r.req.ctx = ctx
	return r
}

// ClientID sets a client identifier that will be passed to the Bankrs API in
// the X-Client-Id header.
func (r *DeveloperSetProfileReq) ClientID(id string) *DeveloperSetProfileReq {
	r.req.clientID = id
	return r
}

// Send sends the request to retrieve the developer's profile.
func (r *DeveloperSetProfileReq) Send() error {
	_, cleanup, err := r.req.putJSON(r.data)
	defer cleanup()
	if err != nil {
		return err
	}
	return nil
}

// ApplicationsService provides access to application related API services that also require an authenticated
// developer session.
type ApplicationsService struct {
	client *DevClient
}

func NewApplicationsService(c *DevClient) *ApplicationsService { return &ApplicationsService{client: c} }

func (d *ApplicationsService) List() *ListApplicationsReq {
	return &ListApplicationsReq{
		req: d.client.newReq(apiV1 + "/developers/applications"),
	}
}

type ListApplicationsReq struct {
	req
}

// Context sets the context to be used during this request. If no context is supplied then
// the request will use context.Background.
func (r *ListApplicationsReq) Context(ctx context.Context) *ListApplicationsReq {
	r.req.ctx = ctx
	return r
}

// ClientID sets a client identifier that will be passed to the Bankrs API in
// the X-Client-Id header.
func (r *ListApplicationsReq) ClientID(id string) *ListApplicationsReq {
	r.req.clientID = id
	return r
}

func (r *ListApplicationsReq) Send() (*ApplicationPage, error) {
	res, cleanup, err := r.req.get()
	defer cleanup()
	if err != nil {
		return nil, err
	}

	var page ApplicationPage
	if err := json.NewDecoder(res.Body).Decode(&page.Applications); err != nil {
		return nil, decodeError(err, res)
	}

	return &page, nil
}

func (d *ApplicationsService) Create(label string) *CreateApplicationsReq {
	return &CreateApplicationsReq{
		req: d.client.newReq(apiV1 + "/developers/applications"),
		data: ApplicationMetadata{
			Label: label,
		},
	}
}

type CreateApplicationsReq struct {
	req
	data ApplicationMetadata
}

// Context sets the context to be used during this request. If no context is supplied then
// the request will use context.Background.
func (r *CreateApplicationsReq) Context(ctx context.Context) *CreateApplicationsReq {
	r.req.ctx = ctx
	return r
}

// ClientID sets a client identifier that will be passed to the Bankrs API in
// the X-Client-Id header.
func (r *CreateApplicationsReq) ClientID(id string) *CreateApplicationsReq {
	r.req.clientID = id
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
		return "", decodeError(err, res)
	}

	return car.ApplicationID, nil
}

type CreateApplicationsResponse struct {
	ApplicationID string `json:"application_id"`
}

func (d *ApplicationsService) Update(applicationID string, label string) *UpdateApplicationReq {
	return &UpdateApplicationReq{
		req: d.client.newReq(apiV1 + "/developers/applications/" + url.PathEscape(applicationID)),
		data: ApplicationMetadata{
			Label: label,
		},
	}
}

type UpdateApplicationReq struct {
	req
	data ApplicationMetadata
}

// Context sets the context to be used during this request. If no context is supplied then
// the request will use context.Background.
func (r *UpdateApplicationReq) Context(ctx context.Context) *UpdateApplicationReq {
	r.req.ctx = ctx
	return r
}

// ClientID sets a client identifier that will be passed to the Bankrs API in
// the X-Client-Id header.
func (r *UpdateApplicationReq) ClientID(id string) *UpdateApplicationReq {
	r.req.clientID = id
	return r
}

func (r *UpdateApplicationReq) Send() error {
	_, cleanup, err := r.req.putJSON(r.data)
	defer cleanup()
	if err != nil {
		return err
	}

	return nil
}

func (d *ApplicationsService) Delete(applicationID string) *DeleteApplicationsReq {
	return &DeleteApplicationsReq{
		req: d.client.newReq(apiV1 + "/developers/applications/" + url.PathEscape(applicationID)),
	}
}

type DeleteApplicationsReq struct {
	req
}

// Context sets the context to be used during this request. If no context is supplied then
// the request will use context.Background.
func (r *DeleteApplicationsReq) Context(ctx context.Context) *DeleteApplicationsReq {
	r.req.ctx = ctx
	return r
}

// ClientID sets a client identifier that will be passed to the Bankrs API in
// the X-Client-Id header.
func (r *DeleteApplicationsReq) ClientID(id string) *DeleteApplicationsReq {
	r.req.clientID = id
	return r
}

func (r *DeleteApplicationsReq) Send() error {
	_, cleanup, err := r.req.delete(nil)
	defer cleanup()
	if err != nil {
		return err
	}

	return nil
}

func (d *ApplicationsService) ListUsers(applicationID string) *ListDevUsersReq {
	r := d.client.newReq(apiV1 + "/developers/users")
	r.headers["x-application-id"] = applicationID
	return &ListDevUsersReq{
		req: r,
	}
}

type ListDevUsersReq struct {
	req
	data PageParams
}

type PageParams struct {
	Cursor string `json:"cursor"`
	Limit  int    `json:"limit"`
}

// Context sets the context to be used during this request. If no context is supplied then
// the request will use context.Background.
func (r *ListDevUsersReq) Context(ctx context.Context) *ListDevUsersReq {
	r.req.ctx = ctx
	return r
}

// ClientID sets a client identifier that will be passed to the Bankrs API in
// the X-Client-Id header.
func (r *ListDevUsersReq) ClientID(id string) *ListDevUsersReq {
	r.req.clientID = id
	return r
}

func (r *ListDevUsersReq) Cursor(cursor string) *ListDevUsersReq {
	r.data.Cursor = cursor
	return r
}

func (r *ListDevUsersReq) Limit(v int) *ListDevUsersReq {
	r.data.Limit = v
	return r
}

func (r *ListDevUsersReq) Send() (*UserListPage, error) {
	if r.data.Limit < 0 {
		return nil, fmt.Errorf("limit must be non-negative")
	}

	var res *http.Response
	var cleanup func()
	var err error
	if r.data.Limit == 0 {
		res, cleanup, err = r.req.get()
	} else {
		res, cleanup, err = r.req.postJSON(r.data)
	}
	defer cleanup()
	if err != nil {
		return nil, err
	}

	var list UserListPage
	if err := json.NewDecoder(res.Body).Decode(&list); err != nil {
		return nil, decodeError(err, res)
	}
	return &list, nil
}

// UserInfo prepares and returns a request to lookup information about a user.
func (d *ApplicationsService) UserInfo(applicationID, id string) *DevUserInfoReq {
	r := d.client.newReq(apiV1 + "/developers/users/" + url.PathEscape(id))
	r.headers["x-application-id"] = applicationID
	return &DevUserInfoReq{
		req: r,
	}
}

type DevUserInfoReq struct {
	req
}

// Context sets the context to be used during this request. If no context is supplied then
// the request will use context.Background.
func (r *DevUserInfoReq) Context(ctx context.Context) *DevUserInfoReq {
	r.req.ctx = ctx
	return r
}

// ClientID sets a client identifier that will be passed to the Bankrs API in
// the X-Client-Id header.
func (r *DevUserInfoReq) ClientID(id string) *DevUserInfoReq {
	r.req.clientID = id
	return r
}

func (r *DevUserInfoReq) Send() (*DevUserInfo, error) {
	res, cleanup, err := r.req.get()
	defer cleanup()

	if err != nil {
		return nil, err
	}

	var info DevUserInfo
	if err := json.NewDecoder(res.Body).Decode(&info); err != nil {
		return nil, decodeError(err, res)
	}
	return &info, nil
}

// ResetUsers prepares and returns a request to reset user data.
func (d *ApplicationsService) ResetUsers(applicationID string, usernames []string) *ResetDevUsersReq {
	r := d.client.newReq(apiV1 + "/developers/users/reset")
	r.headers["x-application-id"] = applicationID
	return &ResetDevUsersReq{
		req:       r,
		usernames: usernames,
	}
}

type ResetDevUsersReq struct {
	req
	usernames []string
}

// Context sets the context to be used during this request. If no context is supplied then
// the request will use context.Background.
func (r *ResetDevUsersReq) Context(ctx context.Context) *ResetDevUsersReq {
	r.req.ctx = ctx
	return r
}

// ClientID sets a client identifier that will be passed to the Bankrs API in
// the X-Client-Id header.
func (r *ResetDevUsersReq) ClientID(id string) *ResetDevUsersReq {
	r.req.clientID = id
	return r
}

// Send sends the request to reset user data.
func (r *ResetDevUsersReq) Send() (*ResetUsersResponse, error) {
	data := struct {
		Usernames []string `json:"usernames"`
	}{
		Usernames: r.usernames,
	}

	res, cleanup, err := r.req.postJSON(data)
	defer cleanup()
	if err != nil {
		return nil, err
	}

	var users ResetUsersResponse
	if err := json.NewDecoder(res.Body).Decode(&users); err != nil {
		return nil, decodeError(err, res)
	}

	return &users, nil
}

// StatsService provides access to statistic related API services.
type StatsService struct {
	client *DevClient
}

func NewStatsService(c *DevClient) *StatsService { return &StatsService{client: c} }

func (d *StatsService) Merchants() *StatsMerchantsReq {
	return &StatsMerchantsReq{
		req: d.client.newReq(apiV1 + "/stats/merchants"),
	}
}

type StatsMerchantsReq struct {
	req
}

// Context sets the context to be used during this request. If no context is supplied then
// the request will use context.Background.
func (r *StatsMerchantsReq) Context(ctx context.Context) *StatsMerchantsReq {
	r.req.ctx = ctx
	return r
}

// ClientID sets a client identifier that will be passed to the Bankrs API in
// the X-Client-Id header.
func (r *StatsMerchantsReq) ClientID(id string) *StatsMerchantsReq {
	r.req.clientID = id
	return r
}

func (r *StatsMerchantsReq) FromDate(date time.Time) *StatsMerchantsReq {
	r.req.par.Set("from_date", date.Format("2006-01-02"))
	return r
}

func (r *StatsMerchantsReq) ToDate(date time.Time) *StatsMerchantsReq {
	r.req.par.Set("to_date", date.Format("2006-01-02"))
	return r
}

func (r *StatsMerchantsReq) Send() (*MerchantsStats, error) {
	// TODO: remove environment parameter
	r.req.par.Set("environment", "sandbox")

	res, cleanup, err := r.req.get()
	defer cleanup()
	if err != nil {
		return nil, err
	}

	var stats MerchantsStats
	if err := json.NewDecoder(res.Body).Decode(&stats); err != nil {
		return nil, decodeError(err, res)
	}

	return &stats, nil
}

func (d *StatsService) Providers() *StatsProvidersReq {
	return &StatsProvidersReq{
		req: d.client.newReq(apiV1 + "/stats/providers"),
	}
}

type StatsProvidersReq struct {
	req
}

// Context sets the context to be used during this request. If no context is supplied then
// the request will use context.Background.
func (r *StatsProvidersReq) Context(ctx context.Context) *StatsProvidersReq {
	r.req.ctx = ctx
	return r
}

// ClientID sets a client identifier that will be passed to the Bankrs API in
// the X-Client-Id header.
func (r *StatsProvidersReq) ClientID(id string) *StatsProvidersReq {
	r.req.clientID = id
	return r
}

func (r *StatsProvidersReq) FromDate(date time.Time) *StatsProvidersReq {
	r.req.par.Set("from_date", date.Format("2006-01-02"))
	return r
}

func (r *StatsProvidersReq) ToDate(date time.Time) *StatsProvidersReq {
	r.req.par.Set("to_date", date.Format("2006-01-02"))
	return r
}

func (r *StatsProvidersReq) Send() (*ProvidersStats, error) {
	// TODO: remove environment parameter
	r.req.par.Set("environment", "sandbox")

	res, cleanup, err := r.req.get()
	defer cleanup()
	if err != nil {
		return nil, err
	}

	var stats ProvidersStats
	if err := json.NewDecoder(res.Body).Decode(&stats); err != nil {
		return nil, decodeError(err, res)
	}

	return &stats, nil
}

func (d *StatsService) Transfers() *StatsTransfersReq {
	return &StatsTransfersReq{
		req: d.client.newReq(apiV1 + "/stats/transfers"),
	}
}

type StatsTransfersReq struct {
	req
}

// Context sets the context to be used during this request. If no context is supplied then
// the request will use context.Background.
func (r *StatsTransfersReq) Context(ctx context.Context) *StatsTransfersReq {
	r.req.ctx = ctx
	return r
}

// ClientID sets a client identifier that will be passed to the Bankrs API in
// the X-Client-Id header.
func (r *StatsTransfersReq) ClientID(id string) *StatsTransfersReq {
	r.req.clientID = id
	return r
}

func (r *StatsTransfersReq) FromDate(date time.Time) *StatsTransfersReq {
	r.req.par.Set("from_date", date.Format("2006-01-02"))
	return r
}

func (r *StatsTransfersReq) ToDate(date time.Time) *StatsTransfersReq {
	r.req.par.Set("to_date", date.Format("2006-01-02"))
	return r
}

func (r *StatsTransfersReq) Send() (interface{}, error) {
	// TODO: remove environment parameter
	r.req.par.Set("environment", "sandbox")

	res, cleanup, err := r.req.get()
	defer cleanup()
	if err != nil {
		return nil, err
	}

	var stats interface{}
	if err := json.NewDecoder(res.Body).Decode(&stats); err != nil {
		return nil, decodeError(err, res)
	}

	fmt.Printf("%+v\n", stats)

	return stats, nil
}

func (d *StatsService) Users() *StatsUsersReq {
	return &StatsUsersReq{
		req: d.client.newReq(apiV1 + "/stats/users"),
	}
}

type StatsUsersReq struct {
	req
}

// Context sets the context to be used during this request. If no context is supplied then
// the request will use context.Background.
func (r *StatsUsersReq) Context(ctx context.Context) *StatsUsersReq {
	r.req.ctx = ctx
	return r
}

// ClientID sets a client identifier that will be passed to the Bankrs API in
// the X-Client-Id header.
func (r *StatsUsersReq) ClientID(id string) *StatsUsersReq {
	r.req.clientID = id
	return r
}

func (r *StatsUsersReq) FromDate(date time.Time) *StatsUsersReq {
	r.req.par.Set("from_date", date.Format("2006-01-02"))
	return r
}

func (r *StatsUsersReq) ToDate(date time.Time) *StatsUsersReq {
	r.req.par.Set("to_date", date.Format("2006-01-02"))
	return r
}

func (r *StatsUsersReq) Send() (*UsersStats, error) {
	// TODO: remove environment parameter
	r.req.par.Set("environment", "sandbox")

	res, cleanup, err := r.req.get()
	defer cleanup()
	if err != nil {
		return nil, err
	}

	var stats UsersStats
	if err := json.NewDecoder(res.Body).Decode(&stats); err != nil {
		return nil, decodeError(err, res)
	}

	return &stats, nil
}

func (d *StatsService) Requests() *StatsRequestsReq {
	return &StatsRequestsReq{
		req: d.client.newReq(apiV1 + "/stats/requests"),
	}
}

type StatsRequestsReq struct {
	req
}

// Context sets the context to be used during this request. If no context is supplied then
// the request will use context.Background.
func (r *StatsRequestsReq) Context(ctx context.Context) *StatsRequestsReq {
	r.req.ctx = ctx
	return r
}

// ClientID sets a client identifier that will be passed to the Bankrs API in
// the X-Client-Id header.
func (r *StatsRequestsReq) ClientID(id string) *StatsRequestsReq {
	r.req.clientID = id
	return r
}

func (r *StatsRequestsReq) FromDate(date time.Time) *StatsRequestsReq {
	r.req.par.Set("from_date", date.Format("2006-01-02"))
	return r
}

func (r *StatsRequestsReq) ToDate(date time.Time) *StatsRequestsReq {
	r.req.par.Set("to_date", date.Format("2006-01-02"))
	return r
}

func (r *StatsRequestsReq) Send() (*RequestsStats, error) {
	// TODO: remove environment parameter
	r.req.par.Set("environment", "sandbox")

	res, cleanup, err := r.req.get()
	defer cleanup()
	if err != nil {
		return nil, err
	}

	var stats RequestsStats
	if err := json.NewDecoder(res.Body).Decode(&stats); err != nil {
		return nil, decodeError(err, res)
	}

	return &stats, nil
}
