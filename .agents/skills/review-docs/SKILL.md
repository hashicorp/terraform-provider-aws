---
name: review-docs
description: "Review a Terraform AWS Provider PR's end-user documentation (website/docs/**/*.markdown): whether docs are needed, description openings, argument/attribute style, section structure, tags wording, code fences, and the Import + Identity Schema section. Use when reviewing documentation changes, or when a PR adds/changes a resource or data source and its docs should be checked."
---

<!-- Copyright IBM Corp. 2014, 2026 -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

# Review: End-User Documentation

Assume the `@maintainer` persona. Scope: resource docs (`r/`), data source docs (`d/`), list resource docs (`list-resources/`) under `website/docs/`, rendered at the Terraform Registry. Loaded from `review-pr`.

Precedence when sources disagree: `swissshepherd` (`make swissshepherd`, configured in `.ci/swissshepherd-weak.hcl`) wins, then the `skaff` templates (`skaff/*/websitedoc.gtpl`), then [docs/end-user-documentation.md](../../../docs/end-user-documentation.md), then this skill. Do not review CHANGELOG updates.

## Should the PR include docs?

- If the PR adds or changes user-facing behavior (new resource/data source/attribute) but has no doc change, say what's missing.
- If the PR removes behavior but leaves stale docs, say what to remove.
- Review only the changed docs; don't review unchanged documentation except as needed for context.

## Description openings

The single-line `description:` in frontmatter and the first paragraph after the H1 must use the verb for the doc type:

| Doc type | Opening verb |
|---|---|
| Resource | `Manages …` |
| Data source | `Provides details about …` |
| List resource | `Lists …` |

Flag openings like "This resource…", "Use this resource…", "Allows you to…", "Resource for…", or "Terraform resource for…".

## Argument and attribute description style

**Forbidden opening words**: `An`, `A`, `The`, `Defines`, `Indicates`, `Specifies`. Rewrite "Indicates the amount of storage" → "Amount of storage." Always propose the corrected wording.

**Don't spell out common abbreviations** (e.g. "Amazon Resource Name (ARN)" → "ARN"). See [Abbreviations](../../../docs/end-user-documentation.md#abbreviations) for the full list and rationale.

**Boolean arguments** must start with `Whether to`:

- ✓ `(Optional) Whether to enable logging.`
- ✗ `(Optional) Enables logging.` / `(Optional) If true, enables logging.`

**Labels** come in parentheses immediately after the hyphen. In Argument Reference, the first label is `(Required)`, `(Optional)`, or `(Read-Only)`, matching the schema: `(Optional)` for Optional+Computed fields and `(Read-Only)` only for computed-only fields. Trailing traits are allowed, for example `(Optional, Forces new resource)` or `(Optional, **Deprecated**)`. Attribute Reference entries carry no label, including no `(Read-Only)`.

**Examples use `example`, not `test`.** Flag any HCL block in docs that uses `"test"` as a resource label or `name = "test"`.

## Section structure

Resource docs, in this order:

1. `# Resource: <aws_resource_name>`
2. `## Example Usage` (with at least `### Basic Usage`)
3. `## Argument Reference`
4. `## Attribute Reference`
5. `## Timeouts` *(only if the resource exposes timeouts)*
6. `## Import` *(only for resources)*

Other types, after the title and Example Usage (`types` in `.ci/swissshepherd-weak.hcl` are authoritative):

| Doc type | Sections, in order |
| --- | --- |
| Data source | Argument Reference, Attribute Reference, Timeouts *(only if the schema has timeouts)*. No Import. |
| Ephemeral resource | Argument Reference, Attribute Reference, Usage Notes *(optional)*. |
| List resource | Argument Reference. No Attribute Reference. |
| Action | Dependency Management *(optional)*, Argument Reference. |
| Function | Signature, Arguments. |

Any other `##` heading is an error.

## Argument Reference

Alphabetize. Use one of these lead-ins (substitute "data source", "ephemeral resource", "list resource", or "action" for "resource"):

```
This resource supports the following arguments:

* `opt_arg` - (Optional) ...
* `req_arg` - (Required) ...
```

or, with required and optional arguments split into separate lists, each alphabetized:

```
The following arguments are required:

* `req_arg` - (Required) ...

The following arguments are optional:

* `opt_arg` - (Optional) ...
```

If there are no required arguments, drop the "required" list; don't write "There are no required arguments."

For tagged resources, `tags` lives under optional arguments with this exact wording:

```
* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
```

## Attribute Reference

Open with `This resource exports the following attributes in addition to the arguments above:` (substitute "data source" / "ephemeral resource"; list resources and actions have no Attribute Reference). Flag pages that re-document arguments here — only computed attributes belong. Alphabetize.

For tagged resources, `tags_all` has this exact wording:

```
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
```

For tagged data sources, `tags` is a top-level computed attribute and there is no `tags_all`.

## Code fences

Use `terraform` for `.tf` blocks and `console` for shell examples. Flag `hcl` (use `terraform`) or unfenced shell commands.

## Style points

- Active voice, present tense. Don't document past or future behavior.
- Single-line description in frontmatter — flag multi-paragraph descriptions.
- Consistent terminology — match the AWS service's nouns; don't invent synonyms.

## Import section & Identity Schema

The `## Import` section (resources only) and its `Identity Schema` subsection have a precise required structure. See [`references/import-section.md`](references/import-section.md) for the full checklist and load it when the PR touches a resource's Import section.
