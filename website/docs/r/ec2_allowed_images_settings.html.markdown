---
subcategory: "EC2 (Elastic Compute Cloud)"
layout: "aws"
page_title: "AWS: aws_ec2_allowed_images_settings"
description: |-
  Manages EC2 Allowed Images Settings for an AWS account.
---

# Resource: aws_ec2_allowed_images_settings

Manages EC2 Allowed Images Settings for an AWS account. This feature allows you to control which AMIs can be used to launch EC2 instances in your account based on specified criteria.

~> **NOTE:** The AWS API does not delete this resource. When you run `destroy`, the provider will attempt to disable the setting.

~> **NOTE:** There is only one Allowed Images Settings configuration per AWS account and region. Creating this resource will configure the account-level settings.

## Example Usage

### Enable with Amazon AMIs only

```terraform
resource "aws_ec2_allowed_images_settings" "example" {
  state = "enabled"

  image_criteria {
    image_providers = ["amazon"]
  }
}
```

### Enable audit mode with specific account IDs

```terraform
resource "aws_ec2_allowed_images_settings" "example" {
  state = "audit-mode"

  image_criteria {
    image_providers = ["amazon", "123456789012"]
  }
}
```

## Argument Reference

This resource supports the following arguments:

- `state` - (Required) State of the Allowed Images Settings. Valid values are `disabled`, `enabled`, or `audit-mode`.
- `image_criteria` - (Optional) List of image criteria blocks. Maximum of 10 criteria blocks allowed. See [`image_criteria`](#image_criteria) below.

### `image_criteria`

The `image_criteria` block supports the following:

- `image_names` - (Optional) Set of AMI name patterns to allow. Maximum of 50 names. Each name must be between 1 and 128 characters and can contain alphanumeric characters, hyphens, underscores, periods, forward slashes, question marks, square brackets, at signs, apostrophes, parentheses, asterisks, and word characters.
- `image_providers` - (Optional) Set of image providers to allow. Maximum of 200 providers. Valid values include `amazon`, `aws-marketplace`, `aws-backup-vault`, `none`, or a 12-digit AWS account ID.
- `marketplace_product_codes` - (Optional) Set of AWS Marketplace product codes to allow. Maximum of 50 product codes. Each code must be between 1 and 25 alphanumeric characters.
- `creation_date_condition` - (Optional) Condition based on AMI creation date. See [`creation_date_condition`](#creation_date_condition) below.
- `deprecation_time_condition` - (Optional) Condition based on AMI deprecation time. See [`deprecation_time_condition`](#deprecation_time_condition) below.

### `creation_date_condition`

The `creation_date_condition` block supports the following:

- `maximum_days_since_created` - (Optional) Maximum number of days since the AMI was created. AMIs older than this will not be allowed.

### `deprecation_time_condition`

The `deprecation_time_condition` block supports the following:

- `maximum_days_since_deprecated` - (Optional) Maximum number of days since the AMI was deprecated. AMIs deprecated longer than this will not be allowed.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

- `state` - Current state of the Allowed Images Settings.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import EC2 Allowed Images Settings. Since there is only one configuration per account and region, no ID is required. For example:

```terraform
import {
  to = aws_ec2_allowed_images_settings.example
  id = "not-used"
}
```

Using `terraform import`, import EC2 Allowed Images Settings. For example:

```console
% terraform import aws_ec2_allowed_images_settings.example not-used
```
