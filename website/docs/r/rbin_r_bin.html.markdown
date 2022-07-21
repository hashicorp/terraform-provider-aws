---
subcategory: "RBin"
layout: "aws"
page_title: "AWS: aws_rbin_rbin"
description: |-
  Terraform resource for managing an AWS RBin RBin.
---

# Resource: aws_rbin_rbin

Terraform resource for managing an AWS RBin RBin.

## Example Usage

### Basic Usage

```terraform
resource "aws_rbin_rbin" "example" {
}
```

## Argument Reference

The following arguments are required:

* `example_arg` - (Required) Concise argument description.

The following arguments are optional:

* `optional_arg` - (Optional) Concise argument description.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of the RBin.
* `example_attribute` - Concise description.

## Timeouts

`aws_rbin_rbin` provides the following [Timeouts](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts) configuration options:

* `create` - (Optional, Default: `60m`)
* `update` - (Optional, Default: `180m`)
* `delete` - (Optional, Default: `90m`)

## Import

RBin RBin can be imported using the `example_id_arg`, e.g.,

```
$ terraform import aws_rbin_rbin.example rft-8012925589
```
