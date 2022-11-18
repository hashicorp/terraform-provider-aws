---
subcategory: "IAM (Identity & Access Management)"
layout: "aws"
page_title: "AWS: aws_iam_attached_user_policies"
description: |-
  Terraform data source that lists all managed policies that are attached to the specified IAM user.
---

# Data Source: aws_iam_attached_user_policies

Terraform data source that lists all managed policies that are attached to the specified IAM user.

## Example Usage

### Basic Usage

```terraform
resource "aws_iam_user" "example" {
  name = "example"
}

resource "aws_iam_policy" "example" {
  name        = "example"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect   = "Allow"
        Action = "ec2:Describe*"
        Resource = "*"
      },
    ],
  })
}

resource "aws_iam_user_policy_attachment" "example" {
  user       = aws_iam_user.example.name
  policy_arn = aws_iam_policy.example.arn
}

data "aws_iam_attached_user_policies" "example" {
  user = aws_iam_user.example.name
  path_prefix = "/"
}
```

## Argument Reference

The following arguments are required:

* `user` - (Required) Name (not ARN) of the user to list attached policies for.

The following arguments are optional:

* `path_prefix` - (Optional) Path prefix for filtering results. Defaults to a slash (`/`), which will list all attached policies.

## Attributes Reference

* `arns` - Set of policy ARNs attached to the user.
* `names` - Set of policy names attached to the user.

[1]: https://awscli.amazonaws.com/v2/documentation/api/latest/reference/iam/list-attached-user-policies.html
