<!-- Copyright IBM Corp. 2014, 2026 -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

# GitHub Workflows

## This README Is Out-of-Date

This README is not maintained. Instead, refer to the Contributor Guide:

* [Continuous integration](https://hashicorp.github.io/terraform-provider-aws/continuous-integration/)
* [Makefile cheat sheet](https://hashicorp.github.io/terraform-provider-aws/makefile-cheat-sheet/)

## Using the `setup-terraform` action

By default, the [`setup-terraform` action](https://github.com/hashicorp/setup-terraform) adds a wrapper for the `terraform` command that allows passing results to subsequent steps. This will prevent using the output of a `terraform` command as the input to another command in the same step.

The wrapper can be turned off by using

```yaml
steps:
- uses: hashicorp/setup-terraform@v4.0.1
  with:
    terraform_wrapper: false
```

## Testing workflows locally

The tool [`act`](https://github.com/nektos/act) can be used to test GitHub workflows locally. The default container [intentionally does not have feature parity](https://github.com/nektos/act#default-runners-are-intentionally-incomplete) with the containers used in GitHub due to the size of a full container.

The file `./actrc` configures `act` to use a fully-featured container.

## Running the static checker on workflows

Check your code for errors in syntax, usage, etc. using the following directive found in the `GNUMakefile` in this repository.

```console
% make gh-workflows-lint
```

## Go module caching

The repository has two independent Go modules with disjoint dependency graphs: the root module (`go.mod`/`go.sum`, the provider itself) and the `.ci/tools` module (`.ci/tools/go.mod`/`.ci/tools/go.sum`, which builds the CI tool binaries such as `tflint`, `terrafmt`, `misspell`, and `actionlint`). The `.ci/tools` module is cached in its own **tools lane**, keyed on `.ci/tools` files so it never shares a cache key with the root module.

### Tools lane

The tools lane caches the **installed tool binaries** (`$(go env GOPATH)/bin`) plus the **tools build cache** (`$(go env GOCACHE)`), keyed on `.ci/tools/go.mod` + `.ci/tools/go.sum`. On a cache hit, jobs skip `go install` entirely.

It follows a **single-writer, many-readers** model:

* **Writer**: the `tools_cache` job in [`provider.yml`](provider.yml) installs the full `.ci/tools` toolset and saves the cache. It writes only on the `main` branch (`if: github.ref == 'refs/heads/main'`), which keeps cache entries stable and prevents faster non-CI jobs from racing to save a poor entry.
* **Readers**: jobs that need a tool call the [`tools_cache` composite action](../actions/tools_cache/action.yml), which restores the cache **read-only** via [`actions/cache/restore`](https://github.com/actions/cache/tree/main/restore) (using `key: nonexistent` with a `restore-keys:` prefix, so it always falls back to the most recent entry written by the writer) and, only on a true cache miss, installs the requested tools as a fallback:

  ```yaml
  steps:
    - uses: actions/setup-go@v7
      with:
        go-version-file: .ci/tools/go.mod
        cache: false
    - uses: ./.github/actions/tools_cache
      with:
        tools: github.com/terraform-linters/tflint
  ```

Because each job runs its own caching (across many workflows, some in parallel), the built-in cache of [`actions/setup-go`](https://github.com/actions/setup-go) is disabled with `cache: false`; otherwise it would duplicate the cache and separately save an untrimmed build cache. This follows the setup-go [advanced usage guidance](https://github.com/actions/setup-go/blob/main/docs/advanced-usage.md#restore-only-caches) for multi-workflow / parallel builds.

> **Note:** `golangci-lint` is cached separately by `golangci-lint-action`'s own internal cache and is not part of the tools lane.
