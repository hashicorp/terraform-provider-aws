---
applyTo: "website/docs/**/*.markdown"
---
<!-- Copyright IBM Corp. 2014, 2026 -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

# User-Facing Documentation

Scope: resource docs (`r/`), data source docs (`d/`), list resource docs (`list-resources/`). Renders at `https://registry.terraform.io/providers/hashicorp/aws/latest/docs`.

Precedence: `swissshepherd` wins, then the `skaff` templates (`skaff/*/websitedoc.gtpl`), then `docs/end-user-documentation.md`, then these instructions.

## Description openings

The single-line `description:` in frontmatter and the first paragraph after the H1 must use the verb for the doc type:

| Doc type | Opening verb |
|---|---|
| Resource | `Manages …` |
| Data source | `Provides details about …` |
| List resource | `Lists …` |

Flag openings like "This resource…", "Use this resource…", "Allows you to…", "Resource for…", or "Terraform resource for…".

## Argument and attribute description style

**Forbidden opening words** for argument/attribute descriptions: `An`, `A`, `The`, `Defines`, `Indicates`, `Specifies`. Rewrite "Indicates the amount of storage" → "Amount of storage." Always propose the corrected wording.

**Boolean arguments** must start with `Whether to`:

- ✓ `(Optional) Whether to enable logging.`
- ✗ `(Optional) Enables logging.` / `(Optional) If true, enables logging.`

**Labels**: every argument gets `(Required)` or `(Optional)` right after the hyphen. Optional+Computed is `(Optional)`. Traits follow the label: `(Optional, Forces new resource)`, `(Optional, **Deprecated**)`. Attribute Reference entries get no label. Use `(Read-Only)` only for read-only fields in a block subsection that also has arguments; flag it anywhere else.

**Examples use `example`, not `test`.** Flag any HCL block in docs that uses `"test"` as a resource label or `name = "test"` etc.

## Section structure

Resource docs have these sections, in this order:

1. `# Resource: <aws_resource_name>`
2. `## Example Usage` (with at least `### Basic Usage`)
3. `## Argument Reference`
4. `## Attribute Reference`
5. `## Timeouts` *(only if the resource exposes timeouts)*
6. `## Import` *(only for resources)*

Other types, after the title and Example Usage:

| Doc type | Sections, in order |
| --- | --- |
| Data source | Argument Reference, Attribute Reference, Timeouts *(only if the schema has timeouts)*. No Import. |
| Ephemeral resource | Argument Reference, Attribute Reference, Usage Notes *(optional)*. |
| List resource | Argument Reference. No Attribute Reference. |
| Action | Dependency Management *(optional)*, Argument Reference. |
| Function | Signature, Arguments. |

Any other `##` heading is an error.

## Argument Reference

Alphabetize. With 5 or more arguments, including at least 2 required and 2 optional, split them:

```
The following arguments are required:

* `req_arg` - (Required) ...

The following arguments are optional:

* `opt_arg` - (Optional) ...
```

Otherwise use one list under `This resource supports the following arguments:`. In bylines, replace "resource" with the documentation type.

For tagged resources, `tags` lives under optional arguments with this exact wording:

```
* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
```

## Attribute Reference

Open with `This resource exports the following attributes in addition to the arguments above:`, replacing "resource" with the documentation type. Flag pages that re-document arguments here — only computed attributes belong. Alphabetize.

For tagged resources, `tags_all` has this exact wording:

```
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
```

For tagged data sources, `tags` is a top-level computed attribute and there is no `tags_all`.

## Code fences

Use `terraform` for `.tf` blocks and `console` for shell examples. Flag `hcl` (use `terraform`) or unfenced shell commands.

## Style points worth flagging

- Active voice, present tense. Don't document past or future behavior.
- Single-line description in frontmatter — flag multi-paragraph descriptions.
- Consistent terminology — if the AWS service uses a specific noun (e.g., "trust store"), match it; don't invent synonyms.
