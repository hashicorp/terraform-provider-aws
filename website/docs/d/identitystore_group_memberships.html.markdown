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

The following arguments are required:

* `group_id` - (Required) The identifier for a group in the Identity Store.
* `identity_store_id` - (Required) Identity Store ID associated with the Single Sign-On Instance.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `group_ids` - Identity Store Group IDs
* `identity_store_ids` - Identity Store IDs
* `member_ids` - List of Identity Store User IDs
* `memberships_ids` - Identity Store Group Membership IDs (the ID of the membership not the member!)
