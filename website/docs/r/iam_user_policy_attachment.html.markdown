---
subcategory: "IAM (Identity & Access Management)"
layout: "aws"
page_title: "AWS: aws_iam_user_policy_attachment"
description: |-
  Attaches a Managed IAM Policy to an IAM user
---

# Resource: aws_iam_user_policy_attachment

Attaches a Managed IAM Policy to an IAM user

~> **NOTE:** The usage of this resource conflicts with the `aws_iam_policy_attachment` resource and will permanently show a difference if both are defined.

## Example Usage

```terraform
resource "aws_iam_user" "user" {
  name = "test-user"
}

resource "aws_iam_policy" "policy" {
  name        = "test-policy"
  description = "A test policy"
  policy      = "{ ... policy JSON ... }"
}

resource "aws_iam_user_policy_attachment" "test-attach" {
  user       = aws_iam_user.user.name
  policy_arn = aws_iam_policy.policy.arn
}
```

## Argument Reference

This resource supports the following arguments:

* `user`        (Required) - The user the policy should be applied to
* `policy_arn`  (Required) - The ARN of the policy you want to apply

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_iam_user_policy_attachment.example
  identity = {
    user       = "test-user"
    policy_arn = "arn:aws:iam::xxxxxxxxxxxx:policy/test-policy"
  }
}

resource "aws_iam_user_policy_attachment" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `user` (String) Name of the IAM user.
* `policy_arn` (String) ARN of the IAM policy.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import IAM user policy attachments using the user name and policy arn separated by `/`. For example:

```terraform
import {
  to = aws_iam_user_policy_attachment.example
  id = "test-user/arn:aws:iam::xxxxxxxxxxxx:policy/test-policy"
}
```

Using `terraform import`, import IAM user policy attachments using the user name and policy arn separated by `/`. For example:

```console
% terraform import aws_iam_user_policy_attachment.example test-user/arn:aws:iam::xxxxxxxxxxxx:policy/test-policy
```
