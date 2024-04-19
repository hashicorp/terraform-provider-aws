---
subcategory: "IAM (Identity & Access Management)"
layout: "aws"
page_title: "AWS: aws_iam_group_policy"
description: |-
  Provides an IAM policy attached to a group.
---

# Resource: aws_iam_group_policy

Provides an IAM policy attached to a group.

~> **NOTE:** We suggest using [`jsonencode()`](https://developer.hashicorp.com/terraform/language/functions/jsonencode) or [`aws_iam_policy_document`](/docs/providers/aws/d/iam_policy_document.html) when assigning a value to `policy`. They seamlessly translate Terraform language into JSON, enabling you to maintain consistency within your configuration without the need for context switches. Also, you can sidestep potential complications arising from formatting discrepancies, whitespace inconsistencies, and other nuances inherent to JSON.

## Example Usage

```terraform
resource "aws_iam_group_policy" "my_developer_policy" {
  name  = "my_developer_policy"
  group = aws_iam_group.my_developers.name

  # Terraform's "jsonencode" function converts a
  # Terraform expression result to valid JSON syntax.
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
          "ec2:Describe*",
        ]
        Effect   = "Allow"
        Resource = "*"
      },
    ]
  })
}

resource "aws_iam_group" "my_developers" {
  name = "developers"
  path = "/users/"
}
```

## Argument Reference

This resource supports the following arguments:

* `policy` - (Required) The policy document. This is a JSON formatted string. For more information about building IAM policy documents with Terraform, see the [AWS IAM Policy Document Guide](https://learn.hashicorp.com/terraform/aws/iam-policy)
* `name` - (Optional) The name of the policy. If omitted, Terraform will
assign a random, unique name.
* `name_prefix` - (Optional) Creates a unique name beginning with the specified
  prefix. Conflicts with `name`.
* `group` - (Required) The IAM group to attach to the policy.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The group policy ID.
* `group` - The group to which this policy applies.
* `name` - The name of the policy.
* `policy` - The policy document attached to the group.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import IAM Group Policies using the `group_name:group_policy_name`. For example:

```terraform
import {
  to = aws_iam_group_policy.mypolicy
  id = "group_of_mypolicy_name:mypolicy_name"
}
```

Using `terraform import`, import IAM Group Policies using the `group_name:group_policy_name`. For example:

```console
% terraform import aws_iam_group_policy.mypolicy group_of_mypolicy_name:mypolicy_name
```
