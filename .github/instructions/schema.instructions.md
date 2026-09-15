---
applyTo: "internal/service/**/*.go"
---
<!-- Copyright IBM Corp. 2014, 2026 -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

# Schema & Resource Shape

Scope: model struct, attributes, blocks, validators, plan modifiers, timeouts on non-test resource code.

## Model struct

```go
type fooResourceModel struct {
    framework.WithRegionModel  // first field; omit for global services only
    // attributes...
    Timeouts timeouts.Value `tfsdk:"timeouts"` // only when timeouts block is present
}
```

- `framework.WithRegionModel` is the first embedded field for regional resources/data sources/list resources. Global services (CloudFront, IAM, Route 53 hosted zones, etc.) omit it and register with `inttypes.ResourceRegionDisabled()`.
- `tfsdk` tags must match schema attribute names exactly (snake_case).
- Use Plugin Framework types (`types.String`, etc.) and `fwtypes.ListNestedObjectValueOf[T]` for nested blocks.
- Time fields use `timetypes.RFC3339`, never `types.String`.

## Attribute rules

- Attribute names are `snake_case`.
- **Alphabetize** within `Attributes` and within `Blocks`. No blank lines between attributes.
- Use `internal/names` constants (`names.AttrARN`, `names.AttrName`, `names.AttrTags`, `names.AttrTagsAll`, `names.AttrTimeouts`, etc.) instead of string literals when one exists.

### Required / Optional / Computed

Only four valid combinations. Flag any other.

| Required | Optional | Computed | When |
|---|---|---|---|
| ✓ | | | User must supply |
| | ✓ | | User may supply; **never use `Default` here** |
| | | ✓ | Read-only |
| | ✓ | ✓ | User or provider supplies; required pairing for `Default` |

Prefer `Optional + Computed` over `Optional + Default` when AWS supplies a server-side default.

### `id` attribute

Plugin Framework does **not** require an `id` attribute. Include `names.AttrID: framework.IDAttribute()` only when the AWS API returns an `Id` field. Flag schemas that include it on ARN-identified resources just by habit.

## Plan modifiers (no `ForceNew`)

Plugin Framework uses plan modifiers, not `ForceNew`. Flag any new Framework code that mentions `ForceNew`.

- `stringplanmodifier.RequiresReplace()` — replaces `ForceNew: true`
- `stringplanmodifier.UseStateForUnknown()` — keeps a computed value stable across plans

## Validators replace `MaxItems`/`MinItems`

Use `listvalidator.SizeAtMost(N)` / `SizeAtLeast(N)` (and `setvalidator` equivalents) on the block, not `MaxItems`/`MinItems` on the attribute. For a single nested object, combine `listvalidator.SizeAtMost(1)` with a list nested block.

## Nested blocks need a custom type

`schema.ListNestedBlock` requires `CustomType: fwtypes.NewListNestedObjectTypeOf[complexArgumentModel](ctx)`. The corresponding model field is `fwtypes.ListNestedObjectValueOf[complexArgumentModel]`. Flag nested blocks that omit `CustomType`.

For sets of strings, use `CustomType: fwtypes.SetOfStringType` with `fwtypes.SetValueOf[types.String]`.

For AWS API map blocks, the provider schema has no map-of-objects type — use `MapBlockKey`. Flag attempts to model AWS map blocks with `schema.MapAttribute`.

## Sensitive attributes

Mark credentials, tokens, and write-only secrets `Sensitive: true`. Flag computed or required attributes that obviously hold secret material but aren't marked sensitive.

## Timeouts

Resources with configurable timeouts use `timeouts.Block(ctx, timeouts.Opts{Create: true, Update: true, Delete: true})` and call `r.SetDefaultCreateTimeout` / `SetDefaultUpdateTimeout` / `SetDefaultDeleteTimeout` with defaults that reflect AWS API behavior — RDS Create can take 30–60 minutes; an EC2 SG takes seconds. Don't use arbitrary round numbers. Data sources and list resources do not expose configurable timeouts.
