---
subcategory: "Cognito IDP (Identity Provider)"
layout: "aws"
page_title: "AWS: aws_cognito_identity_provider"
side_bar_current: "docs-aws-resource-cognito-identity-provider"
description: |-
  Provides a Cognito User Identity Provider resource.
---

# Resource: aws_cognito_identity_provider

Provides a Cognito User Identity Provider resource.

## Example Usage

```terraform
resource "aws_cognito_user_pool" "example" {
  name                     = "example-pool"
  auto_verified_attributes = ["email"]
}

resource "aws_cognito_identity_provider" "example_provider" {
  user_pool_id  = aws_cognito_user_pool.example.id
  provider_name = "Google"
  provider_type = "Google"

  provider_details = {
    authorize_scopes = "email"
    client_id        = "your client_id"
    client_secret    = "your client_secret"
  }

  attribute_mapping = {
    email    = "email"
    username = "sub"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `user_pool_id` (Required) - The user pool id
* `provider_name` (Required) - The provider name
* `provider_type` (Required) - The provider type.  [See AWS API for valid values](https://docs.aws.amazon.com/cognito-user-identity-pools/latest/APIReference/API_CreateIdentityProvider.html#CognitoUserPools-CreateIdentityProvider-request-ProviderType)
* `attribute_mapping` (Optional) - The map of attribute mapping of user pool attributes. [AttributeMapping in AWS API documentation](https://docs.aws.amazon.com/cognito-user-identity-pools/latest/APIReference/API_CreateIdentityProvider.html#CognitoUserPools-CreateIdentityProvider-request-AttributeMapping)
* `idp_identifiers` (Optional) - The list of identity providers.
* `provider_details` (Optional) - The map of identity details, such as access token

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_cognito_identity_provider` resources using their User Pool ID and Provider Name. For example:

```terraform
import {
  to = aws_cognito_identity_provider.example
  id = "us-west-2_abc123:CorpAD"
}
```

Using `terraform import`, import `aws_cognito_identity_provider` resources using their User Pool ID and Provider Name. For example:

```console
% terraform import aws_cognito_identity_provider.example us-west-2_abc123:CorpAD
```
