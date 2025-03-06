---
subcategory: "Global Accelerator"
layout: "aws"
page_title: "AWS: aws_globalaccelerator_accelerator"
description: |-
  Provides a Global Accelerator accelerator.
---

# Resource: aws_globalaccelerator_accelerator

Creates a Global Accelerator accelerator.

## Example Usage

```terraform
resource "aws_globalaccelerator_accelerator" "example" {
  name            = "Example"
  ip_address_type = "IPV4"
  ip_addresses    = ["1.2.3.4"]
  enabled         = true

  attributes {
    flow_logs_enabled   = true
    flow_logs_s3_bucket = "example-bucket"
    flow_logs_s3_prefix = "flow-logs/"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `name` - (Required) The name of the accelerator.
* `ip_address_type` - (Optional) The value for the address type. Defaults to `IPV4`. Valid values: `IPV4`, `DUAL_STACK`.
* `ip_addresses` - (Optional) The IP addresses to use for BYOIP accelerators. If not specified, the service assigns IP addresses. Valid values: 1 or 2 IPv4 addresses.
* `enabled` - (Optional) Indicates whether the accelerator is enabled. Defaults to `true`. Valid values: `true`, `false`.
* `attributes` - (Optional) The attributes of the accelerator. Fields documented below.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

`attributes` supports the following arguments:

* `flow_logs_enabled` - (Optional) Indicates whether flow logs are enabled. Defaults to `false`. Valid values: `true`, `false`.
* `flow_logs_s3_bucket` - (Optional) The name of the Amazon S3 bucket for the flow logs. Required if `flow_logs_enabled` is `true`.
* `flow_logs_s3_prefix` - (Optional) The prefix for the location in the Amazon S3 bucket for the flow logs. Required if `flow_logs_enabled` is `true`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The Amazon Resource Name (ARN) of the accelerator.
* `arn` - The Amazon Resource Name (ARN) of the accelerator.
* `dns_name` - The DNS name of the accelerator. For example, `a5d53ff5ee6bca4ce.awsglobalaccelerator.com`.
* `dual_stack_dns_name` - The Domain Name System (DNS) name that Global Accelerator creates that points to a dual-stack accelerator's four static IP addresses: two IPv4 addresses and two IPv6 addresses. For example, `a1234567890abcdef.dualstack.awsglobalaccelerator.com`.
* `hosted_zone_id` --  The Global Accelerator Route 53 zone ID that can be used to
  route an [Alias Resource Record Set][1] to the Global Accelerator. This attribute
  is simply an alias for the zone ID `Z2BJ6XQ5FK7U4H`.
* `ip_sets` - IP address set associated with the accelerator.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

**ip_sets** exports the following attributes:

* `ip_addresses` - A list of IP addresses in the IP address set.
* `ip_family` - The type of IP addresses included in this IP set.

[1]: https://docs.aws.amazon.com/Route53/latest/APIReference/API_AliasTarget.html

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Global Accelerator accelerators using the `arn`. For example:

```terraform
import {
  to = aws_globalaccelerator_accelerator.example
  id = "arn:aws:globalaccelerator::111111111111:accelerator/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
}
```

Using `terraform import`, import Global Accelerator accelerators using the `arn`. For example:

```console
% terraform import aws_globalaccelerator_accelerator.example arn:aws:globalaccelerator::111111111111:accelerator/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
```
