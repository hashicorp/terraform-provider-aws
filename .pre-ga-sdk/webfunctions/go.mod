module github.com/aws/aws-sdk-go-v2/service/webfunctions

// PRE-GA SHIM: This is a hand-written stand-in for the (not-yet-public)
// generated aws-sdk-go-v2 Web Functions service client. It mirrors the exact
// public surface the generated client will expose (Client, Options,
// NewFromConfig, operation methods, *Input/*Output, and a types subpackage),
// built on the real aws-sdk-go-v2 core + SigV4 so terraform-provider-aws can
// import it via a `replace` directive during the private beta. Replace with the
// official module at GA.

go 1.24

require (
	github.com/aws/aws-sdk-go-v2 v1.42.0
	github.com/aws/smithy-go v1.27.2
)
