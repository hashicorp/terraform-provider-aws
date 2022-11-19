---
subcategory: "IAM (Identity & Access Management)"
layout: "aws"
page_title: "AWS: aws_iam_user_policies"
description: |-
  Terraform data source that lists the names of the inline policies that are embedded in the specified IAM user.
---

# Data Source: aws_iam_user_policies

Terraform data source that lists the names of the inline policies that are embedded in the specified IAM user.

## Example Usage

### Basic Usage

```terraform
resource "aws_iam_user" "example" {
  name = "example"
}

resource "aws_iam_user_policy" "example" {
  user = aws_iam_user.example.name
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect   = "Allow"
        Action   = "ec2:Describe*"
        Resource = "*"
      },
    ],
  })
}

data "aws_iam_user_policies" "example" {
  depends_on = [aws_iam_user_policy.example]

  user = aws_iam_user.example.name
}
```

## Argument Reference

The following arguments are required:

* `user` - (Required) Friendly name of the user to list inline policies for.

## Attributes Reference

* `names` - Set of policy names embedded in the user.

[1]: https://awscli.amazonaws.com/v2/documentation/api/latest/reference/iam/list-user-policies.html
