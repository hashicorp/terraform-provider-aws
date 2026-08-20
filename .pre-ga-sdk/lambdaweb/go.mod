module github.com/aws/aws-sdk-go-v2/service/lambdaweb

// PRE-GA SHIM: This is a hand-written stand-in for the (not-yet-public)
// generated aws-sdk-go-v2 Lambda Web service client. It mirrors the exact
// public surface the generated client will expose (Client, Options,
// NewFromConfig, operation methods, *Input/*Output, and a types subpackage),
// built on the real aws-sdk-go-v2 core + SigV4 so terraform-provider-aws can
// import it via a `replace` directive during the private beta.
//
// GA SWAP: when the official module ships (built from
// LambdaWebFunctionControlPlaneServiceModel via Trebuchet; see the HashiCorp
// Private Feature Workflow wiki), drop the `replace` line in the root go.mod
// (github.com/aws/aws-sdk-go-v2/service/lambdaweb => ./.pre-ga-sdk/lambdaweb),
// pin the real module version in the root require block (replacing
// v0.0.0-preview), delete this .pre-ga-sdk/lambdaweb directory, and run
// `go mod tidy`. The public surface used by the provider (function.go,
// endpoint.go, tags.go, resource_policy.go) matches the generated client, so
// no provider code changes are expected beyond the module swap.

go 1.24

require (
	github.com/aws/aws-sdk-go-v2 v1.42.0
	github.com/aws/smithy-go v1.27.2
)
