---
subcategory: "IAM (Identity & Access Management)"
layout: "aws"
page_title: "AWS: aws_iam_user_policy_attachment"
description: |-
  Lists IAM (Identity & Access Management) User Policy Attachment resources.
---

# List Resource: aws_iam_user_policy_attachment

Lists IAM (Identity & Access Management) User Policy Attachment resources.

## Example Usage

```terraform
list "aws_iam_user_policy_attachment" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `account` - (Optional) AWS account ID to query. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
