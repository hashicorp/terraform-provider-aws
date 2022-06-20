---
subcategory: "AccessAnalyzer"
layout: "aws"
page_title: "AWS: aws_accessanalyzer_archiverule"
description: |-
  Terraform resource for managing an AWS AccessAnalyzer ArchiveRule.
---

# Resource: aws_accessanalyzer_archiverule

Terraform resource for managing an AWS AccessAnalyzer ArchiveRule.

## Example Usage

### Basic Usage

```terraform
resource "aws_accessanalyzer_archiverule" "example" {
}
```

## Argument Reference

The following arguments are required:

* `example_arg` - (Required) Concise argument description.

The following arguments are optional:

* `optional_arg` - (Optional) Concise argument description.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of the ArchiveRule.
* `example_attribute` - Concise description.

## Timeouts

`aws_accessanalyzer_archiverule` provides the following [Timeouts](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts) configuration options:

* `create` - (Optional, Default: `60m`)
* `update` - (Optional, Default: `180m`)
* `delete` - (Optional, Default: `90m`)

## Import

AccessAnalyzer ArchiveRule can be imported using the `example_id_arg`, e.g.,

```
$ terraform import aws_accessanalyzer_archiverule.example rft-8012925589
```
