#!/usr/bin/env bash
# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

set -euo pipefail
umask 077

usage() {
  cat <<'EOF'
Usage: test-odb-db-node-actions.sh [--dry-run] REGION CLOUD_VM_CLUSTER_ID DB_NODE_ID

Stops, starts, then reboots an existing DB node twice. The node and its Exadata
cloud VM cluster must initially be AVAILABLE. Use a disposable maintenance target.
Cleanup attempts to restore AVAILABLE; it does not delete the node or cluster.

Set AWS_PROFILE, or supply AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY (plus
AWS_SESSION_TOKEN for temporary credentials). Container credentials are also supported.
Permissions: odb:GetCloudVmCluster, odb:GetDbNode, odb:StartDbNode,
odb:StopDbNode, and odb:RebootDbNode.
Requires Go from .go-version and Terraform >= 1.14.0. GO_BINARY and
TF_ACC_TERRAFORM_PATH can select executables; otherwise PATH is used. A local
Agent Workspace Go toolchain is used if Go is absent from PATH.
Execute this file with Bash; do not source it into your interactive shell.

--dry-run prints the selected target and command without checking tools or credentials.
EOF
}

fail() { printf 'Error: %s\n' "$*" >&2; exit 1; }

dry_run=false
case "${1:-}" in
  --help|-h) usage; exit 0 ;;
  --dry-run) dry_run=true; shift ;;
esac
if [[ $# -ne 3 ]]; then
  usage >&2
  exit 2
fi

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$repo_root"
export AWS_REGION="$1" AWS_DEFAULT_REGION="$1"
export TF_ACC_ODB_CLOUD_VM_CLUSTER_ID="$2" TF_ACC_ODB_DB_NODE_ID="$3"
export TF_ACC=1 TF_ACC_ODB_DB_NODE_ACTIONS=1

[[ -n "$AWS_REGION" ]] || fail 'REGION must not be empty.'
for identifier in "$TF_ACC_ODB_CLOUD_VM_CLUSTER_ID" "$TF_ACC_ODB_DB_NODE_ID"; do
  [[ ${#identifier} -ge 6 && ${#identifier} -le 64 && "$identifier" =~ ^[a-zA-Z0-9_~.-]+$ ]] ||
    fail 'Cluster and node IDs must be 6–64 letters, digits, underscores, tildes, periods, or hyphens; use AWS IDs, not ARNs or OCIDs.'
done

test_name=TestAccODBDBNodeAction_existingNode
test_args=(test ./internal/service/odb -run "^${test_name}$" -count=1 -parallel=1 -timeout=0 -v)
printf 'Region: %s\nCluster: %s\nNode: %s\n' "$AWS_REGION" "$TF_ACC_ODB_CLOUD_VM_CLUSTER_ID" "$TF_ACC_ODB_DB_NODE_ID"
printf 'Sequence: stop -> start -> reboot -> reboot, with unchanged-apply checks.\n'
if "$dry_run"; then
  printf 'Dry run only; no tools or credentials checked. Command: go'
  printf ' %q' "${test_args[@]}"
  printf '\n'
  exit 0
fi

if [[ -n "${AWS_ACCESS_KEY_ID:-}" && -z "${AWS_SECRET_ACCESS_KEY:-}" ]]; then
  fail 'AWS_ACCESS_KEY_ID is set but AWS_SECRET_ACCESS_KEY is missing.'
fi
if [[ -z "${AWS_PROFILE:-}" && -z "${AWS_ACCESS_KEY_ID:-}" && -z "${AWS_CONTAINER_CREDENTIALS_FULL_URI:-}" && -z "${AWS_CONTAINER_CREDENTIALS_RELATIVE_URI:-}" ]]; then
  fail 'Set AWS_PROFILE (including AWS_PROFILE=default for the default profile), or provide environment/container credentials.'
fi

go_bin="${GO_BINARY:-$(command -v go || true)}"
if [[ -z "$go_bin" ]]; then
  workspace_go="$repo_root/../../.agent-workspace/toolchains/go$(tr -d '[:space:]' < .go-version)/bin/go"
  if [[ -x "$workspace_go" ]]; then
    go_bin="$workspace_go"
  fi
fi
[[ -n "$go_bin" && -x "$go_bin" ]] || fail 'Go is unavailable. Install the version in .go-version or set GO_BINARY to its absolute path.'
go_bin="$(cd "$(dirname "$go_bin")" && pwd)/$(basename "$go_bin")"

terraform_bin="${TF_ACC_TERRAFORM_PATH:-$(command -v terraform || true)}"
[[ -n "$terraform_bin" && -x "$terraform_bin" ]] || fail 'Terraform is unavailable. Install Terraform >= 1.14.0 or set TF_ACC_TERRAFORM_PATH to its absolute path.'
terraform_bin="$(cd "$(dirname "$terraform_bin")" && pwd)/$(basename "$terraform_bin")"
terraform_version="$(CHECKPOINT_DISABLE=1 "$terraform_bin" version | sed -n '1p')"
if [[ ! "$terraform_version" =~ ^Terraform\ v([0-9]+)\.([0-9]+)\.([0-9]+)([+][a-zA-Z0-9.-]+)?$ ]]; then
  fail 'Could not identify a stable Terraform version; Terraform >= 1.14.0 is required.'
fi
if (( BASH_REMATCH[1] < 1 || (BASH_REMATCH[1] == 1 && BASH_REMATCH[2] < 14) )); then
  fail 'Terraform >= 1.14.0 is required; older versions would skip this test.'
fi
export TF_ACC_TERRAFORM_PATH="$terraform_bin"
unset TF_ACC_TERRAFORM_VERSION

log_dir="${TMPDIR:-/tmp}/odb-db-node-live-tests"
mkdir -p "$log_dir"
test_log="$(mktemp "$log_dir/run.XXXXXX")"
printf 'Terraform: %s\nLog: %s\n' "$terraform_version" "$test_log"
"$go_bin" version

# Keep the machine awake through the test and its bounded recovery cleanup.
# An overall Go timeout would terminate cleanup along with the test process.
test_command=("$go_bin" "${test_args[@]}")
if [[ "$(uname -s)" == Darwin ]] && command -v caffeinate >/dev/null 2>&1; then
  test_command=(caffeinate -i "${test_command[@]}")
fi
set +e
"${test_command[@]}" 2>&1 | tee "$test_log"
results=("${PIPESTATUS[@]}")
set -e
if (( results[0] != 0 )); then
  printf 'Live test failed (exit %s). Check the log and node status; cleanup may require manual recovery.\nLog: %s\n' "${results[0]}" "$test_log" >&2
  exit "${results[0]}"
fi
(( results[1] == 0 )) || fail "Could not write the complete test log: $test_log"
if grep -Eq "^[[:space:]]*--- SKIP: ${test_name}(/|[[:space:]])" "$test_log"; then
  fail "The selected test did not report PASS for every required subtest (a test was skipped). See $test_log"
fi
grep -Eq "^--- PASS: ${test_name}[[:space:]]" "$test_log" ||
  fail "The selected test did not report PASS (it may have been skipped). See $test_log"
printf 'Live test passed, including restoration to AVAILABLE. Log: %s\n' "$test_log"
