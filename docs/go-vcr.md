<!-- Copyright IBM Corp. 2014, 2026 -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

# Go-VCR

The Terraform AWS provider utilizes [`go-vcr`](https://github.com/dnaeon/go-vcr) to improve acceptance test performance and reduce costs.

`go-vcr` is a Go library for recording and replaying HTTP requests.
In the context of [Terraform provider acceptance testing](https://developer.hashicorp.com/terraform/plugin/framework/acctests), replaying recorded interactions allows core provider logic to be exercised without provisioning real infrastructure.
The benefits are more pronounced for long-running tests[^1] as the built-in polling mechanisms which would typically wait for resource creation or modification can be by-passed, resulting in quicker feedback loops for developers.

!!! Note
    Maintainers are actively extending `go-vcr` support.
    Certain styles of tests may still have gaps which prevent recording and replaying tests reliably.
    Subscribe to this [meta issue](https://github.com/hashicorp/terraform-provider-aws/issues/25602) for progress updates.

## Using `go-vcr`

The AWS provider supports two VCR modes - record and replay.

To enable `go-vcr`, the `VCR_MODE` and `VCR_PATH` environment variables must both be set.
The valid values for `VCR_MODE` are `RECORD_ONLY` and `REPLAY_ONLY`.
`VCR_PATH` can point to any path on the local file system.

!!! tip
    Always use the same directory for recording and replaying acceptance tests.
    This will maximize re-use of recorded interactions and the corresponding cost savings.

### Recording Tests

`RECORD_ONLY` mode will intercept HTTP interactions made by the provider and write request and response data to a YAML file at the configured path.
A randomness seed is also stored in a separate file, allowing for replayed interactions to generate deterministic resource names.
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
