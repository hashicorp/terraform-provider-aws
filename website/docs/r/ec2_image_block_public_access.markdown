---
subcategory: "EC2 (Elastic Compute Cloud)"
layout: "aws"
page_title: "AWS: aws_ec2_image_block_public_access"
description: |-
  Provides a regional public access block for AMIs. This prevents AMIs from being made publicly accessible.
---

# Resource: aws_ec2_image_block_public_access

Provides a regional public access block for AMIs. This prevents AMIs from being made publicly accessible.

**Note:** if you already have public AMIs, they will remain publicly available.

## Example Usage

```terraform
# Prevent making AMIs publicly accessible in the region and account for which the provider is configured
resource "aws_ec2_image_block_public_access" "test" {
  enabled = true
}
```

## Argument Reference

This resource supports the following arguments:

* `enabled` - (Optional) Whether to prevent the public sharing of your AMI

## Attribute Reference

This resource exports no additional attributes.

## Import

You cannot import this resource.
