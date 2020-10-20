---
subcategory: "SSO Admin"
layout: "aws"
page_title: "AWS: sso_assignment"
description: |-
  Manages an AWS Single Sign-On assignment
---

# Resource: sso_assignment

Provides an AWS Single Sign-On Assignment resource

## Example Usage

```hcl
data "aws_sso_permission_set" "example" {
  instance_arn = data.aws_sso_instance.selected.arn
  name         = "AWSReadOnlyAccess"
}

data "aws_identity_store_group" "example_group" {
  identity_store_id = data.aws_sso_instance.selected.identity_store_id
  display_name      = "Example Group@example.com"
}

resource "aws_sso_assignment" "example" {
  instance_arn       = data.aws_sso_instance.selected.arn
  permission_set_arn = data.aws_sso_permission_set.example.arn

  target_type = "AWS_ACCOUNT"
  target_id   = "012347678910"

  principal_type = "GROUP"
  principal_id   = data.aws_identity_store_group.example_group.group_id
}
```

## Argument Reference

The following arguments are supported:

* `instance_arn` - (Required) The AWS ARN associated with the AWS Single Sign-On Instance.
* `permission_set_arn` - (Required) The AWS ARN associated with the AWS Single Sign-On Permission Set.
* `target_id` - (Required) The identifier of the AWS account to assign to the AWS Single Sign-On Permission Set.
* `principal_type` - (Required) The entity type for which the assignment will be created. Valid values: `USER`, `GROUP`.
* `principal_id` - (Required) An identifier for an object in AWS SSO, such as a user or group. PrincipalIds are GUIDs (For example, f81d4fae-7dec-11d0-a765-00a0c91e6bf6).
* `target_type` - (Optional) Type of AWS Single Sign-On Assignment. Valid values: `AWS_ACCOUNT`.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Identifier of the AWS Single Sign-On Assignment.
* `created_date` - The created date of the AWS Single Sign-On Assignment.

## Import

`aws_sso_assignment` can be imported by using the identifier of the AWS Single Sign-On Assignment, e.g.
identifier = ${InstanceID}/${PermissionSetID}/${TargetType}/${TargetID}/${PrincipalType}/${PrincipalID}
```
$ terraform import aws_sso_assignment.example ssoins-0123456789abcdef/ps-0123456789abcdef/AWS_ACCOUNT/012347678910/GROUP/51b3755f39-e945c18b-e449-4a93-3e95-12231cb7ef96
```
