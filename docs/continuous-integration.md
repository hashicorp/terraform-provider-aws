# Continuous Integration

Continuous integration (CI) includes processes that run when you submit a pull request (PR). These processes can be divided into two broad categories: enrichment and testing.

## CI: Enrichment

The focus of this guide is on CI _testing_, but for completeness, we also mention enrichment processes. These include generating the [changelog](changelog-process.md) when a PR is merged, releasing new versions, labeling a PR based on its size and the AWS service it relates to, and adding automated comments to the PR.

These processes will not usually produce the dreaded red X signifying that a PR has failed CI. For this reason, we shift our focus to CI testing.

## CI: Testing Overview

To help place testing performed as part of CI in context, here is an overview of the Terraform AWS Provider's three types of tests.

1. [**Acceptance tests**](running-and-writing-acceptance-tests.md) are end-to-end evaluations of interactions with AWS. They validate functionalities like creating, reading, and destroying resources within AWS.
2. [**Unit tests**](unit-tests.md) focus on testing isolated units of code within the software, typically at the function level. They assess functionalities solely within the provider itself.
3. **Continuous integration tests** (_You are here!_) encompass a suite of automated tests that are executed on every pull request and include linting, compiling code, running unit tests, and performing static analysis.

## Rationale

Continuous integration (CI) plays a pivotal role in maintaining the health and quality of a large project like the Terraform AWS Provider. CI tests are crucial for automatically assessing code changes for compliance with project standards and functionality expectations, greatly reducing the review burden on maintainers. By executing a battery of tests upon each pull request submission, CI ensures that new contributions integrate seamlessly with the existing codebase, minimizing the risk of regressions and enhancing overall stability.

Additionally, these tests provide rapid feedback to contributors, enabling them to identify and rectify issues early in the development cycle. In essence, CI tests serve as a safeguard, bolstering the reliability and maintainability of the project while fostering a collaborative and iterative development environment.

## Using `make` to Run Specific Tests Locally

**NOTE:** We've made a great effort to ensure that tests running on GitHub have a close-as-possible equivalent in the Makefile. If you notice a difference, please [open an issue](https://github.com/hashicorp/terraform-provider-aws/issues/new/choose) to let us know.

The Makefile included with the Terraform AWS Provider allows you to run many of the CI tests locally before submitting your PR. The file is located in the provider's root directory and is called `GNUmakefile`. You should be able to use `make` with a variety of Linux-type shells that support `bash`, such as a macOS terminal.

**NOTE:** See the [Makefile Cheat Sheet](makefile-cheat-sheet.md) for detailed information about the Makefile.

There are many different tests, and they change often. This guide doesn't cover everything CI does because, as noted above, many of the CI processes enrich the pull request, such as adding labels. If you notice something important that isn't reflected in this documentation, let us know!

**NOTE:** Many tests simply exit without error if passing. "No news is good news."

### Before Running Tests

CI tests run on GitHub when you submit a pull request. However, these tests can take a while to complete. If you prefer, you can run most tests locally. Before running tests locally, you need to clone the repository, which you've likely already done if you're working on a PR, and install the necessary tools.

Use the `tools` target to install a variety of tools used by different CI tests:

```console
make tools
```

### Running All Available CI Tests

Use the `ci` target to run all the tests listed below:

```console
make ci
```

**NOTE:** Depending on your machine, running all the tests can take a long time!

To run most of the tests but exclude the longer-running ones, use the `ci-quick` target. "Quick" may not be _quick_ precisely, but relative to the full `ci` target, it is _quicker_:

```console
make ci-quick
```

Use the `clean-make-tests` target to clean up artifacts left behind by `make` tests, although they should be ignored by Git:

```console
make clean-make-tests
```

### Acceptance Test Linting

Acceptance test linting involves thoroughly testing the Terraform configuration associated with acceptance tests. Currently, this process extracts configuration embedded as strings in Go files. However, as we move testing configurations to `.tf` files, linting will involve testing those files for correctness.

Acceptance test linting has two components: `terrafmt` and `tflint`. The `make` tool provides several targets to help with this.

#### Running All Acceptance Test Linting Checks

Use the `acctest-lint` target to run all the acceptance test linting checks using both `terrafmt` and `tflint`:

```console
make acctest-lint
```

#### Limiting Linting to a Specific Service Package

You can limit the test to a service package by using the `PKG` environment variable:

```console
PKG=rds make acctest-lint
```

The command above is equivalent to using `SVC_DIR` with the full relative path:

```console
SVC_DIR=./internal/service/rds make acctest-lint
```

#### `terrafmt`

Use the `testacc-lint` target to run only the `terrafmt` test. This is useful if you want to skip `tflint`, which takes a long time to run:

```console
make testacc-lint
```

Use the `testacc-lint-fix` target to automatically fix issues found by `terrafmt`:

```console
make testacc-lint-fix
```

#### Validate Acceptance Tests with `tflint`

Use the `testacc-tflint` target to run only the `tflint` test. This is useful if you want to skip `terrafmt`:

```console
make testacc-tflint
```

### Copyright Checks / add headers check

This CI check simply checks to make sure after running the tool, no files have been modified. No modifications signifies that everything already has the proper header.

Use the `copyright` target to add the appropriate copyright headers to all files:

```console
make copyright
```

**NOTE:** Install [tools](#before-running-tests) before running this check.

### Dependency Checks / go_mod

Dependency checks include a variety of tests including who should edit certain types of files. The test that generally trips people up is `go_mod`.

Use the `deps-check` target to make sure that the Go mods files are tidy. This will also install the version of Go defined in the `.go-version` file in the root of the repository.

```console
make deps-check
```

### Documentation Checks

"Documentation" is the context of these checks is the documentation found in the `docs/` directory of the provider. This include contributor and related guides. This is developer-facing unlike the [Examples](#examples-checks) and [Website](#website-checks) checks.

#### markdown-link-check

Use the target `docs-link-check` to check links found in the contributor documentation:

```console
make docs-link-check
```

**NOTE:** Install [Docker](https://docs.docker.com/desktop/install/mac-install/) to run this check.

#### markdown-lint

Use the target `docs-markdown-lint` to lint the contributor documentation:

```console
make docs-markdown-lint
```

**NOTE:** Install [Docker](https://docs.docker.com/desktop/install/mac-install/) to run this check.

#### misspell

Use the target `docs-misspell` to spellcheck the contributor documentation:

```console
make docs-misspell
```

**NOTE:** Install [tools](#before-running-tests) before running this check.

### Examples Checks

These checks help ensure that examples included with the provider are correct.

#### tflint

Use the target `examples-tflint` to lint the examples:

```console
make examples-tflint
```

**NOTE:** Install [tools](#before-running-tests) before running this check.

#### validate-terraform (0.12.31)

This check is not currently available in the Makefile.

#### validate-terraform (1.0.6)

This check is not currently available in the Makefile.

### golangci-lint Checks

golangci-lint checks runs a variety of linters on the provider's code. This is done in two stages with the first stage acting as a gatekeeper since the second stage takes considerably longer to run.

Before running these checks locally, you need to install golangci-lint locally. This can be done in [several ways](https://golangci-lint.run/welcome/install/#local-installation) including using Homebrew on macOS:

```console
brew install golangci-lint
```

Use the target `golangci-lint` to run both checks sequentially:

```console
make golangci-lint
```

You can limit the checks to a specific service package. For example:

```console
PKG=rds make golangci-lint
```

#### 1 of 2

Use the `golangci-lint1` target to run only the first step of these checks:

```console
make golangci-lint1
```

#### 2 of 2

Use the `golangci-lint2` target to run only the second step of these checks:

```console
make golangci-lint2
```

**Tip:** Running the second step against the entire codebase often takes the longest of all CI tests. If you're only working in one service package, you can save a lot of time limiting the scan to that service:

```console
PKG=rds make golangci-lint2
```

### GoReleaser CI / build-32-bit

GoReleaser CI build-32-bit ensures that GoReleaser can build a 32-bit binary. This check catches rare but important edge cases. Currently, we do not offer a `make` target to run this check locally.

### Preferred Library Version Check / `diffgrep`

The Preferred Library Version Check doesn't cause CI to fail but will leave a comment on the pull request.

This check verifies that preferred library versions are used in the development of new resources. It inspects the pull request diff for any occurrence of a non-preferred library name, typically seen in an import block. Currently, the only check is for AWS SDK for Go V1, but this may be extended in the future. If a non-preferred library version is detected, the check will not fail but will leave a comment on the pull request linking to the relevant contributor documentation.

Use the `preferred-lib` target to check your changes against the `origin/main` branch of your Git repository (configurable using `BASE_REF`):

```console
make preferred-lib
```

### Provider Checks

Provider checks are a suite of tests that ensure Go code functions and markdown is correct.

#### go-build

This check determines if the code compiles correctly, there are syntax errors, or there are any unresolved references.

There are two ways to run this check that are basically equivalent.

Use the `go-build` target to build the provider using `go build`, installing the provider in the `terraform-plugin-dir` directory:

```console
make go-build
```

Similarly, use the `build` target to install the provider binary locally using `go install`:

```console
make build
```

#### go_generate

`go_generate` checks to make sure nothing changes after you run the code generators. In other words, generated code and committed code should be in sync or we'll say, "bye bye bye."

Use the `gen-check` target to run the check:

```console
make gen-check
```

Use the `gen` target to run all the generators associated with the provider. Unless you're working on the generators or have inadvertently edited generated code, there should be no changes to the codebase after the generators finish:

```console
make gen
```

**NOTE:** While running the generators, you may see hundreds or thousands of code changes as `make` and the generators delete and recreate files.

#### go_test

`go_test` compiles the code and runs all tests except the [acceptance tests](running-and-writing-acceptance-tests.md). This check may also find higher level code errors than building alone finds.

Use the `test` target to run this test:

```console
make test
```

You can limit `test` to a single service package with the `PKG` environment variable:

```console
PKG=rds make test
```

**NOTE:** `test` and `golangci-lint2` are generally the longest running checks and, depending on your computer, may take considerable time to finish.

#### import-lint

`import-lint` uses [impi](https://github.com/pavius/impi) to make sure that imports in Go files follow the order of _standard_, _third party_, and then _local_. Besides neatness, enforcing the order helps avoid meaningless Git differences. In earlier days of Go, it was possible to order imports more freely. This check may no longer be needed but we need additional verification.

To run this check locally, you will need to install `impi`, which is done as part of `make tools`.

Use the `import-lint` target to run `impi` with the appropriate parameters:

```console
make import-lint
```

#### markdown-lint

`markdown-lint` can be a little confusing since it shows up in CI in three different contexts, each performing slightly different checks:

1. Provider Check / markdown-lint (this check)
2. Documentation Checks / markdown-lint
3. Website Checks / markdown-lint

This particular check uses [markdownlint](https://github.com/DavidAnson/markdownlint) to check all Markdown files in the provider except those in `docs` and `website/docs`, the CHANGELOG, and an example.

Use the `provider-markdown-lint` target to run this test:

```console
make provider-markdown-lint
```

**NOTE:** Install [Docker](https://docs.docker.com/desktop/install/mac-install/) to run this check.

#### misspell

Use `go-misspell` to check the provider code for misspellings:

```console
make go-misspell
```

**NOTE:** Install [tools](#before-running-tests) before running this check.

#### terraform providers schema

This process generates the Terraform AWS Provider schema for use by the `tfproviderdocs` check. In the `make` file, this is done as part of the `tfproviderdocs` target test.

#### tfproviderdocs

**NOTE:** To run this test, you need Terraform installed locally. On macOS, you can use Homebrew to install Terraform:

```console
brew install terraform
```

This test builds the provider binary, loads the provider with Terraform, generates the provider schema, and then uses the tfproviderdocs tool to ensure the provider (via the schema) and documentation are consistent with each other.

Use the `tfproviderdocs` target to run this test:

```console
make tfproviderdocs
```

#### Sweeper Functions Not Linked

This check builds the Terraform AWS Provider in two different configurations, with sweepers and without, to make sure sweepers are properly included or excluded from the builds. The normal build you would receive from the Terraform Registry does not include sweepers and this ensures they aren't accidentally included.

Use the `sweeper-check` target to run both tests:

```console
make sweeper-check
```

You can also run the checks separately.

Use the `sweeper-linked` target to ensure sweeper are included in a sweeper build:

```console
make sweeper-linked
```

Use the `sweeper-unlinked` target to ensure sweeper are not included in a normal build:

```console
make sweeper-unlinked
```

### ProviderLint Checks / providerlint

ProviderLint checks for a variety of best practices. For more details on specific checks and errors, see [providerlint](https://github.com/hashicorp/terraform-provider-aws/tree/main/.ci/providerlint).

Use the `provider-lint` target to run the check just as it runs in CI:

```console
make provider-lint
```

### Semgrep Checks

We use [Semgrep](https://github.com/semgrep/semgrep) for many types of checks and cannot describe all of them here. They are broken into rough groupings for parallel CI processing, as described below.

To locally run Semgrep checks using `make`, you'll need to install Semgrep locally. On macOS, you can do this easily using Homebrew:

```console
brew install semgrep
```

#### Code Quality Scan

This scan looks for a hodgepodge of issues, best practices, and problems we've found over the years.

Use the `semgrep-code-quality` target to run the same check CI runs:

```console
make semgrep-code-quality
```

You can limit the scan to a service package by using the `PKG` environment variable:

```console
PKG=rds make semgrep-code-quality
```

#### Naming Scan Caps/AWS/EC2

Idiomatic Go uses [_mixed caps_](naming.md#mixed_caps) for multiword names, not camel case. In camel case, a name with the words "SMTP thing" would be `SmtpThing`. This is wrong in Go. In mixed caps, and therefore idiomatic Go, `SMTPThing` is correct. This scan ensures that many acronyms and initialisms are capitalized correctly in code.

Use the `semgrep-naming-cae` target to run the same check CI runs:

```console
make semgrep-naming-cae
```

You can limit the scan to a service package by using the `PKG` environment variable:

```console
PKG=rds make semgrep-naming-cae
```

#### Service Name Scan A-Z

This scan ensures that AWS service names are used fairly consistently from one service package to the next.

Use the `semgrep-service-naming` target to run the same check CI runs:

```console
make semgrep-service-naming
```

You can limit the scan to a service package by using the `PKG` environment variable:

```console
PKG=rds make semgrep-service-naming
```

#### Test Configs Scan

This scan checks for consistency in naming of test-related functions.

Use the `semgrep-naming` target to run the same check CI runs:

```console
make semgrep-naming
```

You can limit the scan to a service package by using the `PKG` environment variable:

```console
PKG=rds make semgrep-naming
```

### Skaff Checks / Compile skaff

Use the `skaff-check-compile` target to test building Skaff:

```console
make skaff-check-compile
```

### Website Checks

These checks help ensure that user-facing documentation on the website is correct.

#### markdown-link-check-a-h-markdown

Use the target `website-link-check-markdown` to check links found in the website:

```console
make website-link-check-markdown
```

**NOTE:** Install [Docker](https://docs.docker.com/desktop/install/mac-install/) to run this check.

#### markdown-link-check-i-z-markdown

This range is also checked as part of the "a-h" check above.

#### markdown-link-check-md

Use the target `website-link-check-md` to check links found in the website:

```console
make website-link-check-md
```

**NOTE:** Install [Docker](https://docs.docker.com/desktop/install/mac-install/) to run this check.

#### markdown-lint

Use the target `website-markdown-lint` to lint the website documentation:

```console
make website-markdown-lint
```

**NOTE:** Install [Docker](https://docs.docker.com/desktop/install/mac-install/) to run this check.

#### misspell

Use the target `website-misspell` to spellcheck the documentation:

```console
make website-misspell
```

**NOTE:** Install [tools](#before-running-tests) before running this check.

#### terrafmt

Use the target `website-terrafmt` to check formatting of Terraform configuration in documentation:

```console
make website-terrafmt
```

**NOTE:** Install [tools](#before-running-tests) before running this check.

#### tflint

Use the target `website-tflint` to check formatting of Terraform configuration in documentation:

```console
make website-tflint
```

**NOTE:** Install [tools](#before-running-tests) before running this check.

### Workflow Linting / actionlint

Use the `gh-workflow-lint` target to perform the check:

```console
make gh-workflow-lint
```

**NOTE:** Install [tools](#before-running-tests) before running this check.

### YAML Linting / yamllint

YAMLlint checks the validity of YAML files.

To run YAMLlint locally using `make`, you'll need to install it locally. On macOS, you can install it using Homebrew:

```console
brew install yamllint
```

Use the `yamllint` target to perform the check:

```console
make yamllint
```
