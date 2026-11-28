---
subcategory: "IAM (Identity & Access Management)"
layout: "aws"
page_title: "AWS: aws_iam_openid_connect_provider"
description: |-
  Lists IAM OpenID Connect provider resources.
---

# List Resource: aws_iam_openid_connect_provider

Lists IAM OpenID Connect provider resources.

## Example Usage

```terraform
list "aws_iam_openid_connect_provider" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `account` - (Optional) AWS account ID to query. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
