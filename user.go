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
	"time"
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
	ua            string

	Accesses              *AccessesService
	Jobs                  *JobsService
	Accounts              *AccountsService
	Transactions          *TransactionsService
	ScheduledTransactions *ScheduledTransactionsService
	RepeatedTransactions  *RepeatedTransactionsService
	Transfers             *TransfersService
	RecurringTransfers    *RecurringTransfersService
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
	uc.Accounts = NewAccountsService(uc)
	uc.Transactions = NewTransactionsService(uc)
	uc.ScheduledTransactions = NewScheduledTransactionsService(uc)
	uc.RepeatedTransactions = NewRepeatedTransactionsService(uc)
	uc.Transfers = NewTransfersService(uc)
	uc.RecurringTransfers = NewRecurringTransfersService(uc)

	return uc
}

func (u *UserClient) userAgent() string {
	if u.ua == "" {
		return UserAgent
	}

	return UserAgent + " " + u.ua
}
func (u *UserClient) newReq(path string) req {
	return req{
		hc:   u.hc,
		addr: u.addr,
		path: path,
		headers: headers{
			"User-Agent":       u.userAgent(),
			"x-token":          u.token,
			"x-application-id": u.applicationID,
		},
		par: params{},
	}
}

// SessionToken returns the current session token.
func (u *UserClient) SessionToken() string {
	return u.token
}

// Logout returns a request that may be used to log a user out of the Bankrs
// API. Once this request has been sent the user client is no longer valid and
// should not be used.
func (u *UserClient) Logout() *UserLogoutReq {
	return &UserLogoutReq{
		req: u.newReq(apiV1 + "/users/logout"),
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

// Delete returns a request that may be used to delete a user account and its
// associated data. Once this request has been sent the user client is no
// longer valid and should not be used.
func (u *UserClient) Delete() *UserDeleteReq {
	return &UserDeleteReq{
		req: u.newReq(apiV1 + "/users"),
	}
}

type UserDeleteReq struct {
	req
}

func (r *UserDeleteReq) Context(ctx context.Context) *UserDeleteReq {
	r.req.ctx = ctx
	return r
}

// Send sends the request to delete a user.
func (r *UserDeleteReq) Send() error {
	_, cleanup, err := r.req.delete()
	defer cleanup()
	if err != nil {
		return err
	}
	return nil
}

// AccessesService provides access to bank access related API services.
type AccessesService struct {
	client *UserClient
}

func NewAccessesService(u *UserClient) *AccessesService { return &AccessesService{client: u} }

func (a *AccessesService) List() *ListAccessesReq {
	return &ListAccessesReq{
		req: a.client.newReq(apiV1 + "/accesses"),
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
		return nil, wrap(errContextInvalidServiceResponse, err)
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

func (a *AccessesService) Add(providerID string) *AddAccessReq {
	return &AddAccessReq{
		req:        a.client.newReq(apiV1 + "/accesses"),
		providerID: providerID,
		answers:    ChallengeAnswerList{},
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

// ChallengeAnswer adds an answer to one of the authorisation challenges required to complete addition of the access.
func (r *AddAccessReq) ChallengeAnswer(answer ChallengeAnswer) *AddAccessReq {
	r.answers = append(r.answers, answer)
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
		return nil, wrap(errContextInvalidServiceResponse, err)
	}

	return &job, nil
}

type Job struct {
	URI string `json:"uri"`
}

func (a *AccessesService) Delete(id string) *DeleteAccessReq {
	return &DeleteAccessReq{
		req: a.client.newReq(apiV1 + "/accesses/" + id),
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
		return "", wrap(errContextInvalidServiceResponse, err)
	}

	return deleted.ID, nil
}

func (a *AccessesService) Get(id string) *GetAccessReq {
	return &GetAccessReq{
		req: a.client.newReq(apiV1 + "/accesses/" + id),
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
		return nil, wrap(errContextInvalidServiceResponse, err)
	}

	return &ba, nil
}

type BankAccessWithAccounts struct {
	BankAccess
	Accounts []Account `json:"accounts"`
}

type Account struct {
	ID           int64                `json:"id"`
	BankID       int64                `json:"bank_id"`
	BankAccessID int64                `json:"bank_access_id"`
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

func (a *AccessesService) Update(id string) *UpdateAccessReq {
	return &UpdateAccessReq{
		req:     a.client.newReq(apiV1 + "/accesses/" + id),
		answers: ChallengeAnswerList{},
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

// ChallengeAnswer adds an answer to one of the authorisation challenges required to complete update of the access.
func (r *UpdateAccessReq) ChallengeAnswer(answer ChallengeAnswer) *UpdateAccessReq {
	r.answers = append(r.answers, answer)
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
		return nil, wrap(errContextInvalidServiceResponse, err)
	}

	return &ba, nil
}

func (a *AccessesService) Refresh() *RefreshAccessesReq {
	return &RefreshAccessesReq{
		req: a.client.newReq(apiV1 + "/accesses/refresh"),
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
		return nil, wrap(errContextInvalidServiceResponse, err)
	}

	return jobs, nil
}

// JobsService provides access to jobs related API services.
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
		return nil, wrap(errContextInvalidServiceResponse, err)
	}

	return &status, nil
}

// Answer returns a request that may be used to answer a challenge needed by a job
func (j *JobsService) Answer(uri string) *JobAnswerReq {
	return &JobAnswerReq{
		req:     j.client.newReq("/v1" + uri),
		answers: ChallengeAnswerList{},
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

// ChallengeAnswer adds an answer to one of the authorisation challenges required to complete the job.
func (r *JobAnswerReq) ChallengeAnswer(answer ChallengeAnswer) *JobAnswerReq {
	r.answers = append(r.answers, answer)
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

// Cancel returns a request that may be used to cancel a job.
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
	URI       string          `json:"uri,omitempty"`
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
	ID       int64             `json:"id,omitempty"`
	BankID   int64             `json:"bank_id,omitempty"`
	Name     string            `json:"name,omitempty"`
	Accounts []AccountResponse `json:"accounts,omitempty"`
}

type AccountResponse struct {
	ID        int64  `json:"id,omitempty"`
	Name      string `json:"name"`
	Supported bool   `json:"supported"`
	Number    string `json:"number"`
	IBAN      string `json:"iban"`
}

// AccountsService provides access to account related API services.
type AccountsService struct {
	client *UserClient
}

func NewAccountsService(u *UserClient) *AccountsService { return &AccountsService{client: u} }

func (a *AccountsService) List() *ListAccountsReq {
	return &ListAccountsReq{
		req: a.client.newReq(apiV1 + "/accounts"),
	}
}

type ListAccountsReq struct {
	req
}

func (r *ListAccountsReq) Context(ctx context.Context) *ListAccountsReq {
	r.req.ctx = ctx
	return r
}

func (r *ListAccountsReq) Send() (*AccountPage, error) {
	res, cleanup, err := r.req.get()
	defer cleanup()
	if err != nil {
		return nil, err
	}

	var page AccountPage
	if err := json.NewDecoder(res.Body).Decode(&page.Accounts); err != nil {
		return nil, wrap(errContextInvalidServiceResponse, err)
	}

	return &page, nil
}

type AccountPage struct {
	Accounts []Account
}

func (a *AccountsService) Get(id string) *GetAccountReq {
	return &GetAccountReq{
		req: a.client.newReq(apiV1 + "/accounts/" + id),
	}
}

type GetAccountReq struct {
	req
}

func (r *GetAccountReq) Context(ctx context.Context) *GetAccountReq {
	r.req.ctx = ctx
	return r
}

func (r *GetAccountReq) Send() (*Account, error) {
	res, cleanup, err := r.req.get()
	defer cleanup()
	if err != nil {
		return nil, err
	}

	var account Account
	if err := json.NewDecoder(res.Body).Decode(&account); err != nil {
		return nil, wrap(errContextInvalidServiceResponse, err)
	}

	return &account, nil
}

type Transaction struct {
	ID                    int64            `json:"id"`
	AccessID              int64            `json:"user_bank_access_id,omitempty"`
	UserAccountID         int64            `json:"user_bank_account_id,omitempty"`
	UserAccount           AccountRef       `json:"user_account,omitempty"`
	CategoryID            int64            `json:"category_id,omitempty"`
	RepeatedTransactionID int64            `json:"repeated_transaction_id,omitempty"`
	Counterparty          CounterpartyWrap `json:"counterparty,omitempty"`

	// EntryDate is the time the transaction became known in the account
	EntryDate time.Time `json:"entry_date,omitempty"`

	// SettlementDate is the time the transaction is cleared
	SettlementDate time.Time `json:"settlement_date,omitempty"`

	// Transaction Amount - value and currency
	Amount *MoneyAmount `json:"amount,omitempty"`

	// Usage is the main description field
	Usage string `json:"usage,omitempty"`

	// TransactionType is extracted directly from the finsnap
	TransactionType string `json:"transaction_type,omitempty"`
}

type AccountRef struct {
	ProviderID string `json:"provider_id"`
	IBAN       string `json:"iban,omitempty"`
	Label      string `json:"label,omitempty"`
	ID         string `json:"id,omitempty"`
}

type Merchant struct {
	Name string `json:"name"`
}

type Counterparty struct {
	Name    string     `json:"name"`
	Account AccountRef `json:"account,omitempty"`
}

type CounterpartyWrap struct {
	Counterparty
	Merchant *Merchant `json:"merchant,omitempty"`
}

// TransactionsService provides access to transaction related API services.
type TransactionsService struct {
	client *UserClient
}

func NewTransactionsService(u *UserClient) *TransactionsService {
	return &TransactionsService{client: u}
}

func (a *TransactionsService) List() *ListTransactionsReq {
	return &ListTransactionsReq{
		req: a.client.newReq(apiV1 + "/transactions"),
	}
}

type ListTransactionsReq struct {
	req
}

func (r *ListTransactionsReq) Context(ctx context.Context) *ListTransactionsReq {
	r.req.ctx = ctx
	return r
}

func (r *ListTransactionsReq) Send() (*TransactionPage, error) {
	res, cleanup, err := r.req.get()
	defer cleanup()
	if err != nil {
		return nil, err
	}

	var page TransactionPage
	if err := json.NewDecoder(res.Body).Decode(&page.Transactions); err != nil {
		return nil, wrap(errContextInvalidServiceResponse, err)
	}

	return &page, nil
}

type TransactionPage struct {
	Transactions []Transaction
}

func (a *TransactionsService) Get(id string) *GetTransactionReq {
	return &GetTransactionReq{
		req: a.client.newReq(apiV1 + "/transactions/" + id),
	}
}

type GetTransactionReq struct {
	req
}

func (r *GetTransactionReq) Context(ctx context.Context) *GetTransactionReq {
	r.req.ctx = ctx
	return r
}

func (r *GetTransactionReq) Send() (*Transaction, error) {
	res, cleanup, err := r.req.get()
	defer cleanup()
	if err != nil {
		return nil, err
	}

	var tx Transaction
	if err := json.NewDecoder(res.Body).Decode(&tx); err != nil {
		return nil, wrap(errContextInvalidServiceResponse, err)
	}

	return &tx, nil
}

func (a *TransactionsService) Categorise() *CategoriseTransactionsReq {
	return &CategoriseTransactionsReq{
		req:  a.client.newReq(apiV1 + "/transactions/categorise"),
		cats: []categorisation{},
	}
}

type categorisation struct {
	TransactionID string `json:"id"`
	CategoryID    string `json:"category_id"`
}

type CategoriseTransactionsReq struct {
	req
	cats []categorisation
}

func (r *CategoriseTransactionsReq) Context(ctx context.Context) *CategoriseTransactionsReq {
	r.req.ctx = ctx
	return r
}

func (r *CategoriseTransactionsReq) Category(transactionID string, categoryID string) *CategoriseTransactionsReq {
	r.cats = append(r.cats, categorisation{TransactionID: transactionID, CategoryID: categoryID})
	return r
}

func (r *CategoriseTransactionsReq) Send() error {
	_, cleanup, err := r.req.putJSON(r.cats)
	defer cleanup()
	if err != nil {
		return err
	}

	return nil
}

// ScheduledTransactionsService provides access to scheduled transaction related API services.
type ScheduledTransactionsService struct {
	client *UserClient
}

func NewScheduledTransactionsService(u *UserClient) *ScheduledTransactionsService {
	return &ScheduledTransactionsService{client: u}
}

func (a *ScheduledTransactionsService) List() *ListScheduledTransactionsReq {
	return &ListScheduledTransactionsReq{
		req: a.client.newReq(apiV1 + "/scheduled_transactions"),
	}
}

type ListScheduledTransactionsReq struct {
	req
}

func (r *ListScheduledTransactionsReq) Context(ctx context.Context) *ListScheduledTransactionsReq {
	r.req.ctx = ctx
	return r
}

func (r *ListScheduledTransactionsReq) Send() (*TransactionPage, error) {
	res, cleanup, err := r.req.get()
	defer cleanup()
	if err != nil {
		return nil, err
	}

	var page TransactionPage
	if err := json.NewDecoder(res.Body).Decode(&page.Transactions); err != nil {
		return nil, wrap(errContextInvalidServiceResponse, err)
	}

	return &page, nil
}

func (a *ScheduledTransactionsService) Get(id string) *GetScheduledTransactionReq {
	return &GetScheduledTransactionReq{
		req: a.client.newReq(apiV1 + "/scheduled_transactions/" + id),
	}
}

type GetScheduledTransactionReq struct {
	req
}

func (r *GetScheduledTransactionReq) Context(ctx context.Context) *GetScheduledTransactionReq {
	r.req.ctx = ctx
	return r
}

func (r *GetScheduledTransactionReq) Send() (*Transaction, error) {
	res, cleanup, err := r.req.get()
	defer cleanup()
	if err != nil {
		return nil, err
	}

	var tx Transaction
	if err := json.NewDecoder(res.Body).Decode(&tx); err != nil {
		return nil, wrap(errContextInvalidServiceResponse, err)
	}

	return &tx, nil
}

// RepeatedTransactionsService provides access to repeated transaction related API services.
type RepeatedTransactionsService struct {
	client *UserClient
}

func NewRepeatedTransactionsService(u *UserClient) *RepeatedTransactionsService {
	return &RepeatedTransactionsService{client: u}
}

func (a *RepeatedTransactionsService) List() *ListRepeatedTransactionsReq {
	return &ListRepeatedTransactionsReq{
		req: a.client.newReq(apiV1 + "/repeated_transactions"),
	}
}

type ListRepeatedTransactionsReq struct {
	req
}

func (r *ListRepeatedTransactionsReq) Context(ctx context.Context) *ListRepeatedTransactionsReq {
	r.req.ctx = ctx
	return r
}

func (r *ListRepeatedTransactionsReq) Send() (*RepeatedTransactionPage, error) {
	res, cleanup, err := r.req.get()
	defer cleanup()
	if err != nil {
		return nil, err
	}

	var page RepeatedTransactionPage
	if err := json.NewDecoder(res.Body).Decode(&page.Transactions); err != nil {
		return nil, wrap(errContextInvalidServiceResponse, err)
	}

	return &page, nil
}

type RepeatedTransactionPage struct {
	Transactions []RepeatedTransaction
}

func (a *RepeatedTransactionsService) Get(id string) *GetRepeatedTransactionReq {
	return &GetRepeatedTransactionReq{
		req: a.client.newReq(apiV1 + "/repeated_transactions/" + id),
	}
}

type GetRepeatedTransactionReq struct {
	req
}

func (r *GetRepeatedTransactionReq) Context(ctx context.Context) *GetRepeatedTransactionReq {
	r.req.ctx = ctx
	return r
}

func (r *GetRepeatedTransactionReq) Send() (*RepeatedTransaction, error) {
	res, cleanup, err := r.req.get()
	defer cleanup()
	if err != nil {
		return nil, err
	}

	var tx RepeatedTransaction
	if err := json.NewDecoder(res.Body).Decode(&tx); err != nil {
		return nil, wrap(errContextInvalidServiceResponse, err)
	}

	return &tx, nil
}

type RepeatedTransaction struct {
	ID            int64          `json:"id"`
	UserAccountID int64          `json:"user_bank_account_id"`
	UserAccount   AccountRef     `json:"user_account"`
	RemoteAccount AccountRef     `json:"remote_account"`
	AccessID      int64          `json:"user_bank_access_id"`
	RemoteID      string         `json:"remote_id"`
	Schedule      RecurrenceRule `json:"schedule"`
	Amount        *MoneyAmount   `json:"amount"`
	Usage         string         `json:"usage"`
}

type RecurrenceRule struct {
	Start    time.Time `json:"start"`
	Until    time.Time `json:"until"`
	Freq     string    `json:"frequency"`
	Interval int       `json:"interval"`
	ByDay    int       `json:"by_day"`
}

type MoneyAmount struct {
	Currency string `json:"currency"`
	Value    string `json:"value"`
}

type TransferAddress struct {
	Name string `json:"name"` // required
	Iban string `json:"iban"` // required
}

type TransferType string

const (
	TransferTypeRecurring TransferType = "recurring"
	TransferTypeRegular   TransferType = "regular"
)

type transferParams struct {
	From             string              `json:"from,omitempty"`
	To               TransferAddress     `json:"to,omitempty"`
	Amount           MoneyAmount         `json:"amount,omitempty"`
	Schedule         *RecurrenceRule     `json:"schedule,omitempty"`
	EntryDate        string              `json:"entry_date,omitempty"`
	Usage            string              `json:"usage,omitempty"`
	Type             TransferType        `json:"type,omitempty"`
	ChallengeAnswers ChallengeAnswerList `json:"challenge_answers,omitempty"`
}

type ChallengeAnswerMap map[string]ChallengeAnswer

type RecurringTransferJob struct {
	*RecurringTransfer

	ID      string        `json:"id"`
	Version int           `json:"version"`
	State   string        `json:"state"`
	Errors  []Problem     `json:"errors,omitempty"`
	Step    *TransferStep `json:"step"`
}

type RecurringTransfer struct {
	ID       string           `json:"id"`
	From     *TransferAddress `json:"from"`
	To       *TransferAddress `json:"to"`
	Schedule *RecurrenceRule  `json:"schedule"`
	Amount   *MoneyAmount     `json:"amount"`
	Usage    string           `json:"usage"`
}

type RecurringTransferOptions struct {
	BankHoliday string `json:"bank_holiday"`
}

type PaymentTransferParams struct {
	From             *TransferAddress   `json:"from"`
	To               *TransferAddress   `json:"to"`
	Schedule         *RecurrenceRule    `json:"schedule,omitempty"`
	Amount           *MoneyAmount       `json:"amount"`
	Description      string             `json:"usage"`
	EntryDate        time.Time          `json:"booking_date,omitempty"`
	SettlementDate   time.Time          `json:"effective_date,omitempty"`
	ChallengeAnswers ChallengeAnswerMap `json:"challenge_answers,omitempty"`
	ID               string             `json:"id"`
	Version          int                `json:"version"`
	Step             *TransferStep      `json:"step"`
	State            string             `json:"state"`
	Created          time.Time          `json:"created,omitempty"`
	Updated          time.Time          `json:"updated,omitempty"`
	Errors           []Problem          `json:"errors"`
}

type PaymentTransferCancelParams struct {
	ID      string `json:"id"`
	Version int    `json:"version"`
}

type TransferStep struct {
	Intent string            `json:"intent,omitempty"`
	Data   *TransferStepData `json:"data,omitempty"`
}

type AuthMethod struct {
	ID          string `json:"id"`
	Description string `json:"description"`
}

type TransferStepData struct {
	AuthMethods      []*AuthMethod            `json:"auth_methods,omitempty"`      // Tan Options
	Challenge        string                   `json:"challenge,omitempty"`         // Tan Challenge
	ChallengeMessage string                   `json:"challenge_message,omitempty"` // Tan Challenge Message
	TanType          string                   `json:"tan_type,omitempty"`          // Type of the Tan (optical, itan, unknown)
	Confirm          bool                     `json:"confirm,omitempty"`           // Confirm (similar transfer)
	Transfers        []*PaymentTransferParams `json:"transfers,omitempty"`         // Transfer list (similar transfers)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
//             TRANSFERS SERVICE
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// TransfersService provides access to money transfer related API services.
type TransfersService struct {
	client *UserClient
}

func NewTransfersService(u *UserClient) *TransfersService {
	return &TransfersService{client: u}
}

// Create returns a request that may be used to create a money transfer.
func (t *TransfersService) Create(from string, to TransferAddress, amount MoneyAmount) *CreateTransferReq {
	return &CreateTransferReq{
		req: t.client.newReq(apiV1 + "/transfers"),
		data: transferParams{
			From:   from,
			To:     to,
			Amount: amount,
			Type:   TransferTypeRegular,
		},
	}
}

type CreateTransferReq struct {
	req
	data transferParams
}

// Context sets the context to be used during this request. If no context is supplied then
// the request will use context.Background.
func (r *CreateTransferReq) Context(ctx context.Context) *CreateTransferReq {
	r.req.ctx = ctx
	return r
}

// EntryDate sets the desired date for the transfer to be placed. It cannot be a date in the past.
func (r *CreateTransferReq) EntryDate(date time.Time) *CreateTransferReq {
	r.data.EntryDate = date.Format("2006-01-02")
	return r
}

// Description sets a human readable description for the transfer.
func (r *CreateTransferReq) Description(s string) *CreateTransferReq {
	r.data.Usage = s
	return r
}

// ChallengeAnswer adds an answer to one of the authorisation challenges required to complete the transfer.
func (r *CreateTransferReq) ChallengeAnswer(answer ChallengeAnswer) *CreateTransferReq {
	r.data.ChallengeAnswers = append(r.data.ChallengeAnswers, answer)
	return r
}

// Send sends the request to create a money transfer.
func (r *CreateTransferReq) Send() (*PaymentTransferParams, error) {
	res, cleanup, err := r.req.postJSON(r.data)
	defer cleanup()
	if err != nil {
		return nil, err
	}

	var job PaymentTransferParams
	if err := json.NewDecoder(res.Body).Decode(&job); err != nil {
		return nil, wrap(errContextInvalidServiceResponse, err)
	}

	return &job, nil
}

// Process returns a request that may be used to update information and answer challenges for a transfer.
func (t *TransfersService) Process(id string, version int) *ProcessTransferReq {
	return &ProcessTransferReq{
		req:     t.client.newReq(apiV1 + "/transfers/" + id + "/cancel"),
		version: version,
	}
}

type ProcessTransferReq struct {
	req
	version int
}

// Context sets the context to be used during this request. If no context is supplied then
// the request will use context.Background.
func (r *ProcessTransferReq) Context(ctx context.Context) *ProcessTransferReq {
	r.req.ctx = ctx
	return r
}

// Send sends the request to update information and answer challenges for a transfer.
func (r *ProcessTransferReq) Send() (*PaymentTransferParams, error) {
	data := struct {
		Version int    `json:"version,omitempty"`
		Type    string `json:"type"`
	}{
		Version: r.version,
		Type:    string(TransferTypeRegular),
	}

	res, cleanup, err := r.req.postJSON(&data)
	defer cleanup()
	if err != nil {
		return nil, err
	}

	var pt PaymentTransferParams
	if err := json.NewDecoder(res.Body).Decode(&pt); err != nil {
		return nil, wrap(errContextInvalidServiceResponse, err)
	}

	return &pt, nil
}

// Cancel returns a request that may be used to cancel an ongoing money transfer.
func (t *TransfersService) Cancel(id string, version int) *CancelTransferReq {
	return &CancelTransferReq{
		req:     t.client.newReq(apiV1 + "/transfers/" + id + "/cancel"),
		version: version,
	}
}

type CancelTransferReq struct {
	req
	version int
}

// Context sets the context to be used during this request. If no context is supplied then
// the request will use context.Background.
func (r *CancelTransferReq) Context(ctx context.Context) *CancelTransferReq {
	r.req.ctx = ctx
	return r
}

// Send sends the request to update a money transfer.
func (r *CancelTransferReq) Send() (*PaymentTransferParams, error) {
	data := struct {
		Version int    `json:"version,omitempty"`
		Type    string `json:"type"`
	}{
		Version: r.version,
		Type:    string(TransferTypeRegular),
	}

	res, cleanup, err := r.req.postJSON(&data)
	defer cleanup()
	if err != nil {
		return nil, err
	}

	var pt PaymentTransferParams
	if err := json.NewDecoder(res.Body).Decode(&pt); err != nil {
		return nil, wrap(errContextInvalidServiceResponse, err)
	}

	return &pt, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
//           RECURRING TRANSFERS SERVICE
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// RecurringTransfersService provides access to recurring money transfer related API services.
type RecurringTransfersService struct {
	client *UserClient
}

func NewRecurringTransfersService(u *UserClient) *RecurringTransfersService {
	return &RecurringTransfersService{client: u}
}

// Create returns a request that may be used to create a money transfer.
func (t *RecurringTransfersService) Create(from string, to TransferAddress, amount MoneyAmount, rule RecurrenceRule) *CreateRecurringTransferReq {
	return &CreateRecurringTransferReq{
		req: t.client.newReq(apiV1 + "/transfers"),
		data: transferParams{
			From:     from,
			To:       to,
			Amount:   amount,
			Type:     TransferTypeRecurring,
			Schedule: &rule,
		},
	}
}

type CreateRecurringTransferReq struct {
	req
	data transferParams
}

// Context sets the context to be used during this request. If no context is supplied then
// the request will use context.Background.
func (r *CreateRecurringTransferReq) Context(ctx context.Context) *CreateRecurringTransferReq {
	r.req.ctx = ctx
	return r
}

// EntryDate sets the desired date for the transfer to be placed. It cannot be a date in the past.
func (r *CreateRecurringTransferReq) EntryDate(date time.Time) *CreateRecurringTransferReq {
	r.data.EntryDate = date.Format("2006-01-02")
	return r
}

// Description sets a human readable description for the transfer.
func (r *CreateRecurringTransferReq) Description(s string) *CreateRecurringTransferReq {
	r.data.Usage = s
	return r
}

// ChallengeAnswer adds an answer to one of the authorisation challenges required to complete the transfer.
func (r *CreateRecurringTransferReq) ChallengeAnswer(answer ChallengeAnswer) *CreateRecurringTransferReq {
	r.data.ChallengeAnswers = append(r.data.ChallengeAnswers, answer)
	return r
}

// Send sends the request to create a money transfer.
func (r *CreateRecurringTransferReq) Send() (*RecurringTransferJob, error) {
	res, cleanup, err := r.req.postJSON(r.data)
	defer cleanup()
	if err != nil {
		return nil, err
	}

	var job RecurringTransferJob
	if err := json.NewDecoder(res.Body).Decode(&job); err != nil {
		return nil, wrap(errContextInvalidServiceResponse, err)
	}

	return &job, nil
}

// Delete returns a request that may be used to delete a recurring money transfer.
func (t *RecurringTransfersService) Delete(id string) *DeleteRecurringTransferReq {
	return &DeleteRecurringTransferReq{
		req:     t.client.newReq(apiV1 + "/transfers/" + id),
		answers: ChallengeAnswerList{},
	}
}

type DeleteRecurringTransferReq struct {
	req
	answers ChallengeAnswerList
}

// Context sets the context to be used during this request. If no context is supplied then
// the request will use context.Background.
func (r *DeleteRecurringTransferReq) Context(ctx context.Context) *DeleteRecurringTransferReq {
	r.req.ctx = ctx
	return r
}

// ChallengeAnswer adds an answer to one of the authorisation challenges required to complete the deletion of the transfer.
func (r *DeleteRecurringTransferReq) ChallengeAnswer(answer ChallengeAnswer) *DeleteRecurringTransferReq {
	r.answers = append(r.answers, answer)
	return r
}

// Send sends the request to create a money transfer.
func (r *DeleteRecurringTransferReq) Send() (*RecurringTransferJob, error) {
	data := struct {
		ChallengeAnswers ChallengeAnswerList `json:"challenge_answers"`
	}{
		ChallengeAnswers: r.answers,
	}

	res, cleanup, err := r.req.deleteJSON(&data)
	defer cleanup()
	if err != nil {
		return nil, err
	}

	var job RecurringTransferJob
	if err := json.NewDecoder(res.Body).Decode(&job); err != nil {
		return nil, wrap(errContextInvalidServiceResponse, err)
	}

	return &job, nil
}

// Update returns a request that may be used to update a recurring money transfer.
func (t *RecurringTransfersService) Update(id string, to TransferAddress, amount MoneyAmount) *UpdateRecurringTransferReq {
	return &UpdateRecurringTransferReq{
		req: t.client.newReq(apiV1 + "/transfers/" + id),
		data: transferParams{
			To:     to,
			Amount: amount,
			Type:   TransferTypeRecurring,
		},
	}
}

type UpdateRecurringTransferReq struct {
	req
	data transferParams
}

// Context sets the context to be used during this request. If no context is supplied then
// the request will use context.Background.
func (r *UpdateRecurringTransferReq) Context(ctx context.Context) *UpdateRecurringTransferReq {
	r.req.ctx = ctx
	return r
}

// Schedule sets a recurrence schedule for the transfer.
func (r *UpdateRecurringTransferReq) Schedule(rule RecurrenceRule) *UpdateRecurringTransferReq {
	r.data.Schedule = &rule
	return r
}

// Description sets a human readable description for the transfer.
func (r *UpdateRecurringTransferReq) Description(s string) *UpdateRecurringTransferReq {
	r.data.Usage = s
	return r
}

// ChallengeAnswer adds an answer to one of the authorisation challenges required to complete the transfer.
func (r *UpdateRecurringTransferReq) ChallengeAnswer(answer ChallengeAnswer) *UpdateRecurringTransferReq {
	r.data.ChallengeAnswers = append(r.data.ChallengeAnswers, answer)
	return r
}

// Send sends the request to update a money transfer.
func (r *UpdateRecurringTransferReq) Send() (*RecurringTransferJob, error) {
	res, cleanup, err := r.req.putJSON(r.data)
	defer cleanup()
	if err != nil {
		return nil, err
	}

	var job RecurringTransferJob
	if err := json.NewDecoder(res.Body).Decode(&job); err != nil {
		return nil, wrap(errContextInvalidServiceResponse, err)
	}

	return &job, nil
}

// Process returns a request that may be used to update information and answer challenges for a transfer.
func (t *RecurringTransfersService) Process(id string, version int) *ProcessRecurringTransferReq {
	return &ProcessRecurringTransferReq{
		req:     t.client.newReq(apiV1 + "/transfers/" + id + "/cancel"),
		version: version,
	}
}

type ProcessRecurringTransferReq struct {
	req
	version int
}

// Context sets the context to be used during this request. If no context is supplied then
// the request will use context.Background.
func (r *ProcessRecurringTransferReq) Context(ctx context.Context) *ProcessRecurringTransferReq {
	r.req.ctx = ctx
	return r
}

// Send sends the request to update information and answer challenges for a transfer.
func (r *ProcessRecurringTransferReq) Send() (*RecurringTransferJob, error) {
	data := struct {
		Version int    `json:"version,omitempty"`
		Type    string `json:"type"`
	}{
		Version: r.version,
		Type:    string(TransferTypeRecurring),
	}

	res, cleanup, err := r.req.postJSON(&data)
	defer cleanup()
	if err != nil {
		return nil, err
	}

	var job RecurringTransferJob
	if err := json.NewDecoder(res.Body).Decode(&job); err != nil {
		return nil, wrap(errContextInvalidServiceResponse, err)
	}

	return &job, nil
}

// Cancel returns a request that may be used to cancel an ongoing money transfer.
func (t *RecurringTransfersService) Cancel(id string, version int) *CancelRecurringTransferReq {
	return &CancelRecurringTransferReq{
		req:     t.client.newReq(apiV1 + "/transfers/" + id + "/cancel"),
		version: version,
	}
}

type CancelRecurringTransferReq struct {
	req
	version int
}

// Context sets the context to be used during this request. If no context is supplied then
// the request will use context.Background.
func (r *CancelRecurringTransferReq) Context(ctx context.Context) *CancelRecurringTransferReq {
	r.req.ctx = ctx
	return r
}

// Send sends the request to update a money transfer.
func (r *CancelRecurringTransferReq) Send() (*RecurringTransferJob, error) {
	data := struct {
		Version int    `json:"version,omitempty"`
		Type    string `json:"type"`
	}{
		Version: r.version,
		Type:    string(TransferTypeRecurring),
	}

	res, cleanup, err := r.req.postJSON(&data)
	defer cleanup()
	if err != nil {
		return nil, err
	}

	var job RecurringTransferJob
	if err := json.NewDecoder(res.Body).Decode(&job); err != nil {
		return nil, wrap(errContextInvalidServiceResponse, err)
	}

	return &job, nil
}
