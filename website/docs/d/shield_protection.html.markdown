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

This data source supports the following arguments:

* `account` - (Optional) AWS account ID where this data source will be read. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `protection_id` - (Optional) Unique identifier for the protection.
* `resource_arn` - (Optional) ARN of the resource being protected.

~> Exactly one of `protection_id` or `resource_arn` is required.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `name` - Name of the protection.
* `protection_arn` - ARN of the protection.
