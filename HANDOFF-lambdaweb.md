# Lambda Web (Terraform AWS Provider) - Reviewer Handoff

Private feature. Confidential. Pre-GA.

This document is the entry point for the HashiCorp reviewer. It covers what
the contribution adds, how to build and test it before the official SDK ships,
and the two external items still pending. It is a private-workflow artifact:
it gets dropped from the branch before the public PR opens at GA.

## 1. What this adds

Terraform support for AWS Lambda Web Functions (the service formerly called
Lambda Lite). Plugin Framework, not SDKv2.

Resources:
- `aws_lambdaweb_function` - the web function plus an inline `revision_config`
  (code from an S3 object) and an inline `endpoint_config`.
- `aws_lambdaweb_endpoint` - a standalone endpoint, needed to shift traffic
  with `revision_weights` (the inline `endpoint_config` cannot express weights).
- `aws_lambdaweb_resource_policy` - the function resource policy.

Data sources:
- `aws_lambdaweb_function`
- `aws_lambdaweb_endpoint`

Tagging is managed through the native Lambda Web tagging API
(`@Tags(identifierAttribute="arn")`).

## 2. Branch and layout

- Repo: private fork of `hashicorp/terraform-provider-aws`.
- Branch: `f-lambdaweb` (PR #2, single line against `main`).
- Service package: `internal/service/lambdaweb/`.
- Docs: `website/docs/r/lambdaweb_*.html.markdown` and
  `website/docs/d/lambdaweb_*.html.markdown`.
- Changelog: `.changelog/1.txt` (function/endpoint) and
  `.changelog/lambdaweb-tagging-resourcepolicy.txt`. These use placeholder
  names on purpose. They get renamed to `{PR_NUMBER}.txt` when the public PR
  opens at GA.

## 3. Pre-GA SDK shim (important)

The official `aws-sdk-go-v2/service/lambdaweb` module is not public yet. Until
it ships, the provider builds against a hand-written stand-in that mirrors the
exact public surface the generated client will expose.

- go.mod (root):
  - require: `github.com/aws/aws-sdk-go-v2/service/lambdaweb v0.0.0-preview`
  - replace: `github.com/aws/aws-sdk-go-v2/service/lambdaweb => ./.pre-ga-sdk/lambdaweb`
- The shim lives in `.pre-ga-sdk/lambdaweb/`.

Client operations the shim exposes (same names the generated client will have):
- Functions: `CreateWebFunction`, `GetWebFunction`, `DeleteWebFunction`, `ListWebFunctions`
- Revisions: `CreateWebFunctionRevision`, `GetWebFunctionRevision`, `DeleteWebFunctionRevision`, `ListWebFunctionRevisions`
- Endpoints: `CreateWebFunctionEndpoint`, `GetWebFunctionEndpoint`, `UpdateWebFunctionEndpoint`, `DeleteWebFunctionEndpoint`, `ListWebFunctionEndpoints`
- Tagging: `ListTags`, `TagResource`, `UntagResource`
- Resource policy: `PutResourcePolicy`, `GetResourcePolicy`, `DeleteResourcePolicy`

### GA swap (single dependency-only change)

When the official module ships (Trebuchet build `awsSdkGoV2.zip`):
1. `go mod edit -replace` to drop the local replace, pin the real module version.
2. Delete `.pre-ga-sdk/lambdaweb/`.
3. `go mod tidy`.

Per the contributor guide, this dependency swap is a separate change from the
resource logic. The shim is isolated in its own commit for exactly this reason.

## 4. Build and test before GA

```
# build + vet
GOPROXY=direct go build ./internal/service/lambdaweb/...
GOPROXY=direct go vet ./internal/service/lambdaweb/...

# acceptance (needs credentials in an allowlisted account, two regions)
AWS_DEFAULT_REGION=us-east-1 AWS_ALTERNATE_REGION=eu-west-1 TF_ACC=1 \
  go test ./internal/service/lambdaweb/ -v -timeout 60m \
  -run 'TestAccLambdaWebFunction|TestAccLambdaWebEndpoint|TestAccLambdaWebResourcePolicy'
```

Tests skip automatically in regions where the Lambda Web API is not rolled
out yet: a service-level `testAccPreCheck` makes a cheap `ListWebFunctions`
call and skips on the not-available signature. Run in any of the 17 available
regions (see below).

Last full run (us-east-1 primary, eu-west-1 alternate): 44 acceptance PASS,
0 FAIL, single clean run. `make ci-quick` green (providerlint, golangci-lint
0 issues, import-lint, semgrep 0 findings). 0 orphan resources after runs.

Multi-region validation (2026-08-21): `Function_basic` + `Endpoint_basic` +
`ResourcePolicy_basic` pass in all 17 regions where the API is available:
us-east-1, us-east-2, us-west-1, us-west-2, ca-central-1, sa-east-1,
eu-west-1, eu-west-2, eu-west-3, eu-central-1, eu-north-1, ap-northeast-1/2/3,
ap-south-1, ap-southeast-1/2. The other commercial regions return
`AccessDeniedException` (API not deployed) and tests skip cleanly there.

## 5. Contract notes worth knowing during review

- **Endpoint type and auth type are immutable after creation**: `HomeRegion` /
  `MultiRegion` / `PerRegion`; `ApplicationManaged` / `IamAuth`.
- **`MultiRegion` and `PerRegion` require `auto_deployment_mode = "Disabled"`.**
  The service auto-pins the initial revision at 100%.
- **`PerRegion` requires at least 2 distinct regions** (the home region is
  auto-added). `regions` order is API-owned, so it is modeled as a set.
- **Traffic shifting is only settable through `UpdateWebFunctionEndpoint`**
  (`revision_weights`). The inline `endpoint_config` on the function cannot
  express weights, so publishing a new revision on a `Disabled` endpoint does
  not move traffic by itself. `ModifyPlan` warns about this and points users to
  the standalone `aws_lambdaweb_endpoint` resource.
- **`regions` on the endpoint is `Computed` + `UseStateForUnknown`** (fixes a
  false forced-replacement on out-of-band drift).
- **Code is S3-only** (`code_config.s3_object` = bucket + key + version). The
  provider references an existing object; it does not create the bucket. IAM/S3
  policy eventual consistency (`does not have s3:GetObject`, `S3 versioning
  enabled`) is retried on create.
- **KMS contract**: a customer CMK works end to end only with an unconditioned
  `lambda.amazonaws.com` service principal in the key policy. `aws:SourceAccount`
  breaks it (the service does not propagate source context). Documented on the
  `kms_key_arn` attribute.

## 6. Still pending (external, not code)

1. **Official Go v2 SDK build (Trebuchet)** for `aws-sdk-go-v2/service/lambdaweb`.
   Hard blocker for the GA swap. Requested from the Lambda Web service team.
2. **Account allowlisting for the pre-release API** so the reviewer can run
   acceptance in HashiCorp's own accounts:
   - Primary: 187416307283
   - Alternate: 067819342479
   Requested from the service team.

Everything on the provider side is complete and verified against the shim.
