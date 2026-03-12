<!-- Copyright IBM Corp. 2014, 2026 -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

# Go-VCR

The Terraform AWS provider utilizes [`go-vcr`](https://github.com/dnaeon/go-vcr) to improve acceptance test performance and reduce costs.

`go-vcr` is a Go library for recording and replaying HTTP requests.
In the context of [Terraform provider acceptance testing](https://developer.hashicorp.com/terraform/plugin/framework/acctests), replaying recorded interactions allows core provider logic to be exercised without provisioning real infrastructure.
The benefits are more pronounced for long-running tests[^1] as the built-in polling mechanisms which would typically wait for resource creation or modification can be by-passed, resulting in quicker feedback loops for developers.

!!! Note
    Maintainers are actively rolling out `go-vcr` support across service packages.
    Not all service will support recording and replaying interactions, and those that do may still have gaps for certain styles of tests.
    Subscribe to this [meta issue](https://github.com/hashicorp/terraform-provider-aws/issues/25602) for progress updates.

## Using `go-vcr`

The AWS provider supports two VCR modes - record and replay.

To enable `go-vcr`, the `VCR_MODE` and `VCR_PATH` environment variables must both be set.
The valid values for `VCR_MODE` are `RECORD_ONLY` and `REPLAY_ONLY`.
`VCR_PATH` can point to any path on the local filesystem.

!!! tip
    Always use the same directory for recording and replaying acceptance tests.
    This will maximize re-use of recorded interactions and the corresponding cost savings.

### Recording Tests

`RECORD_ONLY` mode will intercept HTTP interactions made by the provider and write request and response data to a YAML file at the configured path.
A randomness seed is also stored in a separate file, allowing for replayed interactions to generate the same resource names and appropriately match recorded interaction payloads.
The file names will match the test case with a `.yaml` and `.seed` extension, respectively.

To record tests, set `VCR_MODE` to `RECORD_ONLY` and `VCR_PATH` to the test recording directory.
For example, to record Log Group resource tests in the `logs` package:

```sh
make testacc PKG=logs TESTS=TestAccLogsLogGroup_ VCR_MODE=RECORD_ONLY VCR_PATH=/path/to/testdata/ 
```

### Replaying Tests

`REPLAY_ONLY` mode replays recorded HTTP interactions by reading the local interaction and seed files.
Each outbound request is matched with a recorded interaction based on the request headers and body.
When a matching request is found, the recorded response is sent back.
If no matching interaction can be found, an error is thrown and the test will fail.

!!! tip
    A missing interaction likely represents a gap in `go-vcr` support.
    If the underlying cause is not already being tracked (check the open tasks in the [meta issue](https://github.com/hashicorp/terraform-provider-aws/issues/25602)) a new issue should be opened.

To replay tests, set `VCR_MODE` to `REPLAY_ONLY` and `VCR_PATH` to the test recording directory.
For example, to replay Log Group resource tests in the `logs` package:

```sh
make testacc PKG=logs TESTS=TestAccLogsLogGroup_ VCR_MODE=REPLAY_ONLY VCR_PATH=/path/to/testdata/ 
```

## Enabling `go-vcr`

Enabling `go-vcr` support for a service primarily involves replacing certain functions and data structures with "VCR-aware" equivalents.
Broadly this includes service clients, acceptance test data structures, status check functionality (waiters), and any functionality which generates names.

Semgrep rules have been written to automate the majority of these changes.
The `vcr-enable` Make target will apply semgrep rules and then format code and imports for a given package.

```sh
make vcr-enable PKG=logs
```

### Additional Changes

The changes made by semgrep may leave the code in a state which will not compile or conflicts with code generation.
When this occurs some manual intervention may be required before running acceptance tests.

#### Test Check Helper Functions

The most common manual changes required are to acceptance test check helper functions (similar to "check exists" or "check destroy", but not covered via semgrep), which might now reference a `*testing.T` argument within the function body.
Adding a `*testing.T` argument to the function signature will resolve the missing reference.

For example, this was the change applied to the `testAccCheckMetricFilterManyExists` helper function in the `logs` package:

```diff
-func testAccCheckMetricFilterManyExists(ctx context.Context, basename string, n int) resource.TestCheckFunc {
+func testAccCheckMetricFilterManyExists(ctx context.Context, t *testing.T, basename string, n int) resource.TestCheckFunc {
```

#### Generated Tagging Tests

If the service includes resources with generated tagging and Resource Identity tests, two `@Testing` annotations must be removed to ensure regenerating the tests does not remove the `*testing.T` arguments.
Remove the following annotation flags from the resource definition:

```go
// @Testing(existsTakesT=false, destroyTakesT=false)
```

### Validating Changes

#### Compilation Checks

Once code changes are made, do some basic verification to ensure code generation is correct and the provider compiles.

To verify code generation is correct (no diff should be present in generated files):

```sh
go generate ./internal/service/<service-name>
```

To verify the provider compiles:

```sh
make build
```

To verify tests compile:

```sh
go test ./internal/service/<service-name>
```

#### Acceptance Tests

The most time consuming part of enabling `go-vcr` for a service is validating acceptance test results.
**The full acceptance test suite should run in `RECORD_ONLY` mode with no errors.**

There are known support gaps which may result in test failures when running in `REPLAY_ONLY` mode.
This is not a blocker for enabling `go-vcr` in the service, though it is worth verifying the failures are caused by known gaps already documented in the meta-issue.
A new issue should be opened for any failures that appear unrelated to those already being tracked.

Once test validation is complete, a pull request can be opened with the changes and test results.

[^1]:  The full acceptance test suite for certain resources can take upwards of 4 hours to complete. These are typically resources which need to provision compute as part of their lifecycle, such as an [RDS](https://aws.amazon.com/rds/) database or [ElastiCache](https://aws.amazon.com/elasticache/) cluster.
