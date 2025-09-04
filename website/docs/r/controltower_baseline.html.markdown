---
subcategory: "Control Tower"
layout: "aws"
page_title: "AWS: aws_controltower_baseline"
description: |-
  Terraform resource for managing an AWS Control Tower Baseline.
---

# Resource: aws_controltower_baseline

Terraform resource for managing an AWS Control Tower Baseline.

## Example Usage

### Basic Usage

```terraform
resource "aws_controltower_baseline" "example" {
  baseline_identifier = "arn:aws:controltower:us-east-1::baseline/17BSJV3IGJ2QSGA2"
  baseline_version    = "4.0"
  target_identifier   = aws_organizations_organizational_unit.test.arn
  parameters {
    key   = "IdentityCenterEnabledBaselineArn"
    value = "arn:aws:controltower:us-east-1:664418989480:enabledbaseline/XALULM96QHI525UOC"
  }
}
```

## Argument Reference

The following arguments are required:

* `example_arg` - (Required) Concise argument description. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.

The following arguments are optional:

* `optional_arg` - (Optional) Concise argument description. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Baseline. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.
* `example_attribute` - Concise description. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `60m`)
* `update` - (Default `180m`)
* `delete` - (Default `90m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Control Tower Baseline using the `example_id_arg`. For example:

```terraform
import {
  to = aws_controltower_baseline.example
  id = "baseline-id-12345678"
}
```

Using `terraform import`, import Control Tower Baseline using the `example_id_arg`. For example:

```console
% terraform import aws_controltower_baseline.example baseline-id-12345678
```
