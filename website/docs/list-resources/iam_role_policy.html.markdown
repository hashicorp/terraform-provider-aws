---
subcategory: "IAM (Identity & Access Management)"
layout: "aws"
page_title: "AWS: aws_iam_role_policy"
description: |-
  Lists IAM (Identity & Access Management) Role Policy resources.
---

# List Resource: aws_iam_role_policy

Lists IAM (Identity & Access Management) Role Policy resources.

## Example Usage

```terraform
list "aws_iam_role_policy" "example" {
  provider = aws
  config {
    role_name = aws_iam_role.example.name
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `role_name` - (Required) Name of the IAM role to list policies from.
