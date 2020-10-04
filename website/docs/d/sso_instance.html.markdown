---
subcategory: "SSO Admin"
layout: "aws"
page_title: "AWS: aws_sso_instance"
description: |-
  Get information on an AWS Single Sign-On Instance.
---

# Data Source: aws_sso_instance

Use this data source to get the Single Sign-On Instance ARN and Identity Store ID.

## Example Usage

```hcl
data "aws_sso_instance" "selected" {}

output "arn" {
  value = data.aws_sso_instance.selected.arn
}

output "identity_store_id" {
  value = data.aws_sso_instance.selected.identity_store_id
}
```

## Argument Reference

There are no arguments available for this data source.

## Attributes Reference

* `arn` - The AWS ARN associated with the AWS Single Sign-On Instance.
* `identity_store_id` - The Identity Store ID associated with the AWS Single Sign-On Instance.
