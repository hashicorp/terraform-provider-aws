<!-- Copyright IBM Corp. 2014, 2026 -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

# AGENTS.md

This file provides guidance to AI coding agents when working with code in this repository.

## Repository Overview

This is the Go-based Terraform AWS Provider (`github.com/hashicorp/terraform-provider-aws`). It maps AWS API resources to Terraform resources, data sources, ephemeral resources and actions (collectively often referred to as just resources). The primary language is Go; HCL appears in acceptance test configurations and website documentation.

---

## Agent Registry

This project uses specialized personas for different tasks.

### Available Personas
- **`@contributor`**: [Contributor Persona](./.agents/contributor.md) - Contributes code in the form of bugfixes, enhancements to existing resources, and new resources. Makes clarifications and corrections to existing documentation.
- **`@maintainer`**: [Maintainer Persona](./.agents/maintainer.md) - Steward of the project, responsible for both internal and external quality. Reviews contributions. Maintains provider-level features, including new Terraform language constructs.
- **`@tcm`**: [TCM Persona](./.agents/tcm.md) - Triages incoming GitHub issues and PRs. Engages with community members to answer technical and process questions. Suggests workarounds and alternatives to reported bugs.

### Global Rules
- Always use the requested persona for tasks.
- If no persona is specified, default to `@contributor`.
- A persona defines a role with a perspective and responsibilities.
- Personas may invoke skills.

---

## Skills

Skills are loaded from `./.agents/skills`. Each skill supplies step-by-step instructions, code patterns, and guardrails for a specific task.

| Skill | Task |
|---|---|
| [changelog](./.agents/skills/changelog/SKILL.md) | Add a `.changelog/<PR_NUMBER>.txt` entry from a PR URL, commit, and push (with confirmation). |

---

## Global Rules

### Non-negotiables
- Verification is a hard exit criterion for every task. Without it, the task is not done.
- Prefer the boring, obvious solution.
- Touch only what you’re asked to touch.

### AI Usage Policy

Per `docs/ai-usage.md`:

- Disclose AI use in the PR description.
- Include `🤖🤖🤖` in the PR title if an LLM agent is directly involved in submitting it.
- The human PR author is fully responsible for all submitted code and must understand it completely.
- Human reviewers own the final code and must understand it fully.

### Authoritative References

- `GNUmakefile` — canonical list of all targets and variables.
- `docs/naming.md` — naming rules for resources, files, functions, tests.
- `docs/error-handling.md` — error handling patterns.
- `docs/data-handling-and-conversion.md` — flatteners, expanders, AutoFlex.
- `docs/retries-and-waiters.md` — retry and waiter helpers.
- `docs/running-and-writing-acceptance-tests.md` — acceptance test patterns.
- `names/caps.md` — enforced capitalization list for initialisms.
- `.ci/.golangci*.yml` — golangci-lint configuration.
- `docs/ai-agent-guides/` — task-specific agent guides (resource identity, list resources, etc.).

### Environment Setup

```bash
make prereq-go   # install the required Go version (see .go-version)
make tools       # install linters and helper binaries (run once)
make build       # compile the provider
```

Acceptance tests require Terraform CLI 0.12.31+ and AWS credentials in the environment (`AWS_PROFILE` or `AWS_ACCESS_KEY_ID`/`AWS_SECRET_ACCESS_KEY` plus `AWS_DEFAULT_REGION`). The default acceptance test region is `us-west-2`.

### Common Commands

#### Build

```bash
make build          # build provider binary
make go-build       # CI-style build check
```

#### Format and Imports

```bash
make fmt            # fix Go source formatting (gofmt -s)
make fmt-check      # verify formatting
make fix-imports    # fix import ordering (goimports)
```

#### Unit Tests

```bash
make test                                          # all unit tests
make test PKG=iam                                  # one service package
make test PKG=iam TESTARGS='-run TestExpandRole$'  # one test
make test TEST=./internal/create                   # one directory
```

#### Acceptance Tests (require AWS credentials, create real resources)

```bash
make testacc PKG=cloudwatch TESTS=TestAccCloudWatchDashboard_
make testacc PKG=iam TESTS=TestAccIAMRole_basic
make t PKG=iam T=TestAccIAMRole_basic              # alias
```

#### Linting

```bash
make golangci-lint PKG=<service>   # primary Go linter (preferred over make lint)
make golangci-lint1 PKG=<service>  # faster single-shard run
make import-lint PKG=<service>     # import ordering
make provider-lint PKG=<service>   # provider-specific lint rules
make testacc-lint PKG=<service>    # acceptance test HCL lint
```

#### Code Generation

```bash
make gen        # run all generators (required after changing @FrameworkResource/@SDKResource annotations)
make gen-check  # verify generated files are up to date
```

#### Pre-PR Quick Fix

```bash
make quick-fix PKG=<service>   # fmt, imports, testacc-lint, semgrep, terraform-fmt, copyright
make ci-quick                  # broader CI subset
```

### Adding a New Resource or Data Source

Always use `skaff` — do not copy an older resource:

```bash
make skaff
skaff resource --name ExampleThing --service example
```

After scaffolding:

1. Fill in the schema and CRUD handlers.
2. Add `@FrameworkResource("aws_example_thing", name="Example Thing")` annotation.
3. Run `make gen` to register the resource.
4. Write at minimum: `basic`, `disappears`, and key argument tests.

Net-new resources must use Terraform Plugin Framework. Existing SDKv2 resources stay SDKv2 unless migration is the explicit task.

### Code Style and Conventions

#### Naming
- Service packages: `internal/service/<serviceidentifier>` — lowercase, no underscores, prefer the shorter AWS SDK/CLI name.
- Go files: `snake_case.go`; data sources end with `_data_source.go`; tests end with `_test.go`.
- Main constructors: `Resource<Name>()` and `DataSource<Name>()`.
- CRUD helpers: `resource<Name>Create`, `resource<Name>Read`, `resource<Name>Update`, `resource<Name>Delete`.
- Use Go MixedCaps with correct initialisms: `VPCEndpoint`, `ARN`, `IAM`, `URL`, `API`. Never `Id` — always `ID`.
- Do not include the service name in function/variable names within the service package.
- Terraform schema attribute names use `snake_case`.

#### Imports

Order: standard library → third-party → local. Enforced by `make import-lint`.

Required aliases:

- `github.com/hashicorp/terraform-plugin-sdk/v2/helper/id` → `sdkid`
- `github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry` → `sdkretry`
- `github.com/hashicorp/terraform-plugin-testing/helper/acctest` → `sdkacctest`
- `github.com/hashicorp/terraform-provider-aws/internal/types` → `inttypes`

#### Schema
- Prefer `Optional: true` + `Computed: true` for attributes with server-side defaults; avoid provider defaults.
- Use AWS SDK constants instead of duplicating string values.
- For Framework resources, use `internal/framework/flex` AutoFlex before writing custom conversion code.

#### Error Handling
- Name error variables `err`; use early returns.
- Wrap errors: `fmt.Errorf("doing thing: %w", err)`.
- Use `tfawserr.ErrCodeEquals` / `tfawserr.ErrMessageContains` for AWS API error matching.
- In SDKv2 `Read`, only call `d.SetId("")` when `!d.IsNewResource()`.
- Use `internal/retry` for waiters and retries; do not write custom polling loops.

#### Tags

AutoFlex ignores `Tags` by default. Keep tagging logic separate.

#### `nolint` Comments

Must name the specific linter and explain why: `//nolint:staticcheck // reason`.

### Testing Conventions
- Unit tests: `Test<FunctionName>` — no service name, usually no underscores, always call `t.Parallel()`.
- Acceptance tests: `TestAcc<Service><Resource>_<case>` — e.g., `TestAccIAMRole_basic`, `TestAccIAMRole_tags`.
- Serialized acceptance helpers: `testAcc<Resource>_<case>` (lowercase `t`, no service name).
- Acceptance tests use: `acctest.Context(t)`, `acctest.PreCheck`, `acctest.ErrorCheck`, `acctest.ProtoV5ProviderFactories`, destroy checks, and import-state verification.
- Do not hardcode AMI IDs, regions, partitions, or DNS suffixes — use provider helpers.
- Place unit tests before acceptance tests in the same file.

### Pre-PR Checklist
1. `make quick-fix PKG=<service>` — auto-fix formatting, imports, HCL, copyright.
2. `make test PKG=<service>` — unit tests pass.
3. `make golangci-lint PKG=<service>` — no new lint errors.
4. `make import-lint PKG=<service>` — import ordering clean.
5. `make provider-lint PKG=<service>` — provider rules clean.
6. `make gen` — if annotations or generators were changed; then `make gen-check`.
7. Acceptance tests — if AWS credentials are available and the change warrants it.

Escalate to `make ci-quick` only when the above are clean or the change is broad.
