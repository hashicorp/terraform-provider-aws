---
subcategory: "STS (Security Token)"
layout: "aws"
page_title: "AWS: aws_caller_identity"
description: |-
  Get information about the identity of the caller for the provider
  connection to AWS.
---

# Data Source: aws_caller_identity

Use this data source to get the access to the effective Account ID, User ID, and ARN in
which Terraform is authorized.

## Example Usage

```terraform
data "aws_caller_identity" "current" {}

output "account_id" {
  value = data.aws_caller_identity.current.account_id
}

output "caller_arn" {
  value = data.aws_caller_identity.current.arn
}

output "caller_user" {
  value = data.aws_caller_identity.current.user_id
}
```

## Argument Reference

This data source does not support any arguments.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `account_id` - AWS Account ID number of the account that owns or contains the calling entity.
* `arn` - ARN associated with the calling entity.
* `id` - Account ID number of the account that owns or contains the calling entity.
* `user_id` - Unique identifier of the calling entity.
