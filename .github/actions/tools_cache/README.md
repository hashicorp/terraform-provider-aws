<!-- Copyright IBM Corp. 2014, 2026 -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

# Tools Cache

Set up Go for the `.ci/tools` module and restore the **tools cache lane** (installed tool binaries plus the tools build cache) read-only.
This is the reader side of the tools cache lane; see the [workflows README](../../workflows/README.md#go-module-caching) for the overall caching strategy and the writer.

## Usage

### Inputs

| Input | Required | Description |
| ----- | -------- | ----------- |
| `tools` | true | Space- or newline-separated list of `.ci/tools` import paths to install unless the cache matches `.ci/tools` exactly (e.g. `github.com/terraform-linters/tflint`). |

### Example

```yaml
steps:
  - uses: actions/checkout@v7
  - uses: ./.github/actions/tools_cache
    with:
      tools: github.com/terraform-linters/tflint
```

Install more than one tool by passing a multi-line list:

```yaml
  - uses: ./.github/actions/tools_cache
    with:
      tools: |
        github.com/katbyte/terrafmt
        github.com/terraform-linters/tflint
```

## Implementation

The action runs these steps in order:

1. **`actions/setup-go`** with `go-version-file: .ci/tools/go.mod` and `cache: false`.
  The built-in `setup-go` cache is disabled because this action manages the tools cache lane itself;
  leaving it enabled would duplicate the cache and separately save an untrimmed build cache.
  This follows the setup-go [advanced usage guidance](https://github.com/actions/setup-go/blob/main/docs/advanced-usage.md#restore-only-caches) for multi-workflow / parallel builds.

1. **Capture Go cache locations** into the environment (`GOCACHE` from `go env GOCACHE`, `GOBIN_PATH` from `$(go env GOPATH)/bin`).
  These are resolved at runtime because they differ between runner images.

1. **Restore the cache read-only** with [`actions/cache/restore`](https://github.com/actions/cache/tree/main/restore) for `[$GOBIN_PATH, $GOCACHE]`.
  The key is the writer's key for the current `.ci/tools` (`${{ runner.os }}-tools-go-${{ hashFiles('.ci/tools/go.mod', '.ci/tools/go.sum', '.ci/tools/main.go') }}`), with `restore-keys: ${{ runner.os }}-tools-go-` falling back to the most recent entry the writer saved.
  `.ci/tools/main.go` is part of the key because it defines the toolset; adding or removing a tool there produces a new key.
  `actions/cache/restore` never writes to the cache.

1. **Verify restored tools** on an exact hit: check that `$GOBIN_PATH/<last element of each import path>` exists for every requested tool.
  If any binary is missing, the install step runs anyway, so a bad cache entry slows the job instead of failing it.

1. **Install** (`go install <tools>` from `.ci/tools`) unless the exact key hit and every requested binary is present.
  A partial hit restores an entry saved for a different `.ci/tools`, so its binaries may be older versions.
  This happens on a PR that bumps a tool, since PRs can only restore caches saved on `main`, and on the push that merges it, which races the writer.
  The restored build cache still makes the install fast.

The cache is written by the `tools_cache` job in [`provider.yml`](../../workflows/provider.yml), which runs only on `refs/heads/main` and installs every tool imported by [`.ci/tools/main.go`](../../../.ci/tools/main.go) before saving under `${{ runner.os }}-tools-go-${{ hashFiles('.ci/tools/go.mod', '.ci/tools/go.sum', '.ci/tools/main.go') }}`.

> **Note:** `golangci-lint` is cached separately by `golangci-lint-action`'s own internal cache and is not part of the tools lane.
