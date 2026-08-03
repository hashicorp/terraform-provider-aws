---
applyTo: "website/docs/**/*.markdown"
---
<!-- Copyright IBM Corp. 2014, 2026 -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

# Import Section & Identity Schema

Scope: the `## Import` section and `Identity Schema` subsection of resource docs.

## Import section structure (resources only)

Resources with Resource Identity follow this exact structure, in this order:

1. Lead-in `In Terraform v1.12.0 and later, the [\`import\` block](...)` followed by an example using `identity = { ... }`.
2. `### Identity Schema` with `#### Required` and `#### Optional` subsections.
3. Lead-in `In Terraform v1.5.0 and later, use an [\`import\` block](...)` followed by an example using `id = "..."`.
4. Lead-in `Using \`terraform import\`, import ...` followed by a `console` block.

Flag pages whose Import section is missing any of these subsections, or that order them differently.

## Identity Schema content

Required attributes by identity type:

- **ARN identity**: a single `arn` bullet.
- **Parameterized identity**: one bullet per identity attribute.
- **Singleton identity**: no required attributes (the subsection still appears, empty).

Optional attributes are these two lines, verbatim:

```
* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.
```

For **global** resources (CloudFront, IAM, etc.), omit the `region` line.

ARN identity has no Optional attributes — flag pages that add `account_id` / `region` under Optional for an ARN-identified resource.

## List resource docs

List resource docs do not have an Import section. Examples use the `list "..."` configuration block:

```terraform
list "aws_<svc>_<thing>" "example" {
  provider = aws
}
```

The `region` argument is documented as `(Optional) Region to query. Defaults to provider region.`
