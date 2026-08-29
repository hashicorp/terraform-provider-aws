---
name: review-identity
description: "Review Terraform AWS Provider Resource Identity: the identity-strategy annotations (@ArnIdentity, @SingletonIdentity, @IdentityAttribute), multi-attribute @ImportIDHandler parsers, and region opt-out for global services. Use when reviewing a PR that adds or changes identity annotations or import-ID parsing in internal/service/**/*.go."
---

<!-- Copyright IBM Corp. 2014, 2026 -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

# Review: Resource Identity

Assume the `@maintainer` persona. Scope: identity declaration and import-ID parsers. Loaded from `review-pr`.

## Identity strategy — exactly one per resource

Every new resource declares one identity strategy via comment annotations above the factory function. Flag missing identity annotations on new resources.

| Strategy | Annotations | Use when |
|---|---|---|
| ARN | `// @ArnIdentity` (or `// @ArnIdentity("arn_attribute")`) | AWS API uses the ARN as identifier |
| Singleton | `// @SingletonIdentity` | One per region (or one per account for global services) |
| Parameterized | one or more `// @IdentityAttribute("<attr>")` | Composite or non-ARN identifier |

`@IdentityAttribute` supports keywords: `optional`, `resourceAttributeName`, `testNotNull`, `valueType`, and `identityDuplicateAttributes`.

## Multi-attribute identity needs an `ImportIDHandler`

Parameterized identities with more than one attribute require both:

1. An `// @ImportIDHandler("<typeName>")` annotation (alongside the `@IdentityAttribute` annotations) referencing a type that satisfies `inttypes.ImportIDParser`.
2. The implementation:

```go
type fooImportID struct{}

func (fooImportID) Parse(id string) (string, map[string]string, error) {
    // parse a separator-delimited string into named identity attributes
}

var _ inttypes.ImportIDParser = fooImportID{}
```

Flag multi-attribute identity resources that:

- Omit `@ImportIDHandler` — the generator's `Validate()` requires it for multiple parameterized identity and will error.
- Reference an `@ImportIDHandler` whose target type doesn't satisfy `inttypes.ImportIDParser` (the `var _ inttypes.ImportIDParser = ...{}` assertion catches this).
- Implement `Parse` without a clear error message describing the expected import-ID format on malformed input.

`@ImportIDHandler` is meaningful for parameterized identities with multiple attributes. Question its use on singleton or single-attribute parameterized resources.

## Region opt-out for global services

Global services (CloudFront, IAM, Route 53 hosted zones, etc.) omit `framework.WithRegionModel` from the model and register with `inttypes.ResourceRegionDisabled()` in the service-package generator. Identity-Schema docs for these resources omit the `region` attribute.
