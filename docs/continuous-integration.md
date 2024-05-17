# Continuous Integration

Continuous integration (CI) includes processes we run when you submit a pull request (PR). These processes can be divided into two broad categories: enrichment and testing.

## CI: Enrichment

The focus of this guide is on CI _testing_, but for completeness, we also mention enrichment processes. These include generating the [changelog](changelog-process.md) when a PR is merged, releasing new versions, labeling a PR based on its size and the AWS service it relates to, and adding automated comments to the PR.

These processes will not usually produce the dreaded red X signifying that a PR has failed CI. For this reason, we shift our focus to CI testing.

## CI: Testing Overview

To help place testing performed as part of CI in context, here is an overview of the Terraform AWS Provider's three types of tests.

1. [**Acceptance tests**](running-and-writing-acceptance-tests.md) are end-to-end evaluations of interactions with AWS. They validate functionalities like creating, reading, and destroying resources within AWS.
2. [**Unit tests**](unit-tests.md) focus on testing isolated units of code within the software, typically at the function level. They assess functionalities solely within the provider itself.
3. **Continuous integration tests** (_You are here!_) encompass a suite of automated tests that are executed on every pull request and include linting, compiling code, running unit tests, and performing static analysis.

## Rationale

Continuous integration plays a pivotal role in maintaining the health and quality of a large project like the Terraform AWS Provider. CI tests serve as a crucial component in automatically assessing code changes for compliance with project standards and functionality expectations, greatly reducing the review burden on maintainers. By executing a battery of tests upon each pull request submission, continuous integration ensures that new contributions integrate seamlessly with the existing codebase, minimizing the risk of regressions and enhancing overall stability.

Additionally, these tests provide rapid feedback to contributors, enabling them to identify and rectify issues early in the development cycle. In essence, continuous integration tests serve as a safeguard, bolstering the reliability and maintainability of this project while fostering a collaborative and iterative development environment.

## Specific Tests

There are many different tests and they change often. This guide doesn't cover everything CI does because, as noted above, many of the CI processes enrich the pull request, such as adding labels. If you notice something important that isn't reflected in this documentation, we welcome your contribution to fix it!

The makefile included with the Terraform AWS Provider allows you to run many of the CI tests locally before submitting your PR. The file is located in the provider's root directory and is called `GNUmakefile`. You should be able to use `make` with a variety of Linux-type shells that support `sh` and `bash`, such as a MacOS terminal.

**Note:** Many tests will simply exit without error if a test passed. "No news is good news."

### Acceptance Test Linting

Acceptance test linting involves thorough testing of the Terraform configuration associated with acceptance tests. Currently, this extracts configuration embedded as strings in Go files. However, as we move testing configurations to `.tf` files, this will involve testing those files for correctness.

Acceptance Test Linting has two components: `terrafmt` and `tflint`. `make` has several targets to help you with this.

Use the `tools` target before running Acceptance Test Linting:

```sh
% make tools
```

Use the `acctestlint` target to run all the Acceptance Test Linting checks using both `terrafmt` and `tflint`:

```sh
% make acctestlint
```

Limit any of the Acceptance Test Linting checks to a specific directory using `SVC_DIR`:

```sh
% SVC_DIR=./internal/service/rds make acctestlint
```

#### `terrafmt`

Use the `testacc-lint` target to run only the `terrafmt` test (`tflint` takes a long time to run):

```sh
% make testacc-lint
```

Use the `testacc-lint-fix` target to automatically fix issues found with `terrafmt`:

```sh
% make testacc-lint-fix
```

#### Validate Acceptance Tests (`tflint`)

Use the `testacc-tflint` target to run only the `tflint` test (`tflint` takes a long time to run):

```sh
% make testacc-tflint
```

### Copyright Checks

This CI check simply checks to make sure after running the tool, no files have been modified. No modifications signifies that everything already has the proper header.

Use the `tools` target before running Copyright Checks:

```sh
% make tools
```

Use the `copyright` target to add the appropriate copyright headers to all files:

```sh
% make copyright
```

### Dependency Checks

Dependency checks include a variety of tests including who should certain types of files. The test that generally trips people up is `go_mod`.

#### `go_mod`

Use the `depscheck` target to make sure that the Go mods files are tidy. This will also install the version of Go defined in the `.go-version` file in the root of the repository.

```sh
% make depscheck
```

### Preferred Library Version Check

Preferred Library Version Check doesn't cause CI to fail but will leave a comment.

#### `diffgrep`

This check verifies that preferred library versions are used in development of net-new resources. This is done by inspecting the pull request diff for any occurrence of a non-preferred library name, typically seen in an import block. At this time the only check is for AWS SDK for Go V1, but it may be extended in the future. This check will not fail if a non-preferred library version is detected, but will leave a comment on the pull request linking to the relevant contributor documentation.

Use the `preferredlib` target to check your changes against the `origin/main` of your Git repository (configurable using `BASE_REF`):

```sh
% make preferredlib
```

### ProviderLint Checks / providerlint

ProviderLint checks for a variety of best practices. For more details on specific checks and errors, see [providerlint](https://github.com/hashicorp/terraform-provider-aws/tree/main/.ci/providerlint).

Use the `providerlint` target to run the check just as it runs in CI:

```sh
% make providerlint
```

### Semgrep Checks

We use [Semgrep](https://github.com/semgrep/semgrep) for many types of checks and cannot describe all of them here. They are broken into rough groupings for parallel CI processing, as described below.

To locally run Semgrep checks using `make`, you'll need to install Semgrep locally. On MacOS, you can do this easily using Homebrew:

```sh
% brew install semgrep
```

#### Code Quality Scan

Use the `semcodequality` target to run the same check CI runs:

```sh
% make semcodequality
```

You can limit the scan to a service package by using the `PKG_NAME` environment variable:

```sh
% PKG=rds make semcodequality
```

#### Naming Scan Caps/AWS/EC2

Coming soon

#### Test Configs Scan

Coming soon

#### Service Name Scan A-Z

Coming soon

### YAML Linting / yamllint

YAMLlint checks the validity of YAML files.

To run YAMLlist locally using `make`, you'll need to install it locally. On MacOS, you can install it using Homebrew:

```sh
% brew install yamllint
```

Use the `yamllint` target to perform the check:

```sh
% make yamllint
```

### golangci-lint Checks

golangci-lint checks runs a variety of linters on the provider's code. This is done in two stages with the first stage acting as a gatekeeper since the second stage takes considerably longer to run.

Before running these checks locally, you need to install golangci-lint locally. This can be done in [several ways](https://golangci-lint.run/welcome/install/#local-installation) including using Homebrew on MacOS:

```sh
% brew install golangci-lint
```

Use the target `golangci-lint` to run both checks sequentially:

```sh
% make golangci-lint
```

You can limit the checks to a specific service package. For example:

```sh
% PKG=rds make golangci-lint
```

#### 1 of 2

Use the `golangci-lint1` target to run only the first step of these checks:

```sh
% make golangci-lint1
```

#### 2 of 2

Use the `golangci-lint2` target to run only the second step of these checks:

```sh
% make golangci-lint2
```

**Tip:** Running the second step against the entire codebase often takes the longest of all CI tests. If you're only working in one service package, you can save a lot of time limiting the scan to that service:

```sh
% PKG=rds make golangci-lint2
```

### GoReleaser CI / build-32-bit

GoReleaser CI build-32-bit ensures that GoReleaser can build a 32-bit binary. This check catches rare but important edge cases. Currently, we do not offer a `make` target to run this check locally.

### Provider Checks

Provider checks are a suite of tests that ensure Go code functions and markdown is correct.

#### go_build

This check determines if the code compiles correctly, there are syntax errors, or there are any unresolved references.

There are two ways to run this check that are basically equivalent.

Use the `go_build` target to build the provider using `go build`, leaving no binary in the `bin`:

```sh
% make go_build
```

Similarly, use the `build` target to install the provider binary locally using `go install`:

```sh
% make build
```

#### go_generate

`go_generate` checks to make sure nothing changes after you run the code generators. In other words, what the generators generate should be in sync with the committed code to avoid `make` telling you "bye, bye, bye."

Use the `gencheck` target to run the check:

```sh
% make gencheck
```

Use the `gen` target to run all the generators associated with the provider. Unless you're working on the generators or have inadvertently edited generated code, there should be no changes to the codebase after the generators finish:

```sh
% make gen
```

**NOTE:** While running the generators, you may see hundreds or thousands of code changes as `make` and the generators delete and recreate files.

#### go_test

`go_test` compiles the code and runs all tests except the [acceptance tests](running-and-writing-acceptance-tests.md). This check may also find higher level code errors than building alone finds.

Use the `test` target to run this test:

```sh
% make test
```

You can limit `test` to a single service package with the `PKG` environment variable:

```sh
% PKG=rds make test
```

**NOTE:** `test` and `golangci-lint2` are generally the longest running checks and, depending on your computer, may take considerable time to finish.

#### importlint

`importlint` uses [impi](https://github.com/pavius/impi) to make sure that imports in Go files follow the order of _standard_, _third party_, and then _local_. Besides neatness, enforcing the order helps avoid meaningless Git differences. In earlier days of Go, it was possible to order imports more freely. This check may no longer be needed but we need additional verification.

To run this check locally, you will need to install `impi`, which is done as part of `make tools`.

Use the `importlint` target to run `impi` with the appropriate parameters:

```sh
% make importlint
```

#### markdown-lint

`markdown-lint` can be a little confusing since it shows up in CI in three different contexts, each performing slightly different checks:

1. Provider Check / markdown-lint (this check)
2. Documentation Checks / markdown-lint
3. Website Checks / markdown-lint

This particular check uses [markdownlint](https://github.com/DavidAnson/markdownlint) to check all Markdown files in the provider except those in `docs` and `website/docs`, the CHANGELOG, and an example.

**NOTE:** You must have Docker installed to run this check.

Use the `provcheckmarkdownlint` target to run this test:

```sh
% make provcheckmarkdownlint
```

#### terraform_providers_schema

This process generates the Terraform AWS Provider schema for use by the `tfproviderdocs` check.

#### tfproviderdocs

Coming soon

#### validate_sweepers_unlinked

Coming soon
