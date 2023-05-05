---
subcategory: "SSO Identity Store"
layout: "aws"
page_title: "AWS: aws_identitystore_group"
description: |-
  Get information on an Identity Store Group
---

# Data Source: aws_identitystore_group

Use this data source to get an Identity Store Group.

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

output "group_id" {
  value = data.aws_identitystore_group.example.group_id
}
```

## Argument Reference

The following arguments are required:

* `identity_store_id` - (Required) Identity Store ID associated with the Single Sign-On Instance.

The following arguments are optional:

* `alternate_identifier` (Optional) A unique identifier for the group that is not the primary identifier. Conflicts with `group_id` and `filter`. Detailed below.
* `filter` - (Optional, **Deprecated** use the `alternate_identifier` attribute instead) Configuration block for filtering by a unique attribute of the group. Detailed below.
* `group_id` - (Optional) The identifier for a group in the Identity Store.

-> Exactly one of the above arguments must be provided. Passing both `filter` and `group_id` is allowed for backwards compatibility.

### `alternate_identifier` Configuration Block

The following arguments are supported by the `alternate_identifier` configuration block:

* `external_id` - (Optional) Configuration block for filtering by the identifier issued by an external identity provider. Detailed below.
* `unique_attribute` - (Optional) An entity attribute that's unique to a specific entity. Detailed below.

-> Exactly one of the above arguments must be provided.

### `external_id` Configuration Block

The following arguments are supported by the `external_id` configuration block:

* `id` - (Required) The identifier issued to this resource by an external identity provider.
* `issuer` - (Required) The issuer for an external identifier.

### `filter` Configuration Block

~> The `filter` configuration block has been deprecated. Use `alternate_identifier` instead.

The following arguments are supported by the `filter` configuration block:

* `attribute_path` - (Required) Attribute path that is used to specify which attribute name to search. Currently, `DisplayName` is the only valid attribute path.
* `attribute_value` - (Required) Value for an attribute.

### `unique_attribute` Configuration Block

The following arguments are supported by the `unique_attribute` configuration block:

* `attribute_path` - (Required) Attribute path that is used to specify which attribute name to search. For example: `DisplayName`. Refer to the [Group data type](https://docs.aws.amazon.com/singlesignon/latest/IdentityStoreAPIReference/API_Group.html).
* `attribute_value` - (Required) Value for an attribute.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Identifier of the group in the Identity Store.
* `description` - Description of the specified group.
* `display_name` - Group's display name value.
* `external_ids` - List of identifiers issued to this resource by an external identity provider.
    * `id` - The identifier issued to this resource by an external identity provider.
    * `issuer` - The issuer for an external identifier.
