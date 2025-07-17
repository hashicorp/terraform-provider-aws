---
subcategory: "SSO Identity Store"
layout: "aws"
page_title: "AWS: aws_identitystore_group_memberships"
description: |-
  Retrieve list of members for an Identity Store Group.
---

# Data Source: aws_identitystore_group_memberships

Use this data source to get a list of members in an Identity Store Group.

## Example Usage

### Basic Usage

```terraform
data "aws_ssoadmin_instances" "example" {}

data "aws_identitystore_group" "example" {
  identity_store_id = tolist(data.aws_ssoadmin_instances.example.identity_store_ids)[0]

  alternate_identifier {
    unique_attribute {
      attribute_path  = "DisplayName"
      attribute_value = "ExampleGroup"
    }
  }
}

data "aws_identitystore_group_memberships" "example" {
  identity_store_id = tolist(data.aws_ssoadmin_instances.example.identity_store_ids)[0]
  group_id          = data.aws_identitystore_group.example.group_id
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `group_id` - (Required) The identifier for a group in the Identity Store.
* `identity_store_id` - (Required) Identity Store ID associated with the Single Sign-On Instance.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `group_memberships` - A list of group membership objects. See [`group_memberships`](#group_memberships) below.

### `group_memberships`

* `group_id` - Group identifier.
* `identity_store_id` - Identity store identifier.
* `member_id` - An object containing the identifier of a group member. See [`member_id`](#member_id) below.
* `memberships_id` - Group membership identifier.

#### `member_id`

* `user_id` - User identifier of the group member.
