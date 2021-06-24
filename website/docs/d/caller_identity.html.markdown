---
subcategory: ""
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

There are no arguments available for this data source.

## Attributes Reference

* `account_id` - AWS Account ID number of the account that owns or contains the calling entity.
* `arn` - ARN associated with the calling entity.
* `id` - Account ID number of the account that owns or contains the calling entity.
* `source_arn` - Same as `arn` unless `arn` is an assumed role ARN. In that case, `source_arn` is the ARN of the source role of the assumed role. For example, if `arn` is `arn:aws:sts::123456789012:assumed-role/example-role/AWSCLI-Session`, which corresponds to an assumed role, `source_arn` would be `arn:aws:iam::123456789012:role/example-role`.
* `user_id` - Unique identifier of the calling entity.
