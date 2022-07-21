---
subcategory: "RBin"
layout: "aws"
page_title: "AWS: aws_rbin_rbinrule"
description: |-
  Terraform resource for managing an AWS RBin RBinRule.
---

# Resource: aws_rbin_rbinrule

Terraform resource for managing an AWS RBin RBinRule.

## Example Usage

### Basic Usage

```terraform
resource "aws_rbin_rbinrule" "example" {
}
```

## Argument Reference

The following arguments are required:

* `example_arg` - (Required) Concise argument description.

The following arguments are optional:

* `optional_arg` - (Optional) Concise argument description.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of the RBinRule.
* `example_attribute` - Concise description.

## Timeouts

`aws_rbin_rbinrule` provides the following [Timeouts](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts) configuration options:

* `create` - (Optional, Default: `60m`)
* `update` - (Optional, Default: `180m`)
* `delete` - (Optional, Default: `90m`)

## Import

RBin RBinRule can be imported using the `example_id_arg`, e.g.,

```
$ terraform import aws_rbin_rbinrule.example rft-8012925589
```
