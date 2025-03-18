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
    * `addresses` - List of details about the user's address.
        * `country` - Country that this address is in.
        * `formatted` - Name that is typically displayed when the address is shown for display.
        * `locality` - Address locality.
        * `postal_code` - Postal code of the address.
        * `primary` - When `true`, this is the primary address associated with the user.
        * `region` - Region of the address.
        * `street_address` - Street of the address.
        * `type` - Type of address.
    * `display_name` - Name that is typically displayed when the user is referenced.
    * `emails` - List of details about the user's email.
        * `primary` - When `true`, this is the primary email associated with the user.
        * `type` - Type of email.
        * `value` - Email address. This value must be unique across the identity store.
    * `external_ids` - List of identifiers issued to this resource by an external identity provider.
        * `id` - Identifier issued to this resource by an external identity provider.
        * `issuer` - Issuer for an external identifier.
    * `locale` - User's geographical region or location.
    * `name` - Details about the user's full name.
        * `family_name` - Family name of the user.
        * `formatted` - Name that is typically displayed when the name is shown for display.
        * `given_name` - Given name of the user.
        * `honorific_prefix` - Honorific prefix of the user.
        * `honorific_suffix` - Honorific suffix of the user.
        * `middle_name` - Middle name of the user.
    * `nickname` - An alternate name for the user.
    * `phone_numbers` - List of details about the user's phone number.
        * `primary` - When `true`, this is the primary phone number associated with the user.
        * `type` - Type of phone number.
        * `value` - User's phone number.
    * `preferred_language` - Preferred language of the user.
    * `profile_url` - An URL that may be associated with the user.
    * `timezone` - User's time zone.
    * `title` - User's title.
    * `user_id` - Identifier of the user in the Identity Store.
    * `user_name` - User's user name value.
    * `user_type` - User type.
