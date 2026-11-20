---
subcategory: "SSO Identity Store"
layout: "aws"
page_title: "AWS: aws_identitystore_group_membership"
description: |-
  Manages an AWS IdentityStore Group Membership.
---

# Resource: aws_identitystore_group_membership

Terraform resource for managing an AWS IdentityStore Group Membership.

## Example Usage

### Basic Usage

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

This resource supports the following arguments:

* `group_id` - (Required) Identifier for a group in the Identity Store.
* `identity_store_id` - (Required) Identity Store ID associated with the Single Sign-On Instance.
* `member_id` - (Required) Identifier for a user in the Identity Store.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `membership_id` - Identifier of the newly created group membership in the Identity Store.

## Import

In Terraform v1.12.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_identitystore_group_membership` using its identity attributes. For example:

```terraform
import {
  to = aws_identitystore_group_membership.example
  identity = {
    identity_store_id = "d-0000000000"
    membership_id     = "00000000-0000-0000-0000-000000000000"
  }
}
```

### Identity Schema

#### Required

* `identity_store_id` (String) ID of the Identity Store.
* `membership_id` (String) ID of the group membership.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_identitystore_group_membership` using the `identity_store_id/membership_id`. For example:

```terraform
import {
  to = aws_identitystore_group_membership.example
  id = "d-0000000000/00000000-0000-0000-0000-000000000000"
}
```

Using `terraform import`, import `aws_identitystore_group_membership` using the `identity_store_id/membership_id`. For example:

```console
% terraform import aws_identitystore_group_membership.example d-0000000000/00000000-0000-0000-0000-000000000000
```
