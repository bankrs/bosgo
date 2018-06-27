# bosgo - a Bankrs OS Go client

This is the official Go client for accessing the Bankrs OS API.

**Documentation:** [![GoDoc](https://godoc.org/code.bankrs.com/bosgo?status.svg)](https://godoc.org/code.bankrs.com/bosgo)

bosgo requires Go version 1.7 or greater.

## Getting started

Ensure you have a working Go installation and then use go get as follows:

```
go get code.bankrs.com/bosgo
```

## Usage

```go
import "code.bankrs.com/bosgo"
```

There are four types of client that represent the three levels of authorisation required for interacting with Bankrs OS:

* Client - the base client that is used to create a new developer account, login as a developer or request lost passwords.
* DevClient - contains the developer session obtained after a successful login, used for creating and deleting applications and obtaining statistics about the developer account.
* AppClient - provides functionality to manage users for an application and for obtaining information on available financial providers and categories.
* UserClient - contains the user session obtained after logging in via an AppClient. Used for all interactions on behalf of a user including reading transactions and processing payments.

Construct a new client and login as a developer to the Bankrs OS sandbox to obtain recent user statistics:

```go
client := bosgo.New(http.DefaultClient, bosgo.SandboxAddr)
devClient, err := client.Login("email", "password").Send()
if err != nil {
    log.Fatalf("failed to login: %v", err)
}

stats, err := devClient.Stats.Users().Send()
if err != nil {
    log.Fatalf("failed to obtain user stats: %v", err)
}
log.Printf("Total users today: %d", stats.UsersToday.value)
```

Some API services have optional parameters that can be passed. For example, to count the number of active users in the past week the FromDate can be set:

```go
request := devClient.Stats.Users()
stats, err := devClient.Stats.Users().FromDate("2017-06-09").Send()
if err != nil {
    log.Fatalf("failed to obtain user stats: %v", err)
}
log.Printf("Total users in past week: %d", stats.UsersTotal.value)
```

### Create a new developer account and application

```go
client := bosgo.New(http.DefaultClient, bosgo.SandboxAddr)
devClient, err := client.CreateDeveloper("email", "password").Send()
if err != nil {
    log.Fatalf("failed to create developer: %v", err)
}

applicationID, err := devClient.Applications.Create("my application").Send()
if err != nil {
    log.Fatalf("failed to create application: %v", err)
}
```

Once an application has been created, it can be used to create user accounts:

```go
appClient := bosgo.NewAppClient(http.DefaultClient, bosgo.SandboxAddr, applicationID)
userClient, err := appClient.Users.Create("username", "password").Send()
if err != nil {
    log.Fatalf("failed to create user: %v", err)
}
```

### Login on behalf of a user

```go
appClient := bosgo.NewAppClient(http.DefaultClient, bosgo.SandboxAddr, "application")
userClient, err := appClient.Users.Login("username", "password").Send()
if err != nil {
    log.Fatalf("failed to login as user: %v", err)
}
```

### Add a bank access for a user and respond to challenges

Obtain a user client as above then add the bank access, providing challenge answers up front:

```go
req := userClient.Accesses.Add("DE-BIN-10010010")
req.ChallengeAnswer(bosgo.ChallengeAnswer{
    ID: "login",
    Value: "fakebank_investor_1",
    Store: true,
})
req.ChallengeAnswer(bosgo.ChallengeAnswer{
    ID: "pin",
    Value: "1234",
    Store: false,
})
job, err := req.Send()
if err != nil {
    log.Fatalf("failed to add bank access: %v", err)
}
```

Adding an access is an asynchronous operation which can be tracked using the job URI returned from the service call:

```go
status, err := userClient.Jobs.Get(job.URI).Send()
if err != nil {
    log.Fatalf("failed to obtain job status: %v", err)
}
log.Printf("job stage: %s", status.Stage)
```

Some jobs may require further challenges to be provided so they can proceed:

```go
req := userClient.Jobs.Answer(job.URI).Send()
req.ChallengeAnswer(bosgo.ChallengeAnswer{
    ID: "pin",
    Value: "5678",
    Store: false,
})
status, err := req.Send()
if err != nil {
    log.Fatalf("failed to obtain job status: %v", err)
}
log.Printf("job stage: %s", status.Stage)
```
