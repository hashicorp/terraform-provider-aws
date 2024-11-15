---
subcategory: "IAM (Identity & Access Management)"
layout: "aws"
page_title: "AWS: aws_iam_user_policy_attachments_exclusive"
description: |-
  Terraform resource for maintaining exclusive management of customer managed policies assigned to an AWS IAM (Identity & Access Management) user.
---
# Resource: aws_iam_user_policy_attachments_exclusive

Terraform resource for maintaining exclusive management of customer managed policies assigned to an AWS IAM (Identity & Access Management) user.

!> This resource takes exclusive ownership over customer managed policies attached to a user. This includes removal of customer managed policies which are not explicitly configured. To prevent persistent drift, ensure any `aws_iam_user_policy_attachment` resources managed alongside this resource are included in the `policy_arns` argument.

~> Destruction of this resource means Terraform will no longer manage reconciliation of the configured policy attachments. It __will not__ detach the configured policies from the user.

## Example Usage

### Basic Usage

```terraform
resource "aws_iam_user_policy_attachments_exclusive" "example" {
  user_name   = aws_iam_user.example.name
  policy_arns = [aws_iam_policy.example.arn]
}
```

### Disallow Customer Managed Policies

To automatically remove any configured customer managed policies, set the `policy_arns` argument to an empty list.

~> This will not __prevent__ customer managed policies from being assigned to a user via Terraform (or any other interface). This resource enables bringing customer managed policy assignments into a configured state, however, this reconciliation happens only when `apply` is proactively run.

```terraform
resource "aws_iam_user_policy_attachments_exclusive" "example" {
  user_name   = aws_iam_user.example.name
  policy_arns = []
}
```

## Argument Reference

The following arguments are required:

* `user_name` - (Required) IAM user name.
* `policy_arns` - (Required) A list of customer managed policy ARNs to be attached to the user. Policies attached to this user but not configured in this argument will be removed.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to exclusively manage customer managed policy assignments using the `user_name`. For example:

```terraform
import {
  to = aws_iam_user_policy_attachments_exclusive.example
  id = "MyUser"
}
```

Using `terraform import`, import exclusive management of customer managed policy assignments using the `user_name`. For example:

```console
% terraform import aws_iam_user_policy_attachments_exclusive.example MyUser
```
