---
subcategory: "IAM (Identity & Access Management)"
layout: "aws"
page_title: "AWS: aws_iam_role_policy_attachment"
description: |-
  Attaches a Managed IAM Policy to an IAM role
---

# Resource: aws_iam_role_policy_attachment

Attaches a Managed IAM Policy to an IAM role

~> **NOTE:** The usage of this resource conflicts with the `aws_iam_policy_attachment` resource and will permanently show a difference if both are defined.

~> **NOTE:** For a given role, this resource is incompatible with using the [`aws_iam_role` resource](/docs/providers/aws/r/iam_role.html) `managed_policy_arns` argument. When using that argument and this resource, both will attempt to manage the role's managed policy attachments and Terraform will show a permanent difference.

## Example Usage

```terraform
data "aws_iam_policy_document" "assume_role" {
  statement {
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["ec2.amazonaws.com"]
    }

    actions = ["sts:AssumeRole"]
  }
}

resource "aws_iam_role" "role" {
  name               = "test-role"
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

data "aws_iam_policy_document" "policy" {
  statement {
    effect    = "Allow"
    actions   = ["ec2:Describe*"]
    resources = ["*"]
  }
}

resource "aws_iam_policy" "policy" {
  name        = "test-policy"
  description = "A test policy"
  policy      = data.aws_iam_policy_document.policy.json
}

resource "aws_iam_role_policy_attachment" "test-attach" {
  role       = aws_iam_role.role.name
  policy_arn = aws_iam_policy.policy.arn
}
```

## Argument Reference

This resource supports the following arguments:

* `role`  (Required) - The name of the IAM role to which the policy should be applied
* `policy_arn` (Required) - The ARN of the policy you want to apply

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_iam_role_policy_attachment.example
  identity = {
    role       = "test-role"
    policy_arn = "arn:aws:iam::xxxxxxxxxxxx:policy/test-policy"
  }
}

resource "aws_iam_role_policy_attachment" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `role` (String) Name of the IAM role.
* `policy_arn` (String) ARN of the IAM policy.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import IAM role policy attachments using the role name and policy arn separated by `/`. For example:

```terraform
import {
  to = aws_iam_role_policy_attachment.example
  id = "test-role/arn:aws:iam::xxxxxxxxxxxx:policy/test-policy"
}
```

Using `terraform import`, import IAM role policy attachments using the role name and policy arn separated by `/`. For example:

```console
% terraform import aws_iam_role_policy_attachment.example test-role/arn:aws:iam::xxxxxxxxxxxx:policy/test-policy
```
