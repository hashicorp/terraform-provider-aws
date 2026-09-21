---
subcategory: "IAM (Identity & Access Management)"
layout: "aws"
page_title: "AWS: aws_iam_user_policy"
description: |-
  Lists IAM (Identity & Access Management) User Policy resources.
---

# List Resource: aws_iam_user_policy

Lists IAM (Identity & Access Management) User Policy resources.

## Example Usage

```terraform
list "aws_iam_user_policy" "example" {
  provider = aws
  config {
    user_name = aws_iam_user.example.name
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `user_name` - (Required) Name of the IAM user.
