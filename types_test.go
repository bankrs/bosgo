package bosgo

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"
)

// typeMap is a mapping of api blueprint data structure name to bosgo type
var typeMap = map[string]interface{}{
	"Access":                           Access{},
	"AccessCapabilities":               AccessCapabilities{},
	"Account":                          Account{},
	"AccountAllowedOperations":         AllowedOperations{},
	"AccountCapabilities":              AccountCapabilities{},
	"AccountReference":                 AccountRef{},
	"Answer":                           nil,
	"ApplicationID":                    nil,
	"ApplicationItem":                  nil,
	"ApplicationKey":                   ApplicationKey{},
	"AuthMethod":                       AuthMethod{},
	"BadRequestError":                  nil,
	"BadRequestObj":                    nil,
	"Bank":                             nil,
	"BaseAnswer":                       nil,
	"BaseTransaction":                  nil,
	"Beneficiary":                      nil,
	"Category":                         Category{},
	"CategoryName":                     nil,
	"ChallengeEmpty":                   nil,
	"ChallengeField":                   ChallengeField{},
	"Counterparty":                     Counterparty{},
	"Credential":                       nil,
	"CredentialIndex":                  nil,
	"CredentialKeys":                   nil,
	"CredentialNew":                    nil,
	"CredentialProvider":               nil,
	"Credentials":                      nil,
	"CredentialUpdate":                 nil,
	"DailyMerchantObjStats":            DailyMerchantsStats{},
	"DailyProviderObjStats":            DailyProvidersStats{},
	"DailyProviderPingStats":           nil,
	"DailyRequestsStats":               DailyRequestsStats{},
	"DailyTransfersStats":              DailyTransfersStats{},
	"DailyUsersStats":                  DailyUsersStats{},
	"DeveloperConfirmAction":           nil,
	"DeveloperCredentials":             DeveloperCredentials{},
	"DeveloperOAuthLogin":              nil,
	"DeveloperProfile":                 DeveloperProfile{},
	"Error":                            nil,
	"EventPayload":                     nil,
	"FiOperations":                     nil,
	"IBANValidation":                   IBANDetails{},
	"InitialChallenge":                 nil,
	"JobAccess":                        JobAccess{},
	"JobAccount":                       JobAccount{},
	"JobStatus":                        JobStatus{},
	"JobStatusOK":                      nil,
	"JobURI":                           Job{},
	"LinkedAccount":                    nil,
	"LinkedTeam":                       nil,
	"MAUStats":                         nil,
	"Merchant":                         Merchant{},
	"MerchantsStats":                   MerchantsStats{},
	"Money":                            MoneyAmount{},
	"MonthlyUsersStats":                nil,
	"OriginalAmount":                   OriginalAmount{},
	"Problem":                          Problem{},
	"Provider":                         Provider{},
	"ProviderAllowedOperations":        nil,
	"ProviderPingStats":                nil,
	"ProviderSearchResult":             ProviderSearchResult{},
	"ProvidersStats":                   ProvidersStats{},
	"RecurringTransferCapabilities":    RecurringTransferCapabilities{},
	"RecurringTransferPeriod":          Period{},
	"RepeatedTransaction":              RepeatedTransaction{},
	"RequestsStats":                    RequestsStats{},
	"Schedule":                         RecurrenceRule{},
	"ScheduledTransaction":             nil,
	"ScheduledTransferCapabilities":    ScheduledTransferCapabilities{},
	"StatsMoneyAmount":                 StatsMoneyAmount{},
	"StatsValueChange":                 nil,
	"TeamAccess":                       nil,
	"TeamInvite":                       nil,
	"TeamInviteResponse":               nil,
	"TeamInviteToken":                  nil,
	"TeamMember":                       nil,
	"TeamMemberAccess":                 nil,
	"TeamMemberNewAccess":              nil,
	"TeamMemberUpdateAccess":           nil,
	"TeamNew":                          nil,
	"TeamUpdate":                       nil,
	"Transaction":                      Transaction{},
	"TransactionCategorisationRequest": nil,
	"TransferAddress":                  TransferAddress{},
	"TransferBusinessFieldsObject":     nil,
	"TransferRequest":                  nil,
	"TransferResponse":                 Transfer{},
	"TransfersStats":                   TransfersStats{},
	"TransferStatusResponse":           nil,
	"TransferStep":                     TransferStep{},
	"TransferStepData":                 TransferStepData{},
	"Usernames":                        nil,
	"UsersStats":                       UsersStats{},
	"Webhook":                          Webhook{},
	"WebhookEvent":                     Event{},
	"WebhookNew":                       nil,
	"WebhooksTypes":                    nil,
	"WebhookTest":                      nil,
	"WebhookTestResult":                WebhookTestResult{},
}

func TestTypes(t *testing.T) {
	f, err := os.Open("testdata/bankrs.apib")
	if err != nil {
		t.Fatalf("failed to open bankrs.apib: %v", err)
	}
	defer f.Close()

	p := NewParser(f)

	for p.Next() {
		bpType := p.Type()
		x, ok := typeMap[bpType.Name]
		if !ok || x == nil {
			continue
		}

		t.Run(bpType.Name, func(t *testing.T) {
			bosType := reflect.TypeOf(x)
			if bosType.Kind() != reflect.Struct {
				t.Errorf("unexpected kind of value in typeMap, got %s, wanted %s", bosType.Kind(), reflect.Struct)
			}

			fieldsByTag := map[string]reflect.StructField{}
			for i := 0; i < bosType.NumField(); i++ {
				f := bosType.Field(i)
				names := strings.Split(f.Tag.Get("json"), ",")
				fieldsByTag[names[0]] = f
			}

			bpFields := map[string]bool{}
			for _, bpField := range bpType.Fields {
				if _, ok := fieldsByTag[bpField.Name]; !ok {
					t.Errorf("bosgo is missing field %s", bpField.Name)
					continue
				}
				bpFields[bpField.Name] = true
			}
			for bosField := range fieldsByTag {
				if bosField == "-" {
					continue
				}
				if _, ok := bpFields[bosField]; !ok {
					t.Errorf("bosgo has additional field %s", bosField)
					continue
				}
			}

		})
	}

	if p.Err() != nil {
		t.Errorf("unexpected error: %v", p.Err())
	}
}

type Type struct {
	Name   string
	Fields []Field
}

type Field struct {
	Name     string
	BaseType string
	Required bool
}

type Parser struct {
	s     *bufio.Scanner
	err   error
	state int
	typ   Type
	done  bool
	types map[string]Type
}

func NewParser(r io.Reader) *Parser {
	return &Parser{
		s:     bufio.NewScanner(r),
		types: make(map[string]Type),
	}
}

func (p *Parser) Next() bool {
	if p.err != nil {
		return false
	}

	for p.s.Scan() {

		line := p.s.Text()
		switch p.state {
		case 0:
			if strings.HasPrefix(line, "# Data Structures") {
				p.state = 1
			}
		case 1:
			if strings.HasPrefix(line, "##") {
				fields := strings.Fields(line)
				p.state = 2
				p.typ = Type{
					Name: fields[1],
				}

				if len(fields) > 2 {
					if strings.HasPrefix(fields[2], "(") {
						qualifiers := strings.Split(fields[2][1:], ",")
						baseType := strings.TrimSuffix(qualifiers[0], ")")
						if baseType != "object" {
							btyp, ok := p.types[baseType]
							if !ok {
								p.err = fmt.Errorf("found unknown base type %s for type %s", baseType, p.typ.Name)
								return false
							}
							p.typ.Fields = append(p.typ.Fields, btyp.Fields...)

						}

					}
				}

			}
		case 2:
			if len(p.typ.Fields) > 0 && strings.TrimSpace(line) == "" {
				p.state = 1
				p.types[p.typ.Name] = p.typ
				return true
			}
			if strings.HasPrefix(line, "+") {
				fields := strings.Fields(line)

				if fields[1] == "Include" {
					btyp, ok := p.types[fields[2]]
					if !ok {
						p.err = fmt.Errorf("found unknown included type %s for type %s", fields[2], p.typ.Name)
						return false
					}
					p.typ.Fields = append(p.typ.Fields, btyp.Fields...)
					continue
				}

				field := Field{
					Name: strings.TrimSuffix(fields[1], ":"),
				}

				var typeInfo string
				for i := 2; i < len(fields); i++ {
					if strings.HasPrefix(fields[i], "(") {
						if strings.HasSuffix(fields[i], ")") {
							typeInfo = fields[i][1 : len(fields[i])-1]
							break
						}
						typeInfo = fields[i][1:]
					} else if strings.HasSuffix(fields[i], ")") {
						typeInfo += fields[i][:len(fields[i])-1]
						break
					} else if typeInfo != "" {
						typeInfo += fields[i]
					}
				}

				typeParts := strings.Split(typeInfo, ",")

				field.BaseType = typeParts[0]
				if len(typeParts) > 1 && typeParts[1] == "required" {
					field.Required = true
				}
				p.typ.Fields = append(p.typ.Fields, field)
			}
		}

	}

	if p.s.Err() != nil {
		p.err = p.s.Err()
		return false
	}

	if !p.done {
		p.done = true
		return p.typ.Name != ""
	}

	return false
}

func (p *Parser) Err() error {
	return p.err
}

func (p *Parser) Type() Type {
	return p.typ
}

var sample = "" +
	"# Data Structures\n" +
	"\n" +
	"## DeveloperCredentials (object,fixed-type)\n" +
	"+ email:    `john.doe@bankworld.com` (string, required) - Email address for the developer, requires valid email format\n" +
	"+ password: `F6hC>dEgAWNnmRg.7xBE`   (string, required) - Developer's password\n" +
	"\n" +
	"## DeveloperConfirmAction\n" +
	"+ password: `F6hC>dEgAWNnmRg.7xBE`   (string) - Developer's password if registered by email/password\n" +
	"+ email:    `john.doe@bankworld.com` (string) - Primary email address used on provider side if OAuth authorization is used\n" +
	"\n" +
	"## DeveloperOAuthLogin (object,fixed-type)\n" +
	"+ provider:    `github`              (string,required) - Provider ID\n" +
	"+ code                               (string,required) - OAuth code for exchange to an access `token`\n" +
	"\n" +
	"## LinkedAccount (object,fixed-type)\n" +
	"+ type                               (enum[number],required) - Type of account\n" +
	"    + 0                               - Regular\n" +
	"    + 1                               - Used for authorization\n" +
	"+ id:          `github`              (string,required) - Provider ID\n" +
	"+ title:       `GitHub`              (string,required) - Provider title\n" +
	"+ user_name                          (string,required) - User name or identifier in provider system\n" +
	"+ sync_time                          (string,required) - Provider data sync time\n" +
	"\n" +
	"## MerchantsStats (object,fixed-type)\n" +
	"+ from_date                     (string,required) - the from date\n" +
	"+ to_date                       (string,required) - the to date\n" +
	"+ domain:       `merchants`     (string) - the stats domain\n" +
	"+ stats                         (array[DailyMerchantObjStats],optional,fixed-type) - Top most used\n" +
	"\n" +
	"## BaseAnswer (object,fixed-type)\n" +
	"+ id:       login           (string, required) - Identifier of the answer, which answers a challenge with the same id\n" +
	"+ value:    john_doe_hsbc   (string, required) - Value of of the submitted answer\n" +
	"\n" +
	"## Answer (BaseAnswer,fixed-type)\n" +
	"+ store:        true                    (boolean) - Flag indicating whether the submitted answer should be stored\n" +
	"+ valid_until:  `2018-04-16T22:00:00Z`  (string) - Date when the answer should expire\n" +
	"\n" +
	"## AnswerInclude (object,fixed-type)\n" +
	"+ store:        true                    (boolean) - Flag indicating whether the submitted answer should be stored\n" +
	"+ valid_until:  `2018-04-16T22:00:00Z`  (string) - Date when the answer should expire\n" +
	"+ Include BaseAnswer"

func TestParser(t *testing.T) {
	expected := []Type{
		{
			Name: "DeveloperCredentials",
			Fields: []Field{
				{Name: "email", BaseType: "string", Required: true},
				{Name: "password", BaseType: "string", Required: true},
			},
		},
		{
			Name: "DeveloperConfirmAction",
			Fields: []Field{
				{Name: "password", BaseType: "string", Required: false},
				{Name: "email", BaseType: "string", Required: false},
			},
		},
		{
			Name: "DeveloperOAuthLogin",
			Fields: []Field{
				{Name: "provider", BaseType: "string", Required: true},
				{Name: "code", BaseType: "string", Required: true},
			},
		},
		{
			Name: "LinkedAccount",
			Fields: []Field{
				{Name: "type", BaseType: "enum[number]", Required: true},
				{Name: "id", BaseType: "string", Required: true},
				{Name: "title", BaseType: "string", Required: true},
				{Name: "user_name", BaseType: "string", Required: true},
				{Name: "sync_time", BaseType: "string", Required: true},
			},
		},
		{
			Name: "MerchantsStats",
			Fields: []Field{
				{Name: "from_date", BaseType: "string", Required: true},
				{Name: "to_date", BaseType: "string", Required: true},
				{Name: "domain", BaseType: "string", Required: false},
				{Name: "stats", BaseType: "array[DailyMerchantObjStats]", Required: false},
			},
		},
		{
			Name: "BaseAnswer",
			Fields: []Field{
				{Name: "id", BaseType: "string", Required: true},
				{Name: "value", BaseType: "string", Required: true},
			},
		},
		{
			Name: "Answer",
			Fields: []Field{
				{Name: "id", BaseType: "string", Required: true},
				{Name: "value", BaseType: "string", Required: true},
				{Name: "store", BaseType: "boolean", Required: false},
				{Name: "valid_until", BaseType: "string", Required: false},
			},
		},
		{
			Name: "AnswerInclude",
			Fields: []Field{
				{Name: "store", BaseType: "boolean", Required: false},
				{Name: "valid_until", BaseType: "string", Required: false},
				{Name: "id", BaseType: "string", Required: true},
				{Name: "value", BaseType: "string", Required: true},
			},
		},
	}

	i := 0

	p := NewParser(strings.NewReader(sample))

	for p.Next() {
		typ := p.Type()
		if !reflect.DeepEqual(typ, expected[i]) {
			t.Errorf("got %+v, wanted %+v", typ, expected[i])
		}
		i++
	}

	if p.Err() != nil {
		t.Errorf("unexpected error: %v", p.Err())
	}
}
