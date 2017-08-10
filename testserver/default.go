package testserver

import (
	"time"

	"github.com/bankrs/bosgo"
)

const (
	DefaultDeveloperID   = "default-dev"
	DefaultApplicationID = "default-app"
	DefaultUserID        = "default-user"
	DefaultUsername      = "username@example.com"
	DefaultPassword      = "password"
	DefaultProviderID    = "def-provider-id"
	DefaultAccessLogin   = "user"
	DefaultAccessPIN     = "1234"
)

// NewWithDefaults creates a new test server with a default developer, application and user account
func NewWithDefaults() *Server {
	s := New()

	app := App{
		ID:          DefaultApplicationID,
		DeveloperID: DefaultDeveloperID,
	}
	s.setApp(app)

	user := User{
		ID:            DefaultUserID,
		Username:      DefaultUsername,
		Password:      DefaultPassword,
		ApplicationID: DefaultApplicationID,
	}
	s.setUser(user)

	access := s.MakeAccess(DefaultProviderID, "default access")
	txs := []bosgo.Transaction{
		{
			ID:            s.nextID(),
			AccessID:      access.ID,
			UserAccountID: access.Accounts[0].ID,
			UserAccount: bosgo.AccountRef{
				ProviderID: DefaultProviderID,
				IBAN:       access.Accounts[0].IBAN,
			},
			Amount: &bosgo.MoneyAmount{
				Currency: "EUR",
				Value:    "-24.34",
			},
			EntryDate:      time.Date(2017, 7, 31, 0, 0, 0, 0, time.UTC),
			SettlementDate: time.Date(2017, 7, 31, 0, 0, 0, 0, time.UTC),
			Usage:          "Goods bought",
			Counterparty: bosgo.Counterparty{
				Name: "PayPal Europe Sarl",
				Account: bosgo.AccountRef{
					ProviderID: "DE-BIN-75290000",
					IBAN:       "DE84200700245353762745",
				},
				Merchant: &bosgo.Merchant{
					Name: "PayPal",
				},
			},
		},
		{
			ID:            s.nextID(),
			AccessID:      access.ID,
			UserAccountID: access.Accounts[0].ID,
			UserAccount: bosgo.AccountRef{
				ProviderID: DefaultProviderID,
				IBAN:       access.Accounts[0].IBAN,
			},
			Amount: &bosgo.MoneyAmount{
				Currency: "EUR",
				Value:    "0.05",
			},
			EntryDate:      time.Date(2017, 7, 30, 0, 0, 0, 0, time.UTC),
			SettlementDate: time.Date(2017, 7, 30, 0, 0, 0, 0, time.UTC),
			Usage:          "Interest payment",
			Counterparty:   bosgo.Counterparty{},
		},
		{
			ID:            s.nextID(),
			AccessID:      access.ID,
			UserAccountID: access.Accounts[1].ID,
			UserAccount: bosgo.AccountRef{
				ProviderID: DefaultProviderID,
				IBAN:       access.Accounts[1].IBAN,
			},
			Amount: &bosgo.MoneyAmount{
				Currency: "EUR",
				Value:    "60.00",
			},
			EntryDate:      time.Date(2017, 7, 30, 0, 0, 0, 0, time.UTC),
			SettlementDate: time.Date(2017, 7, 30, 0, 0, 0, 0, time.UTC),
			Usage:          "Money transfer",
			Counterparty:   bosgo.Counterparty{},
		},
	}

	rtxs := []bosgo.RepeatedTransaction{
		{
			ID:            s.nextID(),
			AccessID:      access.ID,
			UserAccountID: access.Accounts[0].ID,
			UserAccount: bosgo.AccountRef{
				ProviderID: DefaultProviderID,
				IBAN:       "DE84200700245353762745",
			},
			RemoteAccount: bosgo.AccountRef{
				IBAN: "DE04200800957050250010",
			},
			Schedule: bosgo.RecurrenceRule{
				Start:     time.Date(2017, 1, 1, 0, 0, 0, 0, time.UTC),
				Until:     time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC),
				Frequency: bosgo.FrequencyMonthly,
				Interval:  1,
				ByDay:     24,
			},
			Amount: &bosgo.MoneyAmount{
				Currency: "EUR",
				Value:    "500.00",
			},
			Usage: "Rent",
		},
	}

	s.AddAccess(
		access,
		txs,
		rtxs,
		map[string]string{
			"login": DefaultAccessLogin,
			"pin":   DefaultAccessPIN,
		},
	)

	return s
}

// MakeAccess makes an access with an account
func (s *Server) MakeAccess(providerID, name string) *bosgo.Access {
	accID := s.nextID()
	acc := bosgo.Access{
		ID:         accID,
		ProviderID: providerID,
		Enabled:    true,
		Name:       name,

		Accounts: []bosgo.Account{
			{
				ID:           s.nextID(),
				ProviderID:   providerID,
				BankAccessID: accID,
				Name:         "Account 1",
				Type:         bosgo.AccountTypeBank,
				Number:       "704357300",
				Balance:      "971.20 EUR",
				BalanceDate:  time.Date(2017, 7, 13, 22, 0, 0, 0, time.UTC),
				Enabled:      true,
				Currency:     "EUR",
				IBAN:         "DE84200700245353762745",
				Supported:    true,
				Capabilities: bosgo.AccountCapabilities{
					AccountStatement:  []string{"read"},
					Transfer:          []string{"read"},
					RecurringTransfer: []string{"read"},
				},
			},
			{
				ID:           s.nextID(),
				ProviderID:   providerID,
				BankAccessID: accID,
				Name:         "Account 2",
				Type:         bosgo.AccountTypeBank,
				Number:       "704357301",
				Balance:      "45.00 EUR",
				BalanceDate:  time.Date(2017, 7, 13, 22, 0, 0, 0, time.UTC),
				Enabled:      true,
				Currency:     "EUR",
				IBAN:         "DE56200800950445688921",
				Supported:    true,
				Capabilities: bosgo.AccountCapabilities{
					AccountStatement:  []string{"read"},
					Transfer:          []string{"read"},
					RecurringTransfer: []string{"read"},
				},
			},
		},
	}
	return &acc
}
