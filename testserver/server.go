package testserver

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bankrs/bosgo"
)

type Dev struct {
	ID string
}

type App struct {
	ID          string
	DeveloperID string
}

type User struct {
	ID                   string
	Username             string
	Password             string
	ApplicationID        string
	Accesses             []bosgo.Access
	Transactions         []bosgo.Transaction
	RepeatedTransactions []bosgo.RepeatedTransaction
}

type Job struct {
	ID              string
	UserID          string
	ProviderID      string
	Stage           bosgo.JobStage
	Error           string
	SuppliedAnswers []bosgo.ChallengeAnswer
	AccessDetails   AccessDetails
	Succeeded       bool
}

type OrderOp int

const (
	OrderOpCreate OrderOp = iota
	OrderOpUpdate
	OrderOpDelete
)

type TransferOrder struct {
	UserID        string
	Operation     OrderOp
	Type          bosgo.TransferType
	Transfer      bosgo.Transfer
	AccessDetails AccessDetails
}

type AccessDetails struct {
	Access               bosgo.Access
	Transactions         []bosgo.Transaction
	RepeatedTransactions []bosgo.RepeatedTransaction
	ChallengeMap         map[string]string
	TransferAuths        []TransferAuth
}

type TransferAuth struct {
	Method  string
	Message string
	Answer  string
}

const (
	transferInit = "transfer_init"
)

func (j *Job) isAnswered(id, val string) bool {
	for _, ans := range j.SuppliedAnswers {
		if ans.ID == id && ans.Value == val {
			return true
		}
	}
	return false
}

type Logger interface {
	Logf(format string, args ...interface{})
}

type Server struct {
	Svr *httptest.Server
	mux *http.ServeMux

	mu         sync.Mutex // guards following fields
	id         int64
	logger     Logger
	Devs       map[string]Dev           // map of developers indexed by ID
	Apps       map[string]App           // map of applications indexed by ID
	Users      map[string]User          // map of users indexed by ID
	UserTokens map[string]string        // map of tokens to user ID
	Jobs       map[string]Job           // map of jobs indexed by ID
	Accesses   map[string]AccessDetails // map of access details indexed by provider ID
	Transfers  map[string]TransferOrder // map of transfer orders indexed by ID
}

func New() *Server {
	s := Server{
		Devs:       make(map[string]Dev),
		Apps:       make(map[string]App),
		Users:      make(map[string]User),
		UserTokens: make(map[string]string),
		Jobs:       make(map[string]Job),
		Accesses:   make(map[string]AccessDetails),
		Transfers:  make(map[string]TransferOrder),
	}
	s.Svr = httptest.NewTLSServer(&s)

	s.mux = http.NewServeMux()
	s.mux.HandleFunc("/v1/users", s.handleUsers)
	s.mux.HandleFunc("/v1/users/login", s.handleUsersLogin)
	s.mux.HandleFunc("/v1/users/logout", s.handleUsersLogout)
	s.mux.HandleFunc("/v1/users/reset_password", s.handleUsersResetPassword)

	s.mux.HandleFunc("/v1/accesses", s.handleAccesses)
	s.mux.HandleFunc("/v1/accesses/", s.handleAccess)
	s.mux.HandleFunc("/v1/accounts", s.handleAccounts)
	s.mux.HandleFunc("/v1/jobs/", s.handleJobs)
	s.mux.HandleFunc("/v1/transactions", s.handleTransactions)
	s.mux.HandleFunc("/v1/repeated_transactions", s.handleRepeatedTransactions)
	s.mux.HandleFunc("/v1/transfers", s.handleTransfers)
	s.mux.HandleFunc("/v1/transfers/", s.handleTransfer)

	return &s
}

func (s *Server) URL() string {
	return s.Svr.URL
}

func (s *Server) Addr() string {
	u, _ := url.Parse(s.Svr.URL)

	return u.Host
}

func (s *Server) SetLogger(logger Logger) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.logger = logger
}

func (s *Server) Logf(format string, args ...interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.logger == nil {
		return
	}
	s.logger.Logf(format, args...)
}

func (s *Server) Client() *http.Client {
	cert, err := x509.ParseCertificate(s.Svr.TLS.Certificates[0].Certificate[0])
	if err != nil {
		panic(err)
	}

	pool := x509.NewCertPool()
	pool.AddCert(cert)

	client := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: pool,
			},
		},
	}

	return &client
}

func (s *Server) Close() {
	s.Svr.Close()
	s.Svr = nil
}

func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	s.Logf("received request: %s %s", req.Method, req.URL.Path)
	s.mux.ServeHTTP(w, req)
}

type errorResp struct {
	Errors []apiError `json:"errors"`
}

// APIError represents an error that API may deliver to the customer
type apiError struct {
	Code    string                 `json:"code"`
	Payload map[string]interface{} `json:"payload,omitempty"`
}

func (s *Server) sendError(w http.ResponseWriter, status int, errcode string) {
	resp := errorResp{
		Errors: []apiError{
			{
				Code: errcode,
			},
		},
	}
	s.sendJSON(w, status, resp)
}

func (s *Server) sendJSON(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)

	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(v)
	if err != nil {
		return
	}

	bb := buf.Bytes()
	s.Logf("wrote %d bytes: %.512s", len(bb), string(bb))
	buf.WriteTo(w)
}

func (s *Server) sendNoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) readJSON(w http.ResponseWriter, req *http.Request, v interface{}) bool {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		s.Logf("failed to read body: %v", err)
		s.sendError(w, http.StatusBadRequest, "general")
		return false
	}

	s.Logf("read %d bytes: %.512s", len(body), string(body))
	if err := json.Unmarshal(body, v); err != nil {
		s.Logf("failed to unmarshal body: %v", err)
		s.sendError(w, http.StatusBadRequest, "general")
		return false
	}

	return true
}

func (s *Server) nextIDStr() string {
	id := s.nextID()
	return fmt.Sprintf("%08x", id)
}

func (s *Server) nextID() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.id++
	return s.id
}

func (s *Server) getApp(id string) (App, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.Apps == nil {
		return App{}, false
	}
	app, exists := s.Apps[id]
	return app, exists
}

func (s *Server) setApp(app App) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Apps[app.ID] = app
}

func (s *Server) requireApp(w http.ResponseWriter, req *http.Request) (App, bool) {
	id := req.Header.Get("X-Application-Id")
	if id == "" {
		s.sendError(w, http.StatusUnauthorized, "authentication_app_id_invalid")
		return App{}, false
	}
	app, exists := s.getApp(id)
	if !exists {
		s.sendError(w, http.StatusUnauthorized, "authentication_app_id_invalid")
		return App{}, false
	}
	return app, true
}

func (s *Server) GetUser(id string) (User, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.Users == nil {
		return User{}, false
	}
	user, exists := s.Users[id]
	return user, exists
}

func (s *Server) GetUserByName(name string) (User, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.Users == nil {
		return User{}, false
	}
	for _, user := range s.Users {
		if user.Username == name {
			return user, true
		}
	}
	return User{}, false
}

func (s *Server) setUser(user User) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if user.Accesses == nil {
		user.Accesses = []bosgo.Access{}
	}
	s.Users[user.ID] = user
}

func (s *Server) requireUser(w http.ResponseWriter, req *http.Request) (User, string, bool) {
	app, exists := s.requireApp(w, req)
	if !exists {
		return User{}, "", false
	}

	token := req.Header.Get("X-Token")

	s.mu.Lock()
	id, exists := s.UserTokens[token]
	s.mu.Unlock()

	if !exists {
		s.sendError(w, http.StatusUnauthorized, "authentication_failed")
		return User{}, "", false
	}
	user, found := s.GetUser(id)
	if !found || user.ApplicationID != app.ID {
		s.sendError(w, http.StatusUnauthorized, "authentication_failed")
		return User{}, "", false
	}

	return user, token, found
}

func (s *Server) setUserLoggedIn(id string) string {
	token := s.nextIDStr()
	s.mu.Lock()
	defer s.mu.Unlock()
	s.UserTokens[token] = id
	return token
}

func (s *Server) setUserLoggedOut(token string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.UserTokens, token)
}

func (s *Server) newJob(userID string, providerID string, answers []bosgo.ChallengeAnswer) *bosgo.Job {
	job := Job{
		ID:         s.nextIDStr(),
		UserID:     userID,
		ProviderID: providerID,
		Stage:      bosgo.JobStageAuthenticating,
	}

	s.mu.Lock()
	ad, exists := s.Accesses[providerID]
	s.mu.Unlock()
	if !exists {
		job.Stage = bosgo.JobStageFinished
		job.Error = "unknown_provider"
	} else {
		job.AccessDetails = ad
		s.progressJob(&job, answers)
	}

	s.mu.Lock()
	s.Jobs[job.ID] = job
	s.mu.Unlock()

	return &bosgo.Job{
		URI: "/jobs/" + job.ID,
	}

}

func (s *Server) setJob(job Job) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Jobs[job.ID] = job
}

func (s *Server) getJob(id string) (Job, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	job, exists := s.Jobs[id]
	return job, exists
}

func (s *Server) requireJob(w http.ResponseWriter, req *http.Request) (Job, bool) {
	user, _, found := s.requireUser(w, req)
	if !found {
		return Job{}, false
	}
	if !strings.HasPrefix(req.URL.Path, "/v1/jobs/") {
		s.sendError(w, http.StatusBadRequest, "general")
		return Job{}, false
	}
	jobID := req.URL.Path[9:]

	job, exists := s.getJob(jobID)
	if !exists {
		s.sendError(w, http.StatusNotFound, "resource_not_found")
		return Job{}, false
	}
	if job.UserID != user.ID {
		s.sendError(w, http.StatusUnauthorized, "authentication_failed")
		return Job{}, false
	}

	return job, true
}

func (s *Server) progressJob(j *Job, answers []bosgo.ChallengeAnswer) {
	j.SuppliedAnswers = append(j.SuppliedAnswers, answers...)

	for id, val := range j.AccessDetails.ChallengeMap {
		if !j.isAnswered(id, val) {
			return
		}
	}

	j.Stage = bosgo.JobStageFinished
	j.Succeeded = true

	user, found := s.GetUser(j.UserID)
	if !found {
		return
	}
	user.Accesses = append(user.Accesses, j.AccessDetails.Access)
	user.Transactions = append(user.Transactions, j.AccessDetails.Transactions...)
	user.RepeatedTransactions = append(user.RepeatedTransactions, j.AccessDetails.RepeatedTransactions...)

	s.setUser(user)
}

func (s *Server) requireAccess(w http.ResponseWriter, req *http.Request) (bosgo.Access, bool) {
	user, _, found := s.requireUser(w, req)
	if !found {
		return bosgo.Access{}, false
	}
	if !strings.HasPrefix(req.URL.Path, "/v1/accesses/") {
		s.sendError(w, http.StatusBadRequest, "general")
		return bosgo.Access{}, false
	}

	accessID, err := strconv.ParseInt(req.URL.Path[13:], 10, 64)
	if err != nil {
		s.Logf("failed to parse accessID: %v", err)
		s.sendError(w, http.StatusBadRequest, "general")
		return bosgo.Access{}, false
	}

	for _, acc := range user.Accesses {
		if acc.ID == accessID {
			return acc, true
		}
	}
	s.sendError(w, http.StatusNotFound, "resource_not_found")
	return bosgo.Access{}, false
}

func (s *Server) newTransfer(userID string, providerID string, trp *transferParams) TransferOrder {
	tr := TransferOrder{
		Transfer: bosgo.Transfer{
			ID: s.nextIDStr(),
		},
		UserID: userID,
		Type:   trp.Type,
	}

	s.mu.Lock()
	ad, exists := s.Accesses[providerID]
	s.mu.Unlock()
	if !exists {
		tr.Transfer.State = bosgo.TransferStateFailed
		tr.Transfer.Errors = append(tr.Transfer.Errors, bosgo.Problem{Code: "resource_not_found"})
		s.setTransfer(tr)
		return tr
	}

	tr.AccessDetails = ad
	tr.Transfer.State = bosgo.TransferStateOngoing
	tr.Transfer.Step = bosgo.TransferStep{
		Intent: transferInit,
	}
	s.progressTransfer(&tr, false, trp.ChallengeAnswers)
	s.setTransfer(tr)
	return tr

}

func (s *Server) setTransfer(tr TransferOrder) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Transfers[tr.Transfer.ID] = tr
}

func (s *Server) getTransfer(id string) (TransferOrder, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	tr, exists := s.Transfers[id]
	return tr, exists
}

func (s *Server) requireTransfer(w http.ResponseWriter, req *http.Request) (TransferOrder, bool) {
	user, _, found := s.requireUser(w, req)
	if !found {
		return TransferOrder{}, false
	}
	if !strings.HasPrefix(req.URL.Path, "/v1/transfers/") {
		s.sendError(w, http.StatusBadRequest, "general")
		return TransferOrder{}, false
	}
	id := req.URL.Path[14:]

	tr, exists := s.getTransfer(id)
	if !exists {
		s.sendError(w, http.StatusNotFound, "resource_not_found")
		return TransferOrder{}, false
	}
	if tr.UserID != user.ID {
		s.sendError(w, http.StatusUnauthorized, "authentication_failed")
		return TransferOrder{}, false
	}

	return tr, true
}

func (s *Server) progressTransfer(tr *TransferOrder, confirm bool, answers []bosgo.ChallengeAnswer) {
	switch tr.Transfer.Step.Intent {
	case transferInit:
		tr.Transfer.State = bosgo.TransferStateOngoing
		tr.Transfer.Step = bosgo.TransferStep{
			Intent: bosgo.TransferIntentProvidePIN,
		}
		if len(answers) == 0 {
			break
		}
		fallthrough

	case bosgo.TransferIntentProvidePIN:
		tr.Transfer.State = bosgo.TransferStateOngoing
		for _, ans := range answers {
			if ans.ID == "pin" && ans.Value == tr.AccessDetails.ChallengeMap["pin"] {
				tr.Transfer.Step = bosgo.TransferStep{
					Intent: bosgo.TransferIntentSelectAuthMethod,
					Data: &bosgo.TransferStepData{
						TANType:     bosgo.TANTypeMobile,
						AuthMethods: []bosgo.AuthMethod{},
					},
				}

				for _, ta := range tr.AccessDetails.TransferAuths {
					tr.Transfer.Step.Data.AuthMethods = append(tr.Transfer.Step.Data.AuthMethods, bosgo.AuthMethod{ID: ta.Method})
				}

				return
			}
		}

		// No pin supplied or it didn't match
		tr.Transfer.Errors = append(tr.Transfer.Errors, bosgo.Problem{Code: "fi_invalid_loginname_pin"})
		tr.Transfer.Step = bosgo.TransferStep{
			Intent: bosgo.TransferIntentProvidePIN,
		}

	case bosgo.TransferIntentSelectAuthMethod:
		tr.Transfer.State = bosgo.TransferStateOngoing
		for _, ans := range answers {
			if ans.ID == "auth_method" {
				for _, ta := range tr.AccessDetails.TransferAuths {
					if ans.Value == ta.Method {
						tr.Transfer.Step = bosgo.TransferStep{
							Intent: bosgo.TransferIntentProvideChallengeAnswer,
							Data: &bosgo.TransferStepData{
								ChallengeMessage: ta.Message,
							},
						}
						return
					}
				}
			}
		}
		tr.Transfer.Errors = append(tr.Transfer.Errors, bosgo.Problem{Code: "fi_account_blocked"})
		tr.Transfer.State = bosgo.TransferStateFailed
		tr.Transfer.Step = bosgo.TransferStep{}

	case bosgo.TransferIntentProvideChallengeAnswer:
		for _, ans := range answers {
			if ans.ID == "tan" {
				for _, ta := range tr.AccessDetails.TransferAuths {
					if ans.Value == ta.Answer {
						tr.Transfer.State = bosgo.TransferStateSucceeded
						tr.Transfer.Step = bosgo.TransferStep{}
						now := time.Now()
						tr.Transfer.EntryDate = now
						tr.Transfer.SettlementDate = now
						return
					}
				}
			}
		}

		tr.Transfer.Errors = append(tr.Transfer.Errors, bosgo.Problem{Code: "fi_account_blocked"})
		tr.Transfer.Step = bosgo.TransferStep{}
	case bosgo.TransferIntentConfirmSimilarTransfer:
		tr.Transfer.State = bosgo.TransferStateOngoing
		tr.Transfer.Step = bosgo.TransferStep{
			Intent: bosgo.TransferIntentProvidePIN,
		}
	}
}

// AddAccess adds configuration for an access with its transactions so it can be added to a user via the server API
func (s *Server) AddAccess(ad AccessDetails) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Accesses[ad.Access.ProviderID] = ad
}

// AssignAccess assigns a known access to a user
func (s *Server) AssignAccess(username string, access *bosgo.Access) error {
	user, found := s.GetUserByName(username)
	if !found {
		return fmt.Errorf("unknown user: %s", username)
	}

	user.Accesses = append(user.Accesses, *access)
	s.setUser(user)

	return nil
}

// AssignTransactions assigns a set of transactions to a user, overwriting any existing transactions
func (s *Server) AssignTransactions(username string, txs []bosgo.Transaction) error {
	user, found := s.GetUserByName(username)
	if !found {
		return fmt.Errorf("unknown user: %s", username)
	}

	user.Transactions = txs
	s.setUser(user)

	return nil
}

// AssignRepeatedTransactions assigns a set of repeated transactions to a user, overwriting any existing transactions
func (s *Server) AssignRepeatedTransactions(username string, txs []bosgo.RepeatedTransaction) error {
	user, found := s.GetUserByName(username)
	if !found {
		return fmt.Errorf("unknown user: %s", username)
	}

	user.RepeatedTransactions = txs
	s.setUser(user)

	return nil
}

func (s *Server) handleUsers(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodPost:
		s.handleUserCreate(w, req)
		return
	case http.MethodDelete:
		s.handleUserDelete(w, req)
		return
	}

	http.Error(w, "405 Method Not Allowed", http.StatusMethodNotAllowed)
	return
}

func (s *Server) handleUserCreate(w http.ResponseWriter, req *http.Request) {
	app, proceed := s.requireApp(w, req)
	if !proceed {
		return
	}

	defer req.Body.Close()
	var creds bosgo.UserCredentials
	if !s.readJSON(w, req, &creds) {
		return
	}

	s.mu.Lock()
	for _, u := range s.Users {
		if u.Username == creds.Username {
			s.mu.Unlock()
			s.sendError(w, http.StatusInternalServerError, "server_side")
			return
		}
	}
	s.mu.Unlock()

	user := User{
		ID:            s.nextIDStr(),
		Password:      creds.Password,
		ApplicationID: app.ID,
	}

	s.setUser(user)
	token := s.setUserLoggedIn(user.ID)

	ut := bosgo.UserToken{
		ID:    user.ID,
		Token: token,
	}
	s.sendJSON(w, http.StatusCreated, &ut)
}

func (s *Server) handleUserDelete(w http.ResponseWriter, req *http.Request) {
	s.sendError(w, http.StatusInternalServerError, "not_implemented_by_test_server")
}

func (s *Server) handleUsersLogin(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "405 Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	app, proceed := s.requireApp(w, req)
	if !proceed {
		return
	}

	var creds bosgo.UserCredentials
	if !s.readJSON(w, req, &creds) {
		return
	}

	s.mu.Lock()
	var user User
	for _, u := range s.Users {
		if u.Username != creds.Username {
			continue
		}
		if u.Password != creds.Password {
			break
		}
		user = u
		break
	}
	s.mu.Unlock()

	if user.ApplicationID != app.ID {
		s.sendError(w, http.StatusUnauthorized, "authentication_failed")
		return
	}

	token := s.setUserLoggedIn(user.ID)
	ut := bosgo.UserToken{
		ID:    user.ID,
		Token: token,
	}
	s.sendJSON(w, http.StatusOK, &ut)
}

func (s *Server) handleUsersLogout(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "405 Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	_, token, found := s.requireUser(w, req)
	if !found {
		return
	}
	s.setUserLoggedOut(token)
	s.sendNoContent(w)
}

func (s *Server) handleUsersResetPassword(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "405 Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	s.sendError(w, http.StatusInternalServerError, "not_implemented_by_test_server")
}

func (s *Server) handleAccesses(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodPost:
		s.handleAccessCreate(w, req)
		return
	case http.MethodGet:
		s.handleAccessesList(w, req)
		return
	}

	http.Error(w, "405 Method Not Allowed", http.StatusMethodNotAllowed)
	return
}

func (s *Server) handleAccessCreate(w http.ResponseWriter, req *http.Request) {
	user, _, found := s.requireUser(w, req)
	if !found {
		return
	}

	var data struct {
		ProviderID string                  `json:"provider_id"`
		Answers    []bosgo.ChallengeAnswer `json:"challenge_answers"`
	}
	if !s.readJSON(w, req, &data) {
		return
	}

	job := s.newJob(user.ID, data.ProviderID, data.Answers)

	s.sendJSON(w, http.StatusAccepted, &job)
}

func (s *Server) handleAccessesList(w http.ResponseWriter, req *http.Request) {
	user, _, found := s.requireUser(w, req)
	if !found {
		return
	}

	s.sendJSON(w, http.StatusOK, user.Accesses)
}

func (s *Server) handleJobs(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		s.handleJobStatus(w, req)
		return
	case http.MethodPut:
		s.handleJobAnswer(w, req)
		return
	case http.MethodDelete:
		s.handleJobDelete(w, req)
		return
	}

	http.Error(w, "405 Method Not Allowed", http.StatusMethodNotAllowed)
	return
}

func (s *Server) handleJobStatus(w http.ResponseWriter, req *http.Request) {
	job, found := s.requireJob(w, req)
	if !found {
		return
	}
	s.sendJSON(w, http.StatusOK, s.jobStatus(&job))
}

func (s *Server) handleJobAnswer(w http.ResponseWriter, req *http.Request) {
	job, found := s.requireJob(w, req)
	if !found {
		return
	}

	var answers struct {
		Answers []bosgo.ChallengeAnswer `json:"challenge_answers"`
	}
	if !s.readJSON(w, req, &answers) {
		return
	}

	s.progressJob(&job, answers.Answers)
	s.setJob(job)

	s.sendJSON(w, http.StatusOK, s.jobStatus(&job))
}

func (s *Server) handleJobDelete(w http.ResponseWriter, req *http.Request) {
	job, found := s.requireJob(w, req)
	if !found {
		return
	}
	_ = job
	s.sendError(w, http.StatusInternalServerError, "not_implemented_by_test_server")
}

func (s *Server) jobStatus(job *Job) *bosgo.JobStatus {
	status := bosgo.JobStatus{
		Finished: job.Stage == bosgo.JobStageFinished,
		Stage:    job.Stage,
		URI:      "/jobs/" + job.ID,
		// Challenge *Challenge `json:"challenge,omitempty"`
	}

	if job.Error != "" {
		status.Errors = []bosgo.APIError{{Code: job.Error}}
	}

	if job.Succeeded {
		status.Access = &bosgo.JobAccess{
			ID:         job.AccessDetails.Access.ID,
			ProviderID: job.AccessDetails.Access.ProviderID,
			Name:       job.AccessDetails.Access.Name,
		}
		for _, ac := range job.AccessDetails.Access.Accounts {
			status.Access.Accounts = append(status.Access.Accounts, bosgo.JobAccount{
				ID:        ac.ID,
				Name:      ac.Name,
				Number:    ac.Number,
				IBAN:      ac.IBAN,
				Supported: ac.Supported,
			})
		}

	}

	return &status
}

func (s *Server) handleAccounts(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "405 Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	user, _, found := s.requireUser(w, req)
	if !found {
		return
	}

	var accounts []bosgo.Account
	for _, acc := range user.Accesses {
		accounts = append(accounts, acc.Accounts...)
	}
	s.sendJSON(w, http.StatusOK, accounts)
}

func (s *Server) handleAccess(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		s.handleAccessGet(w, req)
		return
	case http.MethodPost:
		s.sendError(w, http.StatusInternalServerError, "not_implemented_by_test_server")
		return
	case http.MethodDelete:
		s.sendError(w, http.StatusInternalServerError, "not_implemented_by_test_server")
		return
	}

	http.Error(w, "405 Method Not Allowed", http.StatusMethodNotAllowed)
	return
}

func (s *Server) handleAccessGet(w http.ResponseWriter, req *http.Request) {
	access, found := s.requireAccess(w, req)
	if !found {
		return
	}
	s.sendJSON(w, http.StatusOK, access)
}

type txParams struct {
	accessID  int64
	accountID int64
	limit     int64
	offset    int64
	since     time.Time
}

func (s *Server) parseTransactionParams(w http.ResponseWriter, req *http.Request) (txParams, bool) {
	var params txParams
	var err error

	accessIDStr := req.URL.Query().Get("access_id")
	if accessIDStr != "" {
		params.accessID, err = strconv.ParseInt(accessIDStr, 10, 64)
		if err != nil {
			s.Logf("failed to parse access_id: %v", err)
			s.sendError(w, http.StatusBadRequest, "general")
			return txParams{}, false
		}
	}

	accountIDStr := req.URL.Query().Get("account_id")
	if accountIDStr != "" {
		params.accountID, err = strconv.ParseInt(accountIDStr, 10, 64)
		if err != nil {
			s.Logf("failed to parse account_id: %v", err)
			s.sendError(w, http.StatusBadRequest, "general")
			return txParams{}, false
		}
	}

	limitStr := req.URL.Query().Get("limit")
	if limitStr != "" {
		params.limit, err = strconv.ParseInt(limitStr, 10, 64)
		if err != nil {
			s.Logf("failed to parse limit: %v", err)
			s.sendError(w, http.StatusBadRequest, "general")
			return txParams{}, false
		}
	}

	offsetStr := req.URL.Query().Get("offset")
	if offsetStr != "" {
		params.offset, err = strconv.ParseInt(offsetStr, 10, 64)
		if err != nil {
			s.Logf("failed to parse offset: %v", err)
			s.sendError(w, http.StatusBadRequest, "general")
			return txParams{}, false
		}
	}

	sinceStr := req.URL.Query().Get("since")
	if sinceStr != "" {
		params.since, err = time.Parse(time.RFC3339, sinceStr)
		if err != nil {
			s.Logf("failed to parse since: %v", err)
			s.sendError(w, http.StatusBadRequest, "general")
			return txParams{}, false
		}
	}

	if params.limit == 0 {
		params.limit = 50
	} else if params.limit > 300 {
		params.limit = 300
	}

	return params, true
}

func (s *Server) handleTransactions(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "405 Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	user, _, found := s.requireUser(w, req)
	if !found {
		return
	}

	params, ok := s.parseTransactionParams(w, req)
	if !ok {
		return
	}

	page := bosgo.TransactionPage{
		Offset: int(params.offset),
		Limit:  int(params.limit),
	}

	if params.accessID == 0 && params.accountID == 0 && params.since.IsZero() {
		page.Transactions = user.Transactions
	} else {
		page.Transactions = make([]bosgo.Transaction, 0, len(user.Transactions))
		for _, tx := range user.Transactions {
			if (params.accessID == 0 || tx.AccessID == params.accessID) &&
				(params.accountID == 0 || tx.UserAccountID == params.accountID) &&
				(params.since.IsZero() || params.since.Before(tx.EntryDate)) {
				page.Transactions = append(page.Transactions, tx)
			}
		}
	}
	page.Total = len(page.Transactions)

	start := int(params.offset)
	end := int(params.offset + params.limit)
	if end > len(page.Transactions) {
		end = len(page.Transactions)
	}

	if start > 0 || end < len(page.Transactions) {
		page.Transactions = page.Transactions[start:end]
	}

	s.sendJSON(w, http.StatusOK, page)
}

func (s *Server) handleRepeatedTransactions(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "405 Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	user, _, found := s.requireUser(w, req)
	if !found {
		return
	}

	params, ok := s.parseTransactionParams(w, req)
	if !ok {
		return
	}

	page := bosgo.RepeatedTransactionPage{
		Offset: int(params.offset),
		Limit:  int(params.limit),
	}

	if params.accessID == 0 && params.accountID == 0 {
		page.Transactions = user.RepeatedTransactions
	} else {
		page.Transactions = make([]bosgo.RepeatedTransaction, 0, len(user.RepeatedTransactions))
		for _, tx := range user.RepeatedTransactions {
			if (params.accessID == 0 || tx.AccessID == params.accessID) && (params.accountID == 0 || tx.UserAccountID == params.accountID) {
				page.Transactions = append(page.Transactions, tx)
			}
		}
	}
	page.Total = len(page.Transactions)

	start := int(params.offset)
	end := int(params.offset + params.limit)
	if end > len(page.Transactions) {
		end = len(page.Transactions)
	}

	if start > 0 || end < len(page.Transactions) {
		page.Transactions = page.Transactions[start:end]
	}

	s.sendJSON(w, http.StatusOK, page)
}

type transferParams struct {
	From             int64                     `json:"from,omitempty"`
	To               bosgo.TransferAddress     `json:"to,omitempty"`
	Amount           bosgo.MoneyAmount         `json:"amount,omitempty"`
	Schedule         *bosgo.RecurrenceRule     `json:"schedule,omitempty"`
	EntryDate        string                    `json:"entry_date,omitempty"`
	Usage            string                    `json:"usage,omitempty"`
	Type             bosgo.TransferType        `json:"type,omitempty"`
	ChallengeAnswers bosgo.ChallengeAnswerList `json:"challenge_answers,omitempty"`
}

func (s *Server) handleTransfers(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "405 Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	user, _, found := s.requireUser(w, req)
	if !found {
		return
	}

	var data transferParams

	if !s.readJSON(w, req, &data) {
		return
	}

	var providerID string
accessloop:
	for _, acc := range user.Accesses {
		for _, ac := range acc.Accounts {
			if ac.ID == data.From {
				providerID = acc.ProviderID
				break accessloop
			}
		}
	}
	if providerID == "" {
		s.sendError(w, http.StatusNotFound, "resource_not_found")
		return
	}

	tr := s.newTransfer(user.ID, providerID, &data)

	s.sendJSON(w, http.StatusCreated, &tr.Transfer)
}

func (s *Server) handleTransfer(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodPost:
		s.handleTransferProcess(w, req)
		return
	case http.MethodPut:
		s.sendError(w, http.StatusInternalServerError, "not_implemented_by_test_server")
		return
	case http.MethodDelete:
		s.handleTransferDelete(w, req)
		return
	}

	http.Error(w, "405 Method Not Allowed", http.StatusMethodNotAllowed)
	return
}

type transferProcessParams struct {
	Intent           bosgo.TransferIntent      `json:"intent"`
	Version          int                       `json:"version,omitempty"`
	Type             bosgo.TransferType        `json:"type"`
	Confirm          bool                      `json:"confirm,omitempty"`
	ChallengeAnswers bosgo.ChallengeAnswerList `json:"challenge_answers,omitempty"`
}

func (s *Server) handleTransferProcess(w http.ResponseWriter, req *http.Request) {
	tr, found := s.requireTransfer(w, req)
	if !found {
		return
	}

	var data transferProcessParams

	if !s.readJSON(w, req, &data) {
		return
	}

	if data.Version != tr.Transfer.Version {
		tr.Transfer.Errors = append(tr.Transfer.Errors, bosgo.Problem{Code: "versions_mismatch"})
		s.sendJSON(w, http.StatusOK, &tr.Transfer)
		return
	}

	if data.Intent != tr.Transfer.Step.Intent {
		tr.Transfer.Errors = append(tr.Transfer.Errors, bosgo.Problem{Code: "intents_mismatch"})
		s.sendJSON(w, http.StatusOK, &tr.Transfer)
		return
	}

	if tr.Transfer.State != bosgo.TransferStateOngoing {
		tr.Transfer.Errors = append(tr.Transfer.Errors, bosgo.Problem{Code: "state_" + string(tr.Transfer.State) + "_unprocessable"})
		s.sendJSON(w, http.StatusOK, &tr.Transfer)
		return
	}

	s.progressTransfer(&tr, data.Confirm, data.ChallengeAnswers)
	s.setTransfer(tr)

	s.sendJSON(w, http.StatusOK, &tr.Transfer)
}

func (s *Server) handleTransferDelete(w http.ResponseWriter, req *http.Request) {
	tr, found := s.requireTransfer(w, req)
	if !found {
		return
	}
	_ = tr
}

// WriteState writes the current state of the server to w as a series of JSON documents.
func (s *Server) WriteState(w io.Writer) error {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	if err := enc.Encode(s.Devs); err != nil {
		return err
	}
	if err := enc.Encode(s.Apps); err != nil {
		return err
	}
	if err := enc.Encode(s.Users); err != nil {
		return err
	}
	if err := enc.Encode(s.UserTokens); err != nil {
		return err
	}
	if err := enc.Encode(s.Jobs); err != nil {
		return err
	}
	if err := enc.Encode(s.Accesses); err != nil {
		return err
	}
	if err := enc.Encode(s.Transfers); err != nil {
		return err
	}

	if _, err := buf.WriteTo(w); err != nil {
		return err
	}
	return nil
}

// ReadState reads a series of JSON documents from r and replaces the state of the server with the read data.
func (s *Server) ReadState(r io.Reader) error {
	dec := json.NewDecoder(r)

	var tmp Server
	if err := dec.Decode(&tmp.Devs); err != nil {
		return err
	}
	if err := dec.Decode(&tmp.Apps); err != nil {
		return err
	}
	if err := dec.Decode(&tmp.Users); err != nil {
		return err
	}
	if err := dec.Decode(&tmp.UserTokens); err != nil {
		return err
	}
	if err := dec.Decode(&tmp.Jobs); err != nil {
		return err
	}
	if err := dec.Decode(&tmp.Accesses); err != nil {
		return err
	}
	if err := dec.Decode(&tmp.Transfers); err != nil {
		return err
	}

	s.Devs = tmp.Devs
	s.Apps = tmp.Apps
	s.Users = tmp.Users
	s.UserTokens = tmp.UserTokens
	s.Jobs = tmp.Jobs
	s.Accesses = tmp.Accesses
	s.Transfers = tmp.Transfers

	return nil
}
