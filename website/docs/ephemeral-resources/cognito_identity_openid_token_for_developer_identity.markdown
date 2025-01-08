---
subcategory: "Cognito Identity"
layout: "aws"
page_title: "AWS: aws_cognito_identity_openid_token_for_developer_identity"
description: |-
  Terraform ephemeral resource for managing an AWS Cognito Identity Open ID Token for Developer Identity.
---


# Ephemeral: aws_cognito_identity_openid_token_for_developer_identity

Terraform ephemeral resource for managing an AWS Cognito Identity Open ID Token for Developer Identity.

~> Ephemeral resources are a new feature and may evolve as we continue to explore their most effective uses. [Learn more](https://developer.hashicorp.com/terraform/language/v1.10.x/resources/ephemeral).

## Example Usage

### Basic Usage

```terraform
data "aws_cognito_identity_pool" "example" {
  identity_pool_name = "test pool"
}

ephemeral "aws_cognito_identity_openid_token_for_developer_identity" "example" {
  identity_pool_id = data.aws_cognito_identity_pool.example.id
  logins = {
    "login.mycompany.myapp" : "USER_IDENTIFIER"
  }
}
```

## Argument Reference

The following arguments are required:

* `identity_pool_id` - (Required) An identity pool ID in the format REGION:GUID.

The following arguments are optional:

* `identity_id` - (Optional) A unique identifier in the format REGION:GUID.

* `logins` - (Optional) A set of optional name-value pairs that map provider names to provider tokens. Each name-value pair represents a user from a public provider or developer provider. If the user is from a developer provider, the name-value pair will follow the syntax `"developer_provider_name": "developer_user_identifier"`. The developer provider is the "domain" by which Cognito will refer to your users; you provided this domain while creating/updating the identity pool. The developer user identifier is an identifier from your backend that uniquely identifies a user. When you create an identity pool, you can specify the supported logins.

* `principal_tags` - (Optional) Use this operation to configure attribute mappings for custom providers.

* `token_duration` - (Optional) The expiration time of the token, in seconds. You can specify a custom expiration time for the token so that you can cache it. If you don't provide an expiration time, the token is valid for 15 minutes. You can exchange the token with Amazon STS for temporary AWS credentials, which are valid for a maximum of one hour. The maximum token duration you can set is 24 hours. You should take care in setting the expiration time for a token, as there are significant security implications: an attacker could use a leaked token to access your AWS resources for the token's duration.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `token` - An OpenID token.
