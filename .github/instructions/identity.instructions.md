---
applyTo: "internal/service/**/*.go"
---
<!-- Copyright IBM Corp. 2014, 2026 -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

# Resource Identity

Scope: identity declaration and import-ID parsers.

## Identity strategy — exactly one per resource

Every new resource declares one identity strategy via comment annotations above the factory function. Flag missing identity annotations on new resources.

| Strategy | Annotations | Use when |
|---|---|---|
| ARN | `// @ArnIdentity` (or `// @ArnIdentity("arn_attribute")`) | AWS API uses the ARN as identifier |
| Singleton | `// @SingletonIdentity` | One per region (or one per account for global services) |
| Parameterized | one or more `// @IdentityAttribute("<attr>")` | Composite or non-ARN identifier |

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

- Omit `@ImportIDHandler` — generation will emit a build error.
- Reference an `@ImportIDHandler` whose target type doesn't satisfy `inttypes.ImportIDParser` (the `var _ inttypes.ImportIDParser = ...{}` assertion catches this).
- Implement `Parse` without a clear error message describing the expected import-ID format on malformed input.

`@ImportIDHandler` is only valid on parameterized identities with multiple attributes. Flag uses on singleton or single-attribute parameterized resources.

## Region opt-out for global services

Global services (CloudFront, IAM, Route 53 hosted zones, etc.) omit `framework.WithRegionModel` from the model and register with `inttypes.ResourceRegionDisabled()` in the service-package generator. Identity-Schema docs for these resources omit the `region` attribute.
