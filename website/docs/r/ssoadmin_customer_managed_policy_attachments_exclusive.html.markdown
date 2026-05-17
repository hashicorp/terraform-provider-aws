---
subcategory: "SSO Admin"
layout: "aws"
page_title: "AWS: aws_ssoadmin_customer_managed_policy_attachments_exclusive"
description: |-
  Terraform resource for managing exclusive AWS SSO Admin Customer Managed Policy Attachments.
---

# Resource: aws_ssoadmin_customer_managed_policy_attachments_exclusive

Terraform resource for managing exclusive AWS SSO Admin Customer Managed Policy Attachments.

This resource is designed to manage all customer managed policy attachments for an SSO permission set. Using this resource, Terraform will remove any customer managed policies attached to the permission set that are not defined in the configuration.

!> **WARNING:** Do not use this resource together with the `aws_ssoadmin_customer_managed_policy_attachment` resource for the same permission set. Doing so will cause a conflict and will lead to customer managed policies being removed.

~> Destruction of this resource means Terraform will no longer manage the customer managed policy attachments, **but will not detach any policies**. The permission set will retain all customer managed policies that were attached at the time of destruction.

## Example Usage

### Basic Usage

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

resource "aws_ssoadmin_customer_managed_policy_attachments_exclusive" "example" {
  instance_arn       = tolist(data.aws_ssoadmin_instances.example.arns)[0]
  permission_set_arn = aws_ssoadmin_permission_set.example.arn

  customer_managed_policy_reference {
    name = aws_iam_policy.example.name
    path = "/"
  }
}
```

### Disallow Customer Managed Policy Attachments

To disallow all customer managed policy attachments, omit the `customer_managed_policy_reference` block.

~> Any customer managed policies attached to the permission set will be **removed**.

```terraform
resource "aws_ssoadmin_customer_managed_policy_attachments_exclusive" "example" {
  instance_arn       = tolist(data.aws_ssoadmin_instances.example.arns)[0]
  permission_set_arn = aws_ssoadmin_permission_set.example.arn
}
```

## Argument Reference

The following arguments are required:

* `instance_arn` - (Required) ARN of the SSO Instance.
* `permission_set_arn` - (Required) ARN of the Permission Set.

The following arguments are optional:

* `customer_managed_policy_reference` - (Optional) Specifies the names and paths of the customer managed policies to attach. See [Customer Managed Policy Reference](#customer-managed-policy-reference) below.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

### Customer Managed Policy Reference

The `customer_managed_policy_reference` block describes a customer managed IAM policy. You must have an IAM policy that matches the name and path in each AWS account where you want to deploy your specified permission set.

* `name` - (Required) Name of the customer managed IAM Policy to be attached.
* `path` - (Optional) The path to the IAM policy to be attached. The default is `/`. See [IAM Identifiers](https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_identifiers.html#identifiers-friendly-names) for more information.

## Attribute Reference

This resource exports no additional attributes.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `10m`)
* `update` - (Default `10m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_ssoadmin_customer_managed_policy_attachments_exclusive.example
  identity = {
    instance_arn       = "arn:aws:sso:::instance/ssoins-1234567890abcdef"
    permission_set_arn = "arn:aws:sso:::permissionSet/ssoins-1234567890abcdef/ps-1234567890abcdef"
  }
}

resource "aws_ssoadmin_customer_managed_policy_attachments_exclusive" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `instance_arn` (String) ARN of the SSO Instance.
* `permission_set_arn` (String) ARN of the Permission Set.

#### Optional

* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import SSO Admin Customer Managed Policy Attachments Exclusive using the `instance_arn` and `permission_set_arn` arguments, separated by a comma (`,`). For example:

```terraform
import {
  to = aws_ssoadmin_customer_managed_policy_attachments_exclusive.example
  id = "arn:aws:sso:::instance/ssoins-1234567890abcdef,arn:aws:sso:::permissionSet/ssoins-1234567890abcdef/ps-1234567890abcdef"
}
```

Using `terraform import`, import SSO Admin Customer Managed Policy Attachments Exclusive using the `instance_arn` and `permission_set_arn` arguments, separated by a comma (`,`). For example:

```console
% terraform import aws_ssoadmin_customer_managed_policy_attachments_exclusive.example arn:aws:sso:::instance/ssoins-1234567890abcdef,arn:aws:sso:::permissionSet/ssoins-1234567890abcdef/ps-1234567890abcdef
```
