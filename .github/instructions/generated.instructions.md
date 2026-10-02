---
applyTo: "internal/service/**/*_gen.go"
---
<!-- Copyright IBM Corp. 2014, 2026 -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

# Generated Code

These files end in `_gen.go` and are produced by `go generate ./...` (templates and generator code live under `internal/generate/`). Common examples: `service_package_gen.go`, `tags_gen.go`, `service_endpoint_resolver_gen.go`, `*_identity_gen_test.go`, `*_tags_gen_test.go`.

## What to review

- Whether the file should exist at all — was the generator run after the source change that motivated it?
- Diff plausibility — additions/removals match the source changes (a new resource added one annotation should produce a localized diff, not a wholesale rewrite of unrelated registrations).
- Manual edits — almost always a mistake. The legitimate flow is: change the source (annotation, template, or generator code under `internal/generate/`) → run `go generate ./...` → commit the regenerated file. Flag PRs that modify `_gen.go` files without a corresponding source-side change.

## What NOT to review

- Style, naming, formatting, comment wording, import ordering, structural refactors. The generator owns these.
- Suggested simplifications, "extract helper" suggestions, or any cleanup not reachable from the generator's templates.
- Rules from `lifecycle.instructions.md`, `schema.instructions.md`, `helpers.instructions.md`, `identity.instructions.md`, or `tags.instructions.md`. Those target hand-written code.
