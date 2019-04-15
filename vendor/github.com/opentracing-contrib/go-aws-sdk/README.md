[![Build Status][ci-img]][ci]
[![GoDoc]](http://godoc.org/github.com/opentracing-contrib/go-aws-sdk)
[![Apache-2.0 license](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

# OpenTracing support for AWS SDK in Go

The `otaws` package makes it easy to add OpenTracing support for AWS SDK in Go.

## Installation

```
go get github.com/opentracing-contrib/go-aws-sdk
```

## Documentation

See the basic usage examples below and the [package documentation on
godoc.org](https://godoc.org/github.com/opentracing-contrib/go-aws-sdk).

## Usage

```go
// You must have some sort of OpenTracing Tracer instance on hand
var tracer opentracing.Tracer = ...

// Optionally set Tracer as global 
opentracing.SetGlobalTracer(tracer)

// Create AWS Session
sess := session.NewSession(...)

// Create AWS service client e.g. DynamoDB client
dbCient := dynamodb.New(sess)

// Add OpenTracing handlers using global tracer
AddOTHandlers(dbClient.Client)

// Or specify tracer explicitly
AddOTHandlers(dbClient.Client, WithTracer(tracer))

// Call AWS client
result, err := dbClient.ListTables(&dynamodb.ListTablesInput{})

```

## License

[Apache 2.0 License](./LICENSE).

[ci-img]: https://travis-ci.org/opentracing-contrib/go-aws-sdk.svg?branch=master
[ci]: https://travis-ci.org/opentracing-contrib/go-aws-sdk
[GoDoc]: https://godoc.org/github.com/opentracing-contrib/go-aws-sdk?status.svg
