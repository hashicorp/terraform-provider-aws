---
subcategory: "CloudWatch Logs"
layout: "aws"
page_title: "AWS: aws_cloudwatch_log_storage_tier_policy"
description: |-
  Manages a CloudWatch Logs account-level storage tier policy.
---

# Resource: aws_cloudwatch_log_storage_tier_policy

Manages a CloudWatch Logs account-level storage tier policy. When set to `INTELLIGENT_TIERING`, CloudWatch Logs automatically moves log data to the most cost-effective storage tier based on access frequency.

~> Deletion of this resource will reset the storage tier policy to `STANDARD` (the default state).

## Example Usage

### Basic Usage

```terraform
resource "aws_cloudwatch_log_storage_tier_policy" "example" {
  storage_tier = "INTELLIGENT_TIERING"
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `storage_tier` - (Required) Storage tier to set for the account. Valid values are `STANDARD` or `INTELLIGENT_TIERING`.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_cloudwatch_log_storage_tier_policy.example
  identity = {
    region = "us-west-2"
  }
}

resource "aws_cloudwatch_log_storage_tier_policy" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `region` (String) Region where this resource is managed.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import storage tier policies using `region`. For example:

```terraform
import {
  to = aws_cloudwatch_log_storage_tier_policy.example
  id = "us-west-2"
}
```

Using `terraform import`, import storage tier policies using `region`. For example:

```console
% terraform import aws_cloudwatch_log_storage_tier_policy.example us-west-2
```
