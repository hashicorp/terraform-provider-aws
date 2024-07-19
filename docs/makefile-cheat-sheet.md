# Makefile Cheat Sheet

The Terraform AWS Provider Makefile includes a lot of functionality to make working on the provider easier and more efficient. Many contributors are familiar with using the Makefile for running acceptance tests, but there is a lot more functionality hidden in this humble file.

!!! tip
    See [Continuous Integration](continuous-integration.md) for more information about the CI-focused parts of the Makefile.

## Basics

If you're new to our Makefile, this section will bring you up to speed.

### Location

The Makefile is located in the root of the provider repository and is called [GNUmakefile](https://github.com/hashicorp/terraform-provider-aws/blob/main/GNUmakefile).

### Phony Targets

Historically, Makefiles were used to help with the complexities of compiling and linking software projects, managing dependencies, and enabling the creation of various _target_ files. `make` would create a "target" file as determined by the command line:

```console
% make <target>
```

Today, we use _phony_ targets in the Makefile to automate many tasks. "Phony" simply means that the target doesn't define the file we're trying to make but rather the recipe we want to perform, such as running tests.

For example, `testacc` is a phony target to simplify the command for running acceptance tests:

```console
make testacc TESTS=TestAccIAMRole_basic PKG=iam
```

### Meta Targets and Dependent Targets

_Meta_ targets are `make` targets that only run other targets. They aggregate the functionality of other targets for convenience. In the [Cheat Sheet](#cheat-sheet), meta targets are marked with <sup>M</sup>.

_Dependent_ targets also run other targets but, in addition, have their own functionality. A dependent target generally runs the other targets first before executing its own functionality. In the [Cheat Sheet](#cheat-sheet), dependent targets are marked with <sup>D</sup>.

For example, in the cheat sheet, the `ci`, `clean`, and `misspell` targets are meta targets that only run other targets.

On the other hand, examples of dependent targets are `deps-check`, `gen-check`, and `semgrep-code-quality`.

When you call a meta or dependent target and it runs other targets, those targets must complete successfully in order for the target you called to succeed.

## Variables

In the [Cheat Sheet](#cheat-sheet), you can see which variables affect which [targets](#phony-targets). This section describes the variables in more detail.

Variables are often defined before the `make` call on the same line, such as `MY_VAR=42 make my-target`. However, they can also be set on the same line _after_ the `make` call or in your environment, using, for example, `export MY_VAR=42`.

!!! note
    Targets that [meta and dependent targets](#meta-targets-and-dependent-targets) run may not all respect the same set of variables.

* `ACCTEST_PARALLELISM` - (Default: `20`) Number of concurrent acceptance tests to run. Overridden if `P` is set.
* `ACCTEST_TIMEOUT` - (Default: `360m`) Timeout before acceptance tests panic.
* `AWSSDKPATCH_OPTS` - (Default: _None_) See the [awssdkpatch tool](https://github.com/hashicorp/terraform-provider-aws/tree/main/tools/awssdkpatch) for more information.
* `BASE_REF` - (Default: `main`) Origin reference to use for Git `diff` comparison, as in `origin/BASE_REF`.
* `CURDIR` - (Default: Value of `$PWD`) Root path to use for `/.ci/scripts/`.
* `GO_VER` - (Default: Value in `.go-version` file) Version of Go to use. To use the default version on your system, use `GO_VER=go`.
* `K` - (Default: _None_) Name of the service package you want to use, such as `ec2`, `iam`, or `lambda`, limiting Go processing to that package and dependencies. Equivalent to `PKG` variable. Assigns values to `PKG_NAME`, `SVC_DIR`, and `TEST` overridding any values set.
* `P` - (Default: `20`) Number of concurrent acceptance tests to run. Assigns a value to `ACCTEST_PARALLELISM` overridding any value set.
* `PKG` - (Default: _None_) Name of the service package you want to use, such as `ec2`, `iam`, or `lambda`, limiting Go processing to that package and dependencies. Equivalent to `K` variable. Assigns values to `PKG_NAME`, `SVC_DIR`, and `TEST` overridding any values set.
* `PKG_NAME` - (Default: `internal`) Subdirectory (Go package) to use as the basis for Go processing. Overridden if `PKG` or `K` is set.
* `RUNARGS` - (Default: _None_) Raw arguments passed to Go when running acceptance tests. For example, `RUNARGS=-run=TestMyTest`. Overridden if `TESTS` or `T` is set.
* `SEMGREP_ARGS` - (Default: `--error`) Semgrep arguments. See the [Semgrep reference](https://semgrep.dev/docs/cli-reference#semgrep-scan-command-options).
* `SEMGREP_ENABLE_VERSION_CHECK` - (Default: `false`) Whether to check Semgrep servers to verify you are running the latest Semgrep version.
* `SEMGREP_SEND_METRICS` - (Default: `off`) When Semgrep usage metrics are sent to Semgrep.
* `SEMGREP_TIMEOUT` - (Default: `900`) Maximum time to spend running a rule on a single file, in seconds.
* `SVC_DIR` - (Default: `./internal/service`) Directory to as the base for recursive processing. Overridden if `PKG` or `K` is set.
* `SWEEP_DIR` - (Default: `./internal/sweep`) Location of the sweep directory.
* `SWEEP` - (Default: `us-west-2,us-east-1,us-east-2,us-west-1`) Comma-separated list of AWS regions to sweep.
* `SWEEP_TIMEOUT` - (Default: `360m`) Time Go will spend sweeping resources before panicking.
* `SWEEPARGS` - (Default: _None_) Raw arguments that define what to sweep, including dependencies. Similar to `SWEEPERS`. For example, `SWEEPARGS=-sweep-run=aws_example_thing`.
* `SWEEPERS` - (Default: _None_) Resources to sweep, including dependencies. Similar to `SWEEPARGS`. For example, `SWEEPERS=aws_example_thing`. Assigns a value to `SWEEPARGS` overridding any value set.
* `T` - (Default: _None_) Names of tests to run. Equivalent to `TESTS`. Assigns a value to `RUNARGS` overridding any value set.
* `TEST` - (Default: `./...`) Limit tests to this directory and dependencies. Overridden if `PKG` or `K` is set.
* `TEST_COUNT` - (Default: `1`) Number of times to run each acceptance or unit test.
* `TESTS` - (Default: _None_) Names of tests to run. Equivalent to `T`. Assigns a value to `RUNARGS` overridding any value set.
* `TESTARGS` - (Default: _None_) Raw arguments passed to Go when running tests. Unlike `RUNARGS`, this is _not_ overridden if `TESTS` or `T` is set.

## Cheat Sheet

* **Target**: Use as a subcommand to `make`, such as `make gen`. [Meta and dependent targets](#meta-targets-and-dependent-targets) are marked with <sup>M</sup> and <sup>D</sup>, respectively.
* **Description**: When CI-related, this aligns with the name of the check as seen on GitHub.
* **CI?**: Indicates whether the target is equivalent or largely equivalent to a check run on the GitHub repository for a pull request. See [continuous integration](continuous-integration.md) for more details.
* **Legacy?**: Indicates whether the target is a legacy holdover. Use caution with a legacy target! It may not work, or it may perform checks or fixes that do _not_ align with current practices. In the future, this target should be removed, modernized, or verified to still have value.
* **Vars**: [Variables](#variables) that you can set when using the target, such as `MY_VAR=42 make my-target`. [Meta and dependent targets](#meta-targets-and-dependent-targets) run other targets that may not respect the same variables.

!!! tip
    Makefile autocompletion works out of the box on Zsh (the default shell for Terminal on macOS) and Fish shells. For Bash, the `bash-completion` package, among others, provides Makefile autocompletion. Using autocompletion allows you, for example, to type `make ac`, press _tab_, and the shell autocompletes `make acctest-lint`.

| Target | Description | CI? | Legacy? | Vars |
| --- | --- | --- | --- | --- |
| `acctest-lint`<sup>M</sup> | Run all CI acceptance test checks | ✔️ |  | `K`, `PKG`, `SVC_DIR` |
| `awssdkpatch`<sup>D</sup> | Install [awssdkpatch](https://github.com/hashicorp/terraform-provider-aws/tree/main/tools/awssdkpatch) |  |  | `GO_VER` |
| `awssdkpatch-apply`<sup>D</sup> | Apply a patch generated with [awssdkpatch](https://github.com/hashicorp/terraform-provider-aws/tree/main/tools/awssdkpatch) |  |  | `AWSSDKPATCH_OPTS`, `GO_VER`, `K`, `PKG`, `PKG_NAME` |
| `awssdkpatch-gen`<sup>D</sup> | Generate a patch file using [awssdkpatch](https://github.com/hashicorp/terraform-provider-aws/tree/main/tools/awssdkpatch) |  |  | `AWSSDKPATCH_OPTS`, `GO_VER`, `K`, `PKG`, `PKG_NAME` |
| `build`<sup>D</sup> | Build the provider |  |  | `GO_VER` |
| `changelog-misspell` | CHANGELOG Misspell / misspell | ✔️ |  |  |
| `ci`<sup>M</sup> | Run all CI checks | ✔️ |  | `BASE_REF`, `GO_VER`, `K`, `PKG`, `SEMGREP_ARGS`, `SVC_DIR`, `TEST`, `TESTARGS` |
| `ci-quick`<sup>M</sup> | Run quicker CI checks | ✔️ |  | `BASE_REF`, `GO_VER`, `K`, `PKG`, `SEMGREP_ARGS`, `SVC_DIR`, `TEST`, `TESTARGS` |
| `clean`<sup>M</sup> | Clean up Go cache, tidy and re-install tools |  |  | `GO_VER` |
| `clean-go`<sup>D</sup> | Clean up Go cache |  |  | `GO_VER` |
| `clean-make-tests` | Clean up artifacts from make tests |  |  |  |
| `clean-tidy`<sup>D</sup> | Clean up tidy |  |  | `GO_VER` |
| `copyright` | Copyright Checks / add headers check | ✔️ |  |  |
| _default_ | = `build` |  |  | `GO_VER` |
| `deps-check`<sup>D</sup> | Dependency Checks / go_mod | ✔️ |  | `GO_VER` |
| `docs`<sup>M</sup> | Run all CI documentation checks | ✔️ |  |  |
| `docs-check` | Check provider documentation |  | ✔️ |  |
| `docs-link-check` | Documentation Checks / markdown-link-check | ✔️ |  |  |
| `docs-lint` | Lint documentation |  | ✔️ |  |
| `docs-lint-fix` | Fix documentation linter findings |  | ✔️ |  |
| `docs-markdown-lint` | Documentation Checks / markdown-lint | ✔️ |  |  |
| `docs-misspell` | Documentation Checks / misspell | ✔️ |  |  |
| `examples-tflint` | Examples Checks / tflint | ✔️ |  |  |
| `fix-constants`<sup>M</sup> | Use Semgrep to fix constants |  |  | `K`, `PKG`, `PKG_NAME`, `SEMGREP_ARGS` |
| `fix-imports` | Fixing source code imports with goimports |  |  |  |
| `fmt` | Fix Go source formatting |  |  | `K`, `PKG`, `PKG_NAME` |
| `fmt-check` | Verify Go source is formatted |  |  | `CURDIR` |
| `fumpt` | Run gofumpt |  |  | `K`, `PKG`, `PKG_NAME` |
| `gen`<sup>D</sup> | Run all Go generators |  |  | `GO_VER` |
| `gen-check`<sup>D</sup> | Provider Checks / go_generate | ✔️ |  |  |
| `generate-changelog` | Generate changelog |  |  | `CURDIR` |
| `gh-workflow-lint` | Workflow Linting / actionlint | ✔️ |  |  |
| `go-build` | Provider Checks / go-build | ✔️ |  |  |
| `go-misspell` | Provider Checks / misspell | ✔️ |  |  |
| `golangci-lint`<sup>M</sup> | All golangci-lint Checks | ✔️ |  | `K`, `PKG`, `TEST` |
| `golangci-lint1` | golangci-lint Checks / 1 of 2 | ✔️ |  | `K`, `PKG`, `TEST` |
| `golangci-lint2` | golangci-lint Checks / 2 of 2 | ✔️ |  | `K`, `PKG`, `TEST` |
| `help` | Display help |  |  |  |
| `import-lint` | Provider Checks / import-lint | ✔️ |  | `K`, `PKG`, `TEST` |
| `install`<sup>M</sup> | = `build` |  |  | `GO_VER` |
| `lint`<sup>M</sup> | Legacy target, use caution |  | ✔️ |  |
| `lint-fix`<sup>M</sup> | Fix acceptance test, website, and docs linter findings |  | ✔️ |  |
| `misspell`<sup>M</sup> | Run all CI misspell checks | ✔️ |  |  |
| `preferred-lib` | Preferred Library Version Check / diffgrep | ✔️ |  | `BASE_REF` |
| `prereq-go` | Install the project's Go version |  |  | `GO_VER` |
| `provider-lint` | ProviderLint Checks / providerlint | ✔️ |  | `K`, `PKG`, `SVC_DIR` |
| `provider-markdown-lint` | Provider Check / markdown-lint | ✔️ |  |  |
| `sane`<sup>D</sup> | Run sane check |  |  | `ACCTEST_PARALLELISM`, `ACCTEST_TIMEOUT`, `GO_VER`, `TEST_COUNT` |
| `sanity`<sup>D</sup> | Run sanity check (failures allowed) |  |  | `ACCTEST_PARALLELISM`, `ACCTEST_TIMEOUT`, `GO_VER`, `TEST_COUNT` |
| `semgrep`<sup>M</sup> | Run all CI Semgrep checks | ✔️ |  | `K`, `PKG`, `PKG_NAME`, `SEMGREP_ARGS` |
| `semgrep-all`<sup>D</sup> | Run semgrep on all files |  |  | `K`, `PKG`, `PKG_NAME`, `SEMGREP_ARGS` |
| `semgrep-code-quality`<sup>D</sup> | Semgrep Checks / Code Quality Scan | ✔️ |  | `K`, `PKG`, `PKG_NAME`, `SEMGREP_ARGS` |
| `semgrep-constants`<sup>D</sup> | Fix constants with Semgrep --autofix |  |  | `K`, `PKG`, `PKG_NAME`, `SEMGREP_ARGS` |
| `semgrep-docker`<sup>D</sup> | Run Semgrep |  | ✔️ |  |
| `semgrep-fix`<sup>D</sup> | Fix Semgrep issues that have fixes |  |  | `K`, `PKG`, `PKG_NAME`, `SEMGREP_ARGS` |
| `semgrep-naming`<sup>D</sup> | Semgrep Checks / Test Configs Scan | ✔️ |  | `K`, `PKG`, `PKG_NAME`, `SEMGREP_ARGS` |
| `semgrep-naming-cae`<sup>D</sup> | Semgrep Checks / Naming Scan Caps/`AWS`/EC2 | ✔️ |  | `K`, `PKG`, `PKG_NAME`, `SEMGREP_ARGS` |
| `semgrep-service-naming`<sup>D</sup> | Semgrep Checks / Service Name Scan A-Z | ✔️ |  | `K`, `PKG`, `PKG_NAME`, `SEMGREP_ARGS` |
| `semgrep-validate` | Validate Semgrep configuration files |  |  |  |
| `skaff`<sup>D</sup> | Install skaff |  |  | `GO_VER` |
| `skaff-check-compile` | Skaff Checks / Compile skaff | ✔️ |  |  |
| `sweep`<sup>D</sup> | Run sweepers |  |  | `GO_VER`, `SWEEP_DIR`, `SWEEP_TIMEOUT`, `SWEEP`, `SWEEPARGS` |
| `sweeper`<sup>D</sup> | Run sweepers with failures allowed |  |  | `GO_VER`, `SWEEP_DIR`, `SWEEP_TIMEOUT`, `SWEEP` |
| `sweeper-check`<sup>M</sup> | Provider Checks / Sweeper Linked, Unlinked | ✔️ |  |  |
| `sweeper-linked` | Provider Checks / Sweeper Functions Linked | ✔️ |  |  |
| `sweeper-unlinked`<sup>D</sup> | Provider Checks / Sweeper Functions Not Linked | ✔️ |  |  |
| `t`<sup>D</sup> | Run acceptance tests  (similar to `testacc`) |  |  | `ACCTEST_PARALLELISM`, `ACCTEST_TIMEOUT`, `GO_VER`, `K`, `PKG`, `PKG_NAME`, `RUNARGS`, `TEST_COUNT`, `TESTARGS` |
| `test`<sup>D</sup> | Run unit tests |  |  | `GO_VER`, `K`, `PKG`, `TEST`, `TESTARGS` |
| `test-compile`<sup>D</sup> | Test package compilation |  |  | `GO_VER`, `K`, `PKG`, `PKG_NAME`, `TEST`, `TESTARGS` |
| `testacc`<sup>D</sup> | Run acceptance tests |  |  | `ACCTEST_PARALLELISM`, `ACCTEST_TIMEOUT`, `GO_VER`, `K`, `PKG`, `PKG_NAME`, `RUNARGS`, `TEST_COUNT`, `TESTARGS` |
| `testacc-lint` | Acceptance Test Linting / terrafmt | ✔️ |  | `K`, `PKG`, `SVC_DIR` |
| `testacc-lint-fix` | Fix acceptance test linter findings |  |  | `K`, `PKG`, `SVC_DIR` |
| `testacc-short`<sup>D</sup> | Run acceptace tests with the -short flag |  |  | `ACCTEST_PARALLELISM`, `ACCTEST_TIMEOUT`, `GO_VER`, `K`, `PKG`, `PKG_NAME`, `RUNARGS`, `TEST_COUNT`, `TESTARGS` |
| `testacc-tflint` | Acceptance Test Linting / tflint | ✔️ |  | `K`, `PKG`, `SVC_DIR` |
| `tfproviderdocs`<sup>D</sup> | Provider Checks / tfproviderdocs | ✔️ |  |  |
| `tfsdk2fw`<sup>D</sup> | Install tfsdk2fw |  |  | `GO_VER` |
| `tools`<sup>D</sup> | Install tools |  |  | `GO_VER` |
| `ts`<sup>M</sup> | Alias to `testacc-short` |  |  |  |
| `website`<sup>M</sup> | Run all CI website checks | ✔️ |  |  |
| `website-link-check` | Check website links |  | ✔️ |  |
| `website-link-check-ghrc` | Check website links with ghrc |  | ✔️ |  |
| `website-link-check-markdown` | Website Checks / markdown-link-check-a-z-markdown | ✔️ |  |  |
| `website-link-check-md` | Website Checks / markdown-link-check-md | ✔️ |  |  |
| `website-lint` | Lint website files |  | ✔️ |  |
| `website-lint-fix` | Fix website linter findings |  | ✔️ |  |
| `website-markdown-lint` | Website Checks / markdown-lint | ✔️ |  |  |
| `website-misspell` | Website Checks / misspell | ✔️ |  |  |
| `website-terrafmt` | Website Checks / terrafmt | ✔️ |  |  |
| `website-tflint` | Website Checks / tflint | ✔️ |  |  |
| `yamllint` | `YAML` Linting / yamllint | ✔️ |  |  |
