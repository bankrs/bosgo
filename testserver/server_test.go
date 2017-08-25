package testserver

import (
	"fmt"
	"testing"

	"github.com/bankrs/bosgo"
)

func TestUserLogin(t *testing.T) {
	s := NewWithDefaults()
	if testing.Verbose() {
		s.SetLogger(t)
	}
	defer s.Close()

	appClient := bosgo.NewAppClient(s.Client(), s.Addr(), DefaultApplicationID)
	_, err := appClient.Users.Login(DefaultUsername, DefaultPassword).Send()
	if err != nil {
		t.Fatalf("failed to login as user: %v", err)
	}
}

func TestUserLoginFail(t *testing.T) {
	s := NewWithDefaults()
	if testing.Verbose() {
		s.SetLogger(t)
	}
	defer s.Close()

	appClient := bosgo.NewAppClient(s.Client(), s.Addr(), DefaultApplicationID)
	_, err := appClient.Users.Login(DefaultUsername, "foo").Send()
	if err == nil {
		t.Fatalf("got no error, wanted one")
	}
}

func TestUserLoginWrongApp(t *testing.T) {
	s := NewWithDefaults()
	if testing.Verbose() {
		s.SetLogger(t)
	}
	defer s.Close()

	appClient := bosgo.NewAppClient(s.Client(), s.Addr(), "fooapp")
	_, err := appClient.Users.Login(DefaultUsername, DefaultPassword).Send()
	if err == nil {
		t.Fatalf("got no error, wanted one")
	}
}

func TestAccessesList(t *testing.T) {
	s := NewWithDefaults()
	if testing.Verbose() {
		s.SetLogger(t)
	}
	defer s.Close()

	appClient := bosgo.NewAppClient(s.Client(), s.Addr(), DefaultApplicationID)
	userClient, err := appClient.Users.Login(DefaultUsername, DefaultPassword).Send()
	if err != nil {
		t.Fatalf("failed to login as user: %v", err)
	}

	ac, err := userClient.Accesses.List().Send()
	if err != nil {
		t.Fatalf("failed to retrieve accesses: %v", err)
	}
	if len(ac.Accesses) != 0 {
		t.Errorf("got %d accesses, wanted 0", len(ac.Accesses))
	}

}

func TestUserCreate(t *testing.T) {
	s := NewWithDefaults()
	if testing.Verbose() {
		s.SetLogger(t)
	}
	defer s.Close()

	appClient := bosgo.NewAppClient(s.Client(), s.Addr(), DefaultApplicationID)
	userClient, err := appClient.Users.Create("scooby", "sandwich").Send()
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	// Test user is authorised to see their accesses (even though there are none for the new user)
	ac, err := userClient.Accesses.List().Send()
	if err != nil {
		t.Fatalf("failed to retrieve accesses: %v", err)
	}
	if len(ac.Accesses) != 0 {
		t.Errorf("got %d accesses, wanted 0", len(ac.Accesses))
	}

}

func TestAccessCreateNoLogin(t *testing.T) {
	s := NewWithDefaults()
	if testing.Verbose() {
		s.SetLogger(t)
	}
	defer s.Close()

	appClient := bosgo.NewAppClient(s.Client(), s.Addr(), DefaultApplicationID)
	userClient, err := appClient.Users.Login(DefaultUsername, DefaultPassword).Send()
	if err != nil {
		t.Fatalf("failed to login as user: %v", err)
	}

	job, err := userClient.Accesses.Add(DefaultProviderID).Send()
	if err != nil {
		t.Fatalf("failed to add access: %v", err)
	}
	t.Logf("job URI: %s", job.URI)

	status, err := userClient.Jobs.Get(job.URI).Send()
	if err != nil {
		t.Fatalf("failed to get job status: %v", err)
	}

	if status.Finished != false {
		t.Errorf("got finished %v, wanted false", status.Finished)
	}

	if status.Stage != bosgo.JobStageAuthenticating {
		t.Errorf("got stage %v, wanted %v", status.Stage, bosgo.JobStageAuthenticating)
	}
}

func TestAccessCreateWithLogin(t *testing.T) {
	s := NewWithDefaults()
	if testing.Verbose() {
		s.SetLogger(t)
	}
	defer s.Close()

	appClient := bosgo.NewAppClient(s.Client(), s.Addr(), DefaultApplicationID)
	userClient, err := appClient.Users.Login(DefaultUsername, DefaultPassword).Send()

	if err != nil {
		t.Fatalf("failed to login as user: %v", err)
	}

	req := userClient.Accesses.Add(DefaultProviderID)

	req.ChallengeAnswer(bosgo.ChallengeAnswer{
		ID:    ChallengeLogin,
		Value: DefaultAccessLogin,
	})
	req.ChallengeAnswer(bosgo.ChallengeAnswer{
		ID:    ChallengePIN,
		Value: DefaultAccessPIN,
	})

	job, err := req.Send()
	if err != nil {
		t.Fatalf("failed to add access: %v", err)
	}
	t.Logf("job URI: %s", job.URI)

	status, err := userClient.Jobs.Get(job.URI).Send()
	if err != nil {
		t.Fatalf("failed to get job status: %v", err)
	}

	if status.Finished != true {
		t.Errorf("got finished %v, wanted true", status.Finished)
	}
	if status.Stage != bosgo.JobStageFinished {
		t.Errorf("got stage %v, wanted %v", status.Stage, bosgo.JobStageFinished)
	}
	if status.Access == nil {
		t.Fatalf("got nil access, wanted non-nil")
	}

	if status.Access.ProviderID != DefaultProviderID {
		t.Errorf("got provider id %s, wanted %s", status.Access.ProviderID, DefaultProviderID)
	}
}

func TestAccessCreateUnknownProvider(t *testing.T) {
	s := NewWithDefaults()
	if testing.Verbose() {
		s.SetLogger(t)
	}
	defer s.Close()

	appClient := bosgo.NewAppClient(s.Client(), s.Addr(), DefaultApplicationID)
	userClient, err := appClient.Users.Login(DefaultUsername, DefaultPassword).Send()
	if err != nil {
		t.Fatalf("failed to login as user: %v", err)
	}

	providerID := "bogus_provider"
	job, err := userClient.Accesses.Add(providerID).Send()
	if err != nil {
		t.Fatalf("failed to add access: %v", err)
	}
	t.Logf("job URI: %s", job.URI)

	status, err := userClient.Jobs.Get(job.URI).Send()
	if err != nil {
		t.Fatalf("failed to get job status: %v", err)
	}

	if status.Finished != true {
		t.Errorf("got finished %v, wanted true", status.Finished)
	}

	if status.Stage != bosgo.JobStageFinished {
		t.Errorf("got stage %v, wanted %v", status.Stage, bosgo.JobStageFinished)
	}

	if len(status.Errors) == 0 {
		t.Errorf("got no errors, wanted at least one")
	}
}

func TestAccessCreateMultiStep(t *testing.T) {
	s := NewWithDefaults()
	if testing.Verbose() {
		s.SetLogger(t)
	}
	defer s.Close()

	appClient := bosgo.NewAppClient(s.Client(), s.Addr(), DefaultApplicationID)
	userClient, err := appClient.Users.Login(DefaultUsername, DefaultPassword).Send()

	if err != nil {
		t.Fatalf("failed to login as user: %v", err)
	}

	job, err := userClient.Accesses.Add(DefaultProviderID).Send()
	if err != nil {
		t.Fatalf("failed to add access: %v", err)
	}
	t.Logf("job URI: %s", job.URI)

	status, err := userClient.Jobs.Get(job.URI).Send()
	if err != nil {
		t.Fatalf("failed to get job status: %v", err)
	}
	if status.Stage != bosgo.JobStageAuthenticating {
		t.Errorf("got stage %v, wanted %v", status.Stage, bosgo.JobStageAuthenticating)
	}

	req := userClient.Jobs.Answer(job.URI)
	req.ChallengeAnswer(bosgo.ChallengeAnswer{
		ID:    ChallengeLogin,
		Value: DefaultAccessLogin,
	})
	if err := req.Send(); err != nil {
		t.Fatalf("failed to answer first challenge: %v", err)
	}
	status, err = userClient.Jobs.Get(job.URI).Send()
	if err != nil {
		t.Fatalf("failed to get job status: %v", err)
	}
	if status.Stage != bosgo.JobStageAuthenticating {
		t.Errorf("got stage %v, wanted %v", status.Stage, bosgo.JobStageAuthenticating)
	}

	req = userClient.Jobs.Answer(job.URI)
	req.ChallengeAnswer(bosgo.ChallengeAnswer{
		ID:    ChallengePIN,
		Value: DefaultAccessPIN,
	})

	if err := req.Send(); err != nil {
		t.Fatalf("failed to answer second challenge: %v", err)
	}
	status, err = userClient.Jobs.Get(job.URI).Send()
	if err != nil {
		t.Fatalf("failed to get job status: %v", err)
	}
	if status.Stage != bosgo.JobStageFinished {
		t.Errorf("got stage %v, wanted %v", status.Stage, bosgo.JobStageFinished)
	}
	if status.Finished != true {
		t.Errorf("got finished %v, wanted true", status.Finished)
	}

}

func addDefaultAccess(userClient *bosgo.UserClient) (int64, int64, error) {
	req := userClient.Accesses.Add(DefaultProviderID)
	req.ChallengeAnswer(bosgo.ChallengeAnswer{
		ID:    ChallengeLogin,
		Value: DefaultAccessLogin,
	})
	req.ChallengeAnswer(bosgo.ChallengeAnswer{
		ID:    ChallengePIN,
		Value: DefaultAccessPIN,
	})

	job, err := req.Send()
	if err != nil {
		return 0, 0, err
	}

	status, err := userClient.Jobs.Get(job.URI).Send()
	if err != nil {
		return 0, 0, err
	}
	if status.Access == nil {
		return 0, 0, fmt.Errorf("no access found")
	}

	return status.Access.ID, status.Access.Accounts[0].ID, nil
}

func TestAccessCreateAddsAccessToList(t *testing.T) {
	s := NewWithDefaults()
	if testing.Verbose() {
		s.SetLogger(t)
	}
	defer s.Close()

	appClient := bosgo.NewAppClient(s.Client(), s.Addr(), DefaultApplicationID)
	userClient, err := appClient.Users.Login(DefaultUsername, DefaultPassword).Send()
	if err != nil {
		t.Fatalf("failed to login as user: %v", err)
	}

	accessID, _, err := addDefaultAccess(userClient)
	if err != nil {
		t.Fatalf("failed to add access: %v", err)
	}

	ac, err := userClient.Accesses.List().Send()
	if err != nil {
		t.Fatalf("failed to retrieve accesses: %v", err)
	}
	if len(ac.Accesses) != 1 {
		t.Errorf("got %d accesses, wanted 1", len(ac.Accesses))
	}

	if ac.Accesses[0].ID != accessID {
		t.Errorf("got access id %d, wanted %d", ac.Accesses[0].ID, accessID)
	}
}

func TestGetAccess(t *testing.T) {
	s := NewWithDefaults()
	if testing.Verbose() {
		s.SetLogger(t)
	}
	defer s.Close()

	appClient := bosgo.NewAppClient(s.Client(), s.Addr(), DefaultApplicationID)
	userClient, err := appClient.Users.Login(DefaultUsername, DefaultPassword).Send()
	if err != nil {
		t.Fatalf("failed to login as user: %v", err)
	}

	accessID, _, err := addDefaultAccess(userClient)
	if err != nil {
		t.Fatalf("failed to add access: %v", err)
	}

	acc, err := userClient.Accesses.Get(accessID).Send()
	if err != nil {
		t.Fatalf("failed to retrieve accesses: %v", err)
	}
	if acc.Name != "default access" {
		t.Errorf("got access %s, wanted %s", acc.Name, "default access")
	}
}

func TestListTransactions(t *testing.T) {
	s := NewWithDefaults()
	if testing.Verbose() {
		s.SetLogger(t)
	}
	defer s.Close()

	appClient := bosgo.NewAppClient(s.Client(), s.Addr(), DefaultApplicationID)
	userClient, err := appClient.Users.Login(DefaultUsername, DefaultPassword).Send()
	if err != nil {
		t.Fatalf("failed to login as user: %v", err)
	}

	_, _, err = addDefaultAccess(userClient)
	if err != nil {
		t.Fatalf("failed to add access: %v", err)
	}

	txs, err := userClient.Transactions.List().Send()
	if err != nil {
		t.Fatalf("failed to retrieve accesses: %v", err)
	}
	if len(txs.Transactions) != 3 {
		t.Errorf("got %d transactions, wanted 3", len(txs.Transactions))
	}
}

func TestListRepeatedTransactions(t *testing.T) {
	s := NewWithDefaults()
	if testing.Verbose() {
		s.SetLogger(t)
	}
	defer s.Close()

	appClient := bosgo.NewAppClient(s.Client(), s.Addr(), DefaultApplicationID)
	userClient, err := appClient.Users.Login(DefaultUsername, DefaultPassword).Send()
	if err != nil {
		t.Fatalf("failed to login as user: %v", err)
	}

	_, _, err = addDefaultAccess(userClient)
	if err != nil {
		t.Fatalf("failed to add access: %v", err)
	}

	txs, err := userClient.RepeatedTransactions.List().Send()
	if err != nil {
		t.Fatalf("failed to retrieve accesses: %v", err)
	}
	if len(txs.Transactions) != 1 {
		t.Errorf("got %d transactions, wanted 1", len(txs.Transactions))
	}
}

func TestCreateTransfer(t *testing.T) {
	s := NewWithDefaults()
	if testing.Verbose() {
		s.SetLogger(t)
	}
	defer s.Close()

	appClient := bosgo.NewAppClient(s.Client(), s.Addr(), DefaultApplicationID)
	userClient, err := appClient.Users.Login(DefaultUsername, DefaultPassword).Send()
	if err != nil {
		t.Fatalf("failed to login as user: %v", err)
	}

	_, accountID, err := addDefaultAccess(userClient)
	if err != nil {
		t.Fatalf("failed to add access: %v", err)
	}

	amount := bosgo.MoneyAmount{
		Currency: "EUR",
		Value:    "12.50",
	}

	addr := bosgo.TransferAddress{
		Name: "Jane Doe",
		IBAN: "DE28500105175552834822",
	}

	// Create the transfer
	transfer, err := userClient.Transfers.Create(accountID, addr, amount).Send()
	if err != nil {
		t.Fatalf("failed to create transfer: %v", err)
	}

	if len(transfer.Errors) > 0 {
		for _, pr := range transfer.Errors {
			t.Logf("got unexpected error: %s %s", pr.Code, pr.Message)
		}
		t.Fatalf("got %d errors, wanted none", len(transfer.Errors))
	}

	if transfer.Step.Intent != bosgo.TransferIntentProvidePIN {
		t.Errorf("got intent %v, wanted %v", transfer.Step.Intent, bosgo.TransferIntentProvidePIN)
	}

	// Send the pin
	req := userClient.Transfers.Process(transfer.ID, transfer.Step.Intent, transfer.Version)
	req.ChallengeAnswer(bosgo.ChallengeAnswer{
		ID:    "pin",
		Value: DefaultAccessPIN,
	})
	transfer, err = req.Send()
	if err != nil {
		t.Fatalf("failed to process pin: %v", err)
	}
	if transfer.Step.Intent != bosgo.TransferIntentSelectAuthMethod {
		t.Errorf("got intent %v, wanted %v", transfer.Step.Intent, bosgo.TransferIntentSelectAuthMethod)
	}
	if transfer.Step.Data == nil {
		t.Fatalf("got nil step data, wanted non-nil")
	}
	if len(transfer.Step.Data.AuthMethods) != 1 {
		t.Fatalf("got %d auth methods, wanted 1", len(transfer.Step.Data.AuthMethods))
	}
	if transfer.Step.Data.AuthMethods[0].ID != DefaultAuthMethod {
		t.Errorf("got auth method %v, wanted %v", transfer.Step.Data.AuthMethods[0].ID, DefaultAuthMethod)
	}

	// Send the auth method
	req = userClient.Transfers.Process(transfer.ID, transfer.Step.Intent, transfer.Version)
	req.ChallengeAnswer(bosgo.ChallengeAnswer{
		ID:    "auth_method",
		Value: DefaultAuthMethod,
	})
	transfer, err = req.Send()
	if err != nil {
		t.Fatalf("failed to process auth method: %v", err)
	}
	if transfer.Step.Intent != bosgo.TransferIntentProvideChallengeAnswer {
		t.Errorf("got intent %v, wanted %v", transfer.Step.Intent, bosgo.TransferIntentProvideChallengeAnswer)
	}
	if transfer.Step.Data == nil {
		t.Fatalf("got nil step data, wanted non-nil")
	}
	if transfer.Step.Data.ChallengeMessage != DefaultAuthMessage {
		t.Errorf("got challenge message %v, wanted %v", transfer.Step.Data.ChallengeMessage, DefaultAuthMessage)
	}

	// Send the auth answer
	req = userClient.Transfers.Process(transfer.ID, transfer.Step.Intent, transfer.Version)
	req.ChallengeAnswer(bosgo.ChallengeAnswer{
		ID:    "tan",
		Value: DefaultAuthAnswer,
	})
	transfer, err = req.Send()
	if err != nil {
		t.Fatalf("failed to process auth answer: %v", err)
	}
	if transfer.State != bosgo.TransferStateSucceeded {
		t.Errorf("got state %v, wanted %v", transfer.State, bosgo.TransferStateSucceeded)
	}
}
