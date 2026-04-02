---
subcategory: "IAM (Identity & Access Management)"
layout: "aws"
page_title: "AWS: aws_iam_role_policies"
description: |-
  Get the names of inline policies associated with an IAM role.
---

# Data Source: aws_iam_role_policies

Use this data source to get the names of inline policies associated with an IAM role.

## Example Usage

```terraform
data "aws_iam_role_policies" "example" {
  role_name = "my-role-name"
}
```

## Argument Reference

This data source supports the following arguments:

* `role_name` - (Required) Name of the IAM role.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `policy_names` - Set of inline policy names associated with the role.
