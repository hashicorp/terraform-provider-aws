---
subcategory: "IAM (Identity & Access Management)"
layout: "aws"
page_title: "AWS: aws_iam_attached_role_policies"
description: |-
  Terraform data source that lists all managed policies that are attached to the specified IAM role.
---

# Data Source: aws_iam_attached_role_policies

Terraform data source that lists all managed policies that are attached to the specified IAM role.

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

resource "aws_iam_policy" "example" {
  name = "example"

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

resource "aws_iam_role_policy_attachment" "example" {
  role       = aws_iam_role.example.name
  policy_arn = aws_iam_policy.example.arn
}

data "aws_iam_attached_role_policies" "example" {
  role        = aws_iam_role.example.name
  path_prefix = "/"
}
```

## Argument Reference

The following arguments are required:

* `role` - (Required) Name (not ARN) of the role to list attached policies for.

The following arguments are optional:

* `path_prefix` - (Optional) Path prefix for filtering results. Defaults to a slash (`/`), which will list all attached policies.

## Attributes Reference

* `arns` - Set of policy ARNs attached to the role.
* `names` - Set of policy names attached to the role.

[1]: https://awscli.amazonaws.com/v2/documentation/api/latest/reference/iam/list-attached-role-policies.html
