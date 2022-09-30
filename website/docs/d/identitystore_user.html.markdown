---
subcategory: "SSO Identity Store"
layout: "aws"
page_title: "AWS: aws_identitystore_user"
description: |-
  Get information on an Identity Store User
---

# Data Source: aws_identitystore_user

Use this data source to get an Identity Store User.

## Example Usage

```terraform
data "aws_ssoadmin_instances" "example" {}

data "aws_identitystore_user" "example" {
  identity_store_id = tolist(data.aws_ssoadmin_instances.example.identity_store_ids)[0]

  filter {
    attribute_path  = "UserName"
    attribute_value = "ExampleUser"
  }
}

output "user_id" {
  value = data.aws_identitystore_user.example.user_id
}
```

## Argument Reference

The following arguments are required:

* `identity_store_id` - (Required) Identity Store ID associated with the Single Sign-On Instance.

The following arguments are optional:

* `external_id` - (Optional) Configuration block for filtering by the identifier issued by an external identity provider. At most 1 filter can be provided. Conflicts with `filter`. Detailed below.
* `filter` - (Optional) Configuration block for filtering by a unique attribute of the user. At most 1 filter can be provided. Conflicts with `external_id`. Detailed below.
* `user_id` - (Optional) The identifier for a user in the Identity Store.

-> At least one of `external_id`, `filter`, or `user_id` must be set.

### `external_id` Configuration Block

The following arguments are supported by the `external_id` configuration block:

* `id` - (Required) The identifier issued to this resource by an external identity provider.
* `issuer` - (Required) The issuer for an external identifier.

### `filter` Configuration Block

The following arguments are supported by the `filter` configuration block:

* `attribute_path` - (Required) Attribute path that is used to specify which attribute name to search. For example: `UserName`. Refer to the [User data type](https://docs.aws.amazon.com/singlesignon/latest/IdentityStoreAPIReference/API_User.html).
* `attribute_value` - (Required) Value for an attribute.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Identifier of the user in the Identity Store.
* `user_name` - User's user name value.
