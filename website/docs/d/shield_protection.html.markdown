---
subcategory: "Shield"
layout: "aws"
page_title: "AWS: aws_shield_protection"
description: |-
  Terraform data source for managing an AWS Shield Protection.
---

# Data Source: aws_shield_protection

Terraform data source for managing an AWS Shield Protection.

## Example Usage

### Basic Usage

```terraform
data "aws_shield_protection" "example" {
  protection_id = "abc123"
}
```

### By Resource ARN

```terraform
data "aws_shield_protection" "example" {
  resource_arn = "arn:aws:globalaccelerator::123456789012:accelerator/1234abcd-abcd-1234-abcd-1234abcdefgh"
}
```

## Argument Reference

~> Exactly one of `protection_id` or `resource_arn` is required.

The following arguments are optional:

* `protection_id` - (Optional) Unique identifier for the protection.
* `resource_arn` - (Optional) ARN (Amazon Resource Name) of the resource being protected.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `name` - Name of the protection.
* `protection_arn` - ARN of the protection.
