package bosgo

import (
	"context"
	"encoding/json"
	"net/http"
)

// UserClient is a client used for interacting with services in the context of
// a registered application and a valid user session. It is safe for
// concurrent use by multiple goroutines.
type UserClient struct {
	// never modified once they have been set
	hc            *http.Client
	addr          string
	token         string // session token
	applicationID string

	Accesses *AccessesService
	Jobs     *JobsService
}

// NewUserClient creates a new user client, ready to use.
func NewUserClient(client *http.Client, addr string, token string, applicationID string) *UserClient {
	uc := &UserClient{
		hc:            client,
		addr:          addr,
		token:         token,
		applicationID: applicationID,
	}
	uc.Accesses = NewAccessesService(uc)
	uc.Jobs = NewJobsService(uc)

	return uc
}

func (u *UserClient) newReq(path string) req {
	return req{
		hc:   u.hc,
		addr: u.addr,
		path: path,
		headers: headers{
			"x-token":          u.token,
			"x-application-id": u.applicationID,
		},
		par: params{},
	}
}

// Logout returns a request that may be used to log a user out of the Bankrs
// API. Once this request has been sent the user client is no longer valid and
// should not be used.
func (u *UserClient) Logout() *UserLogoutReq {
	return &UserLogoutReq{
		req: u.newReq("/v1/users/logout"),
	}
}

type UserLogoutReq struct {
	req
}

func (r *UserLogoutReq) Context(ctx context.Context) *UserLogoutReq {
	r.req.ctx = ctx
	return r
}

func (r *UserLogoutReq) Send() error {
	_, cleanup, err := r.req.postJSON(nil)
	defer cleanup()
	if err != nil {
		return err
	}
	return nil
}

type AccessesService struct {
	client *UserClient
}

func NewAccessesService(u *UserClient) *AccessesService { return &AccessesService{client: u} }

func (a *AccessesService) List() *ListAccessesReq {
	return &ListAccessesReq{
		req: a.client.newReq("/v1/accesses"),
	}
}

type ListAccessesReq struct {
	req
}

func (r *ListAccessesReq) Context(ctx context.Context) *ListAccessesReq {
	r.req.ctx = ctx
	return r
}

func (r *ListAccessesReq) Send() (*BankAccessPage, error) {
	res, cleanup, err := r.req.get()
	defer cleanup()
	if err != nil {
		return nil, err
	}

	var page BankAccessPage
	if err := json.NewDecoder(res.Body).Decode(&page.Accesses); err != nil {
		return nil, err
	}

	return &page, nil
}

type BankAccessPage struct {
	Accesses []BankAccess
}

type BankAccess struct {
	ID         int64  `json:"id"`
	BankID     int64  `json:"bank_id"`
	Name       string `json:"name"`
	IsPinSaved bool   `json:"is_pin_saved"`
	Enabled    bool   `json:"enabled"`
}

type ChallengeAnswerList []ChallengeAnswer

type ChallengeAnswer struct {
	ID    string `json:"id"`
	Value string `json:"value"`
	Store bool   `json:"store"`
}

func (a *AccessesService) Add(providerID string, answers ChallengeAnswerList) *AddAccessReq {
	return &AddAccessReq{
		req:        a.client.newReq("/v1/accesses"),
		providerID: providerID,
		answers:    answers,
	}
}

type AddAccessReq struct {
	req
	providerID string
	answers    ChallengeAnswerList
}

func (r *AddAccessReq) Context(ctx context.Context) *AddAccessReq {
	r.req.ctx = ctx
	return r
}

func (r *AddAccessReq) Send() (*Job, error) {
	data := struct {
		ProviderID       string              `json:"provider_id"`
		ChallengeAnswers ChallengeAnswerList `json:"challenge_answers"`
	}{
		ProviderID:       r.providerID,
		ChallengeAnswers: r.answers,
	}

	res, cleanup, err := r.req.postJSON(&data)
	defer cleanup()
	if err != nil {
		return nil, err
	}

	var job Job
	if err := json.NewDecoder(res.Body).Decode(&job); err != nil {
		return nil, err
	}

	return &job, nil
}

type Job struct {
	URI string `json:"uri"`
}

func (a *AccessesService) Delete(id string) *DeleteAccessReq {
	return &DeleteAccessReq{
		req: a.client.newReq("/v1/accesses/" + id),
	}
}

type DeleteAccessReq struct {
	req
}

func (r *DeleteAccessReq) Context(ctx context.Context) *DeleteAccessReq {
	r.req.ctx = ctx
	return r
}

// Send sends the request to get details of a bank access.
func (r *DeleteAccessReq) Send() (string, error) {
	res, cleanup, err := r.req.delete()
	defer cleanup()
	if err != nil {
		return "", err
	}

	var deleted struct {
		ID string `json:"deleted_access_id"`
	}
	if err := json.NewDecoder(res.Body).Decode(&deleted); err != nil {
		return "", err
	}

	return deleted.ID, nil
}

func (a *AccessesService) Get(id string) *GetAccessReq {
	return &GetAccessReq{
		req: a.client.newReq("/v1/accesses/" + id),
	}
}

type GetAccessReq struct {
	req
}

func (r *GetAccessReq) Context(ctx context.Context) *GetAccessReq {
	r.req.ctx = ctx
	return r
}

// Send sends the request to get details of a bank access.
func (r *GetAccessReq) Send() (*BankAccessWithAccounts, error) {
	res, cleanup, err := r.req.get()
	defer cleanup()
	if err != nil {
		return nil, err
	}

	var ba BankAccessWithAccounts
	if err := json.NewDecoder(res.Body).Decode(&ba); err != nil {
		return nil, err
	}

	return &ba, nil
}

type BankAccessWithAccounts struct {
	BankAccess
	Accounts []Account `json:"accounts"`
}

type Account struct {
	ID           int64 `json:"id"`
	BankId       int64 `json:"bank_id"`
	BankAccessId int64 `json:"bank_access_id"`

	Name         string               `json:"name"`
	Type         string               `json:"type"`
	Number       string               `json:"number"`
	Balance      string               `json:"balance"`
	BalanceDate  string               `json:"balance_date"`
	Enabled      bool                 `json:"enabled"`
	Currency     string               `json:"currency"`
	Iban         string               `json:"iban"`
	Supported    bool                 `json:"supported"`
	Alias        string               `json:"alias"`
	Capabilities *AccountCapabilities `json:"capabilities" `
	Bin          string               `json:"bin"`
}

type AccountCapabilities struct {
	AccountStatement  []string `json:"account_statement"`
	Transfer          []string `json:"transfer"`
	RecurringTransfer []string `json:"recurring_transfer"`
}

func (a *AccessesService) Update(id string, answers ChallengeAnswerList) *UpdateAccessReq {
	return &UpdateAccessReq{
		req:     a.client.newReq("/v1/accesses/" + id),
		answers: answers,
	}
}

type UpdateAccessReq struct {
	req
	answers ChallengeAnswerList
}

func (r *UpdateAccessReq) Context(ctx context.Context) *UpdateAccessReq {
	r.req.ctx = ctx
	return r
}

func (r *UpdateAccessReq) Send() (*BankAccessWithAccounts, error) {
	data := struct {
		ChallengeAnswers ChallengeAnswerList `json:"challenge_answers"`
	}{
		ChallengeAnswers: r.answers,
	}

	res, cleanup, err := r.req.postJSON(&data)
	defer cleanup()
	if err != nil {
		return nil, err
	}

	var ba BankAccessWithAccounts
	if err := json.NewDecoder(res.Body).Decode(&ba); err != nil {
		return nil, err
	}

	return &ba, nil
}

func (a *AccessesService) Refresh() *RefreshAccessesReq {
	return &RefreshAccessesReq{
		req: a.client.newReq("/v1/accesses/refresh"),
	}
}

type RefreshAccessesReq struct {
	req
}

func (r *RefreshAccessesReq) Context(ctx context.Context) *RefreshAccessesReq {
	r.req.ctx = ctx
	return r
}

func (r *RefreshAccessesReq) Send() ([]Job, error) {
	res, cleanup, err := r.req.postJSON(nil)
	defer cleanup()
	if err != nil {
		return nil, err
	}

	var jobs []Job
	if err := json.NewDecoder(res.Body).Decode(&jobs); err != nil {
		return nil, err
	}

	return jobs, nil
}

type JobsService struct {
	client *UserClient
}

func NewJobsService(u *UserClient) *JobsService { return &JobsService{client: u} }

// Get returns a request that may be used to get the details of a job.
func (j *JobsService) Get(uri string) *JobGetReq {
	return &JobGetReq{
		req: j.client.newReq("/v1" + uri),
	}
}

type JobGetReq struct {
	req
}

// Context sets the context to be used during this request. If no context is supplied then
// the request will use context.Background.
func (r *JobGetReq) Context(ctx context.Context) *JobGetReq {
	r.req.ctx = ctx
	return r
}

// Send sends the request to get details of a job.
func (r *JobGetReq) Send() (*JobStatus, error) {
	res, cleanup, err := r.req.get()
	defer cleanup()
	if err != nil {
		return nil, err
	}

	var status JobStatus
	if err := json.NewDecoder(res.Body).Decode(&status); err != nil {
		return nil, err
	}

	return &status, nil
}

// Answer returns a request that may be used to answer a challenge needed by a job
func (j *JobsService) Answer(uri string, answers ChallengeAnswerList) *JobAnswerReq {
	return &JobAnswerReq{
		req:     j.client.newReq("/v1" + uri),
		answers: answers,
	}
}

type JobAnswerReq struct {
	req
	answers ChallengeAnswerList
}

// Context sets the context to be used during this request. If no context is supplied then
// the request will use context.Background.
func (r *JobAnswerReq) Context(ctx context.Context) *JobAnswerReq {
	r.req.ctx = ctx
	return r
}

// Send sends the request to get answer a challenge needed by a job.
func (r *JobAnswerReq) Send() error {
	data := struct {
		Answers ChallengeAnswerList `json:"challenge_answers"`
	}{
		Answers: r.answers,
	}

	_, cleanup, err := r.req.putJSON(&data)
	defer cleanup()
	if err != nil {
		return err
	}

	return nil
}

// Get returns a request that may be used to cancel a job.
func (j *JobsService) Cancel(uri string) *JobCancelReq {
	return &JobCancelReq{
		req: j.client.newReq("/v1" + uri),
	}
}

type JobCancelReq struct {
	req
}

// Context sets the context to be used during this request. If no context is supplied then
// the request will use context.Background.
func (r *JobCancelReq) Context(ctx context.Context) *JobCancelReq {
	r.req.ctx = ctx
	return r
}

// Send sends the request to cancel a job.
func (r *JobCancelReq) Send() error {
	_, cleanup, err := r.req.delete()
	defer cleanup()
	if err != nil {
		return err
	}

	return nil
}

type JobStatus struct {
	Finished  bool            `json:"finished"`
	Stage     string          `json:"stage"`
	Challenge *Challenge      `json:"challenge,omitempty"`
	Uri       string          `json:"uri,omitempty"`
	Errors    []APIError      `json:"errors,omitempty"`
	Access    *AccessResponse `json:"access,omitempty"`
}

type APIError struct {
	Code    string                 `json:"code"`
	Payload map[string]interface{} `json:"payload,omitempty"`
}

type Challenge struct {
	CanContinue    bool             `json:"cancontinue"`
	MaxSteps       uint             `json:"maxsteps"`
	CurStep        uint             `json:"curstep"`
	NextChallenges []ChallengeField `json:"nextchallenges"`
	LastProblems   []Problem        `json:"lastproblems"`
	Hint           string           `json:"hint"`
}

type Problem struct {
	Domain  string            `json:"domain"`
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Info    map[string]string `json:"info"`
}

type ChallengeField struct {
	ID          string   `json:"id"`
	Description string   `json:"description"`
	Type        string   `json:"type"`
	Previous    string   `json:"previous"`
	Stored      bool     `json:"stored"`
	Reset       bool     `json:"reset"`
	Secure      bool     `json:"secure"`
	Optional    bool     `json:"optional"`
	UnStoreable bool     `json:"unstoreable"`
	Methods     []string `json:"methods"`
}

type AccessResponse struct {
	Id       int64             `json:"id,omitempty"`
	BankId   int64             `json:"bank_id,omitempty"`
	Name     string            `json:"name,omitempty"`
	Accounts []AccountResponse `json:"accounts,omitempty"`
}

type AccountResponse struct {
	Id        int64  `json:"id,omitempty"`
	Name      string `json:"name"`
	Supported bool   `json:"supported"`
	Number    string `json:"number"`
	IBAN      string `json:"iban"`
}
