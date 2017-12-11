package testserver

import (
	"bytes"
	"fmt"
	"strconv"
	"testing"
	"time"

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

	if status.Stage != bosgo.JobStageChallenge {
		t.Errorf("got stage %v, wanted %v", status.Stage, bosgo.JobStageChallenge)
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
	if status.Stage != bosgo.JobStageImported {
		t.Errorf("got stage %v, wanted %v", status.Stage, bosgo.JobStageImported)
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

	if !status.Finished {
		t.Errorf("got finished %v, wanted true", status.Finished)
	}

	if status.Stage != bosgo.JobStageProblem {
		t.Errorf("got stage %v, wanted %v", status.Stage, bosgo.JobStageProblem)
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
	if status.Stage != bosgo.JobStageChallenge {
		t.Errorf("got stage %v, wanted %v", status.Stage, bosgo.JobStageChallenge)
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
	if status.Stage != bosgo.JobStageChallenge {
		t.Errorf("got stage %v, wanted %v", status.Stage, bosgo.JobStageChallenge)
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
	if status.Stage != bosgo.JobStageImported {
		t.Errorf("got stage %v, wanted %v", status.Stage, bosgo.JobStageImported)
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

func TestListScheduledTransactions(t *testing.T) {
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

	txs, err := userClient.ScheduledTransactions.List().Send()
	if err != nil {
		t.Fatalf("failed to retrieve scheduled  transactions: %v", err)
	}
	if len(txs) != 2 {
		t.Errorf("got %d transactions, wanted 2", len(txs))
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
		t.Fatalf("failed to retrieve transactions: %v", err)
	}
	if len(txs.Transactions) != 3 {
		t.Errorf("got %d transactions, wanted 3", len(txs.Transactions))
	}
}

func TestListTransactionsLimit(t *testing.T) {
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

	txs, err := userClient.Transactions.List().Limit(2).Send()
	if err != nil {
		t.Fatalf("failed to retrieve transactions: %v", err)
	}
	if len(txs.Transactions) != 2 {
		t.Errorf("got %d transactions, wanted 3", len(txs.Transactions))
	}
}

func TestListTransactionsOffset(t *testing.T) {
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

	txs, err := userClient.Transactions.List().Offset(2).Send()
	if err != nil {
		t.Fatalf("failed to retrieve transactions: %v", err)
	}
	if len(txs.Transactions) != 1 {
		t.Fatalf("got %d transactions, wanted 1", len(txs.Transactions))
	}

	if txs.Transactions[0].Amount.Value != "60.00" {
		t.Errorf("got value %s, wanted %s", txs.Transactions[0].Amount.Value, "60.00")
	}

}

func TestListTransactionsSince(t *testing.T) {
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

	txs, err := userClient.Transactions.List().Since(time.Date(2017, 7, 28, 0, 0, 0, 0, time.UTC)).Send()
	if err != nil {
		t.Fatalf("failed to retrieve transactions: %v", err)
	}
	if len(txs.Transactions) != 2 {
		t.Fatalf("got %d transactions, wanted 2", len(txs.Transactions))
	}

	if txs.Transactions[0].Amount.Value != "-24.34" {
		t.Errorf("got value %s, wanted %s", txs.Transactions[0].Amount.Value, "-24.34")
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
			t.Logf("got unexpected error: %s", pr.Code)
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

func TestCreateRecurringTransfer(t *testing.T) {
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

	rule := bosgo.RecurrenceRule{
		Start:     time.Now(),
		Frequency: bosgo.FrequencyMonthly,
		Interval:  2,
	}
	// Create the transfer
	transfer, err := userClient.RecurringTransfers.Create(accountID, addr, amount, rule, "description/usage").Send()
	if err != nil {
		t.Fatalf("failed to create transfer: %v", err)
	}

	if len(transfer.Errors) > 0 {
		for _, pr := range transfer.Errors {
			t.Logf("got unexpected error: %s", pr.Code)
		}
		t.Fatalf("got %d errors, wanted none", len(transfer.Errors))
	}

	if transfer.Step.Intent != bosgo.TransferIntentProvidePIN {
		t.Errorf("got intent %v, wanted %v", transfer.Step.Intent, bosgo.TransferIntentProvidePIN)
	}

	// Send the pin
	req := userClient.RecurringTransfers.Process(transfer.ID, transfer.Step.Intent, transfer.Version)
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
	req = userClient.RecurringTransfers.Process(transfer.ID, transfer.Step.Intent, transfer.Version)
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
	req = userClient.RecurringTransfers.Process(transfer.ID, transfer.Step.Intent, transfer.Version)
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

func TestDeleteRecurringTransfer(t *testing.T) {
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

	repeatedTxs, err := userClient.RepeatedTransactions.List().Send()
	if err != nil {
		t.Fatalf("failed to get repeated txs: %v", err)
	}

	if len(repeatedTxs.Transactions) == 0 {
		t.Fatalf("missing repeated txs on default access")
	}

	// Delete the repeated tx with unknown ID
	recTrf, err := userClient.RepeatedTransactions.Delete("9999").Send()
	if err == nil {
		t.Fatalf("didn't fail to delete an unknown repeated transaction: %v", err)
	}

	// Delete the repeated tx with correct ID should
	rtxID := strconv.Itoa(int(repeatedTxs.Transactions[0].ID))
	recTrf, err = userClient.RepeatedTransactions.Delete(rtxID).Send()
	if err != nil {
		t.Fatalf("failed to delete recTrf: %v", err)
	}

	if len(recTrf.Errors) > 0 {
		for _, pr := range recTrf.Errors {
			t.Logf("got unexpected error: %s", pr.Code)
		}
		t.Fatalf("got %d errors, wanted none", len(recTrf.Errors))
	}

	if recTrf.Step.Intent != bosgo.TransferIntentProvidePIN {
		t.Errorf("got intent %v, wanted %v", recTrf.Step.Intent, bosgo.TransferIntentProvidePIN)
	}

	// Send the pin
	req := userClient.RecurringTransfers.Process(recTrf.ID, recTrf.Step.Intent, recTrf.Version)
	req.ChallengeAnswer(bosgo.ChallengeAnswer{
		ID:    "pin",
		Value: DefaultAccessPIN,
	})
	recTrf, err = req.Send()
	if err != nil {
		t.Fatalf("failed to process pin: %v", err)
	}
	if recTrf.Step.Intent != bosgo.TransferIntentSelectAuthMethod {
		t.Errorf("got intent %v, wanted %v", recTrf.Step.Intent, bosgo.TransferIntentSelectAuthMethod)
	}
	if recTrf.Step.Data == nil {
		t.Fatalf("got nil step data, wanted non-nil")
	}
	if len(recTrf.Step.Data.AuthMethods) != 1 {
		t.Fatalf("got %d auth methods, wanted 1", len(recTrf.Step.Data.AuthMethods))
	}
	if recTrf.Step.Data.AuthMethods[0].ID != DefaultAuthMethod {
		t.Errorf("got auth method %v, wanted %v", recTrf.Step.Data.AuthMethods[0].ID, DefaultAuthMethod)
	}

	// Send the auth method
	req = userClient.RecurringTransfers.Process(recTrf.ID, recTrf.Step.Intent, recTrf.Version)
	req.ChallengeAnswer(bosgo.ChallengeAnswer{
		ID:    "auth_method",
		Value: DefaultAuthMethod,
	})
	recTrf, err = req.Send()
	if err != nil {
		t.Fatalf("failed to process auth method: %v", err)
	}
	if recTrf.Step.Intent != bosgo.TransferIntentProvideChallengeAnswer {
		t.Errorf("got intent %v, wanted %v", recTrf.Step.Intent, bosgo.TransferIntentProvideChallengeAnswer)
	}
	if recTrf.Step.Data == nil {
		t.Fatalf("got nil step data, wanted non-nil")
	}
	if recTrf.Step.Data.ChallengeMessage != DefaultAuthMessage {
		t.Errorf("got challenge message %v, wanted %v", recTrf.Step.Data.ChallengeMessage, DefaultAuthMessage)
	}

	// Send the auth answer
	req = userClient.RecurringTransfers.Process(recTrf.ID, recTrf.Step.Intent, recTrf.Version)
	req.ChallengeAnswer(bosgo.ChallengeAnswer{
		ID:    "tan",
		Value: DefaultAuthAnswer,
	})
	recTrf, err = req.Send()
	if err != nil {
		t.Fatalf("failed to process auth answer: %v", err)
	}
	if recTrf.State != bosgo.TransferStateSucceeded {
		t.Errorf("got state %v, wanted %v", recTrf.State, bosgo.TransferStateSucceeded)
	}
}

func TestUpdateRecurringTransfer(t *testing.T) {
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

	repeatedTxs, err := userClient.RepeatedTransactions.List().Send()
	if err != nil {
		t.Fatalf("failed to get repeated txs: %v", err)
	}

	if len(repeatedTxs.Transactions) == 0 {
		t.Fatalf("missing repeated txs on default access")
	}

	// update the repeated tx with unknown ID should fail
	recTrf, err := userClient.RepeatedTransactions.Delete("9999").Send()
	if err == nil {
		t.Fatalf("didn't fail to delete an unknown repeated transaction: %v", err)
	}

	rtx := repeatedTxs.Transactions[0]
	rtx.Amount.Value = "99.32"

	addr := bosgo.TransferAddress{
		Name: "Jane Doe",
		IBAN: "DE28500105175552834822",
	}

	// Update the repeated tx with correct ID
	rtxID := strconv.Itoa(int(repeatedTxs.Transactions[0].ID))
	recTrf, err = userClient.RepeatedTransactions.Update(rtxID, addr, *rtx.Amount, "this is the updated usage").Send()
	if err != nil {
		t.Fatalf("failed to delete recTrf: %v", err)
	}

	if len(recTrf.Errors) > 0 {
		for _, pr := range recTrf.Errors {
			t.Logf("got unexpected error: %s", pr.Code)
		}
		t.Fatalf("got %d errors, wanted none", len(recTrf.Errors))
	}

	if recTrf.Step.Intent != bosgo.TransferIntentProvidePIN {
		t.Errorf("got intent %v, wanted %v", recTrf.Step.Intent, bosgo.TransferIntentProvidePIN)
	}

	// Send the pin
	req := userClient.RecurringTransfers.Process(recTrf.ID, recTrf.Step.Intent, recTrf.Version)
	req.ChallengeAnswer(bosgo.ChallengeAnswer{
		ID:    "pin",
		Value: DefaultAccessPIN,
	})
	recTrf, err = req.Send()
	if err != nil {
		t.Fatalf("failed to process pin: %v", err)
	}
	if recTrf.Step.Intent != bosgo.TransferIntentSelectAuthMethod {
		t.Errorf("got intent %v, wanted %v", recTrf.Step.Intent, bosgo.TransferIntentSelectAuthMethod)
	}
	if recTrf.Step.Data == nil {
		t.Fatalf("got nil step data, wanted non-nil")
	}
	if len(recTrf.Step.Data.AuthMethods) != 1 {
		t.Fatalf("got %d auth methods, wanted 1", len(recTrf.Step.Data.AuthMethods))
	}
	if recTrf.Step.Data.AuthMethods[0].ID != DefaultAuthMethod {
		t.Errorf("got auth method %v, wanted %v", recTrf.Step.Data.AuthMethods[0].ID, DefaultAuthMethod)
	}

	// Send the auth method
	req = userClient.RecurringTransfers.Process(recTrf.ID, recTrf.Step.Intent, recTrf.Version)
	req.ChallengeAnswer(bosgo.ChallengeAnswer{
		ID:    "auth_method",
		Value: DefaultAuthMethod,
	})
	recTrf, err = req.Send()
	if err != nil {
		t.Fatalf("failed to process auth method: %v", err)
	}
	if recTrf.Step.Intent != bosgo.TransferIntentProvideChallengeAnswer {
		t.Errorf("got intent %v, wanted %v", recTrf.Step.Intent, bosgo.TransferIntentProvideChallengeAnswer)
	}
	if recTrf.Step.Data == nil {
		t.Fatalf("got nil step data, wanted non-nil")
	}
	if recTrf.Step.Data.ChallengeMessage != DefaultAuthMessage {
		t.Errorf("got challenge message %v, wanted %v", recTrf.Step.Data.ChallengeMessage, DefaultAuthMessage)
	}

	// Send the auth answer
	req = userClient.RecurringTransfers.Process(recTrf.ID, recTrf.Step.Intent, recTrf.Version)
	req.ChallengeAnswer(bosgo.ChallengeAnswer{
		ID:    "tan",
		Value: DefaultAuthAnswer,
	})
	recTrf, err = req.Send()
	if err != nil {
		t.Fatalf("failed to process auth answer: %v", err)
	}
	if recTrf.State != bosgo.TransferStateSucceeded {
		t.Errorf("got state %v, wanted %v", recTrf.State, bosgo.TransferStateSucceeded)
	}
}

func TestWriteReadState(t *testing.T) {
	s := NewWithDefaults()
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

	var buf bytes.Buffer
	if err := s.WriteState(&buf); err != nil {
		t.Fatalf("unexpected error writing state: %v", err)
	}

	s2 := New()
	defer s2.Close()
	s2.ReadState(&buf)

	appClient2 := bosgo.NewAppClient(s2.Client(), s2.Addr(), DefaultApplicationID)
	userClient2, err := appClient2.Users.Login(DefaultUsername, DefaultPassword).Send()
	if err != nil {
		t.Fatalf("failed to login to new server as user: %v", err)
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
	transfer, err := userClient2.Transfers.Create(accountID, addr, amount).Send()
	if err != nil {
		t.Fatalf("failed to create transfer: %v", err)
	}

	if len(transfer.Errors) > 0 {
		for _, pr := range transfer.Errors {
			t.Logf("got unexpected error: %s", pr.Code)
		}
		t.Fatalf("got %d errors, wanted none", len(transfer.Errors))
	}

}

func TestAccessRefreshMultiStep(t *testing.T) {
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

	job, err := userClient.Accesses.Refresh(accessID).Send()
	if err != nil {
		t.Fatalf("failed to refresh access: %v", err)
	}
	t.Logf("job URI: %s", job.URI)

	status, err := userClient.Jobs.Get(job.URI).Send()
	if err != nil {
		t.Fatalf("failed to get job status: %v", err)
	}
	if status.Stage != bosgo.JobStageChallenge {
		t.Errorf("got stage %v, wanted %v", status.Stage, bosgo.JobStageChallenge)
	}

	if status.Challenge == nil {
		t.Fatalf("expecting job to have unanswered challenges")
	}

	// Nothing is stored so we need to answer 2 challenges
	if len(status.Challenge.NextChallenges) != 2 {
		for _, ch := range status.Challenge.NextChallenges {
			t.Logf("got challenge id %s", ch.ID)
		}
		t.Errorf("got %d challenges, wanted %d", len(status.Challenge.NextChallenges), 2)
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
	if status.Stage != bosgo.JobStageChallenge {
		t.Errorf("got stage %v, wanted %v", status.Stage, bosgo.JobStageChallenge)
	}

	if status.Challenge == nil {
		t.Fatalf("expecting job to have unanswered challenges")
	}

	if len(status.Challenge.NextChallenges) != 2 {
		for _, ch := range status.Challenge.NextChallenges {
			t.Logf("got challenge id %s", ch.ID)
		}
		t.Fatalf("got %d challenges, wanted %d", len(status.Challenge.NextChallenges), 2)
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
	if status.Stage != bosgo.JobStageImported {
		t.Errorf("got stage %v, wanted %v", status.Stage, bosgo.JobStageImported)
	}
	if status.Finished != true {
		t.Errorf("got finished %v, wanted true", status.Finished)
	}

	ac, err := userClient.Accesses.List().Send()
	if err != nil {
		t.Fatalf("failed to retrieve accesses: %v", err)
	}
	if len(ac.Accesses) != 1 {
		t.Errorf("got %d accesses, wanted 1", len(ac.Accesses))
	}

}

func TestAccessRefreshWrongPin(t *testing.T) {
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

	job, err := userClient.Accesses.Refresh(accessID).Send()
	if err != nil {
		t.Fatalf("failed to refresh access: %v", err)
	}
	t.Logf("job URI: %s", job.URI)

	status, err := userClient.Jobs.Get(job.URI).Send()
	if err != nil {
		t.Fatalf("failed to get job status: %v", err)
	}
	if status.Stage != bosgo.JobStageChallenge {
		t.Errorf("got stage %v, wanted %v", status.Stage, bosgo.JobStageChallenge)
	}

	if status.Challenge == nil {
		t.Fatalf("expecting job to have unanswered challenges")
	}

	if len(status.Challenge.NextChallenges) != 2 {
		for _, ch := range status.Challenge.NextChallenges {
			t.Logf("got challenge id %s", ch.ID)
		}
		t.Fatalf("got %d challenges, wanted %d", len(status.Challenge.NextChallenges), 2)
	}

	req := userClient.Jobs.Answer(job.URI)
	req.ChallengeAnswer(bosgo.ChallengeAnswer{
		ID:    ChallengeLogin,
		Value: DefaultAccessLogin,
	})
	req.ChallengeAnswer(bosgo.ChallengeAnswer{
		ID:    ChallengePIN,
		Value: "wrongpin",
	})
	if err := req.Send(); err != nil {
		t.Fatalf("failed to answer first challenge: %v", err)
	}
	status, err = userClient.Jobs.Get(job.URI).Send()
	if err != nil {
		t.Fatalf("failed to get job status: %v", err)
	}
	if status.Stage != bosgo.JobStageChallenge {
		t.Errorf("got stage %v, wanted %v", status.Stage, bosgo.JobStageChallenge)
	}

	if status.Challenge == nil {
		t.Fatalf("expecting job to have unanswered challenges")
	}

	if len(status.Errors) == 0 {
		t.Fatalf("got 0 errors, wanted at least one")
	}

	gotWrongPinProblem := false
	for _, p := range status.Errors {
		if p.Code == "user_wrong_pin" {
			gotWrongPinProblem = true
			break
		}
	}

	if !gotWrongPinProblem {
		t.Errorf("did not get problem 'user_wrong_pin'")
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
	if status.Stage != bosgo.JobStageImported {
		t.Errorf("got stage %v, wanted %v", status.Stage, bosgo.JobStageImported)
	}
	if !status.Finished {
		t.Errorf("got finished %v, wanted true", status.Finished)
	}

}

func TestAccessRefreshCustomProblems(t *testing.T) {
	s := NewWithDefaults()

	access := s.MakeAccess(DefaultProviderID+"_problems", "access with problems")
	ad := AccessDetails{
		Access:               *access,
		Transactions:         []bosgo.Transaction{},
		RepeatedTransactions: []bosgo.RepeatedTransaction{},
		ChallengeMap: map[string]string{
			ChallengeLogin: DefaultAccessLogin + "_problems",
			ChallengePIN:   DefaultAccessPIN,
		},
		TransferAuths: []TransferAuth{
			{
				Method:  DefaultAuthMethod,
				Message: DefaultAuthMessage,
				Answer:  DefaultAuthAnswer,
			},
		},
		StageProblems: map[bosgo.JobStage][]bosgo.Problem{
			bosgo.JobStageChallenge: {
				{
					Code: "fi_accountnumber_10_digits_pin_between_5_to_10",
				},
			},
		},
	}
	s.AddAccess(ad)

	if testing.Verbose() {
		s.SetLogger(t)
	}
	defer s.Close()

	appClient := bosgo.NewAppClient(s.Client(), s.Addr(), DefaultApplicationID)
	userClient, err := appClient.Users.Login(DefaultUsername, DefaultPassword).Send()

	if err != nil {
		t.Fatalf("failed to login as user: %v", err)
	}

	accessID, _, err := addAccess(userClient, DefaultProviderID+"_problems", DefaultAccessLogin+"_problems", DefaultAccessPIN)
	if err != nil {
		t.Fatalf("failed to add access: %v", err)
	}

	job, err := userClient.Accesses.Refresh(accessID).Send()
	if err != nil {
		t.Fatalf("failed to refresh access: %v", err)
	}
	t.Logf("job URI: %s", job.URI)

	status, err := userClient.Jobs.Get(job.URI).Send()
	if err != nil {
		t.Fatalf("failed to get job status: %v", err)
	}
	if status.Stage != bosgo.JobStageChallenge {
		t.Errorf("got stage %v, wanted %v", status.Stage, bosgo.JobStageChallenge)
	}

	if len(status.Errors) == 0 {
		t.Fatal("missing status errors")
	}

	if status.Errors[0].Code != "fi_accountnumber_10_digits_pin_between_5_to_10" {
		t.Fatalf("got bad error code %s, wanted %s", status.Errors[0].Code, "fi_accountnumber_10_digits_pin_between_5_to_10")
	}

	if status.Stage != bosgo.JobStageChallenge {
		t.Errorf("got stage %v, wanted %v", status.Stage, bosgo.JobStageChallenge)
	}
	if status.Finished {
		t.Errorf("got finished %v, wanted false", status.Finished)
	}
}

func addAccess(userClient *bosgo.UserClient, provider, login, pin string) (int64, int64, error) {
	req := userClient.Accesses.Add(provider)
	req.ChallengeAnswer(bosgo.ChallengeAnswer{
		ID:    ChallengeLogin,
		Value: login,
	})
	req.ChallengeAnswer(bosgo.ChallengeAnswer{
		ID:    ChallengePIN,
		Value: pin,
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
