---
subcategory: "SSO Admin"
layout: "aws"
page_title: "AWS: aws_ssoadmin_customer_managed_policy_attachment"
description: |-
  Manages a customer managed policy for a Single Sign-On (SSO) Permission Set
---

# Resource: aws_ssoadmin_customer_managed_policy_attachment

Provides a customer managed policy attachment for a Single Sign-On (SSO) Permission Set resource

~> **NOTE:** Creating this resource will automatically [Provision the Permission Set](https://docs.aws.amazon.com/singlesignon/latest/APIReference/API_ProvisionPermissionSet.html) to apply the corresponding updates to all assigned accounts.

## Example Usage

```terraform
data "aws_ssoadmin_instances" "example" {}

resource "aws_ssoadmin_permission_set" "example" {
  name         = "Example"
  instance_arn = tolist(data.aws_ssoadmin_instances.example.arns)[0]
}

resource "aws_iam_policy" "example" {
  name        = "TestPolicy"
  description = "My test policy"
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

resource "aws_ssoadmin_customer_managed_policy_attachment" "example" {
  instance_arn       = aws_ssoadmin_permission_set.example.instance_arn
  permission_set_arn = aws_ssoadmin_permission_set.example.arn
  customer_managed_policy_reference {
    name = aws_iam_policy.example.name
    path = "/"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `instance_arn` - (Required, Forces new resource) The Amazon Resource Name (ARN) of the SSO Instance under which the operation will be executed.
* `permission_set_arn` - (Required, Forces new resource) The Amazon Resource Name (ARN) of the Permission Set.
* `customer_managed_policy_reference` - (Required, Forces new resource) Specifies the name and path of a customer managed policy. See below.

### Customer Managed Policy Reference

The `customer_managed_policy_reference` config block describes a customer managed IAM policy. You must have an IAM policy that matches the name and path in each AWS account where you want to deploy your specified permission set.

* `name` - (Required, Forces new resource) Name of the customer managed IAM Policy to be attached.
* `path` - (Optional, Forces new resource) The path to the IAM policy to be attached. The default is `/`. See [IAM Identifiers](https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_identifiers.html#identifiers-friendly-names) for more information.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Policy Name, Policy Path, Permission Set Amazon Resource Name (ARN), and SSO Instance ARN, each separated by a comma (`,`).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import SSO Managed Policy Attachments using the `name`, `path`, `permission_set_arn`, and `instance_arn` separated by a comma (`,`). For example:

```terraform
import {
  to = aws_ssoadmin_customer_managed_policy_attachment.example
  id = "TestPolicy,/,arn:aws:sso:::permissionSet/ssoins-2938j0x8920sbj72/ps-80383020jr9302rk,arn:aws:sso:::instance/ssoins-2938j0x8920sbj72"
}
```

Using `terraform import`, import SSO Managed Policy Attachments using the `name`, `path`, `permission_set_arn`, and `instance_arn` separated by a comma (`,`). For example:

```console
% terraform import aws_ssoadmin_customer_managed_policy_attachment.example TestPolicy,/,arn:aws:sso:::permissionSet/ssoins-2938j0x8920sbj72/ps-80383020jr9302rk,arn:aws:sso:::instance/ssoins-2938j0x8920sbj72
```
