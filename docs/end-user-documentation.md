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
│   ├── index.html.markdown    # Home page and all provider level documentation, including provider configuration
│   ├── list-resources/        # Documentation for list resources
│   └── r/                     # Documentation for resources
└── examples/                  # Large example configurations
```

## Guidelines

Follow these guidelines to keep [provider documentation](https://registry.terraform.io/providers/hashicorp/aws/latest/docs) consistent. Unless noted otherwise, _resource_ refers to resources, data sources, list resources, ephemeral resources and provider functions.

### Examples

Each resource must include at least one example Terraform configuration.

- Examples must be functional.
- Examples that includes several resource definitions should be added to the repository `examples` directory instead of an individual resource documentation page. Each directory under `examples` should be self-contained; if a user run `terraform plan` on the example, no errors should be returned.
- Terraform configuration should use `terraform` code fences. Do not use `hcl` code fences.
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

#### Bylines

Valid bylines are:

- "This resource supports the following arguments:" if required and optional arguments are in one list, or
- "The following arguments are required:" and "The following arguments are optional:" if required and optional arguments are split into separate lists.

#### Description

Every argument must have a description.

- Descriptions are concise and follow US English spelling and grammar rules.
- Descriptions start with a verb or noun — never "A ", "An ", "The ", "Specifies ", "Indicates ".
- Boolean argument descriptions start with "Whether to ".
- Don't spell out common abbreviations. See [Abbreviations](#abbreviations) below.
- Information about AWS service features and valid argument values that are likely to update over time should link to AWS service user guides and API references where possible.
- If an argument has validation allowing only specific inputs, these must be documented. For example:
    - `` Allowed values are: `value1`, `value2`, and `value3`. ``.
    - `` Valid value is between `0` and `100`. ``
- If the argument has a default value, this must be documented, e.g. `` Default value: `ENABLED`. ``.

#### Blocks

Each block argument must have two entries in the documentation:

1. The initial entry, with a link to the subsection describing the block's arguments. e.g. `` `ip_rule` - (Optional) IP rules. See [`ip_rule` Block](#ip_rule-block) below. ``.
1. A subsection, added after all top-level arguments. If the resource has multiple blocks, these subsections should be ordered alphabetically.

Within the subsection, arguments follow the ordering rules above. Additionally any Computed attributes are listed alphabetically after any Optional arguments.

For example:

```
## Argument Reference

* `name` - (Required) Name of the thing.
* `ip_rule` - (Optional) IP rules. See [`ip_rule` Block](#ip_rule-block) below.

### `ip_rule` Block

The `ip_rule` block supports:

* `ip_range` - (Required) IP range of the rule.
* `description` - (Optional) Description of the rule.
* `status` - (Computed) Status of the rule.
```

### Attributes

Every attribute must be documented.

#### Ordering

In documentation, the order of attributes is:

1. Any `id` attribute
1. The remaining attributes, sorted alphabetically.

#### Bylines

Valid bylines are:

- "In addition to all arguments above, the following attributes are exported:", or
- "This resource exports no additional attributes." if no attributes are exported.

#### Description

Every attribute must have a description.

- Descriptions are concise and follow US English spelling and grammar rules.
- Valid values and default values are not documented.
- Don't spell out common abbreviations. See [Abbreviations](#abbreviations) below.

#### Blocks

Each block attribute must have two entries in the documentation:

1. The initial entry, with a link to the subsection describing the block's attributes. e.g. `` `health` - Health status. See [`health`](#health) below. ``.
1. A subsection, added after all top-level arguments. If the resource has multiple blocks, these subsections should be ordered alphabetically.

Within the subsection, arguments follow the ordering rules above.

For example:

```
## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - ID of the thing.
* `health` - Health status. See [`health`](#health) below.

### `health`

`health` supports:

* `status_code` - Health status code.
* `status_message` - Health status message.
```

### Abbreviations

Don't spell out an abbreviation in a description if the abbreviation has effectively become a lexical item — a term readers understand directly rather than by expanding it. Spelling it out can make documentation less readable without improving understanding.

For example, a typical reader understands "ARN" directly; expanding it to "Amazon Resource Name" does not help them understand the concept.

**Test for treating an abbreviation this way:** Does spelling it out materially improve the intended reader's understanding? If not, use the abbreviation.

This applies to argument and attribute descriptions, as well as general prose (e.g. the resource description, `Note` blocks). It does not apply to the first use of a term that a typical reader of AWS provider documentation would not already recognize.

Abbreviations to use instead of spelling out (`swissshepherd`'s `banned_glosses` check enforces this list; see `.ci/swissshepherd-weak.hcl`):

| Abbreviation | Do not spell out as |
| --- | --- |
| AMI | Amazon Machine Image |
| ARN | Amazon Resource Name |
| ARNs | Amazon Resource Names |
| API | Application Programming Interface |
| CPU | Central Processing Unit |
| CLI | Command-Line Interface |
| DNS | Domain Name System |
| EC2 | Elastic Compute Cloud |
| XML | Extensible Markup Language |
| GPU | Graphics Processing Unit |
| HTML | HyperText Markup Language |
| HTTP | Hypertext Transfer Protocol |
| IP | Internet Protocol |
| JSON | JavaScript Object Notation |
| KMS | Key Management Service |
| RDS | Relational Database Service |
| S3 | Simple Storage Service |
| SDK | Software Development Kit |
| SQL | Structured Query Language |
| TCP | Transmission Control Protocol |
| TLS | Transport Layer Security |
| UTF | Unicode Transformation Format |
| URI | Uniform Resource Identifier |
| URL | Uniform Resource Locator |
| USB | Universal Serial Bus |
| VPC | Virtual Private Cloud |
| VPN | Virtual Private Network |
| YAML | YAML Ain't Markup Language |

### Notes

Notes provide information beyond the basic description of a resource, argument or attribute.
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
