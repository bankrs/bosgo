# testserver - a server to help testing Bankrs OS clients

This is an experimental test server package for testing access to the Bankrs OS API. 
It is intended to provide basic support for testing the main features of client applications. 
The test server starts an http server which can be accessed using the bosgo client library or directly via HTTP requests.

Current API coverage is minimal and focussed on the happy path:

 - [x] Create a user
 - [x] User login
 - [x] User logout
 - [x] Add an access to a user
 - [x] List user accesses
 - [x] List user accounts
 - [x] List transactions

**Documentation:** [![GoDoc](https://godoc.org/github.com/bankrs/bosgo/testserver?status.svg)](https://godoc.org/github.com/bankrs/bosgo/testserver)  

bosgo testserver requires Go version 1.7 or greater.

## Getting started

Ensure you have a working Go installation and then use go get as follows:

```
go get -u github.com/bankrs/bosgo
```

## Usage

This package contains several unit tests that demonstrate how to use the test server. The basic approach is as follows:

```Go
    // Create a test server with some default data
    s := testserver.NewWithDefaults()
    defer s.Close()


    // Create a client using the server-supplied HTTP client which is configured to accept the test server's TLS config
    appClient := bosgo.NewAppClient(s.Client(), s.Addr(), DefaultApplicationID)
 
   // The rest of this code is standard usage of bosgo and does not depend on the test server

    // Login as the default user
    userClient, err := appClient.Users.Login(DefaultUsername, DefaultPassword).Send()
    if err != nil {
         log.Fatalf("failed to login: %v", err)
    }

    // Prepare the request to add the access with challenge answers
    req := userClient.Accesses.Add(DefaultProviderID)
    req.ChallengeAnswer(bosgo.ChallengeAnswer{
        ID:    "login",
        Value: DefaultAccessLogin,
    })
    req.ChallengeAnswer(bosgo.ChallengeAnswer{
        ID:    "pin",
        Value: DefaultAccessPIN,
    })

    // Send the request
    job, err := req.Send()
    if err != nil {
         log.Fatalf("failed to send the add access request: %v", err)
    }

    // Check the status of the job to see if it has finished
    status, err := userClient.Jobs.Get(job.URI).Send()
    if err != nil {
         log.Fatalf("failed to get job status: %v", err)
    }
    if status.Finished != true {
        log.Fatalf("job not finished")
    }
    if len(status.Errors) != 0 {
        log.Fatalf("job status error: %v", status.Errors[0])
    }

    // List accesses for the user
    ac, err := userClient.Accesses.List().Send()
    if err != nil {
        log.Fatalf("failed to retrieve accesses: %v", err)
    }
    if len(ac.Accesses) != 1 {
        log.Fatalf("got %d accesses, wanted 1", len(ac.Accesses))
    }
```
