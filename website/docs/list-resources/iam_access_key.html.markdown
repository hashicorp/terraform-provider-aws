---
subcategory: "IAM (Identity & Access Management)"
layout: "aws"
page_title: "AWS: aws_iam_access_key"
description: |-
  Lists IAM (Identity & Access Management) Access Key resources.
---

# List Resource: aws_iam_access_key

Lists IAM (Identity & Access Management) Access Key resources.

## Example Usage

```terraform
list "aws_iam_access_key" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `account` - (Optional) AWS account ID to query. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
