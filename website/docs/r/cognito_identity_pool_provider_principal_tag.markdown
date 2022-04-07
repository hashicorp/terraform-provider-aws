---
subcategory: "Cognito"
layout: "aws"
page_title: "AWS: aws_cognito_identity_pool_provider_principal_tag"
description: |-
  Provides an AWS Cognito Identity Principal Mapping.
---

# Resource: aws_cognito_identity_pool_provider_principal_tag

Provides an AWS Cognito Identity Principal Mapping.

## Example Usage

```terraform
resource "aws_cognito_user_pool" "example" {
  name                     = "user pool"
  auto_verified_attributes = ["email"]
}

resource "aws_cognito_user_pool_client" "example" {
  name         = "client"
  user_pool_id = aws_cognito_user_pool.example.id
  supported_identity_providers = compact([
    "COGNITO",
  ])
}

resource "aws_cognito_identity_pool" "example" {
  identity_pool_name               = "identity pool"
  allow_unauthenticated_identities = false
  cognito_identity_providers {
    client_id               = aws_cognito_user_pool_client.example.id
    provider_name           = aws_cognito_user_pool.example.endpoint
    server_side_token_check = false
  }
}

resource "aws_cognito_identity_pool_provider_principal_tag" "example" {
  identity_pool_id       = aws_cognito_identity_pool.example.id
  identity_provider_name = aws_cognito_user_pool.example.endpoint
  use_defaults           = false
  principal_tags = {
    test = "value"
  }
}
```

## Argument Reference

The following arguments are supported:

* `identity_pool_id` (Required) - An identity pool ID.
* `identity_provider_name` (Required) - The name of the identity provider.
* `principal_tags`: (Optional: []) - String to string map of variables.
* `use_defaults`: (Optional: true) use default (username and clientID) attribute mappings.

## Attributes Reference

No additional attributes are exported.

## Import

Cognito Identity Pool Roles Attachment can be imported using the Identity Pool ID and provider name, e.g.,

```
$ terraform import aws_cognito_identity_pool_provider_principal_tag.example us-west-2_abc123:CorpAD
```
