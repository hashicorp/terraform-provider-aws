---
subcategory: "Recycle Bin (RBin)"
layout: "aws"
page_title: "AWS: aws_rbin_rbinrule"
description: |-
  Terraform data source for managing an AWS RBin RBinRule.
---

# Data Source: aws_rbin_rbinrule

Terraform data source for managing an AWS RBin RBinRule.

## Example Usage

### Basic Usage

```terraform
data "aws_rbin_rbinrule" "example" {
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
