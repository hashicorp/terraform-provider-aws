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
| [breaking-changes](./.agents/skills/breaking-changes/SKILL.md) | Review a PR for possible breaking changes. |
| [changelog](./.agents/skills/changelog/SKILL.md) | Add a `.changelog/<PR_NUMBER>.txt` entry from a PR URL, commit, and push (with confirmation). |
| [fixdocs](./.agents/skills/fixdocs/SKILL.md) | Fix end user documentation with `swissshepherd`. |
| [reviewdocs](./.agents/skills/reviewdocs/SKILL.md) | Review a PR's end user documentation updates. |

---

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

## Guiding Principles

### Follow existing conventions
Consistency is key to maintaining a readable and maintainable codebase.
Before writing any code, analyze the existing codebase to understand and adopt its naming conventions, coding style, and language usage.

### Avoid code duplication and minimize dependencies
This repository contains a comprehensive set of utility packages. Look for opportunities to use them before writing new code.
- Look in the `internal/` directory (excluding `internal/generate/` and `internal/services/`) for broadly reusable utilities
- Only add new dependencies as a last resort, or when explicitly requested

### Code quality must not be compromised
- Every change must build successfully and pass all tests. Use `make test` to run unit tests.
- Code must be lint-free. Use `make lint` to check for linting issues.

## Additional Guidelines

### Go language usage
- **GO USES TAB (`\t`) CHARACTERS TO INDENT**
- **Use elegant Go, modern (Go 1.26+) idioms** (e.g., `slices.Contains()`)
- **Go nuance**: Don't build single files, **build a package**

### Code generation
- Run `make gen` after making changes to any annotations (`// @...` comments in Go files), any `internal/service/*/generate.go` source files, or `names/data/names_data.hcl`.

### Working in the repository

#### Running commands
- Confirm that any commands you intend to run are safe before running them.
- Commands that are safe to run in the repository are:
  - Any command that invokes `make`.
  - Any command that invokes `go`.
- Do not prompt for confirmation before running any of the safe commands above.
  - Any other commands may be unsafe and should not be run without confirmation.

#### Running tests
- Every PR must leave tests in a passing state. This is non-negotiable.
- All existing tests must pass.
  - Use `make test` to run unit tests.
  - If your change breaks an existing test, fix it.
- TODO acceptance tests?
  - New features require new tests.
- CI is the gate. Run `make ci`.
  - PRs with failing tests do not merge.

### Documentation Checklist
- Authoritative reference: `./docs/end-user-documentation.md`.
- New features require new documentation.
- Correct spelling and grammar are important.
- Run `make swissshepherd` to verify.

### Copyright headers
- All applicable files must have a copyright header.
- Run `make copyright-fix` to ensure headers are correct.

### Commit messages
- Each commit should be small and address a single change.
- The commit message describes the change.

### CHANGELOG entries
- CHANGELOG entries are required for:
  - New resources, data sources, ephemeral resources, action, list resources and functions.
  - Bug fixes.
  - Enhancements.
- Do not edit `CHANGELOG.md` directly.
- TODO
