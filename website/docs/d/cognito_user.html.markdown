---
subcategory: "Cognito IDP (Identity Provider)"
layout: "aws"
page_title: "AWS: aws_cognito_user"
description: |-
  Retrieves an Amazon Cognito user.
---

# Data Source: aws_cognito_user

Retrieves an Amazon Cognito user by username or exact email address. Both lookup paths use the `AdminGetUser` API operation to return full user details.

~> **Note:** Each read calls `AdminGetUser` and contributes to the user pool's monthly active user (MAU) count for billing purposes. `ListUsers`, used for email lookup, does not contribute to MAU billing. See the [`AdminGetUser` API reference](https://docs.aws.amazon.com/cognito-user-identity-pools/latest/APIReference/API_AdminGetUser.html).

## Example Usage

### Username Lookup

```hcl
data "aws_cognito_user" "example" {
  user_pool_id = "us-west-2_aaaaaaaaa"
  username     = "example-user"
}
```

### Exact Email Lookup

```hcl
data "aws_cognito_user" "example" {
  user_pool_id = "us-west-2_aaaaaaaaa"
  email        = "user@example.com"
}
```

Email lookup first calls the eventually consistent [`ListUsers` API](https://docs.aws.amazon.com/cognito-user-identity-pools/latest/APIReference/API_ListUsers.html) with an exact `email` filter. Exactly one user must match. The data source then calls `AdminGetUser` with the matched user's canonical username. Custom attributes cannot be used as lookup selectors.

An active email alias can be supplied with `username`. Use the explicit `email` selector to find users whose stored email is unverified or otherwise not an active alias.

The caller needs `cognito-idp:AdminGetUser` permission for username lookup. Email lookup needs both `cognito-idp:ListUsers` and `cognito-idp:AdminGetUser`.

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `user_pool_id` - (Required) ID of the user pool containing the user.
* `email` - (Optional) Exact value of the user's standard `email` attribute. Conflicts with `username`. Exactly one of `email` or `username` must be configured.
* `username` - (Optional) Username or configured active alias attribute of the user. Length must be between 1 and 128 characters. Conflicts with `email`. Exactly one of `email` or `username` must be configured. When `email` is configured, this attribute exports the resolved canonical username.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - ID combining `user_pool_id` and the configured `username` selector, or `user_pool_id` and the resolved canonical username for email lookup.
* `attributes` - Map of user attributes.
* `creation_date` - Date and time when the user was created in RFC3339 format.
* `enabled` - Whether the user is enabled.
* `last_modified_date` - Date and time when the user was last modified in RFC3339 format.
* `mfa_setting_list` - MFA methods activated for the user.
* `preferred_mfa_setting` - Preferred MFA method for the user.
* `status` - User status.
* `sub` - Unique user identifier.
