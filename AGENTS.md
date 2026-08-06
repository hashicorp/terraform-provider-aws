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
| [review-pr](./.agents/skills/review-pr/SKILL.md) | Review a Terraform AWS Provider PR. Router: holds cross-cutting principles and routes to the scoped `review-*` leaf skills below based on the files a PR changes. |
| [review-lifecycle](./.agents/skills/review-lifecycle/SKILL.md) | Review resource CRUD, errors, and AutoFlex (`internal/service/**/*.go`). |
| [review-schema](./.agents/skills/review-schema/SKILL.md) | Review Plugin Framework schema shape (`internal/service/**/*.go`). |
| [review-helpers](./.agents/skills/review-helpers/SKILL.md) | Review finders, waiters, sweepers, data sources, list resources (`internal/service/**/*.go`). |
| [review-identity](./.agents/skills/review-identity/SKILL.md) | Review Resource Identity annotations and import-ID handlers (`internal/service/**/*.go`). |
| [review-tags](./.agents/skills/review-tags/SKILL.md) | Review tag schema attributes, wiring, and the `@Tags` annotation (`internal/service/**/*.go`). |
| [review-generated](./.agents/skills/review-generated/SKILL.md) | Review generated code (`internal/service/**/*_gen.go`). |
| [review-tests](./.agents/skills/review-tests/SKILL.md) | Review acceptance/unit test basics (`internal/service/**/*_test.go`). |
| [review-tests-helpers](./.agents/skills/review-tests-helpers/SKILL.md) | Review Exists/Destroy, data source, list, and unit tests (`internal/service/**/*_test.go`). |
| [review-docs](./.agents/skills/review-docs/SKILL.md) | Review a PR's end user documentation updates (`website/docs/**/*.markdown`). |

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
│   ├── smerr/              # Smarterr utilities
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
- Verification is a hard exit criterion for every PR (see [Development workflow](#development-workflow)). Without it, the task is not done.
- Prefer the boring, obvious solution. Touch only what you're asked to touch.
- Every PR must build, pass tests, and be lint-free.
- Follow existing conventions for naming, style, and idioms.
- Follow current best practices and conventions for naming, style, idioms. Legacy patterns should be avoided.
- Reuse the repository's utility packages (in `internal/`, excluding `internal/generate/` and `internal/service/`) before writing new utility code. Add new dependencies only after exhausting these.

### Coding Conventions (Follow These)

#### Go language usage
- **GO USES TAB (`\t`) CHARACTERS TO INDENT**
- **Use elegant Go, modern (Go 1.26+) idioms** (e.g., `slices.Contains()`)
- **Go nuance**: Don't build single files, **build a package**

#### Error handling
- Use smarterr/smerr
- Use `retry.NotFound()` to check for missing resources during Read.
- Return early on error; don't accumulate diagnostics past the first fatal error.

### Common Patterns

#### Resource file naming
- `internal/service/{service}/{thing}.go` — thing resource implementation
- `internal/service/{service}/{thing}_test.go` — thing resource acceptance tests
- `internal/service/{service}/{thing}_data_source.go` — thing data source
- `website/docs/r/{service}_{thing}.html.markdown` — thing resource documentation
- `website/docs/d/{service}_{thing}.html.markdown` — thing data source documentation

#### Resource implementation pattern (Framework)
New resources use the Terraform Plugin Framework pattern:
- Implement `resource.Resource` interface
- Use AutoFlex for flattening/expanding where possible
- Use `retry.RetryContext` for eventual consistency

For example:
```go
func (r *thingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
    // 1. Read model from state
    // 2. Call AWS API
    // 3. Handle NotFound → remove from state
    // 4. AutoFlex response into model
    // 5. Write model to state
}
```

## Development workflow

### Overview
- Substantive changes and correctness first, lint after: run `make quick-fix PKG=<service>` near the end, before raising PR. Avoid the tiny change → lint → tiny change → lint loop. `make fmt` is the cheap exception — run it freely.
- Scope most commands to the package changed — the provider is very large. CI is the provider-wide gate for build, lint, and semgrep.

### AI usage
When you help prepare a PR, disclose the AI's role in the description and add `🤖🤖🤖` to the title. See [`docs/ai-usage.md`](docs/ai-usage.md) for the full policy. **Humans are fully responsible for the code regardless of AI usage.**

### Running commands
- `make t` and `make testacc`: Run acceptance tests and create real AWS resources. Get explicit approval before running.
- `make …` (except acceptance tests), `go …`, and read-only commands (`awk`, `grep`, `ls`, `rg`) are safe to run without confirmation.

### Regenerate, test, and verify (scoped to your package)
- **Regenerate** after changing annotations or a service's `generate.go`: `make gen PKG=<service>`. Run the provider-wide `make gen` only after changing `names/data/names_data.hcl`, anything under `internal/generate/` — it affects every service and takes many minutes.
- **Test** with `make test PKG=<service>` (unit test) (`T=<pattern>` filters by name); for non-service changes, e.g., `go test ./internal/conns/...`.
- **Fix and verify** with `make quick-fix PKG=<service>` — the default final pass. It applies formatting, imports, lint, semgrep fixes, and `copyright-fix`, and fails if the build is broken (no separate build step needed).
- **Documentation**: run `make swissshepherd` to verify changes align with docs. Run `make swissshepherd-refresh` only once at the beginning of a session.

### Commits, CHANGELOG, and docs
- Keep each commit small, atomic, and single-purpose; the message describes the change.
- Add a `.changelog/` entry for new features, bug fixes, and enhancements.
- New features require new documentation; `./docs/end-user-documentation.md` is authoritative.

## Boundaries
- Never edit `CHANGELOG.md` directly — use `.changelog/` entries.
- Never edit generated files by hand — modify the generator or annotations, then run `make gen PKG=<service>` or `make gen` (provider level).
- Do not modify `go.mod`/`go.sum` without running `go mod tidy`.
- Do not add new external dependencies without explicit approval.
- The `website/` directory follows different conventions; see `docs/end-user-documentation.md`.
