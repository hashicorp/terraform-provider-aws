---
subcategory: "CloudWatch Logs"
layout: "aws"
page_title: "AWS: aws_cloudwatch_log_storage_tier_policy"
description: |-
  Manages a CloudWatch Logs account-level storage tier policy.
---

# Resource: aws_cloudwatch_log_storage_tier_policy

Manages a CloudWatch Logs account-level storage tier policy. When set to `INTELLIGENT_TIERING`, CloudWatch Logs automatically moves log data to the most cost-effective storage tier based on access frequency.

## Example Usage

### Basic Usage

```terraform
resource "aws_cloudwatch_log_storage_tier_policy" "example" {
  storage_tier = "INTELLIGENT_TIERING"
}
```

### Standard Storage Tier

```terraform
resource "aws_cloudwatch_log_storage_tier_policy" "example" {
  storage_tier = "STANDARD"
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `storage_tier` - (Required) Storage tier to set for the account. Valid values are `STANDARD` or `INTELLIGENT_TIERING`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `region` - AWS region where the storage tier policy is configured.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

Import CloudWatch Logs Storage Tier Policy using the region. For example:

```terraform
import {
  to = aws_cloudwatch_log_storage_tier_policy.example
  id = "us-west-2"
}
```

**CLI:**

```console
% terraform import aws_cloudwatch_log_storage_tier_policy.example us-west-2
```

### Identity Schema

Import using the identity configuration:

```terraform
import {
  to = aws_cloudwatch_log_storage_tier_policy.example
  identity = {
    region = "us-west-2"
  }
}
```

This is a regional singleton resource - only one storage tier policy can exist per AWS account per region. When this resource is destroyed, the storage tier policy is reset to `STANDARD` (the default state). The storage tier policy applies to all log groups in the account within the specified region. Setting the policy to `INTELLIGENT_TIERING` enables automatic cost optimization by moving log data to appropriate storage tiers based on access frequency.
