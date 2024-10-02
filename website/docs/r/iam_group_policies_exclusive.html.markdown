---
subcategory: "IAM (Identity & Access Management)"
layout: "aws"
page_title: "AWS: aws_iam_group_policies_exclusive"
description: |-
  Terraform resource for maintaining exclusive management of inline policies assigned to an AWS IAM (Identity & Access Management) group.
---
# Resource: aws_iam_group_policies_exclusive

Terraform resource for maintaining exclusive management of inline policies assigned to an AWS IAM (Identity & Access Management) group.

!> This resource takes exclusive ownership over inline policies assigned to a group. This includes removal of inline policies which are not explicitly configured. To prevent persistent drift, ensure any `aws_iam_group_policy` resources managed alongside this resource are included in the `policy_names` argument.

## Example Usage

### Basic Usage

```terraform
resource "aws_iam_group_policies_exclusive" "example" {
  group_name   = aws_iam_group.example.name
  policy_names = [aws_iam_group_policy.example.name]
}
```

### Disallow Inline Policies

To automatically remove any configured inline policies, set the `policy_names` argument to an empty list.

~> This will not __prevent__ inline policies from being assigned to a group via Terraform (or any other interface). This resource enables bringing inline policy assignments into a configured state, however, this reconciliation happens only when `apply` is proactively run.

```terraform
resource "aws_iam_group_policies_exclusive" "example" {
  group_name   = aws_iam_group.example.name
  policy_names = []
}
```

## Argument Reference

The following arguments are required:

* `group_name` - (Required) IAM group name.
* `policy_names` - (Required) A list of inline policy names to be assigned to the group. Policies attached to this group but not configured in this argument will be removed.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to exclusively manage inline policy assignments using the `group_name`. For example:

```terraform
import {
  to = aws_iam_group_policies_exclusive.example
  id = "MyGroup"
}
```

Using `terraform import`, import exclusive management of inline policy assignments using the `group_name`. For example:

```console
% terraform import aws_iam_group_policies_exclusive.example MyGroup
```
