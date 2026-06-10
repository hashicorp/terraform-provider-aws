# GitHub Copilot Instructions
<!-- Copyright IBM Corp. 2014, 2026 -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

This repository implements the Terraform AWS Provider.

## Compatibility is non-negotiable

Changes must preserve Terraform state compatibility, upgrade behavior, import behavior, and existing user workflows. Schema changes that force replacement, rename attributes, or break state migration are flagged unless explicitly justified.

## Favor recent patterns; do not enforce legacy ones

The provider has both modern (Plugin Framework) and legacy (Plugin SDK v2) code. New work follows recent patterns. Do not ask contributors to mimic legacy patterns just because nearby code uses them. Do not suggest refactors or new abstractions unless they address a correctness, maintainability, or compatibility issue.

## Go style

Write modern Go (Go 1.25+ idioms): `slices`, `maps`, `cmp`, `iter`, `errors.Is` / `errors.As`. Prefer return-early. Use AWS SDK for Go v2 only. Detect AWS API exceptions with `errs.IsA[*awstypes.<Exception>]`, not type assertions or string matching.

## Review tone

Be specific and actionable. Cite the rule and propose the corrected code. Skip generic Go advice that adds no value beyond what a competent Go reviewer already knows.

Frame comments as guidance for maintainers to weigh, not as required changes for contributors. Contributors are not obligated to address every Copilot comment. Prefer "consider…", "this could…", "one option would be…" over imperatives like "change this", "you must", or "this needs to be fixed".

Begin each comment with this line on its own, before any other text:

> 🤖 _This Copilot comment is not a required action for the contributor._

## Broader agent context

For Copilot cloud agent and Copilot CLI sessions that work across the repository, see [`AGENTS.md`](../AGENTS.md). It covers personas, skills, build/test commands, and the AI-usage policy. Copilot code review and Copilot Chat in IDEs do not read `AGENTS.md`; rules those surfaces need live in this file or the path-specific instruction files below.

## Scoped rules

More detailed rules live in path-specific instruction files. Multiple files may apply to the same path.

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

The `internal/service/**/*.go` glob also matches `*_test.go`. Most rules in the non-test files don't fire on test files, but Copilot will see all matching files.
