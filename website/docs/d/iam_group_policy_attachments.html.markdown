---
subcategory: "IAM (Identity & Access Management)"
layout: "aws"
page_title: "AWS: aws_iam_group_policy_attachments"
description: |-
  Terraform data source that lists all managed policies that are attached to the specified IAM group.
---

# Data Source: aws_iam_group_policy_attachments

Terraform data source that lists all managed policies that are attached to the specified IAM group.

## Example Usage

### Basic Usage

```terraform
resource "aws_iam_group" "example" {
  name = "example"
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

resource "aws_iam_group_policy_attachment" "example" {
  group      = aws_iam_group.example.name
  policy_arn = aws_iam_policy.example.arn
}

data "aws_iam_group_policy_attachments" "example" {
  group       = aws_iam_group.example.name
  path_prefix = "/"
}
```

## Argument Reference

The following arguments are required:

* `group` - (Required) Name (not ARN) of the group to list attached policies for.

The following arguments are optional:

* `path_prefix` - (Optional) Path prefix for filtering results. Defaults to a slash (`/`), which will list all attached policies.

## Attributes Reference

* `arns` - Set of policy ARNs attached to the group.
* `names` - Set of policy names attached to the group.

[1]: https://awscli.amazonaws.com/v2/documentation/api/latest/reference/iam/list-attached-group-policies.html
