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

	Accesses              *AccessesService
	Jobs                  *JobsService
	Accounts              *AccountsService
	Transactions          *TransactionsService
	ScheduledTransactions *ScheduledTransactionsService
	RepeatedTransactions  *RepeatedTransactionsService
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
	ID           int64                `json:"id"`
	BankId       int64                `json:"bank_id"`
	BankAccessId int64                `json:"bank_access_id"`
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

type AccountsService struct {
	client *UserClient
}

func NewAccountsService(u *UserClient) *AccountsService { return &AccountsService{client: u} }

func (a *AccountsService) List() *ListAccountsReq {
	return &ListAccountsReq{
		req: a.client.newReq("/v1/accounts"),
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
		return nil, err
	}

	return &page, nil
}

type AccountPage struct {
	Accounts []Account
}

func (a *AccountsService) Get(id string) *GetAccountReq {
	return &GetAccountReq{
		req: a.client.newReq("/v1/accounts/" + id),
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
		return nil, err
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
	Amount *Money `json:"amount,omitempty"`

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

// TODO: unmarshal money
type Money struct {
	E uint8  `json:"exp,omitempty"`  // negative exponent, or inverse exponent 10^-E
	S int64  `json:"val,omitempty"`  // significand, or mantissa
	C uint16 `json:"code,omitempty"` // currency number (iso-4217:2008) and bitmasked status flag - highest 3 bits are reserved for error/statusflags
}

type TransactionsService struct {
	client *UserClient
}

func NewTransactionsService(u *UserClient) *TransactionsService {
	return &TransactionsService{client: u}
}

func (a *TransactionsService) List() *ListTransactionsReq {
	return &ListTransactionsReq{
		req: a.client.newReq("/v1/transactions"),
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
		return nil, err
	}

	return &page, nil
}

type TransactionPage struct {
	Transactions []Transaction
}

func (a *TransactionsService) Get(id string) *GetTransactionReq {
	return &GetTransactionReq{
		req: a.client.newReq("/v1/transactions/" + id),
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
		return nil, err
	}

	return &tx, nil
}

type ScheduledTransactionsService struct {
	client *UserClient
}

func NewScheduledTransactionsService(u *UserClient) *ScheduledTransactionsService {
	return &ScheduledTransactionsService{client: u}
}

func (a *ScheduledTransactionsService) List() *ListScheduledTransactionsReq {
	return &ListScheduledTransactionsReq{
		req: a.client.newReq("/v1/scheduled_transactions"),
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
		return nil, err
	}

	return &page, nil
}

func (a *ScheduledTransactionsService) Get(id string) *GetScheduledTransactionReq {
	return &GetScheduledTransactionReq{
		req: a.client.newReq("/v1/scheduled_transactions/" + id),
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
		return nil, err
	}

	return &tx, nil
}

type RepeatedTransactionsService struct {
	client *UserClient
}

func NewRepeatedTransactionsService(u *UserClient) *RepeatedTransactionsService {
	return &RepeatedTransactionsService{client: u}
}

func (a *RepeatedTransactionsService) List() *ListRepeatedTransactionsReq {
	return &ListRepeatedTransactionsReq{
		req: a.client.newReq("/v1/repeated_transactions"),
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
		return nil, err
	}

	return &page, nil
}

type RepeatedTransactionPage struct {
	Transactions []RepeatedTransaction
}

func (a *RepeatedTransactionsService) Get(id string) *GetRepeatedTransactionReq {
	return &GetRepeatedTransactionReq{
		req: a.client.newReq("/v1/repeated_transactions/" + id),
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
		return nil, err
	}

	return &tx, nil
}

type RepeatedTransaction struct {
	ID            int64      `json:"id"`
	UserAccountID int64      `json:"user_bank_account_id"`
	UserAccount   AccountRef `json:"user_account"`
	RemoteAccount AccountRef `json:"remote_account"`
	AccessID      int64      `json:"user_bank_access_id"`
	RemoteID      string     `json:"remote_id"`
	Schedule      RRule      `json:"schedule"`
	Amount        *Money     `json:"amount"`
	Usage         string     `json:"usage"`
}

type RRule struct {
	Start    time.Time `json:"start"`
	Until    time.Time `json:"until"`
	Freq     string    `json:"frequency"`
	Interval int       `json:"interval"`
	ByDay    int       `json:"by_day"`
}
