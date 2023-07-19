---
subcategory: "IAM (Identity & Access Management)"
layout: "aws"
page_title: "AWS: aws_iam_group_policy_attachment"
description: |-
  Attaches a Managed IAM Policy to an IAM group
---

# Resource: aws_iam_group_policy_attachment

Attaches a Managed IAM Policy to an IAM group

~> **NOTE:** The usage of this resource conflicts with the `aws_iam_policy_attachment` resource and will permanently show a difference if both are defined.

## Example Usage

```terraform
resource "aws_iam_group" "group" {
  name = "test-group"
}

resource "aws_iam_policy" "policy" {
  name        = "test-policy"
  description = "A test policy"
  policy      = "{ ... policy JSON ... }"
}

resource "aws_iam_group_policy_attachment" "test-attach" {
  group      = aws_iam_group.group.name
  policy_arn = aws_iam_policy.policy.arn
}
```

## Argument Reference

This resource supports the following arguments:

* `group`  (Required) - The group the policy should be applied to
* `policy_arn`  (Required) - The ARN of the policy you want to apply

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import IAM group policy attachments using the group name and policy arn separated by `/`. For example:

```terraform
import {
  to = aws_iam_group_policy_attachment.test-attach
  id = "test-group/arn:aws:iam::xxxxxxxxxxxx:policy/test-policy"
}
```

Using `terraform import`, import IAM group policy attachments using the group name and policy arn separated by `/`. For example:

```console
% terraform import aws_iam_group_policy_attachment.test-attach test-group/arn:aws:iam::xxxxxxxxxxxx:policy/test-policy
```
