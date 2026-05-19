<!-- Copyright IBM Corp. 2014, 2026 -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

# AGENTS.md

This file provides guidance to AI coding agents when working with code in this repository.

## Repository Overview

This is the Go-based Terraform AWS Provider (`github.com/hashicorp/terraform-provider-aws`). It maps AWS API resources to Terraform resources, data sources, ephemeral resources and actions (collectively often referred to as just resources). The primary language is Go; HCL appears in acceptance test configurations and website documentation.

## Agent Registry

This project uses specialized personas for different tasks.

### Available Personas
- **`@contributor`**: [Contributor Persona](./.agents/contributor.md) - Contributes code in the form of bugfixes, enhancements to existing resources, and new resources. Makes clarifications and corrections to existing documentation.
- **`@maintainer`**: [Maintainer Persona](./.agents/maintainer.md) - Steward of the project, responsible for both internal and external quality. Reviews contributions. Maintains provider-level features, including new Terraform language constructs.
- **`@tcm`**: [TCM Persona](./.agents/tcm.md) - Triages incoming GitHub issues and PRs. Engages with community members to answer technical and process questions. Suggests workarounds and alternatives to reported bugs.

### Registry Rules
- Always use the requested persona for tasks.
- If no persona is specified, default to `@contributor`.
- A persona defines a role with a perspective and responsibilities.
- Personas may invoke skills.

## Skills

Skills are loaded from `./.agents/skills`. Each skill supplies step-by-step instructions, code patterns, and guardrails for a specific task.

| Skill | Task |
|---|---|
| [breaking-changes](./.agents/skills/breaking-changes/SKILL.md) | Review a PR for possible breaking changes. |
| [changelog](./.agents/skills/changelog/SKILL.md) | Add a `.changelog/<PR_NUMBER>.txt` entry from a PR URL, commit, and push (with confirmation). |
| [fixdocs](./.agents/skills/fixdocs/SKILL.md) | Fix end user documentation with `swissshepherd`. |
| [reviewdocs](./.agents/skills/reviewdocs/SKILL.md) | Review a PR's end user documentation updates. |

## Stack
- Go 1.26+, AWS SDK for Go v2.
- Terraform Plugin Framework + Terraform Plugin SDKv2 ([muxed](https://developer.hashicorp.com/terraform/plugin/mux) provider).
- Code generators in `internal/generate/`.
- Build system: GNU Make (see `GNUmakefile`).
- Testing: Go standard `testing` package + [`terraform-plugin-testing` acceptance test framework](https://developer.hashicorp.com/terraform/plugin/testing/acceptance-tests).

## Code Structure (The important parts)

```
terraform-provider-aws/
├── .changelog/             # CHANGELOG entries
├── internal/
│   ├── acctest/            # Acceptance test helpers
│   ├── backoff/            # Low-level backoff loop implementation
│   ├── conns/              # Provider-level global state, including provider configuration
│   ├── enum/               # AWS SDK for Go v2 enumeration utilities
│   ├── errs/               # Go `error` utilities
│   │   ├── fwdiag/         # Terraform Plugin Framework `Diagnostic` utilities
│   │   └── sdkdiag/        # Terraform Plugin SDKv2 `Diagnostic` utilities
│   ├── flex/               # General and Terraform Plugin SDKv2-specific flatteners and expanders
│   ├── framework/          # Terraform Plugin Framework utilities
│   │   ├── flex/           # Flatteners and expanders, including AutoFlex
│   │   ├── types/          # Custom type implementations
│   │   └── validators/     # Validator implementations
│   ├── function/           # Provider functions
│   ├── generate/           # Code generators
│   ├── iter/               # Go iterator utilities
│   ├── json/               # JSON utilities
│   ├── maps/               # Go `map` utilities
│   ├── provider/           # Provider initialization and configuration
│   │   ├── framework/      # Terraform Plugin Framework-specific initialization and configuration plus interceptors
│   │   ├── interceptors/   # Common interceptor utilities
│   │   └── sdkv2/          # Terraform Plugin SDKv2-specific initialization and configuration plus interceptors
│   ├── reflect/            # Go reflection utilities
│   ├── retry/              # Generic operation retry functionality
│   │   └── state.go        # Resource wait-for-state functionality
│   ├── sdkv2/              # Terraform Plugin SDKv2 utilities
│   ├── service/*/          # Per-service resource implementations
│   │   ├── exports.go      # Functions and variables used by other Go packages
│   │   ├── exports_test.go # Functions and variables used by acceptance tests for this Go package
│   │   ├── generate.go     # Code generation instructions
│   │   └── sweep.go        # This service's resource sweepers
│   ├── slices/             # Go slice utilities
│   ├── smerr/              # Smart error utilities
│   ├── sweep/              # Resource sweeper utilities
│   ├── tags/               # Resource tagging utilities
│   ├── types/              # Go types
│   ├── vcr/                # VCR testing utilities
│   └── verify/             # Terraform Plugin SDKv2-specific attribute validation
├── go.mod
├── go.sum
├── GNUmakefile             # Build and test commands
└── main.go                 # Entry point
```

## Important: Dual Framework

This provider uses TWO Terraform plugin frameworks simultaneously:
- **Terraform Plugin SDKv2** (older resources) — uses `schema.Resource`, `d.Set()`, `d.Get()`
- **Terraform Plugin Framework** (newer resources) — uses `resource.Resource`, plan modifiers, AutoFlex

When modifying an existing resource, use the SAME framework it already uses.
When creating a new resource, use the Terraform Plugin Framework.

## Conventions

### Non-negotiable Rules
- Verification is a hard exit criterion for every task. Without it, the task is not done.
- Prefer the boring, obvious solution.
- Touch only what you’re asked to touch.
- Code quality must not be compromised.
  - Every change must build successfully and pass all tests. Use `make test` to run unit tests.
  - Code must be lint-free. Use `make lint` to check for linting issues.
- Follow existing conventions.
  - Consistency is key to maintaining a readable and maintainable codebase.
  - Before writing any code, analyze the existing codebase to understand and adopt its naming conventions, coding style, and language usage.
- This repository contains a comprehensive set of utility packages. Look for opportunities to use them before writing new code.
  - Look in the `internal/` directory (excluding `internal/generate/` and `internal/services/`) for broadly reusable utilities.
  - Only add new dependencies as a last resort, or when explicitly requested.

#### AI usage policy

Per `docs/ai-usage.md`:
- Disclose AI use in the PR description.
- Include `🤖🤖🤖` in the PR title if an LLM agent is directly involved in submitting it.
- The human PR author is fully responsible for all submitted code and must understand it completely.
- Human reviewers own the final code and must understand it fully.

### Coding Conventions (Follow These)

#### Go language usage
- **GO USES TAB (`\t`) CHARACTERS TO INDENT**
- **Use elegant Go, modern (Go 1.26+) idioms** (e.g., `slices.Contains()`)
- **Go nuance**: Don't build single files, **build a package**

#### Code generation
- Run `make gen` after making changes to any annotations (`// @...` comments in Go files), any `internal/service/*/generate.go` source files, or `names/data/names_data.hcl`.

#### Error handling
- Wrap AWS errors with `fmt.Errorf("reading X (%s): %w", id, err)`.
- Use `retry.NotFound()` to check for missing resources during Read.
- Return early on error; don't accumulate diagnostics past the first fatal error.

### Guidelines

#### Running commands
- Confirm that any commands you intend to run are safe before running them.
- Commands that are safe to run in the repository are:
  - Any command that invokes `make`.
  - Any command that invokes `go`.
- Do not prompt for confirmation before running any of the safe commands above.
  - Any other commands may be unsafe and should not be run without confirmation.

#### Running tests
- Every PR must leave tests in a passing state.
- All existing tests must pass.
  - Use `make test` to run unit tests.
  - If your change breaks an existing test, fix it.
- CI is the gate. Run `make ci-quick`.
  - PRs with failing tests do not merge.

#### Documentation checklist
- Authoritative reference: `./docs/end-user-documentation.md`.
- New features require new documentation.
- Correct spelling and grammar are important.
- Run `make swissshepherd` to verify.

#### Copyright headers
- All applicable files must have a copyright header.
- Run `make copyright-fix` to ensure headers are correct.

#### Commit messages
- Each commit should be small and address a single change.
- The commit message describes the change.

#### CHANGELOG entries
- CHANGELOG entries are required for:
  - New resources, data sources, ephemeral resources, action, list resources and functions.
  - Bug fixes.
  - Enhancements.

### Common Patterns

#### Resource file naming
- `internal/service/{service}/{thing}.go` — thing resource implementation
- `internal/service/{service}/{thing}_test.go` — thing resource acceptance tests
- `internal/service/{service}/{thing}_data_source.go` — thing data source
- `website/docs/r/{thing}.html.markdown` — thing resource documentation
- `website/docs/d/{thing}.html.markdown` — thing data source documentation

#### Resource implementation pattern (Framework)
New resources use the Terraform Plugin Framework pattern:
- Implement `resource.Resource` interface
- Use AutoFlex for flattening/expanding where possible
- Use `retry.RetryContext` for eventual consistency

## Boundaries
- Never edit `CHANGELOG.md` directly — use `.changelog/` entries.
- Never edit generated files by hand — modify the generator or annotations, then run `make gen`.
- Do not modify `go.mod`/`go.sum` without running `go mod tidy`.
- Do not add new external dependencies without explicit approval.
- Beware of running acceptance tests (`make testacc`) without explicit approval — they create real AWS resources.
- The `website/` directory follows different conventions; see `docs/end-user-documentation.md`.

## Verification

Before finishing:
1. `make build` — must compile cleanly.
2. `make ci-quick` — zero warnings.
3. `make test` — all unit tests pass.
4. `make gen` — if you changed annotations or generators.
5. `make copyright-fix` — if you added new files.
6. `make swissshepherd` — if you changed documentation.
