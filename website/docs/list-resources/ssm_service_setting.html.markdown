---
subcategory: "SSM (Systems Manager)"
layout: "aws"
page_title: "AWS: aws_ssm_service_setting"
description: |-
  Lists SSM (Systems Manager) Service Setting resources.
---

# List Resource: aws_ssm_service_setting

Lists SSM (Systems Manager) Service Setting resources.

An SSM Service Setting is a per-account, per-Region singleton, and the AWS API provides no operation to list them.
This list resource always returns a single result for the setting with the specified ARN.

## Example Usage

```terraform
list "aws_ssm_service_setting" "example" {
  provider = aws

  config {
    arn = "arn:aws:ssm:us-east-1:123456789012:servicesetting/ssm/parameter-store/high-throughput-enabled"
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `arn` - (Required) ARN of the SSM Service Setting to list.
* `region` - (Optional) Region to query. Defaults to provider region.
