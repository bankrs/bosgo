package testserver

import (
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

	s.AddAccess(
		s.makeDefaultAccess(),
		map[string]string{
			"login": DefaultAccessLogin,
			"pin":   DefaultAccessPIN,
		},
	)

	return s
}

func (s *Server) makeDefaultAccess() *bosgo.Access {
	accID := s.nextID()
	acc := bosgo.Access{
		ID:         accID,
		ProviderID: DefaultProviderID,
		Enabled:    true,
		Name:       "default access",

		Accounts: []bosgo.Account{
			{
				ID:           s.nextID(),
				ProviderID:   DefaultProviderID,
				BankAccessID: accID,
				Name:         "Account 1",
				Type:         bosgo.AccountTypeBank,
				Number:       "704357300",
				Balance:      "971.20 EUR",
				BalanceDate:  "2017-07-13T22:00:00Z",
				Enabled:      true,
				Currency:     "EUR",
				IBAN:         "DE75524206009411376450",
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
