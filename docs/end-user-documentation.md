<!-- Copyright IBM Corp. 2014, 2026 -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

# End User Documentation

## Code Structure

All end user documentation is found in the `/website` folder of the repository.

```
├── website/docs
│   ├── actions/               # Documentation for actions
│   ├── d/                     # Documentation for data sources
│   ├── ephemeral-resources/   # Documentation for ephemeral resources
│   ├── function/              # Documentation for provider functions
│   ├── guides/                # Long format guides for provider level configuration or provider upgrades
│   ├── index.html.markdown    # Home page and all provider level documentation, including provder configuration
│   ├── list-resources/        # Documentation for list resources
│   └── r/                     # Documentation for resources
└── examples/                  # Large example configurations
```

## Guidelines

Follow these guidelines to keep [provider documentation](https://registry.terraform.io/providers/hashicorp/aws/latest/docs) consistent. Unless noted otherwise, _resource_ refers to resources, data sources, list resources, ephemeral resources and provider functions.

### Examples

Each resource must include a at least one example Terraform configuration.

- Examples must be functional.
- Examples that includes several resource definitions should be added to the repository `examples` directory instead of an individual resource documentation page. Each directory under `examples` should be self-contained; if a user run `terraform plan` on the example, no errors should be returned.
- Terraform configuration should use `hcl` code fences. Do not use `terraform` code fences.
- Examples should not define `terraform` or `provider` blocks.
- Examples should refrain from highlighting particular Terraform configuration language syntax workarounds or features such as `variable`, `local`, `count`, and built-in functions.
- Generally the resource instance name should simply be `example`, e.g. `resource "aws_instance" "example"`.
- All name arguments within the example configuration should use simple example values that match the resource being defined. Where attribute validation allows, prefer values prefixed with `example-`, e.g. `name = "example-instance"`. Avoid overly complex naming.
- Examples do not need to include every argument. A basic example should use the same configuration as the resource's basic acceptance test.

### Arguments

Every argument must be documented.

#### Ordering

In documentation, the order of arguments is:

1. Any arguments that make up the resource's identity.
1. Required arguments, sorted alphabetically.
1. Optional arguments, sorted alphabetically.

#### Description

Every argument must have a description.

- Descriptions are concise.
- Information about AWS service features and valid argument values that are likely to update over time should link to AWS service user guides and API references where possible.
- If an argument has validation allowing only specific inputs, these must be documented. For example:
  - `` Allowed values are: `value1`, `value2`, and `value3`. ``.
  - `` Valid value is between `0` and `100`. ``
- If the argument has a default value, this must be documented, e.g. `` Default value: `ENABLED`. ``.

#### Blocks

Each block argument must have two entries in the documentation:

1. The initial entry, with a link to the subsection describing the block's arguments. e.g. `` `ip_rule` - (Optional) IP rules. See [ip_rule](#ip-rule) below. ``.
1. A subsection, added after all top-level arguments. If the resource has multiple blocks, these subsections should be ordered alphabetically.
1. Within the subsection, arguments follow the ordering rules above.

For example:

```
## Argument Reference

* `name` - (Required) Name of the thing.
* `ip_rule` - (Optional) IP rules. See [ip_rule](#ip-rule) below.

### `ip_rule`

`ip_rule` supports the following:

* `ip_range` - (Required) IP range of the rule.
* `description` - (Optional) Description of the rule.
```

### Attributes

### Notes

Note blocks provide information beyond the basic description of a resource, argument or attribute.
Notes follow the format (`(->|~>|!>) **Note:**`). Level of importance is documented below.

#### Informational Note

Provides additional useful information, recommendations and/or tips to the user.

Use the `-> **Note:**` format. The Terraform registry will template this note as a block with an info icon.

For example:

```markdown
-> **Note:** The `activation_code` argument cannot be imported.
```

#### Warning Note

Provides information that the user will need to avoid certain errors. These errors are non-breaking and do not cause irreversable changes.

Use the `~> **Note:**` format. The Terraform registry will template this note as a block with a warning icon.

For example:

```markdown
~> **Note:** All arguments including the username and password will be stored in the raw state as plain-text.
```

#### Caution Note

Provides critical information on potential irreversible changes, including data loss and other negative effects.

Use the `!> **Note:**` format. The Terraform registry will template this note as a block with a caution icon.

For example:

```markdown
!> **Note:** This will destroy and recreate the table, possibly resulting in data loss.
```
