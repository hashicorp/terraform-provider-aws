---
subcategory: "SSO Identity Store"
layout: "aws"
page_title: "AWS: aws_identitystore_group_membership"
description: |-
  Terraform resource for managing an AWS IdentityStore Group Membership.
---

# Resource: aws_identitystore_group_membership

Terraform resource for managing an AWS IdentityStore Group Membership.

## Example Usage

```terraform
data "aws_ssoadmin_instances" "example" {}

resource "aws_identitystore_user" "example" {
  identity_store_id = tolist(data.aws_ssoadmin_instances.example.identity_store_ids)[0]
  display_name      = "John Doe"
  user_name         = "john.doe@example.com"

  name {
    family_name = "Doe"
    given_name  = "John"
  }
}

resource "aws_identitystore_group" "example" {
  identity_store_id = tolist(data.aws_ssoadmin_instances.example.identity_store_ids)[0]
  display_name      = "MyGroup"
  description       = "Some group name"
}

resource "aws_identitystore_group_membership" "example" {
  identity_store_id = tolist(data.aws_ssoadmin_instances.example.identity_store_ids)[0]
  group_id          = aws_identitystore_group.example.group_id
  member_id         = aws_identitystore_user.example.user_id
}
```

## Argument Reference

The following arguments are supported:

* `member_id` - (Required) The identifier for a user in the Identity Store.
* `group_id` - (Required)  The identifier for a group in the Identity Store.
* `identity_store_id` - (Required) Identity Store ID associated with the Single Sign-On Instance.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `membership_id` - The identifier of the newly created group membership in the Identity Store.

## Import

`aws_identitystore_group_membership` can be imported using the `identity_store_id/membership_id`, e.g.,

```
$ terraform import aws_identitystore_group_membership.example d-0000000000/00000000-0000-0000-0000-000000000000
```
