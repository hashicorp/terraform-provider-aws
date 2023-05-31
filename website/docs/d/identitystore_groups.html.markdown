---
subcategory: "SSO Identity Store"
layout: "aws"
page_title: "AWS: aws_identitystore_groups"
description: |-
  Retrieve list of groups for an Identity Store instance.
---

# Data Source: aws_identitystore_groups

Use this data source to get a list of groups in an Identity Store instance.

## Example Usage

### Basic Usage

```terraform
data "aws_ssoadmin_instances" "example" {}

data "aws_identitystore_groups" "example" {
  identity_store_id = tolist(data.aws_ssoadmin_instances.example.identity_store_ids)[0]
}
```

## Argument Reference

The following arguments are required:

* `identity_store_id` - (Required) Identity Store ID associated with the Single Sign-On Instance.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `groups` - List of Identity Store Groups
    * `group_id` - Identifier of the group in the Identity Store.
    * `description` - Description of the specified group.
    * `display_name` - Group's display name value.
    * `external_ids` - List of identifiers issued to this resource by an external identity provider.
        * `id` - The identifier issued to this resource by an external identity provider.
        * `issuer` - The issuer for an external identifier.
