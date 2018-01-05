---
layout: "aws"
page_title: "AWS: aws_cognito_user_pool_app-client"
side_bar_current: "docs-aws-resource-cognito-user-pool-app-client"
description: |-
  Provides a Cognito User Pool App Client resource.
---

# aws_cognito_user_pool_app_client

Provides a Cognito User Pool App Client resource.

## Example Usage

### Basic configuration

```hcl
data "aws_region" "current" {
  current = true
}

resource "aws_cognito_user_pool" "user_pool" {
  name = "my-user-pool"
}

resource "aws_cognito_user_pool_app_client" "app_client" {
  name         = "my-user-pool-client"
  user_pool_id = "${aws_cognito_user_pool.user_pool.id}"
}

resource "aws_cognito_identity_pool" "identity_pool" {
  identity_pool_name               = "identitypool"
  allow_unauthenticated_identities = true

  cognito_identity_providers {
    client_id               = "${aws_cognito_user_pool_app_client.app_client.id}"
    provider_name           = "cognito-idp.${data.aws_region.current.name}.amazonaws.com/${aws_cognito_user_pool.user_pool.id}"
    server_side_token_check = false
  }
}
```

### Setting attributes to be read and written by users

```hcl
resource "aws_cognito_user_pool" "user_pool" {
  name = "my-user-pool"

  schema {
    attribute_data_type      = "String"
    developer_only_attribute = false
    mutable                  = false
    name                     = "foo"
    required                 = false
  }

  schema {
    attribute_data_type      = "String"
    developer_only_attribute = false
    mutable                  = false
    name                     = "bar"
    required                 = false
  }
}

resource "aws_cognito_user_pool_app_client" "app_client" {
  name         = "my-user-pool-client"
  user_pool_id = "${aws_cognito_user_pool.user_pool.id}"

  generate_secret        = false
  refresh_token_validity = 7

  read_attributes = [
    "email",
    "email_verified",
    "name",
    "custom:foo",
    "custom:bar",
  ]

  write_attributes = [
    "email",
    "name",
    "custom:bar",
  ]
}
```

## Argument Reference

The following arguments are supported:

* `generate_secret` - (Optional) Whether to generate a secret for the user pool app client.
* `name` - (Required) The name of the user pool app client.
* `refresh_token_validity` - (Optional) The time limit, in days, after which the refresh token is no longer valid and cannot be used.
* `read_attributes` - (Optional) A list of attributes that users may read. Custom attributes defined in the user pool must be prefixed with `custom:`.
* `write_attributes` - (Optional) A list of attributes that users may write. Custom attributes defined in the user pool must be prefixed with `custom:`.

## Attribute Reference

The following additional attributes are exported:

* `id` - The id of the user pool app client.
* `client_secret` - The user pool app client's secret. This is only exported if `generate_secret` is set to true.
