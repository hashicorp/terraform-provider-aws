<!-- Copyright IBM Corp. 2014, 2026 -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

# Oracle Database@AWS DB node management actions

**Reviewed:** 2026-10-06

**Status:** Implemented and live validated for cloud VM cluster nodes; release pending

## Decision

Expose three native Terraform actions: `aws_odb_start_db_node`, `aws_odb_stop_db_node`, and `aws_odb_reboot_db_node`. Each invocation addresses one existing cloud VM cluster and DB node. A lifecycle operation does not own, create, replace, or delete either resource.

Actions fit the explicit, repeatable nature of a reboot. A resource that permanently declares `reboot = true` would need extra state conventions to distinguish another request from an unchanged apply. Native actions instead run only through an explicit CLI invocation or a configured resource lifecycle trigger. Planning, reading data sources, and refreshing do not invoke a lifecycle API. A stable `terraform_data.input` revision requests one operation for a successful create or update; another revision requests another operation. An unchanged apply after success produces no new invocation. A failed apply may be retried and is not an exactly-once guarantee.

The implementation uses the existing ODB client and the AWS SDK for Go v2. It does not invoke the AWS CLI, depend on a provisioner, or change the dependency versions in `go.mod`.

## CLI/API coverage review

The review compares the five node APIs with the provider at upstream commit `b71a8748fd` and the current SDK dependency, `github.com/aws/aws-sdk-go-v2/service/odb v1.24.0`.

| AWS operation | Provider support before this change | Delivery in this change | Remaining gap |
| --- | --- | --- | --- |
| [ListDbNodes](https://docs.aws.amazon.com/odb/latest/APIReference/API_ListDbNodes.html) | `aws_odb_db_nodes` data source, including pagination, for a cloud VM cluster | Existing discovery remains available | Exascale cluster selector is not exposed by the data source |
| [GetDbNode](https://docs.aws.amazon.com/odb/latest/APIReference/API_GetDbNode.html) | `aws_odb_db_node` data source for a cloud VM cluster | Used for preflight and completion polling | Exascale cluster selector is not exposed by the data source |
| [StartDbNode](https://docs.aws.amazon.com/odb/latest/APIReference/API_StartDbNode.html) | No native node action | `aws_odb_start_db_node` | Exascale cluster selector requires a separate provider extension |
| [StopDbNode](https://docs.aws.amazon.com/odb/latest/APIReference/API_StopDbNode.html) | No native node action | `aws_odb_stop_db_node` | Exascale cluster selector requires a separate provider extension |
| [RebootDbNode](https://docs.aws.amazon.com/odb/latest/APIReference/API_RebootDbNode.html) | No native node action | `aws_odb_reboot_db_node` | Exascale cluster selector and stronger completion evidence, if AWS exposes it |

The pinned SDK v1.24.0 exposes `exadbVmClusterId` as an alternative to `cloudVmClusterId` in all five generated request types. Exascale support is therefore a provider scope gap, not an unavailable SDK feature. The AWS API reference pages for start, stop, reboot, and get still show only the cloud VM cluster selector at review time, while the [ListDbNodes API](https://docs.aws.amazon.com/odb/latest/APIReference/API_ListDbNodes.html) and [AWS SDK for Ruby operation reference](https://docs.aws.amazon.com/sdk-for-ruby/v3/api/Aws/Odb/Client.html#start_db_node-instance_method) show both selectors. The story's explicit `cloudVmClusterId` flow is implemented first. Supporting the alternative requires exactly-one-selector validation, discovery changes, corresponding IAM review, and an eligible Exascale test environment.

These actions do not accept Autonomous VM cluster identifiers. Autonomous database lifecycle operations use distinct AWS APIs and are outside this DB node change. Listing a node or finding an IAM permission with a node-related name does not establish that every node operation is supported for that cluster type. In the SDK, the node-specific operation set is Get, List, Start, Stop, and Reboot; `odb:CreateDbNode` and `odb:DeleteDbNode` are authorization permissions used with cluster operations, not standalone DB node creation/deletion APIs to implement here.

## Inputs and permissions

Each action requires `cloud_vm_cluster_id` and `db_node_id`, accepts an optional provider region override, and has an optional `timeout` in seconds. The default timeout is 3600 seconds; the supported range is 1–86400 seconds. Both identifiers must be 6–64 characters and match `^[a-zA-Z0-9_~.-]+$`. They are AWS resource identifiers, not ARNs or OCI OCIDs. The API validates that the node belongs to the selected cluster.

The caller needs the matching mutation permission (`odb:StartDbNode`, `odb:StopDbNode`, or `odb:RebootDbNode`) plus `odb:GetDbNode`. The [service authorization reference](https://docs.aws.amazon.com/service-authorization/latest/reference/list_odb.html) lists both `cloud-vm-cluster` and `db-node` as required resource types for these four permissions. The corresponding ARN forms are `arn:PARTITION:odb:REGION:ACCOUNT:cloud-vm-cluster/CLUSTER_ID` and `arn:PARTITION:odb:REGION:ACCOUNT:db-node/NODE_ID`. `odb:ListDbNodes` is additionally needed when discovering nodes. The live test also reads the parent cluster and requires `odb:GetCloudVmCluster`.

No new credential or database password input is introduced. The action uses the provider's configured AWS credentials and endpoint handling. Its added progress and logging messages should contain only operational metadata and validated node status, without dumping requests, node objects, hostnames, IP addresses, OCI metadata, or service-supplied free text such as `statusReason`.

## Lifecycle and completion

The AWS [DB node status model](https://docs.aws.amazon.com/odb/latest/APIReference/API_DbNode.html) has `AVAILABLE`, `FAILED`, `PROVISIONING`, `TERMINATED`, `TERMINATING`, `UPDATING`, `STOPPING`, `STOPPED`, and `STARTING`. It has no distinct `REBOOTING` state. The following initial-state policy is a conservative provider guard; the API reference does not publish a complete state-transition contract.

| Requested operation | Initial eligible state | Already complete | Completion state |
| --- | --- | --- | --- |
| Start | `STOPPED` | `AVAILABLE` succeeds without a mutation | `AVAILABLE` |
| Stop | `AVAILABLE` | `STOPPED` succeeds without a mutation | `STOPPED` |
| Reboot | `AVAILABLE` | No no-op; each explicit invocation requests a reboot | Transition observed, then two consecutive `AVAILABLE` polls |

Other initial states fail before dispatch. This avoids sending overlapping lifecycle commands during maintenance, provisioning, or another operation. During a start or stop, stale reads of the starting state can occur before the expected `STARTING` or `STOPPING` state. For reboot, `UPDATING`, `STOPPING`, `STOPPED`, and `STARTING` can provide evidence of a transition. Unexpected states, service errors, cancellation, and timeout produce a diagnostic rather than a successful invocation.

The request and polling share a bounded context. Mutation calls disable automatic SDK retries because these APIs have no client idempotency token, and a network error could follow a request AWS already accepted. Status reads retain the SDK's retry behavior. No automatic second reboot is attempted after an ambiguous result.

Reboot cannot use a simple `AVAILABLE` waiter: that state is also present before reboot starts. The response and subsequent polls must provide evidence of a transition before two consecutive final-state reads can complete the action. Consecutive reads reduce the risk of a single stale `AVAILABLE` response. They cannot prove operation identity or guarantee that AWS reads are current. The API has no operation handle, last-reboot time, or other completion token; a reboot that completes without an observed transition therefore times out conservatively. Live validation must establish the actual sequence for the supported environment.

Success means the ODB service reports the requested node lifecycle state. It does not assert database connection or application readiness. A timeout or canceled Terraform process does not roll back an accepted AWS request. An operator must inspect node status before requesting another disruptive operation after such an outcome.

## Versions and validation

Terraform actions require [Terraform v1.14.0 or later](https://developer.hashicorp.com/terraform/plugin/framework/actions/testing). The first released AWS provider version containing these actions is not yet assigned. Until release, testing requires a provider built from this change; the current published provider must not be assumed to include it. This branch requires Go 1.26.8 according to `go.mod`.

Offline fixtures cover all three operations, schema validation, API routing, initial-state guards, no-op start/stop behavior, status progression, stale reads, malformed responses, failures, deadlines, cancellation, and redaction. The live Terraform acceptance test additionally checks that plan/refresh and unchanged applies do not invoke operations, and that advancing the revision requests another reboot.

Validation recorded on 2026-10-06:

* `make build` passed for the full provider.
* `go test ./internal/provider/framework -run '^TestProviderInit$' -count=1 -p=4` passed provider registration and schema checks.
* The full offline ODB package suite passed: 53 top-level tests and 295 subtests, including 154 new action tests and subtests. The new action implementation has 106 of 107 statements covered (99.1%). Acceptance tests requiring live AWS resources were skipped.
* `go vet -p=4 ./internal/service/odb` passed.
* All four Terraform documentation snippets passed HCL parsing and formatting checks. Repository-version Terrafmt checks passed for the three action reference pages and the two new test files; spelling checks passed for the added documentation and changelog.
* Repository golangci-lint v2.12.2 configurations 1, 2, 3, and 5 passed for the entire ODB package, including tests. Configuration 4 passed for ODB production code with `--tests=false`; its run including tests was interrupted after exceeding the configured timeout. No findings were suppressed to obtain these results.
* After correcting the partition precheck, the actual acceptance test passed with Terraform 1.14.0 against a loopback API simulator using fake credentials and blocked outbound proxies. All nine steps completed, with exactly one stop, one start, and two reboot requests, including unchanged-plan checks and cleanup. This validates the Terraform harness, not real ODB lifecycle behavior. The DB-node unit tests and full `make build` also passed again.
* The contributor ran `TestAccODBDBNodeAction_existingNode` successfully against an existing cloud VM cluster node in `us-east-1` through a configured ODB service endpoint. The run used commit `3ab846ded7`, Terraform 1.14.0, and Go 1.26.8 on Darwin ARM64. The test passed in 309.29 seconds (314.909 seconds for the package), covering stop, start, two reboots, and unchanged-apply checks. The runner confirmed restoration to `AVAILABLE`. The contributor supplied the run log and a redacted screenshot as evidence; no account credentials or endpoint details are recorded here.

* After adopting the repository's serial acceptance-test pattern and replacing `PlanOnly` with pre-apply plan checks, the revised eight-step sequence passed against the loopback simulator in 100.38 seconds, again issuing exactly one stop, one start, and two reboots. The live-test runner rejects skipped child tests even when the parent reports a pass; all eight retained runner regression cases passed.
* `make semgrep-code-quality semgrep-constants PKG=odb` passed with CI's Semgrep v1.168.0. DB-node unit tests, ShellCheck for both runner scripts, and Terraform configuration formatting checks passed after these test changes.

`make swissshepherd` passed with the required Swiss Shepherd v0.24.0 after moving operation behavior and permissions guidance into each action page's introduction. The check rebuilt the full provider and generated its schema with Terraform 1.14.0; all documentation checks passed. The `quick-fix` target was not run because even `PKG=odb` rewrites unrelated copyright, example, and website files; verification used read-only checks instead.

The opt-in `TestAccODBDBNodeAction_existingNode` exercises an existing, initially `AVAILABLE` cloud VM cluster and DB node. It stops, starts, reboots, and explicitly requests another reboot, verifying the resulting status and action plan counts. Terraform manages only a `terraform_data` trigger. The test registers cleanup before mutations and attempts to return a stopped node to `AVAILABLE` using a fresh bounded context; it does not delete or recreate the node or cluster. Cleanup cannot guarantee restoration if AWS is unavailable or a lifecycle operation fails.

Preflight verifies service access and resource eligibility through `GetCloudVmCluster` and `GetDbNode` on the configured endpoint. The partition service catalog bundled in `aws-sdk-go-base` omits ODB and would incorrectly skip this test, including in the standard AWS partition. The test therefore uses those resource reads to establish availability and reports read failures before requesting an action.

Run only against a disposable or explicitly approved maintenance target, with suitable AWS credentials and Terraform v1.14.0 or later:

```console
AWS_PROFILE='YOUR_PROFILE' ./.ci/scripts/test-odb-db-node-actions.sh REGION CLOUD_VM_CLUSTER_ID DB_NODE_ID
```

Execute the runner with Bash; do not source it into an interactive shell. It sets both AWS region variables, checks Terraform's version, runs serially, captures a private log, and treats a skipped test as a failure. On macOS, it uses `caffeinate` when available to keep the machine awake through cleanup. Each action and recovery operation has its own deadline; the runner disables Go's overall test timeout so it does not kill the recovery process. Use `--dry-run` before the three positional arguments to inspect the target and command without executing the test. The equivalent direct command is:

```console
TF_ACC=1 \
TF_ACC_ODB_DB_NODE_ACTIONS=1 \
TF_ACC_ODB_CLOUD_VM_CLUSTER_ID='CLOUD_VM_CLUSTER_ID' \
TF_ACC_ODB_DB_NODE_ID='DB_NODE_ID' \
AWS_REGION='REGION' \
AWS_DEFAULT_REGION='REGION' \
go test ./internal/service/odb -run '^TestAccODBDBNodeAction_existingNode$' -count=1 -parallel=1 -timeout=0 -v
```

The two explicit node identifiers and action opt-in guard prevent a general acceptance test run from selecting an arbitrary node for disruptive testing. The successful live run above validates the tested cloud VM cluster environment. For subsequent runs, record the actual environment, command result, observed transitions, and restoration result.

## Delivery assessment for the story

The cloud VM cluster implementation covers three action registrations, one shared execution path, offline regression coverage, successful live acceptance validation, and three user-facing reference pages. Provider release scheduling remains a separate completion gate. Exascale parity adds two discovery interfaces, selector validation and request-routing coverage for all five operations, and a second eligible live environment; it should be estimated and tracked explicitly if included in the delivery target.

The live result is recorded above; no provider release number or deployment date is assigned. The delivery target remains subject to acceptance of the cloud VM cluster scope and upstream release scheduling. Use this review and the final build/test results as the basis for the story update and target confirmation.
