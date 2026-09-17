---
applyTo: "internal/service/**/*.go"
---
<!-- Copyright IBM Corp. 2014, 2026 -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

# Tags

Scope: tag schema attributes, lifecycle wiring, and the `@Tags` annotation.

## Schema attributes

Tagged resources must include both:

```go
names.AttrTags:    tftags.TagsAttribute(),
names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
```

The model includes both `Tags` and `TagsAll` of type `tftags.Map`. Flag tagged resources that include only one.

Tagged data sources expose a single computed `tags` attribute via `tftags.TagsAttributeComputedOnly()` and no `tags_all`.

## Resource-level annotation

Tagged resources need `// @Tags(identifierAttribute="arn")` (or the attribute holding the resource's tag-attaching identifier) above the factory function. This drives tag-test generation. Flag tagged resources that omit `@Tags`.

## Wiring tags into Create / Update

After `flex.Expand` populates the AWS SDK input struct, set tags from the framework's helper:

```go
input.Tags = getTagsIn(ctx)
```

AutoFlex does **not** copy tags through. Missing this is a silent bug — the resource is created without tags and downstream tag-tests will fail. Flag tagged resources whose Create / Update doesn't set `input.Tags = getTagsIn(ctx)` after `flex.Expand`.

## Hand-written tag tests

Tag tests are **generated** for resources with `@Tags`. Flag PRs that add hand-written `_tags*` tests for new resources — they should be regenerated instead.
