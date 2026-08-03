---
subcategory: "EC2 (Elastic Compute Cloud)"
layout: "aws"
page_title: "AWS: aws_ec2_allowed_images_settings"
description: |-
  Provides EC2 allowed images settings.
---

# Resource: aws_ec2_allowed_images_settings

Provides EC2 allowed images settings for an AWS account. This feature allows you to control which AMIs can be used to launch EC2 instances in your account based on specified criteria.

For more information about the image criteria that can be set, see the [AWS documentation on Allowed AMIs JSON configuration](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-allowed-amis.html#allowed-amis-json-configuration).

~> **NOTE:** The AWS API does not delete this resource. When you run `destroy`, the provider will attempt to disable the setting.

~> **NOTE:** There is only one allowed images settings configuration per AWS account and region. Creating this resource will configure the account-level settings.

## Example Usage

### Enable with Amazon AMIs only

```terraform
resource "aws_ec2_allowed_images_settings" "example" {
  state = "enabled"

  image_criterion {
    image_providers = ["amazon"]
  }
}
```

### Enable audit mode with specific account IDs

```terraform
resource "aws_ec2_allowed_images_settings" "example" {
  state = "audit-mode"

  image_criterion {
    image_providers = ["amazon", "123456789012"]
  }
}
```

## Argument Reference

This resource supports the following arguments:

- `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
- `state` - (Required) State of the allowed images settings. Valid values are `enabled` or `audit-mode`.
- `image_criterion` - (Optional) List of image criteria. Maximum of 10 criterion blocks allowed. See [`image_criterion`](#image_criterion) below.

### `image_criterion`

The `image_criterion` block supports the following:

- `image_names` - (Optional) Set of AMI name patterns to allow. Maximum of 50 names.
- `image_providers` - (Optional) Set of image providers to allow. Maximum of 200 providers. Valid values include `amazon`, `aws-marketplace`, `aws-backup-vault`, `none`, or a 12-digit AWS account ID.
- `marketplace_product_codes` - (Optional) Set of AWS Marketplace product codes to allow. Maximum of 50 product codes.
- `creation_date_condition` - (Optional) Condition based on AMI creation date. See [`creation_date_condition`](#creation_date_condition) below.
- `deprecation_time_condition` - (Optional) Condition based on AMI deprecation time. See [`deprecation_time_condition`](#deprecation_time_condition) below.

### `creation_date_condition`

The `creation_date_condition` block supports the following:

- `maximum_days_since_created` - (Required) Maximum number of days since the AMI was created.

### `deprecation_time_condition`

The `deprecation_time_condition` block supports the following:

- `maximum_days_since_deprecated` - (Required) Maximum number of days since the AMI was deprecated. Setting this to `0` means no deprecated images are allowed.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import EC2 allowed images settings. Since there is only one configuration per account and region, region is used as the resource ID. For example:

```terraform
import {
  to = aws_ec2_allowed_images_settings.example
  id = "us-east-1"
}
```

Using `terraform import`, import EC2 allowed images settings. For example:

```console
% terraform import aws_ec2_allowed_images_settings.example us-east-1
```
