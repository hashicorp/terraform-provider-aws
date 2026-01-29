---
subcategory: "SSO Admin"
layout: "aws"
page_title: "AWS: aws_ssoadmin_managed_policy_attachments_exclusive"
description: |-
  Terraform resource for managing exclusive AWS SSO Admin Managed Policy Attachments.
---

# Resource: aws_ssoadmin_managed_policy_attachments_exclusive

Terraform resource for managing exclusive AWS SSO Admin Managed Policy Attachments.

This resource is designed to manage all managed policy attachments for an SSO permission set. Using this resource, Terraform will remove any managed policies attached to the permission set that are not defined in the configuration.

!> **WARNING:** Do not use this resource together with the `aws_ssoadmin_managed_policy_attachment` resource for the same permission set. Doing so will cause a conflict and will lead to managed policies being removed.

~> Destruction of this resource means Terraform will no longer manage the managed policy attachments, **but will not detach any policies**. The permission set will retain all managed policies that were attached at the time of destruction.

## Example Usage

### Basic Usage

```terraform
data "aws_ssoadmin_instances" "example" {}

resource "aws_ssoadmin_permission_set" "example" {
  name         = "Example"
  instance_arn = tolist(data.aws_ssoadmin_instances.example.arns)[0]
}

resource "aws_ssoadmin_managed_policy_attachments_exclusive" "example" {
  instance_arn       = tolist(data.aws_ssoadmin_instances.example.arns)[0]
  permission_set_arn = aws_ssoadmin_permission_set.example.arn

  managed_policy_arns = [
    "arn:aws:iam::aws:policy/ReadOnlyAccess",
  ]
}
```

### Disallow Managed Policy Attachments

To disallow all managed policy attachments, set `managed_policy_arns` to an empty list.

~> Any managed policies attached to the permission set will be **removed**.

```terraform
resource "aws_ssoadmin_managed_policy_attachments_exclusive" "example" {
  instance_arn       = tolist(data.aws_ssoadmin_instances.example.arns)[0]
  permission_set_arn = aws_ssoadmin_permission_set.example.arn

  managed_policy_arns = []
}
```

## Argument Reference

The following arguments are required:

* `instance_arn` - (Required) ARN of the SSO Instance.
* `managed_policy_arns` - (Required) Set of ARNs of IAM managed policies to attach to the Permission Set.
* `permission_set_arn` - (Required) ARN of the Permission Set.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

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
  to = aws_ssoadmin_managed_policy_attachments_exclusive.example
  identity = {
    instance_arn       = "arn:aws:sso:::instance/ssoins-1234567890abcdef"
    permission_set_arn = "arn:aws:sso:::permissionSet/ssoins-1234567890abcdef/ps-1234567890abcdef"
  }
}

resource "aws_ssoadmin_managed_policy_attachments_exclusive" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `instance_arn` (String) ARN of the SSO Instance.
* `permission_set_arn` (String) ARN of the Permission Set.

#### Optional

* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import SSO Admin Managed Policy Attachments Exclusive using the `instance_arn` and `permission_set_arn` arguments, separated by a comma (`,`). For example:

```terraform
import {
  to = aws_ssoadmin_managed_policy_attachments_exclusive.example
  id = "arn:aws:sso:::instance/ssoins-1234567890abcdef,arn:aws:sso:::permissionSet/ssoins-1234567890abcdef/ps-1234567890abcdef"
}
```

Using `terraform import`, import SSO Admin Managed Policy Attachments Exclusive using the `instance_arn` and `permission_set_arn` arguments, separated by a comma (`,`). For example:

```console
% terraform import aws_ssoadmin_managed_policy_attachments_exclusive.example arn:aws:sso:::instance/ssoins-1234567890abcdef,arn:aws:sso:::permissionSet/ssoins-1234567890abcdef/ps-1234567890abcdef
```
