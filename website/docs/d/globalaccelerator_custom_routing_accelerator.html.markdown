---
subcategory: "Global Accelerator"
layout: "aws"
page_title: "AWS: aws_globalaccelerator_custom_routing_accelerator"
description: |-
  Provides a Global Accelerator custom routing accelerator data source.
---

# Data Source: aws_globalaccelerator_custom_routing_accelerator

Provides information about a Global Accelerator custom routing accelerator.

## Example Usage

```terraform
variable "accelerator_arn" {
  type    = string
  default = ""
}

variable "accelerator_name" {
  type    = string
  default = ""
}

data "aws_globalaccelerator_custom_routing_accelerator" "example" {
  arn  = var.accelerator_arn
  name = var.accelerator_name
}
```

## Argument Reference

This data source supports the following arguments:

* `arn` - (Optional) Full ARN of the custom routing accelerator.
* `name` - (Optional) Unique name of the custom routing accelerator.

~> **NOTE:** When both `arn` and `name` are specified, `arn` takes precedence.

## Attribute Reference

See the [`aws_globalaccelerator_custom_routing_accelerator` resource](/docs/providers/aws/r/globalaccelerator_custom_routing_accelerator.html) for details on the
returned attributes - they are identical.
