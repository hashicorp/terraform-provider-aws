---
subcategory: "IAM (Identity & Access Management)"
layout: "aws"
page_title: "AWS: aws_iam_group_policies"
description: |-
  Terraform data source that lists the names of the inline policies that are embedded in the specified IAM group.
---

# Data Source: aws_iam_group_policies

Terraform data source that lists the names of the inline policies that are embedded in the specified IAM group.

## Example Usage

### Basic Usage

```terraform
resource "aws_iam_group" "example" {
  name = "example"
}

resource "aws_iam_group_policy" "example" {
  group = aws_iam_group.example.name
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

data "aws_iam_group_policies" "example" {
  depends_on = [aws_iam_group_policy.example]

  group = aws_iam_group.example.name
}
```

## Argument Reference

The following arguments are required:

* `group` - (Required) Friendly name of the group to list inline policies for.

## Attributes Reference

* `names` - Set of policy names embedded in the group.

[1]: https://awscli.amazonaws.com/v2/documentation/api/latest/reference/iam/list-group-policies.html
