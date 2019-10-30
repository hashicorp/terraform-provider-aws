---
layout: "aws"
page_title: "AWS: aws_cognito_user_pool_identity_provider"
description: |-
  Provides a Cognito User Pool Identity Provider resource.
---

# Resource: aws_cognito_user_pool_identity_provider

Provides a Cognito User Pool Identity Provider resource.

[Cognito User Pool Identity Providers allow](https://docs.aws.amazon.com/cognito/latest/developerguide/cognito-user-pools-identity-federation.html)
sign in from third party identity providers via the hosted web UI.
This feature is independent of federation through Amazon Cognito identity pools (federated identities).

## Example Usage

### Basic configuration

```hcl
resource "aws_cognito_user_pool" "pool" {
  name = "example"
}

resource "aws_cognito_user_pool_identity_provider" "google" {
  provider_name    = "Google"
  user_pool_id     = "${aws_cognito_user_pool.pool.id}"
  provider_type    = "Google"

  provider_details = {
    authorize_scopes              = "email"
    attributes_url_add_attributes = "true"
    authorize_url                 = "https://accounts.google.com/o/oauth2/v2/auth"
    token_url                     = "https://www.googleapis.com/oauth2/v4/token"
    oidc_issuer                   = "https://accounts.google.com"
    client_id                     = "123456789012-a1b2c3d4f5g6h7i8j9k0l1m2n3o4p5q6.apps.googleusercontent.com"
    attributes_url                = "https://people.googleapis.com/v1/people/me?personFields="
    client_secret                 = "rAnDoMly_GeNeRaTeD_sEcReT"
    token_request_method          = "POST"
  }
}
```

## Argument Reference

The following arguments are supported:

* `user_pool_id` - (Required) The user pool that this identity provider should be attached to.
* `provider_name` - (Required) The provider name. For provider types `Facebook`, `Google` and `LoginWithAmazon` this should be set the same.
* `provider_type` - (Required) The provider type. Possible values: `Facebook`, `Google`, `LoginWithAmazon`, `OIDC`, `SAML`.
* `provider_details` - (Required) A map of provider attributes. Each provider type requires different fields. 
* `attribute_mapping` - (Optional) A mapping of identity provider attributes to standard and custom user pool attributes.
* `idp_identifiers` - (Optional) A list of identity provider identifiers.

## Import

Cognito User Pool Identity Providers can be imported using `user_pool_id:provider_name`, e.g.

```
$ terraform import aws_cognito_user_pool_identity_provider.google us-west-2_pGw0UpvKt:Google
```
