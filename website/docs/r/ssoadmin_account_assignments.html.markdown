---
subcategory: "SSO Admin"
layout: "aws"
page_title: "AWS: aws_ssoadmin_account_assignments"
description: |-
  Manages multiple Single Sign-On (SSO) Account Assignments Authoritatively
---

# Resource: aws_ssoadmin_account_assignments

Manages multiple Single Sign-On (SSO) Account Assignments Authoritatively

## Example Usage

```terraform
data "aws_ssoadmin_instances" "example" {}

data "aws_ssoadmin_permission_set" "example" {
  instance_arn = tolist(data.aws_ssoadmin_instances.example.arns)[0]
  name         = "AWSReadOnlyAccess"
}

data "aws_identitystore_group" "example1" {
  identity_store_id = tolist(data.aws_ssoadmin_instances.example.identity_store_ids)[0]

  filter {
    attribute_path  = "DisplayName"
    attribute_value = "ExampleGroup1"
  }
}

data "aws_identitystore_group" "example2" {
  identity_store_id = tolist(data.aws_ssoadmin_instances.example.identity_store_ids)[0]

  filter {
    attribute_path  = "DisplayName"
    attribute_value = "ExampleGroup2"
  }
}

resource "aws_ssoadmin_account_assignments" "example" {
  instance_arn       = data.aws_ssoadmin_permission_set.example.instance_arn
  permission_set_arn = data.aws_ssoadmin_permission_set.example.arn

  principal_ids = [
    data.aws_identitystore_group.example1.group_id,
    data.aws_identitystore_group.example2.group_id
  ]

  principal_type = "GROUP"

  target_id   = "012347678910"
  target_type = "AWS_ACCOUNT"
}
```

## Argument Reference

The following arguments are supported:

* `instance_arn` - (Required, Forces new resource) The Amazon Resource Name (ARN) of the SSO Instance.
* `permission_set_arn` - (Required, Forces new resource) The Amazon Resource Name (ARN) of the Permission Set that the admin wants to grant the principal access to.
* `principal_ids` - (Required) A list of identifiers for objects in SSO, such as a user or group. PrincipalIds are GUIDs (For example, `f81d4fae-7dec-11d0-a765-00a0c91e6bf6`).
* `principal_type` - (Required, Forces new resource) The entity type for which the assignment will be created. Valid values: `USER`, `GROUP`.
* `target_id` - (Required, Forces new resource) An AWS account identifier, typically a 10-12 digit string.
* `target_type` - (Optional, Forces new resource) The entity type for which the assignment will be created. Valid values: `AWS_ACCOUNT`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The identifier of the Account Assignment i.e., `principal_type`, `target_id`, `target_type`, `permission_set_arn`, `instance_arn` separated by commas (`,`).

## Import

SSO Account Assignments can be imported using the `principal_type`, `target_id`, `target_type`, `permission_set_arn`, `instance_arn` separated by commas (`,`) e.g.,

```
$ terraform import aws_ssoadmin_account_assignment.example GROUP,1234567890,AWS_ACCOUNT,arn:aws:sso:::permissionSet/ssoins-0123456789abcdef/ps-0123456789abcdef,arn:aws:sso:::instance/ssoins-0123456789abcdef
```
