---
subcategory: "Global Accelerator"
layout: "aws"
page_title: "AWS: aws_globalaccelerator_custom_routing_accelerator"
description: |-
  Provides a Global Accelerator custom routing accelerator.
---

# Resource: aws_globalaccelerator_custom_routing_accelerator

Creates a Global Accelerator custom routing accelerator.

## Example Usage

```terraform
resource "aws_globalaccelerator_custom_routing_accelerator" "example" {
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

* `attributes` - (Optional) Configuration block with accelerator attributes. See [`attributes`](#attributes-block) below.
* `enabled` - (Optional) Whether the accelerator is enabled. Defaults to `true`. Valid values: `true`, `false`.
* `ip_address_type` - (Optional) IP address type that an accelerator supports. For a custom routing accelerator, the value must be `"IPV4"`.
* `ip_addresses` - (Optional) IP addresses to use for BYOIP accelerators. If not specified, the service assigns IP addresses. Valid values: 1 or 2 IPv4 addresses.
* `name` - (Required) Name of a custom routing accelerator.
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `attributes` Block

* `flow_logs_enabled` - (Optional) Whether flow logs are enabled. Defaults to `false`. Valid values: `true`, `false`.
* `flow_logs_s3_bucket` - (Optional) Name of the Amazon S3 bucket for the flow logs. Required if `flow_logs_enabled` is `true`.
* `flow_logs_s3_prefix` - (Optional) Prefix for the location in the Amazon S3 bucket for the flow logs. Required if `flow_logs_enabled` is `true`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the custom accelerator.
* `dns_name` - DNS name of the accelerator. For example, `a5d53ff5ee6bca4ce.awsglobalaccelerator.com`.
* `hosted_zone_id` - Global Accelerator Route 53 zone ID used to route an [Alias Resource Record Set](https://docs.aws.amazon.com/Route53/latest/APIReference/API_AliasTarget.html) to the accelerator. This is an alias for the zone ID `Z2BJ6XQ5FK7U4H`.
* `id` - ARN of the custom accelerator.
* `ip_sets` - IP address set associated with the accelerator. See [`ip_sets`](#ip_sets-block) below.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

### `ip_sets` Block

* `ip_addresses` - List of IP addresses in the IP address set.
* `ip_family` - Type of IP addresses included in this IP set.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_globalaccelerator_custom_routing_accelerator.example
  identity = {
    "arn" = "arn:aws:globalaccelerator::123456789012:accelerator/1234abcd-abcd-1234-abcd-1234abcdefgh"
  }
}

resource "aws_globalaccelerator_custom_routing_accelerator" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

- `arn` (String) ARN of the Global Accelerator custom routing accelerator.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Global Accelerator custom routing accelerators using the `arn`. For example:

```terraform
import {
  to = aws_globalaccelerator_custom_routing_accelerator.example
  id = "arn:aws:globalaccelerator::111111111111:accelerator/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
}
```

Using `terraform import`, import Global Accelerator custom routing accelerators using the `arn`. For example:

```console
% terraform import aws_globalaccelerator_custom_routing_accelerator.example arn:aws:globalaccelerator::111111111111:accelerator/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
```
