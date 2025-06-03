---
subcategory: "SSO Identity Store"
layout: "aws"
page_title: "AWS: aws_identitystore_user"
description: |-
  Terraform resource for managing an AWS Identity Store User.
---

# Resource: aws_identitystore_user

This resource manages a User resource within an Identity Store.

-> **Note:** If you use an external identity provider or Active Directory as your identity source,
use this resource with caution. IAM Identity Center does not support outbound synchronization,
so your identity source does not automatically update with the changes that you make to
users using this resource.

## Example Usage

### Basic Usage

```terraform
resource "aws_identitystore_user" "example" {
  identity_store_id = tolist(data.aws_ssoadmin_instances.example.identity_store_ids)[0]

  display_name = "John Doe"
  user_name    = "johndoe"

  name {
    given_name  = "John"
    family_name = "Doe"
  }

  emails {
    value = "john@example.com"
  }
}
```

## Argument Reference

The following arguments are required:

* `display_name` - (Required) The name that is typically displayed when the user is referenced.
* `identity_store_id` - (Required, Forces new resource) The globally unique identifier for the identity store that this user is in.
* `name` - (Required) Details about the user's full name. Detailed below.
* `user_name` - (Required, Forces new resource) A unique string used to identify the user. This value can consist of letters, accented characters, symbols, numbers, and punctuation. This value is specified at the time the user is created and stored as an attribute of the user object in the identity store. The limit is 128 characters.

The following arguments are optional:

* `addresses` - (Optional) Details about the user's address. At most 1 address is allowed. Detailed below.
* `emails` - (Optional) Details about the user's email. At most 1 email is allowed. Detailed below.
* `locale` - (Optional) The user's geographical region or location.
* `nickname` - (Optional) An alternate name for the user.
* `phone_numbers` - (Optional) Details about the user's phone number. At most 1 phone number is allowed. Detailed below.
* `preferred_language` - (Optional) The preferred language of the user.
* `profile_url` - (Optional) An URL that may be associated with the user.
* `timezone` - (Optional) The user's time zone.
* `title` - (Optional) The user's title.
* `user_type` - (Optional) The user type.

-> Unless specified otherwise, all fields can contain up to 1024 characters of free-form text.

### addresses Configuration Block

* `country` - (Optional) The country that this address is in.
* `formatted` - (Optional) The name that is typically displayed when the address is shown for display.
* `locality` - (Optional) The address locality.
* `postal_code` - (Optional) The postal code of the address.
* `primary` - (Optional) When `true`, this is the primary address associated with the user.
* `region` - (Optional) The region of the address.
* `street_address` - (Optional) The street of the address.
* `type` - (Optional) The type of address.

### emails Configuration Block

* `primary` - (Optional) When `true`, this is the primary email associated with the user.
* `type` - (Optional) The type of email.
* `value` - (Optional) The email address. This value must be unique across the identity store.

### name Configuration Block

The following arguments are required:

* `family_name` - (Required) The family name of the user.
* `given_name` - (Required) The given name of the user.

The following arguments are optional:

* `formatted` - (Optional) The name that is typically displayed when the name is shown for display.
* `honorific_prefix` - (Optional) The honorific prefix of the user.
* `honorific_suffix` - (Optional) The honorific suffix of the user.
* `middle_name` - (Optional) The middle name of the user.

### phone_numbers Configuration Block

* `primary` - (Optional) When `true`, this is the primary phone number associated with the user.
* `type` - (Optional) The type of phone number.
* `value` - (Optional) The user's phone number.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `external_ids` - A list of identifiers issued to this resource by an external identity provider.
    * `id` - The identifier issued to this resource by an external identity provider.
    * `issuer` - The issuer for an external identifier.
* `user_id` - The identifier for this user in the identity store.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import an Identity Store User using the combination `identity_store_id/user_id`. For example:

```terraform
import {
  to = aws_identitystore_user.example
  id = "d-9c6705e95c/065212b4-9061-703b-5876-13a517ae2a7c"
}
```

Using `terraform import`, import an Identity Store User using the combination `identity_store_id/user_id`. For example:

```console
% terraform import aws_identitystore_user.example d-9c6705e95c/065212b4-9061-703b-5876-13a517ae2a7c
```
