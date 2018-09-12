package bosgo

import (
	"time"
)

type DeveloperCredentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
type DeveloperProfile struct {
	Company             string `json:"company"`
	HasProductionAccess bool   `json:"has_production_access"`
}

type ApplicationPage struct {
	Applications []ApplicationMetadata `json:"applications,omitempty"`
}

type ApplicationMetadata struct {
	ApplicationID string    `json:"id,omitempty"`
	Label         string    `json:"label,omitempty"`
	CreatedAt     time.Time `json:"created_at,omitempty"` // Deprecated: no longer used
}

type ApplicationKeyPage struct {
	Keys []ApplicationKey `json:"keys,omitempty"`
}

type ApplicationKey struct {
	Key       string    `json:"key,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
}

type StatsPeriod struct {
	From   string `json:"from_date"`
	To     string `json:"to_date"`
	Domain string `json:"domain"`
}

type UsersStats struct {
	StatsPeriod
	UsersTotal StatsValue        `json:"users_total"` // with weekly relative change
	UsersToday StatsValue        `json:"users_today"` // with daily relative change
	Stats      []DailyUsersStats `json:"stats"`
}

type StatsValue struct {
	Value int64 `json:"value"`
}

type DailyUsersStats struct {
	Date        string `json:"date"`
	UsersTotal  int64  `json:"users_total"`
	NewUsers    int64  `json:"new_users"`
	ActiveUsers int64  `json:"active_users"`
}

type TransfersStats struct {
	StatsPeriod
	TotalOut StatsMoneyAmount      `json:"total_out"`
	TodayOut StatsMoneyAmount      `json:"today_out"`
	Stats    []DailyTransfersStats `json:"stats"`
}

type DailyTransfersStats struct {
	Date string           `json:"date"`
	Out  StatsMoneyAmount `json:"out"`
}

type MerchantsStats struct {
	StatsPeriod
	Stats []DailyMerchantsStats `json:"stats"`
}

type DailyMerchantsStats struct {
	Date      string      `json:"date"`
	Merchants []NameValue `json:"merchants"`
}

type ProvidersStats struct {
	StatsPeriod
	Stats []DailyProvidersStats `json:"stats"`
}

type DailyProvidersStats struct {
	Date      string      `json:"date"`
	Providers []NameValue `json:"providers"`
}

type StatsMoneyAmount struct {
	Value    float64 `json:"value"`
	Currency string  `json:"currency"`
}

type NameValue struct {
	Name  string `json:"name"`
	Value int64  `json:"value"`
}

type RequestsStats struct {
	StatsPeriod
	RequestsTotal StatsValue           `json:"requests_total"`
	RequestsToday StatsValue           `json:"requests_today"`
	Stats         []DailyRequestsStats `json:"stats"`
}

type DailyRequestsStats struct {
	Date          string `json:"date"`
	RequestsTotal int64  `json:"requests_total"`
}

type CategoryList []Category

type Category struct {
	ID    int64             `json:"id"`
	Names map[string]string `json:"names"`
	Group string            `json:"group"`
}

type ProviderSearchResults []ProviderSearchResult

type ProviderSearchResult struct {
	Score    float64  `json:"score"`
	Provider Provider `json:"provider"`
}

type Provider struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Country     string          `json:"country"`
	URL         string          `json:"url"`
	Address     string          `json:"address"`
	PostalCode  string          `json:"postal_code"`
	Challenges  []ChallengeSpec `json:"challenges"`
}

type ChallengeSpec struct {
	ID          string            `json:"id"`
	Description string            `json:"description"`
	Type        ChallengeType     `json:"type"`
	Secure      bool              `json:"secure"`
	UnStoreable bool              `json:"unstoreable"`
	Info        map[string]string `json:"info,omitempty"`
}

type ChallengeType string

const (
	ChallengeTypeAlpha        ChallengeType = "alpha"
	ChallengeTypeNumeric      ChallengeType = "numeric"
	ChallengeTypeAlphaNumeric ChallengeType = "alphanumeric"
)

type ChallengeAnswerList []ChallengeAnswer

type ChallengeAnswer struct {
	ID         string    `json:"id"`
	Value      string    `json:"value"`
	Store      bool      `json:"store"`
	ValidUntil time.Time `json:"valid_until"`
}

type UserListPage struct {
	Users      []string `json:"users,omitempty"`
	NextCursor string   `json:"next,omitempty"`
}

type UserCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type UserToken struct {
	ID    string `json:"id"`    // globally unique identifier for a user
	Token string `json:"token"` // session token
}

type AccessPage struct {
	Accesses []Access `json:"accesses"`
}

type Access struct {
	ID           int64              `json:"id"`
	Name         string             `json:"name"`
	Enabled      bool               `json:"enabled"`
	IsPinSaved   bool               `json:"is_pin_saved"`
	AuthPossible bool               `json:"auth_possible"`
	ProviderID   string             `json:"provider_id"`
	Accounts     []Account          `json:"accounts,omitempty"`
	Capabilities AccessCapabilities `json:"capabilities"`
}

type Account struct {
	ID                int64               `json:"id"`
	ProviderID        string              `json:"provider_id"`
	BankAccessID      int64               `json:"bank_access_id"`
	Name              string              `json:"name"`
	Type              AccountType         `json:"type"`
	Number            string              `json:"number"`
	Balance           string              `json:"balance"`
	BalanceDate       time.Time           `json:"balance_date"`
	AvailableBalance  string              `json:"available_balance"`
	CreditLine        string              `json:"credit_line"`
	Removed           bool                `json:"removed"`
	Currency          string              `json:"currency"`
	IBAN              string              `json:"iban"`
	Alias             string              `json:"alias"`
	Capabilities      AccountCapabilities `json:"capabilities"`
	AllowedOperations AllowedOperations   `json:"allowed_operations"`
	Bin               string              `json:"bin"`
}

type AccountCapabilities struct {
	AccountStatement  []string `json:"account_statement"`
	Transfer          []string `json:"transfer"`
	RecurringTransfer []string `json:"recurring_transfer"`
}

type AllowedOperations struct {
	PaymentTransfer     bool `json:"transfer"`
	AccountStatement    bool `json:"statement"`
	AccountBalance      bool `json:"balance"`
	CreditCardStatement bool `json:"-"`
	CreditCardBalance   bool `json:"-"`
	CreateRecTrf        bool `json:"create_recurring_transfer"`
	ReadRecTrf          bool `json:"read_recurring_transfer"`
	UpdateRecTrf        bool `json:"update_recurring_transfer"`
	DeleteRecTrf        bool `json:"delete_recurring_transfer"`
}

type AccountType string

const (
	AccountTypeCurrent    AccountType = "current"
	AccountTypeSavings    AccountType = "savings"
	AccountTypeCreditCard AccountType = "creditcard"
	AccountTypeLoan       AccountType = "loan"
)

type Job struct {
	URI string `json:"uri"`
}

type JobStatus struct {
	Finished  bool       `json:"finished"`
	Stage     JobStage   `json:"stage"`
	Challenge *Challenge `json:"challenge,omitempty"`
	URI       string     `json:"uri,omitempty"`
	Errors    []Problem  `json:"errors,omitempty"`
	Access    *JobAccess `json:"access,omitempty"`
}

type JobStage string

const (
	JobStageUnauthenticated JobStage = "unauthenticated"
	JobStageAuthenticated   JobStage = "authenticated"
	JobStageChallenge       JobStage = "challenge"
	JobStageImported        JobStage = "imported"
	JobStageCancelled       JobStage = "cancelled"
	JobStageProblem         JobStage = "problem"
)

type Challenge struct {
	NextChallenges []ChallengeField `json:"next_challenges"`
	LastProblems   []Problem        `json:"last_problems"`
}

type ChallengeField struct {
	ID            string            `json:"id"`
	Description   string            `json:"description"`
	ChallengeType string            `json:"type"`
	Previous      string            `json:"previous"`
	Stored        bool              `json:"stored"`
	Reset         bool              `json:"reset"`
	Secure        bool              `json:"secure"`
	Optional      bool              `json:"optional"`
	UnStoreable   bool              `json:"unstoreable"`
	Methods       []string          `json:"methods"`
	Info          map[string]string `json:"info"`
}

type Problem struct {
	Domain string                 `json:"domain"`
	Code   string                 `json:"code"`
	Info   map[string]interface{} `json:"info"`
}

type JobAccess struct {
	ID         int64        `json:"id,omitempty"`
	ProviderID string       `json:"provider_id,omitempty"`
	Name       string       `json:"name,omitempty"`
	Accounts   []JobAccount `json:"accounts,omitempty"`
}

type JobAccount struct {
	ID     int64     `json:"id,omitempty"`
	Name   string    `json:"name"`
	Number string    `json:"number"`
	IBAN   string    `json:"iban"`
	Errors []Problem `json:"errors"`
}

type TransactionPage struct {
	Transactions []Transaction `json:"data"`
	Total        int           `json:"total"`
	Limit        int           `json:"limit"`
	Offset       int           `json:"offset"`
}

type Transaction struct {
	ID                    int64           `json:"id"`
	AccessID              int64           `json:"user_bank_access_id,omitempty"`
	UserAccountID         int64           `json:"user_bank_account_id,omitempty"`
	UserAccount           AccountRef      `json:"user_account,omitempty"`
	CategoryID            int64           `json:"category_id,omitempty"`
	RepeatedTransactionID int64           `json:"repeated_transaction_id,omitempty"`
	Counterparty          Counterparty    `json:"counterparty,omitempty"`
	RemoteID              string          `json:"remote_id"`
	EntryDate             time.Time       `json:"entry_date,omitempty"`
	SettlementDate        time.Time       `json:"settlement_date,omitempty"`
	Amount                *MoneyAmount    `json:"amount,omitempty"`
	OriginalAmount        *OriginalAmount `json:"original_amount,omitempty"`
	Usage                 string          `json:"usage,omitempty"`
	TransactionType       string          `json:"transaction_type,omitempty"`
	Gvcode                string          `json:"gvcode,omitempty"`
}

type AccountRef struct {
	ProviderID string `json:"provider_id"`
	IBAN       string `json:"iban,omitempty"`
	Label      string `json:"label,omitempty"`
	Number     string `json:"id,omitempty"`
	Type       string `json:"type,omitempty"`
}

type OriginalAmount struct {
	Value        *MoneyAmount `json:"value"`
	ExchangeRate string       `json:"exchange_rate"`
}

type Merchant struct {
	Name string `json:"name"`
}

type Counterparty struct {
	Name     string     `json:"name"`
	Account  AccountRef `json:"account,omitempty"`
	Merchant *Merchant  `json:"merchant,omitempty"`
}

type RepeatedTransactionPage struct {
	Transactions []RepeatedTransaction `json:"data"`
	Total        int                   `json:"total"`
	Limit        int                   `json:"limit"`
	Offset       int                   `json:"offset"`
}

type RepeatedTransaction struct {
	ID            int64          `json:"id"`
	AccessID      int64          `json:"user_bank_access_id,omitempty"`
	UserAccountID int64          `json:"user_bank_account_id,omitempty"`
	UserAccount   AccountRef     `json:"user_account"`
	RemoteAccount AccountRef     `json:"remote_account"`
	RemoteID      string         `json:"remote_id"`
	Schedule      RecurrenceRule `json:"schedule"`
	Amount        *MoneyAmount   `json:"amount"`
	Usage         string         `json:"usage"`
}

type RecurrenceRule struct {
	Start     time.Time `json:"start"`
	Until     time.Time `json:"until"`
	Frequency Frequency `json:"frequency"`
	Interval  int       `json:"interval"`
	ByDay     int       `json:"by_day"`
}

type Frequency string

const (
	FrequencyOnce    Frequency = "once"
	FrequencyDaily   Frequency = "daily"
	FrequencyWeekly  Frequency = "weekly"
	FrequencyMonthly Frequency = "monthly"
	FrequencyYearly  Frequency = "yearly"
)

type MoneyAmount struct {
	Currency string `json:"currency"`
	Value    string `json:"value"`
}

type TransferAddress struct {
	Name      string `json:"name"`
	IBAN      string `json:"iban"`
	AccessID  int64  `json:"bank_access_id,omitempty"`
	AccountID int64  `json:"bank_account_id,omitempty"`
}

type TransferType string

const (
	TransferTypeRecurring TransferType = "recurring"
	TransferTypeRegular   TransferType = "regular"
)

type ChallengeAnswerMap map[string]ChallengeAnswer

type Transfer struct {
	ID               string             `json:"id"`
	From             TransferAddress    `json:"from"`
	To               TransferAddress    `json:"to"`
	Amount           *MoneyAmount       `json:"amount"`
	Usage            string             `json:"usage"`
	Version          int                `json:"version"`
	Step             TransferStep       `json:"step"`
	State            TransferState      `json:"state"`
	Schedule         *RecurrenceRule    `json:"schedule,omitempty"`
	EntryDate        time.Time          `json:"booking_date,omitempty"`
	SettlementDate   time.Time          `json:"effective_date,omitempty"`
	Created          time.Time          `json:"created,omitempty"`
	Updated          time.Time          `json:"updated,omitempty"`
	RemoteID         string             `json:"remote_id"`
	ChallengeAnswers ChallengeAnswerMap `json:"challenge_answers,omitempty"`
	Errors           []Problem          `json:"errors"`
}

type RecurringTransfer struct {
	ID               string             `json:"id"`
	From             TransferAddress    `json:"from"`
	To               TransferAddress    `json:"to"`
	Amount           MoneyAmount        `json:"amount"`
	Usage            string             `json:"usage"`
	Version          int                `json:"version"`
	Step             TransferStep       `json:"step"`
	State            TransferState      `json:"state"`
	Schedule         *RecurrenceRule    `json:"schedule,omitempty"`
	RemoteID         string             `json:"remote_id"`
	ChallengeAnswers ChallengeAnswerMap `json:"challenge_answers,omitempty"`
	Errors           []Problem          `json:"errors,omitempty"`
}

type TransferState string

const (
	TransferStateOngoing   TransferState = "ongoing"
	TransferStateSucceeded TransferState = "succeeded"
	TransferStateFailed    TransferState = "failed"
	TransferStateCancelled TransferState = "cancelled"
)

type TransferIntent string

const (
	TransferIntentProvidePIN             TransferIntent = "provide_pin"
	TransferIntentProvideCredentials     TransferIntent = "provide_credentials"
	TransferIntentSelectAuthMethod       TransferIntent = "select_auth_method"
	TransferIntentProvideChallengeAnswer TransferIntent = "provide_challenge_answer"
	TransferIntentConfirmSimilarTransfer TransferIntent = "confirm_similar_transfer"
)

type PaymentTransferCancelParams struct {
	ID      string `json:"id"`
	Version int    `json:"version"`
}

type TransferStep struct {
	Intent TransferIntent    `json:"intent,omitempty"`
	Data   *TransferStepData `json:"data,omitempty"`
}

type AuthMethod struct {
	ID          string `json:"id"`
	Description string `json:"description"`
}

type TransferStepData struct {
	AuthMethods      []AuthMethod `json:"auth_methods,omitempty"`      // TAN Options
	Challenge        string       `json:"challenge,omitempty"`         // TAN Challenge
	ChallengeMessage string       `json:"challenge_message,omitempty"` // TAN Challenge Message
	TANType          TANType      `json:"tan_type,omitempty"`          // Type of the TAN (optical, itan, unknown)
	Confirm          bool         `json:"confirm,omitempty"`           // Confirm (similar transfer)
	Transfers        []Transfer   `json:"transfers,omitempty"`         // Transfer list (similar transfers)
}

type TANType string

const (
	// TANTypePaymentPin4 is an abstract 4-chars string used to authorise payment
	TANTypePaymentPin4 TANType = "paymentPIN4"
	// TANTypeOptical indicates an optical TAN such as flickering barcodes
	TANTypeOptical TANType = "optical"
	// TANTypeITAN indicates an iTAN (aka indexed TAN) such as a list of TAN numbers with a sequence
	TANTypeITAN TANType = "itan"
	// TANTypeMobile indicates a mobileTAN such as an SMS with a passcode
	TANTypeMobile TANType = "mobile"
	// TANTypeChip indicates a chipTAN provided from a calculator device
	TANTypeChip TANType = "chip"
	// TANTypePush indicates a push push notification to a mobile app
	TANTypePush TANType = "push"
	// TANTypeOTP indicates a one-time password
	TANTypeOTP TANType = "otp"
	// TypeUSB indicate a usb based TAN
	TANTypeUSB TANType = "usb"
	// TANTypePhoto indicates a colorised matrix barcode
	TANTypePhoto TANType = "photo"

	TANTypeUnknown TANType = "unknown"
)

type DeletedUser struct {
	DeletedUserID string `json:"deleted_user_id"`
}

type IBANDetails struct {
	Account IBANAccount `json:"acc_ref"`
	Banks   []IBANBank  `json:"fis"`
}

type IBANBank struct {
	ID             string `json:"id"`              // the bank identity assigned by the identity provider
	Label          string `json:"label"`           // the bank name
	Country        string `json:"country"`         // the country (e.g. DE)
	Provider       string `json:"provider"`        //  the identity provider (e.g. BIC)
	ServiceContext string `json:"service_context"` // the service context, (e.g. SEPA)
}

type IBANAccount struct {
	IBAN     string `json:"IBAN"`     // the validated IBAN
	Provider string `json:"provider"` // the authoritative provider, IBO for IBANs
}

type ResetUsersResponse struct {
	Users []ResetUserOutcome `json:"users"`
}

type ResetUserOutcome struct {
	Username string    `json:"username"`
	Problems []Problem `json:"problems"`
}

type DevUserInfo struct {
	Username string `json:"username"`
}

type Webhook struct {
	ID          string    `json:"id"`
	URL         string    `json:"url"`
	Events      []string  `json:"events"`
	APIVersion  int       `json:"api_version"`
	Enabled     bool      `json:"enabled"`
	Environment string    `json:"environment"`
	CreatedAt   time.Time `json:"created_at"`
}

type WebhookPage struct {
	Webhooks []Webhook `json:"webhooks,omitempty"`
}

type WebhookTestResult struct {
	Payload  EventPayload  `json:"payload"`
	Response EventResponse `json:"response"`
}

type EventPayload struct {
	Event Event                  `json:"event"`
	Data  map[string]interface{} `json:"data"`
}

type Event struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	URL         string    `json:"url"`
	APIVersion  int       `json:"api_version"`
	Environment string    `json:"environment"`
	CreatedAt   time.Time `json:"created_at"`
}

type EventResponse struct {
	ID     string `json:"id"`
	Code   int    `json:"code"`
	Status string `json:"status"`
}

type ApplicationSettings struct {
	BackgroundRefresh bool `json:"background_refresh"`
}

type AccessCapabilities struct {
	RecurringTransfer RecurringTransferCapabilities `json:"recurring_transfer"`
	ScheduledTransfer ScheduledTransferCapabilities `json:"scheduled_transfer"`
	Trading           bool                          `json:"trading"`
}

type RecurringTransferCapabilities struct {
	Periods                      []Period `json:"periods"`
	MinimumLeadTimeCreate        int      `json:"minimum_lead_time_create"`
	MaximumLeadTimeCreate        int      `json:"maximum_lead_time_create"`
	MinimumLeadTimeEdit          int      `json:"minimum_lead_time_edit"`
	MaximumLeadTimeEdit          int      `json:"maximum_lead_time_edit"`
	MinimumLeadTimeDelete        int      `json:"minimum_lead_time_delete"`
	MaximumLeadTimeDelete        int      `json:"maximum_lead_time_delete"`
	LastDayOfMonthEnabled        bool     `json:"last_day_of_month_enabled"`
	FirstScheduledDateModifiable bool     `json:"first_scheduled_date_modifiable"`
	TimeUnitModifiable           bool     `json:"time_unit_modifiable"`
	PeriodLengthModifiable       bool     `json:"period_length_modifiable"`
	ScheduledDateModifiable      bool     `json:"scheduled_date_modifiable"`
	LastScheduleDateModifiable   bool     `json:"last_schedule_date_modifiable"`
}

type ScheduledTransferCapabilities struct {
	Supported bool `json:"supported"`
}

type Period struct {
	Type   string `json:"type"`
	Repeat int    `json:"repeat"`
}
