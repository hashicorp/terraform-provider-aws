---
subcategory: "SSO Identity Store"
layout: "aws"
page_title: "AWS: aws_identitystore_group_membership"
description: |-
  Lists SSO Identity Store Group Membership resources.
---

# List Resource: aws_identitystore_group_membership

Lists SSO Identity Store Group Membership resources.

## Example Usage

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

list "aws_identitystore_group_membership" "example" {
  provider = aws

  config {
    identity_store_id = tolist(data.aws_ssoadmin_instances.example.identity_store_ids)[0]
    group_id          = data.aws_identitystore_group.example.group_id
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `group_id` - (Required) ID of the group.
* `identity_store_id` - (Required) ID of the Identity Store.
* `region` - (Optional) Region to query. Defaults to provider region.
