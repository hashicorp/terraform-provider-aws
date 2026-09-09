---
name: review-pr
description: "Review a Terraform AWS Provider pull request for correctness and conventions. Start here for any PR review: it holds the cross-cutting review principles and routes to scoped sub-skills (review-lifecycle, review-schema, review-helpers, review-identity, review-tags, review-generated, review-tests, review-tests-helpers, review-docs) based on which files the PR changes. Use when given a hashicorp/terraform-provider-aws PR URL and asked to review it."
---

<!-- Copyright IBM Corp. 2014, 2026 -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

# Skill: Review a Pull Request

Assume the `@maintainer` persona.

Review a GitHub Pull Request against the Terraform AWS Provider's conventions. This skill carries the cross-cutting principles and a routing table; load the scoped leaf skills for the file types the PR actually changes.

## When to use

Trigger this skill when the user:
- Provides a `https://github.com/hashicorp/terraform-provider-aws/pull/<N>` URL and asks for a review.
- Says "review PR", "review this pull request", or similar, with a PR URL.

## Inputs

Required:
- A GitHub PR URL. Extract `<PR_NUMBER>` with the regex `/pull/(\d+)`.

If the user provides only a PR number, ask for the full URL (or confirm the repo is `hashicorp/terraform-provider-aws`).

## Procedure

1. Fetch the PR's changed files and diff.
2. Categorize each changed path against the routing table below.
3. Load the matching leaf skill(s) and apply their rules to the changed lines only.
4. Always consider compatibility (see below) and, when docs or a changelog are expected but missing, say so.
5. Consolidate findings into a single review. Cite the rule and propose corrected code.

## Cross-cutting principles

**Compatibility is non-negotiable.** Changes must preserve Terraform state compatibility, upgrade behavior, import behavior, and existing user workflows. Schema changes that force replacement, rename attributes, or break state migration must be flagged unless explicitly justified. For a dedicated compatibility pass, use the `breaking-changes` skill. Do **not** rely on any `breaking-change` label.

**Favor recent patterns; do not enforce legacy ones.** The provider has both modern (Plugin Framework) and legacy (Plugin SDKv2) code. New work follows recent patterns. Do not ask contributors to mimic legacy patterns just because nearby code uses them. When modifying an existing resource, keep the framework it already uses.

**Go style.** Expect modern Go (1.26+): `slices`, `maps`, `cmp`, `iter`, `errors.Is`/`errors.As`, range-over-int/func. Prefer return-early. AWS SDK for Go v2 only. Detect AWS API exceptions with `errs.IsA[*awstypes.<Exception>]` — never type assertions or `strings.Contains`.

**Review tone.** Be specific, accurate, and to the point. Cite the rule and propose the corrected code. Prioritize correctness, design, and compatibility. Don't hand-author comments for mechanical issues that `make lint`, semgrep, or formatters already enforce — at most, note that automated checks will flag them. Spend the review on what tools can't catch: logic, AWS API usage, state and upgrade compatibility, and test coverage.

## Routing table

| Changed path | Concern | Leaf skill |
|---|---|---|
| `internal/service/**/*.go` (non-test) | CRUD, errors, AutoFlex | `review-lifecycle` |
| `internal/service/**/*.go` (non-test) | Model/attrs/blocks/validators | `review-schema` |
| `internal/service/**/*.go` (non-test) | Finders, waiters, sweepers, data sources, list resources | `review-helpers` |
| `internal/service/**/*.go` (non-test) | Resource Identity | `review-identity` |
| `internal/service/**/*.go` (non-test) | Tags | `review-tags` |
| `internal/service/**/*_gen.go` | Generated code | `review-generated` |
| `internal/service/**/*_test.go` | Acceptance/unit test basics | `review-tests` |
| `internal/service/**/*_test.go` | Exists/Destroy, data source/list/unit tests | `review-tests-helpers` |
| `website/docs/**/*.markdown` | End-user documentation | `review-docs` |

Multiple leaves may apply to one file. The `internal/service/**/*.go` glob matches both production and test files; the non-test leaves generally don't apply to tests.

## Broader context

See [`AGENTS.md`](../../../AGENTS.md) for personas, the skills registry, build/test commands, and the AI-usage policy.
