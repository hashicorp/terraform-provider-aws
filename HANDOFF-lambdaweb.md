# Lambda Web (Terraform AWS Provider) - Reviewer Handoff

Private feature. Confidential. Pre-GA.

This document is the entry point for the HashiCorp reviewer. It covers what
the contribution adds, how to build and test it before the official SDK ships,
and the one external item still pending. It is a private-workflow artifact:
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

## 3. Vendored generated SDK module (important)

The official `aws-sdk-go-v2/service/lambdaweb` module is not on the public
proxy yet. The provider builds against the **generated** module from the
Trebuchet drop `awsSdkGoV2.zip` (`service/lambdaweb`, version
`1.0.0-zeta.1587f42719a3`, confirmed identical to the GA build), vendored
locally. The earlier hand-written shim is gone.

- go.mod (root):
  - require: `github.com/aws/aws-sdk-go-v2/service/lambdaweb v0.0.0-preview`
  - replace: `github.com/aws/aws-sdk-go-v2/service/lambdaweb => ./.pre-ga-sdk/lambdaweb`
  - core bumped to what the generated module requires:
    `aws-sdk-go-v2 v1.45.1`, `smithy-go v1.28.1`,
    `internal/configsources v1.5.1`, `internal/endpoints/v2 v2.8.1`
- `.pre-ga-sdk/lambdaweb/` is the unzipped `service/lambdaweb/` minus
  `snapshot/`, `request_snapshot/`, `response_snapshot/` and `*_test.go`, with
  the three local `replace` lines removed from its own `go.mod`.
- Generated types use `*int32` for `Weight`, `MaxEnvironments`,
  `MaxConcurrencyPerEnvironment`, `TimeoutSeconds` and `RateLimit`; the
  resource code and tests already match.

Client operations the module exposes:
- Functions: `CreateWebFunction`, `GetWebFunction`, `DeleteWebFunction`, `ListWebFunctions`
- Revisions: `CreateWebFunctionRevision`, `GetWebFunctionRevision`, `DeleteWebFunctionRevision`, `ListWebFunctionRevisions`
- Endpoints: `CreateWebFunctionEndpoint`, `GetWebFunctionEndpoint`, `UpdateWebFunctionEndpoint`, `DeleteWebFunctionEndpoint`, `ListWebFunctionEndpoints`
- Tagging: `ListTags`, `TagResource`, `UntagResource`
- Resource policy: `PutResourcePolicy`, `GetResourcePolicy`, `DeleteResourcePolicy`

### GA swap (single dependency-only change)

When the module is published on the proxy (post-GA):
1. `go mod edit -dropreplace` to drop the local replace, then pin the published tag.
2. Delete `.pre-ga-sdk/lambdaweb/`.
3. `go mod tidy`. No code changes expected: the vendored copy is the same generated code.

The exact commands are in section 7.

Per the contributor guide, this dependency swap is a separate change from the
resource logic, so it stays in its own commit.

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
call and skips on the not-available signature. Run in any of the 17 regions
verified on 2026-08-21 (see below).

Last full run (us-east-1, 2026-08-27): 45 acceptance PASS, 0 FAIL, single
clean run. `make ci-quick` green (providerlint, golangci-lint 0 issues,
import-lint, semgrep 0 findings). 0 orphan resources after runs.

Acceptance re-run 2026-10-05 against the generated SDK module (account
279826012703): the full `TestAcc` suite (46 tests, including identity, tags
and data-source tests) passed 46/46 in us-east-1 (alternate eu-west-1, 307 s)
and 46/46 in eu-west-1 (alternate us-east-1, 241 s), 0 skips, 0 orphan
resources. Then re-run after the `auth_type` / `regions` contract fixes below
(see section 5), 46/46 again in us-east-1.

Multi-region validation (2026-08-21): `Function_basic` + `Endpoint_basic` +
`ResourcePolicy_basic` pass in all 17 regions where the API is available:
us-east-1, us-east-2, us-west-1, us-west-2, ca-central-1, sa-east-1,
eu-west-1, eu-west-2, eu-west-3, eu-central-1, eu-north-1, ap-northeast-1/2/3,
ap-south-1, ap-southeast-1/2. The other commercial regions return
`AccessDeniedException` (API not deployed) and tests skip cleanly there.

## 5. Contract notes worth knowing during review

- **Endpoint type is immutable after creation** (`HomeRegion` /
  `MultiRegion` / `PerRegion`) and carries `RequiresReplace` on the standalone
  endpoint and on the inline `endpoint_config`.
- **`auth_type` is editable in place** (`ApplicationManaged` <-> `IamAuth`).
  The launch-contract wording said it was fixed at creation, but
  `UpdateWebFunctionEndpoint` accepts `authType` in the generated SDK and the
  CLI model, and a live probe on 2026-10-05 flipped it both ways with
  `updateStatus: Successful` and the same `domainName`. The provider updates
  it in place (both resources) and `TestAccLambdaWebEndpoint_basic` asserts an
  `Update` action plus an unchanged domain name. If the service docs still say
  "immutable", the docs are behind the API.
- **`auto_deployment_mode` defaults to `LatestRevision` for `HomeRegion`.**
  `CreateWebFunction` and `CreateWebFunctionEndpoint` default a missing mode
  to `Disabled`, which pins the endpoint to revision 1, so both resources send
  `LatestRevision` explicitly when the mode is unset. The value is read back
  into state.
- **`MultiRegion` and `PerRegion` require `auto_deployment_mode = "Disabled"`.**
  The provider requires it explicitly at plan time. The service auto-pins the
  initial revision at 100%.
- **`PerRegion` requires at least 2 distinct regions after the home region is
  auto-added** (live probes 2026-10-05): `regions = ["eu-west-1"]` from
  us-east-1 is accepted and becomes `[eu-west-1, us-east-1]`; `["us-east-1"]`
  and an omitted `regions` are both rejected ("requires at least 2 distinct
  regions ... The request results in the regions [us-east-1]"). The provider
  requires `regions` for `PerRegion` at plan time and leaves the single-region
  case to the API (it cannot know the home region during validation). No
  minimum is confirmed for `MultiRegion`. The per-endpoint region count is a
  service quota, so the schema only applies the model limit (1-100). `regions`
  order is API-owned, so it is modeled as a set.
- **`revision_config` and `endpoint_config` are required**, as are the nested
  `build_config`, `runtime_config`, `code_config`, `s3_object` and
  `service_config` blocks (`listvalidator.IsRequired`).
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
  breaks it (the service does not propagate source context). Alias ARNs are
  rejected, so `kms_key_arn` is validated against the key ARN pattern
  `arn:(aws[a-z-]*):kms:[a-z0-9-]+:\d{12}:key/[a-z0-9-]+`. Both are documented
  on the `kms_key_arn` attribute.
- **`scaling_config` and `throttle_config`**: endpoint-level `maxEnvironments`
  and `rateLimit`. Modeled as Optional+Computed `ObjectAttribute`s (blocks
  cannot be Computed; proto5 carries object attribute types fine). Live
  behavior: the API echoes them only when explicitly set — unset endpoints
  report nothing and account-level defaults apply invisibly. `rateLimit`
  accepts only quantized values (0, 100-1000 step 100, 2000-10000 step 1000);
  the service rejects anything else with a `ValidationException` listing the
  values.
- **The service assigns default `telemetryConfig` to every new revision**
  (a log group and INFO log levels). Revisions are immutable, so this is a
  server default, not drift: `restoreTelemetryConfig` keeps `telemetry_config`
  as configuration-only state (the configured value replaces the server echo
  on create and refresh; import leaves it unset). Without it every apply fails
  with "block count changed from 0 to 1".

## 6. Still pending (external, not code)

1. **Published `aws-sdk-go-v2/service/lambdaweb` module on the Go proxy.** The
   generated code is already vendored (section 3); only the `replace` has to
   go once a tag exists.

HashiCorp's test accounts (primary 187416307283, alternate 067819342479) are
allowlisted for the pre-release API in us-east-1 and eu-west-1, so acceptance
tests run there directly. Re-confirm after GA.

The provider-side changes are done and verified against the generated SDK
module (vendored copy of the GA build) with full build, vet and unit tests.
What remains is pointing go.mod at the published tag and an acceptance re-run
(section 7).

## 7. Remaining: pin the published SDK tag and acceptance re-run

1. Pin the published module, in its own commit (replace `vX.Y.Z` with the
   version on the proxy; if the module path is not `service/lambdaweb`, stop:
   the package, `names/data/names_data.hcl`, resource names and docs need a
   rename first):

   ```
   go mod edit -dropreplace=github.com/aws/aws-sdk-go-v2/service/lambdaweb
   go mod edit -require=github.com/aws/aws-sdk-go-v2/service/lambdaweb@vX.Y.Z
   git rm -r .pre-ga-sdk/lambdaweb
   GOPROXY=direct go mod tidy
   ```

2. Confirm nothing drifted between the vendored zeta build and the published
   tag (`int32` fields, `AccessDeniedException` in `testAccPreCheck` are
   already aligned):

   ```
   GOPROXY=direct go build ./...
   GOPROXY=direct go vet ./internal/service/lambdaweb/...
   go test ./internal/service/lambdaweb/... -run 'Test[^A]'
   make gen
   ```

3. Re-run the acceptance suite in both regions (done 2026-10-05 against the
   vendored module, 46/46 each; repeat once against the published tag), then
   CI:

   ```
   AWS_DEFAULT_REGION=us-east-1 AWS_ALTERNATE_REGION=eu-west-1 TF_ACC=1 \
     go test ./internal/service/lambdaweb/ -v -timeout 90m \
     -run 'TestAccLambdaWebFunction|TestAccLambdaWebEndpoint|TestAccLambdaWebResourcePolicy'
   AWS_DEFAULT_REGION=eu-west-1 AWS_ALTERNATE_REGION=us-east-1 TF_ACC=1 \
     go test ./internal/service/lambdaweb/ -v -timeout 90m \
     -run 'TestAccLambdaWebFunction|TestAccLambdaWebEndpoint|TestAccLambdaWebResourcePolicy'
   make ci-quick
   ```

4. PR prep: rename `.changelog/1.txt` and
   `.changelog/lambdaweb-tagging-resourcepolicy.txt` to `{PR_NUMBER}.txt`,
   remove the "Pre-GA" header comments (`function.go`, `endpoint.go`,
   `resource_policy.go`) and the pre-GA wording in `function_test.go`, delete
   this file, then open the PR.
