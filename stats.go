package bosgo

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

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
