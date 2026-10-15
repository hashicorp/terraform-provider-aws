---
name: review-generated
description: "Review generated code in the Terraform AWS Provider (files ending in _gen.go produced by go generate). Use when a PR changes internal/service/**/*_gen.go — check whether the generator was run, whether the diff is plausible for the source change, and whether files were hand-edited. Do NOT apply hand-written-code rules to these files."
---

<!-- Copyright IBM Corp. 2014, 2026 -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

# Review: Generated Code

Assume the `@maintainer` persona. Scope: files ending in `_gen.go`, produced by the code generators (`make gen`; templates and generator code live under `internal/generate/`). Common examples: `service_package_gen.go`, `tags_gen.go`, `service_endpoint_resolver_gen.go`, `*_identity_gen_test.go`, `*_tags_gen_test.go`. Loaded from `review-pr`.

## What to review

- Whether the file should exist at all — was the generator run after the source change that motivated it?
- Diff plausibility — additions/removals match the source changes (a new resource added one annotation should produce a localized diff, not a wholesale rewrite of unrelated registrations).
- Manual edits — almost always a mistake. The legitimate flow is: change the source (annotation, template, or generator code under `internal/generate/`) → regenerate (`make gen PKG=<service>`, or the provider-wide `make gen` when the change is under `internal/generate/`) → commit the regenerated file. Flag PRs that modify `_gen.go` files without a corresponding source-side change.

## What NOT to review

- Style, naming, formatting, comment wording, import ordering, structural refactors. The generator owns these.
- Suggested simplifications, "extract helper" suggestions, or any cleanup not reachable from the generator's templates.
- Rules from `review-lifecycle`, `review-schema`, `review-helpers`, `review-identity`, or `review-tags`. Those target hand-written code.
