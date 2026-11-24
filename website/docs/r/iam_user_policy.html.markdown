---
subcategory: "IAM (Identity & Access Management)"
layout: "aws"
page_title: "AWS: aws_iam_user_policy"
description: |-
  Provides an IAM policy attached to a user.
---

# Resource: aws_iam_user_policy

Provides an IAM policy attached to a user.

~> **NOTE:** We suggest using [`jsonencode()`](https://developer.hashicorp.com/terraform/language/functions/jsonencode) or [`aws_iam_policy_document`](/docs/providers/aws/d/iam_policy_document.html) when assigning a value to `policy`. They seamlessly translate Terraform language into JSON, enabling you to maintain consistency within your configuration without the need for context switches. Also, you can sidestep potential complications arising from formatting discrepancies, whitespace inconsistencies, and other nuances inherent to JSON.

## Example Usage

```terraform
resource "aws_iam_user_policy" "lb_ro" {
  name = "test"
  user = aws_iam_user.lb.name

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

resource "aws_iam_user" "lb" {
  name = "loadbalancer"
  path = "/system/"
}

resource "aws_iam_access_key" "lb" {
  user = aws_iam_user.lb.name
}
```

## Argument Reference

This resource supports the following arguments:

* `policy` - (Required) The policy document. This is a JSON formatted string. For more information about building AWS IAM policy documents with Terraform, see the [AWS IAM Policy Document Guide](https://learn.hashicorp.com/terraform/aws/iam-policy).
* `name` - (Optional) The name of the policy. If omitted, Terraform will assign a random, unique name.
* `name_prefix` - (Optional, Forces new resource) Creates a unique name beginning with the specified prefix. Conflicts with `name`.
* `user` - (Required) IAM user to which to attach this policy.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The user policy ID, in the form of `user_name:user_policy_name`.
* `name` - The name of the policy (always set).

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_iam_user_policy.example
  identity = {
    user = "my-user"
    name = "my-policy"
  }
}

resource "aws_iam_user_policy" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `user` (String) Name of the IAM user.
* `name` (String) Name of the user policy.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import IAM User Policies using a `:` delimited string separating `user` and `name`. For example:

```terraform
import {
  to = aws_iam_user_policy.example
  id = "my-user:my-policy"
}
```

Using `terraform import`, import IAM User Policies using a `:` delimited string separating `user` and `name`. For example:

```console
% terraform import aws_iam_user_policy.example my-user:my-policy
```
