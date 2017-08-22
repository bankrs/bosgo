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
	"strconv"
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
		return DefaultUserAgent
	}

	return DefaultUserAgent + " " + u.ua
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

func (r *ListAccessesReq) Send() (*AccessPage, error) {
	res, cleanup, err := r.req.get()
	defer cleanup()
	if err != nil {
		return nil, err
	}

	var page AccessPage
	if err := json.NewDecoder(res.Body).Decode(&page.Accesses); err != nil {
		return nil, wrap(errContextInvalidServiceResponse, err)
	}

	return &page, nil
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

func (a *AccessesService) Delete(id int64) *DeleteAccessReq {
	return &DeleteAccessReq{
		req: a.client.newReq(apiV1 + "/accesses/" + strconv.FormatInt(id, 10)),
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

func (a *AccessesService) Get(id int64) *GetAccessReq {
	return &GetAccessReq{
		req: a.client.newReq(apiV1 + "/accesses/" + strconv.FormatInt(id, 10)),
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
func (r *GetAccessReq) Send() (*Access, error) {
	res, cleanup, err := r.req.get()
	defer cleanup()
	if err != nil {
		return nil, err
	}

	var ba Access
	if err := json.NewDecoder(res.Body).Decode(&ba); err != nil {
		return nil, wrap(errContextInvalidServiceResponse, err)
	}

	return &ba, nil
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

func (r *UpdateAccessReq) Send() (*Access, error) {
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

	var ba Access
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

func (r *ListTransactionsReq) AccountID(id int64) *ListTransactionsReq {
	r.req.par["account_id"] = []string{strconv.FormatInt(id, 10)}
	return r
}

func (r *ListTransactionsReq) AccessID(id int64) *ListTransactionsReq {
	r.req.par["access_id"] = []string{strconv.FormatInt(id, 10)}
	return r
}

func (r *ListTransactionsReq) Since(t time.Time) *ListTransactionsReq {
	r.req.par["since"] = []string{t.Format(time.RFC3339)}
	return r
}

func (r *ListTransactionsReq) Limit(limit int) *ListTransactionsReq {
	r.req.par["limit"] = []string{fmt.Sprintf("%d", limit)}
	return r
}

func (r *ListTransactionsReq) Offset(offset int) *ListTransactionsReq {
	r.req.par["offset"] = []string{fmt.Sprintf("%d", offset)}
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

func (r *RepeatedTransactionsService) List() *ListRepeatedTransactionsReq {
	return &ListRepeatedTransactionsReq{
		req: r.client.newReq(apiV1 + "/repeated_transactions"),
	}
}

type ListRepeatedTransactionsReq struct {
	req
}

func (r *ListRepeatedTransactionsReq) Context(ctx context.Context) *ListRepeatedTransactionsReq {
	r.req.ctx = ctx
	return r
}

func (r *ListRepeatedTransactionsReq) AccountID(id int64) *ListRepeatedTransactionsReq {
	r.req.par["account_id"] = []string{strconv.FormatInt(id, 10)}
	return r
}

func (r *ListRepeatedTransactionsReq) AccessID(id int64) *ListRepeatedTransactionsReq {
	r.req.par["access_id"] = []string{strconv.FormatInt(id, 10)}
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

func (r *RepeatedTransactionsService) Get(id string) *GetRepeatedTransactionReq {
	return &GetRepeatedTransactionReq{
		req: r.client.newReq(apiV1 + "/repeated_transactions/" + id),
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

// Delete returns a request that may be used to delete a repeated transaction.
func (r *RepeatedTransactionsService) Delete(id string) *DeleteRepeatedTransactionReq {
	return &DeleteRepeatedTransactionReq{
		req:     r.client.newReq(apiV1 + "/repeated_transaction/" + id),
		answers: ChallengeAnswerList{},
	}
}

type DeleteRepeatedTransactionReq struct {
	req
	answers ChallengeAnswerList
}

// Context sets the context to be used during this request. If no context is supplied then
// the request will use context.Background.
func (r *DeleteRepeatedTransactionReq) Context(ctx context.Context) *DeleteRepeatedTransactionReq {
	r.req.ctx = ctx
	return r
}

// ChallengeAnswer adds an answer to one of the authorisation challenges required to complete the deletion.
func (r *DeleteRepeatedTransactionReq) ChallengeAnswer(answer ChallengeAnswer) *DeleteRepeatedTransactionReq {
	r.answers = append(r.answers, answer)
	return r
}

// Send sends the request to delete a repeated transaction. It returns
// information about the long running recurring transfer job that may be used
// to track and progress the deletion.
func (r *DeleteRepeatedTransactionReq) Send() (*RecurringTransfer, error) {
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

	var tr RecurringTransfer
	if err := json.NewDecoder(res.Body).Decode(&tr); err != nil {
		return nil, wrap(errContextInvalidServiceResponse, err)
	}

	return &tr, nil
}

// Update returns a request that may be used to update a repeated transaction.
func (r *RepeatedTransactionsService) Update(id string, to TransferAddress, amount MoneyAmount) *UpdateRepeatedTransactionReq {
	return &UpdateRepeatedTransactionReq{
		req: r.client.newReq(apiV1 + "/repeated_transaction/" + id),
		data: transferParams{
			To:     to,
			Amount: amount,
			Type:   TransferTypeRecurring,
		},
	}
}

type UpdateRepeatedTransactionReq struct {
	req
	data transferParams
}

// Context sets the context to be used during this request. If no context is supplied then
// the request will use context.Background.
func (r *UpdateRepeatedTransactionReq) Context(ctx context.Context) *UpdateRepeatedTransactionReq {
	r.req.ctx = ctx
	return r
}

// Schedule sets a recurrence schedule for the transaction.
func (r *UpdateRepeatedTransactionReq) Schedule(rule RecurrenceRule) *UpdateRepeatedTransactionReq {
	r.data.Schedule = &rule
	return r
}

// Description sets a human readable description for the transaction.
func (r *UpdateRepeatedTransactionReq) Description(s string) *UpdateRepeatedTransactionReq {
	r.data.Usage = s
	return r
}

// ChallengeAnswer adds an answer to one of the authorisation challenges required to complete the update.
func (r *UpdateRepeatedTransactionReq) ChallengeAnswer(answer ChallengeAnswer) *UpdateRepeatedTransactionReq {
	r.data.ChallengeAnswers = append(r.data.ChallengeAnswers, answer)
	return r
}

// Send sends the request to update a repeated transaction. It returns
// information about the long running recurring transfer job that may be used
// to track and progress the update.
func (r *UpdateRepeatedTransactionReq) Send() (*RecurringTransfer, error) {
	res, cleanup, err := r.req.putJSON(r.data)
	defer cleanup()
	if err != nil {
		return nil, err
	}

	var tr RecurringTransfer
	if err := json.NewDecoder(res.Body).Decode(&tr); err != nil {
		return nil, wrap(errContextInvalidServiceResponse, err)
	}

	return &tr, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
//             TRANSFERS SERVICE
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// TransfersService provides access to money transfer related API services.
type TransfersService struct {
	client *UserClient
}

type transferParams struct {
	From             int64               `json:"from,omitempty"`
	To               TransferAddress     `json:"to,omitempty"`
	Amount           MoneyAmount         `json:"amount,omitempty"`
	Schedule         *RecurrenceRule     `json:"schedule,omitempty"`
	EntryDate        string              `json:"entry_date,omitempty"`
	Usage            string              `json:"usage,omitempty"`
	Type             TransferType        `json:"type,omitempty"`
	ChallengeAnswers ChallengeAnswerList `json:"challenge_answers,omitempty"`
}

type transferProcessParams struct {
	Intent           TransferIntent      `json:"intent"`
	Version          int                 `json:"version,omitempty"`
	Type             TransferType        `json:"type"`
	Confirm          bool                `json:"confirm,omitempty"`
	ChallengeAnswers ChallengeAnswerList `json:"challenge_answers,omitempty"`
}

func NewTransfersService(u *UserClient) *TransfersService {
	return &TransfersService{client: u}
}

// Create returns a request that may be used to create a money transfer.
func (t *TransfersService) Create(from int64, to TransferAddress, amount MoneyAmount) *CreateTransferReq {
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
func (r *CreateTransferReq) Send() (*Transfer, error) {
	res, cleanup, err := r.req.postJSON(r.data)
	defer cleanup()
	if err != nil {
		return nil, err
	}

	var tr Transfer
	if err := json.NewDecoder(res.Body).Decode(&tr); err != nil {
		return nil, wrap(errContextInvalidServiceResponse, err)
	}

	return &tr, nil
}

// Process returns a request that may be used to update information and answer challenges for a transfer.
func (t *TransfersService) Process(id string, intent TransferIntent, version int) *ProcessTransferReq {
	return &ProcessTransferReq{
		req: t.client.newReq(apiV1 + "/transfers/" + id),
		data: transferProcessParams{
			Intent:  intent,
			Version: version,
			Type:    TransferTypeRegular,
		},
	}
}

type ProcessTransferReq struct {
	req
	data transferProcessParams
}

// Context sets the context to be used during this request. If no context is supplied then
// the request will use context.Background.
func (r *ProcessTransferReq) Context(ctx context.Context) *ProcessTransferReq {
	r.req.ctx = ctx
	return r
}

// Confirm sets whether the user has confirmed a transfer that appears to be similar to another that was recently sent.
func (r *ProcessTransferReq) Confirm(confirm bool) *ProcessTransferReq {
	r.data.Confirm = confirm
	return r
}

// ChallengeAnswer adds an answer to one of the authorisation challenges required to complete the transfer.
func (r *ProcessTransferReq) ChallengeAnswer(answer ChallengeAnswer) *ProcessTransferReq {
	r.data.ChallengeAnswers = append(r.data.ChallengeAnswers, answer)
	return r
}

// Send sends the request to update information and answer challenges for a transfer.
func (r *ProcessTransferReq) Send() (*Transfer, error) {
	res, cleanup, err := r.req.postJSON(&r.data)
	defer cleanup()
	if err != nil {
		return nil, err
	}

	var tr Transfer
	if err := json.NewDecoder(res.Body).Decode(&tr); err != nil {
		return nil, wrap(errContextInvalidServiceResponse, err)
	}

	return &tr, nil
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
func (r *CancelTransferReq) Send() (*Transfer, error) {
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

	var tr Transfer
	if err := json.NewDecoder(res.Body).Decode(&tr); err != nil {
		return nil, wrap(errContextInvalidServiceResponse, err)
	}

	return &tr, nil
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

// Create returns a request that may be used to create a money transfer. from is an account id belonging to the user.
func (t *RecurringTransfersService) Create(from int64, to TransferAddress, amount MoneyAmount, rule RecurrenceRule) *CreateRecurringTransferReq {
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
func (r *CreateRecurringTransferReq) Send() (*RecurringTransfer, error) {
	res, cleanup, err := r.req.postJSON(r.data)
	defer cleanup()
	if err != nil {
		return nil, err
	}

	var tr RecurringTransfer
	if err := json.NewDecoder(res.Body).Decode(&tr); err != nil {
		return nil, wrap(errContextInvalidServiceResponse, err)
	}

	return &tr, nil
}

// Process returns a request that may be used to update information and answer challenges for a transfer.
func (t *RecurringTransfersService) Process(id string, intent TransferIntent, version int) *ProcessRecurringTransferReq {
	return &ProcessRecurringTransferReq{
		req: t.client.newReq(apiV1 + "/transfers/" + id),
		data: transferProcessParams{
			Intent:  intent,
			Version: version,
			Type:    TransferTypeRecurring,
		},
	}
}

type ProcessRecurringTransferReq struct {
	req
	data transferProcessParams
}

// Context sets the context to be used during this request. If no context is supplied then
// the request will use context.Background.
func (r *ProcessRecurringTransferReq) Context(ctx context.Context) *ProcessRecurringTransferReq {
	r.req.ctx = ctx
	return r
}

// Confirm sets whether the user has confirmed a transfer that appears to be similar to another that was recently sent.
func (r *ProcessRecurringTransferReq) Confirm(confirm bool) *ProcessRecurringTransferReq {
	r.data.Confirm = confirm
	return r
}

// ChallengeAnswer adds an answer to one of the authorisation challenges required to complete the transfer.
func (r *ProcessRecurringTransferReq) ChallengeAnswer(answer ChallengeAnswer) *ProcessRecurringTransferReq {
	r.data.ChallengeAnswers = append(r.data.ChallengeAnswers, answer)
	return r
}

// Send sends the request to update information and answer challenges for a transfer.
func (r *ProcessRecurringTransferReq) Send() (*RecurringTransfer, error) {
	res, cleanup, err := r.req.postJSON(&r.data)
	defer cleanup()
	if err != nil {
		return nil, err
	}

	var tr RecurringTransfer
	if err := json.NewDecoder(res.Body).Decode(&tr); err != nil {
		return nil, wrap(errContextInvalidServiceResponse, err)
	}

	return &tr, nil
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
func (r *CancelRecurringTransferReq) Send() (*RecurringTransfer, error) {
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

	var tr RecurringTransfer
	if err := json.NewDecoder(res.Body).Decode(&tr); err != nil {
		return nil, wrap(errContextInvalidServiceResponse, err)
	}

	return &tr, nil
}
