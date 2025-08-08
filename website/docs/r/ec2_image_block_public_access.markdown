---
subcategory: "EC2 (Elastic Compute Cloud)"
layout: "aws"
page_title: "AWS: aws_ec2_image_block_public_access"
description: |-
  Provides a regional public access block for AMIs. This prevents AMIs from being made publicly accessible.
---

# Resource: aws_ec2_image_block_public_access

Provides a regional public access block for AMIs. This prevents AMIs from being made publicly accessible.
If you already have public AMIs, they will remain publicly available.

~> **NOTE:** Deleting this resource does not change the block public access value, the resource in simply removed from state instead.

## Example Usage

```terraform
# Prevent making AMIs publicly accessible in the region and account for which the provider is configured
resource "aws_ec2_image_block_public_access" "test" {
  state = "block-new-sharing"
}
```

## Argument Reference

This resource supports the following arguments:

* `state` - (Required) The state of block public access for AMIs at the account level in the configured AWS Region. Valid values: `unblocked` and `block-new-sharing`.

## Attribute Reference

This resource exports no additional attributes.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `update` - (Default `10m`)

## Import

You cannot import this resource.
