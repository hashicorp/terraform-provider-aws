<!-- Copyright IBM Corp. 2014, 2026 -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

# GitHub Copilot Instructions

This repository implements the Terraform AWS Provider.

## Compatibility is non-negotiable

Changes must preserve Terraform state compatibility, upgrade behavior, import behavior, and existing user workflows. Schema changes that force replacement, rename attributes, or break state migration are flagged unless explicitly justified.

## Favor recent patterns; do not enforce legacy ones

The provider has both modern (Plugin Framework) and legacy (Plugin SDK v2) code. New work follows recent patterns. Do not ask contributors to mimic legacy patterns just because nearby code uses them. Do not suggest refactors or new abstractions unless they address correctness, maintainability, or compatibility.

## Go style

Write modern Go (Go 1.25+): `slices`, `maps`, `cmp`, `iter`, `errors.Is` / `errors.As`, `range` over int/func. Prefer return-early. Use AWS SDK for Go v2 only. Detect AWS API exceptions with `errs.IsA[*awstypes.<Exception>]`.

## Review tone

Be specific and actionable. Cite the rule and propose the corrected code. Focus on substance over style: avoid minor comments that create noise. Frame comments as guidance for maintainers to weigh, not required changes for contributors.

**Begin each comment with this line on its own, before any other text:**

> 🤖 _This Copilot comment is not a required action for the contributor._

## Broader agent context

See [`AGENTS.md`](../AGENTS.md) for personas, skills, build/test commands, and the AI-usage policy.

## Scoped rules

Path-specific instruction files. Multiple files may apply to the same path.

| Concern | File | Applies to |
|---|---|---|
| Resource lifecycle (CRUD, errors, AutoFlex) | `instructions/lifecycle.instructions.md` | `internal/service/**/*.go` |
| Helpers (finders, waiters, sweepers, data sources, list resources) | `instructions/helpers.instructions.md` | `internal/service/**/*.go` |
| Schema / resource shape | `instructions/schema.instructions.md` | `internal/service/**/*.go` |
| Resource Identity | `instructions/identity.instructions.md` | `internal/service/**/*.go` |
| Tags (schema attrs, wiring, annotation) | `instructions/tags.instructions.md` | `internal/service/**/*.go` |
| Generated code | `instructions/generated.instructions.md` | `internal/service/**/*_gen.go` |
| Acceptance test basics | `instructions/acceptance-tests.instructions.md` | `internal/service/**/*_test.go` |
| Test helpers (Exists/Destroy, list/data source/unit tests) | `instructions/acceptance-tests-helpers.instructions.md` | `internal/service/**/*_test.go` |
| User-facing documentation | `instructions/docs.instructions.md` | `website/docs/**/*.markdown` |
| Import section + Identity Schema docs | `instructions/docs-import.instructions.md` | `website/docs/**/*.markdown` |

The `internal/service/**/*.go` glob matches both production and test files; non-test rules generally don't apply to tests.
