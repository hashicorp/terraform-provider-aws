---
subcategory: "IAM (Identity & Access Management)"
layout: "aws"
page_title: "AWS: aws_iam_role_policies_exclusive"
description: |-
  Terraform resource for maintaining exclusive management of inline policies assigned to an AWS IAM (Identity & Access Management) role.
---
# Resource: aws_iam_role_policies_exclusive

Terraform resource for maintaining exclusive management of inline policies assigned to an AWS IAM (Identity & Access Management) role.

!> This resource takes exclusive ownership over inline policies assigned to a role. This includes removal of inline policies which are not explicitly configured. To prevent persistent drift, ensure any `aws_iam_role_policy` resources managed alongside this resource are included in the `policy_names` argument.

~> Destruction of this resource means Terraform will no longer manage reconciliation of the configured inline policy assignments. It __will not__ delete the configured policies from the role.

## Example Usage

### Basic Usage

```terraform
resource "aws_iam_role_policies_exclusive" "example" {
  role_name    = aws_iam_role.example.name
  policy_names = [aws_iam_role_policy.example.name]
}
```

### Disallow Inline Policies

To automatically remove any configured inline policies, set the `policy_names` argument to an empty list.

~> This will not __prevent__ inline policies from being assigned to a role via Terraform (or any other interface). This resource enables bringing inline policy assignments into a configured state, however, this reconciliation happens only when `apply` is proactively run.

```terraform
resource "aws_iam_role_policies_exclusive" "example" {
  role_name    = aws_iam_role.example.name
  policy_names = []
}
```

## Argument Reference

The following arguments are required:

* `role_name` - (Required) IAM role name.
* `policy_names` - (Required) A list of inline policy names to be assigned to the role. Policies attached to this role but not configured in this argument will be removed.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to exclusively manage inline policy assignments using the `role_name`. For example:

```terraform
import {
  to = aws_iam_role_policies_exclusive.example
  id = "MyRole"
}
```

Using `terraform import`, import exclusive management of inline policy assignments using the `role_name`. For example:

```console
% terraform import aws_iam_role_policies_exclusive.example MyRole
```
