---
subcategory: "IAM (Identity & Access Management)"
layout: "aws"
page_title: "AWS: aws_iam_role_policies"
description: |-
  Terraform data source that lists the names of the inline policies that are embedded in the specified IAM role.
---

# Data Source: aws_iam_role_policies

Terraform data source that lists the names of the inline policies that are embedded in the specified IAM role.

## Example Usage

### Basic Usage

```terraform
resource "aws_iam_role" "example" {
  name = "example"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "ec2.amazonaws.com"
        }
      },
    ]
  })
}

resource "aws_iam_role_policy" "example" {
  role = aws_iam_role.example.name
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

data "aws_iam_role_policies" "example" {
  depends_on = [aws_iam_role_policy.example]

  role = aws_iam_role.example.name
}
```

## Argument Reference

The following arguments are required:

* `role` - (Required) Friendly name of the role to list inline policies for.

## Attributes Reference

* `names` - Set of policy names embedded in the role.

[1]: https://awscli.amazonaws.com/v2/documentation/api/latest/reference/iam/list-role-policies.html
