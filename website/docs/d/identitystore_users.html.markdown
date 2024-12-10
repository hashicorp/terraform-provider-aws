---
subcategory: "SSO Identity Store"
layout: "aws"
page_title: "AWS: aws_identitystore_users"
description: |-
  Retrieve list of users for an Identity Store instance.
---

# Data Source: aws_identitystore_users

Use this data source to get a list of users in an Identity Store instance.

## Example Usage

### Basic Usage

```terraform
data "aws_ssoadmin_instances" "example" {}

data "aws_identitystore_users" "example" {
  identity_store_id = tolist(data.aws_ssoadmin_instances.example.identity_store_ids)[0]
}
```

## Argument Reference

The following arguments are required:

* `identity_store_id` - (Required) Identity Store ID associated with the Single Sign-On Instance.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `users` - List of Identity Store Users
    * `id` - Identifier of the user in the Identity Store.
    * `addresses` - List of details about the user's address.
        * `country` - The country that this address is in.
        * `formatted` - The name that is typically displayed when the address is shown for display.
        * `locality` - The address locality.
        * `postal_code` - The postal code of the address.
        * `primary` - When `true`, this is the primary address associated with the user.
        * `region` - The region of the address.
        * `street_address` - The street of the address.
        * `type` - The type of address.
    * `display_name` - The name that is typically displayed when the user is referenced.
    * `emails` - List of details about the user's email.
        * `primary` - When `true`, this is the primary email associated with the user.
        * `type` - The type of email.
        * `value` - The email address. This value must be unique across the identity store.
    * `external_ids` - List of identifiers issued to this resource by an external identity provider.
        * `id` - The identifier issued to this resource by an external identity provider.
        * `issuer` - The issuer for an external identifier.
    * `locale` - The user's geographical region or location.
    * `name` - Details about the user's full name.
        * `family_name` - The family name of the user.
        * `formatted` - The name that is typically displayed when the name is shown for display.
        * `given_name` - The given name of the user.
        * `honorific_prefix` - The honorific prefix of the user.
        * `honorific_suffix` - The honorific suffix of the user.
        * `middle_name` - The middle name of the user.
    * `nickname` - An alternate name for the user.
    * `phone_numbers` - List of details about the user's phone number.
        * `primary` - When `true`, this is the primary phone number associated with the user.
        * `type` - The type of phone number.
        * `value` - The user's phone number.
    * `preferred_language` - The preferred language of the user.
    * `profile_url` - An URL that may be associated with the user.
    * `timezone` - The user's time zone.
    * `title` - The user's title.
    * `user_name` - User's user name value.
    * `user_type` - The user type.
