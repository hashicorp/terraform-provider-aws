---
subcategory: "IAM (Identity & Access Management)"
layout: "aws"
page_title: "AWS: aws_iam_group_policy_attachments_exclusive"
description: |-
  Terraform resource for maintaining exclusive management of managed IAM policies assigned to an AWS IAM (Identity & Access Management) group.
---

# Resource: aws_iam_group_policy_attachments_exclusive

Terraform resource for maintaining exclusive management of managed IAM policies assigned to an AWS IAM (Identity & Access Management) group.

!> This resource takes exclusive ownership over managed IAM policies attached to a group. This includes removal of managed IAM policies which are not explicitly configured. To prevent persistent drift, ensure any `aws_iam_group_policy_attachment` resources managed alongside this resource are included in the `policy_arns` argument.

~> Destruction of this resource means Terraform will no longer manage reconciliation of the configured policy attachments. It **will not** detach the configured policies from the group.

## Example Usage

### Basic Usage

```terraform
resource "aws_iam_group_policy_attachments_exclusive" "example" {
  group_name  = aws_iam_group.example.name
  policy_arns = [aws_iam_policy.example.arn]
}
```

### Disallow Managed IAM Policies

To automatically remove any configured managed IAM policies, set the `policy_arns` argument to an empty list.

~> This will not **prevent** managed IAM policies from being assigned to a group via Terraform (or any other interface). This resource enables bringing managed IAM policy assignments into a configured state, however, this reconciliation happens only when `apply` is proactively run.

```terraform
resource "aws_iam_group_policy_attachments_exclusive" "example" {
  group_name  = aws_iam_group.example.name
  policy_arns = []
}
```

## Argument Reference

The following arguments are required:

* `group_name` - (Required) IAM group name.
* `policy_arns` - (Required) A list of managed IAM policy ARNs to be attached to the group. Policies attached to this group but not configured in this argument will be removed.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to exclusively manage managed IAM policy assignments using the `group_name`. For example:

```terraform
import {
  to = aws_iam_group_policy_attachments_exclusive.example
  id = "MyGroup"
}
```

Using `terraform import`, import exclusive management of managed IAM policy assignments using the `group_name`. For example:

```console
% terraform import aws_iam_group_policy_attachments_exclusive.example MyGroup
```
