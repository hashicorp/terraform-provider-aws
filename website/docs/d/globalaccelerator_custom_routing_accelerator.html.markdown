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

This data source exports the following attributes in addition to the arguments above:

* `attributes` - Attributes of the accelerator. See [`attributes`](#attributes-block) below.
* `dns_name` - DNS name of the accelerator. For example, `a5d53ff5ee6bca4ce.awsglobalaccelerator.com`.
* `enabled` - Whether the accelerator is enabled.
* `hosted_zone_id` - Global Accelerator Route 53 zone ID used to route an [Alias Resource Record Set](https://docs.aws.amazon.com/Route53/latest/APIReference/API_AliasTarget.html) to the accelerator.
* `ip_address_type` - Value for the address type.
* `ip_sets` - IP address set associated with the accelerator. See [`ip_sets`](#ip_sets-block) below.
* `tags` - Map of tags assigned to the resource.

### `attributes` Block

* `flow_logs_enabled` - Whether flow logs are enabled.
* `flow_logs_s3_bucket` - Name of the Amazon S3 bucket for the flow logs.
* `flow_logs_s3_prefix` - Prefix for the location in the Amazon S3 bucket for the flow logs.

### `ip_sets` Block

* `ip_addresses` - List of IP addresses in the IP address set.
* `ip_family` - Type of IP addresses included in this IP set.
