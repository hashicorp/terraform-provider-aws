---
subcategory: "SSO Admin"
layout: "aws"
page_title: "AWS: aws_ssoadmin_application_grant"
description: |-
  Terraform resource for managing an AWS SSO Admin Application Grant.
---
# Resource: aws_ssoadmin_application_grant

Terraform resource for managing an AWS SSO Admin Application Grant.

Grants are authorization configurations that allow an IAM Identity Center application to request specific types of tokens. Each grant type defines a different OAuth 2.0 flow the application can use.

## Example Usage

### Authorization Code Grant (PKCE)

```terraform
data "aws_ssoadmin_instances" "example" {}

resource "aws_ssoadmin_application" "example" {
  name                     = "example"
  application_provider_arn = "arn:aws:sso::aws:applicationProvider/custom"
  instance_arn             = tolist(data.aws_ssoadmin_instances.example.arns)[0]
}

resource "aws_ssoadmin_application_grant" "authorization_code" {
  application_arn = aws_ssoadmin_application.example.arn
  grant_type      = "authorization_code"

  grant {
    authorization_code {
      redirect_uris = ["http://127.0.0.1/auth/callback"]
    }
  }
}
```

### Token Exchange Grant

```terraform
data "aws_ssoadmin_instances" "example" {}

resource "aws_ssoadmin_application" "example" {
  name                     = "example"
  application_provider_arn = "arn:aws:sso::aws:applicationProvider/custom"
  instance_arn             = tolist(data.aws_ssoadmin_instances.example.arns)[0]
}

resource "aws_ssoadmin_application_grant" "token_exchange" {
  application_arn = aws_ssoadmin_application.example.arn
  grant_type      = "urn:ietf:params:oauth:grant-type:token-exchange"

  grant {
    token_exchange {}
  }
}
```

### JWT Bearer Grant

```terraform
data "aws_ssoadmin_instances" "example" {}

resource "aws_ssoadmin_application" "example" {
  name                     = "example"
  application_provider_arn = "arn:aws:sso::aws:applicationProvider/custom"
  instance_arn             = tolist(data.aws_ssoadmin_instances.example.arns)[0]
}

resource "aws_ssoadmin_trusted_token_issuer" "example" {
  name                      = "example"
  instance_arn              = tolist(data.aws_ssoadmin_instances.example.arns)[0]
  trusted_token_issuer_type = "OIDC_JWT"

  trusted_token_issuer_configuration {
    oidc_jwt_configuration {
      claim_attribute_path          = "email"
      identity_store_attribute_path = "emails.value"
      issuer_url                    = "https://example.com"
      jwks_retrieval_option         = "OPEN_ID_DISCOVERY"
    }
  }
}

resource "aws_ssoadmin_application_grant" "jwt_bearer" {
  application_arn = aws_ssoadmin_application.example.arn
  grant_type      = "urn:ietf:params:oauth:grant-type:jwt-bearer"

  grant {
    jwt_bearer {
      authorized_token_issuers {
        trusted_token_issuer_arn = aws_ssoadmin_trusted_token_issuer.example.arn
        authorized_audiences     = [aws_ssoadmin_application.example.client_id]
      }
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `application_arn` - (Required, Forces new resource) ARN of the application to configure the grant for.
* `grant_type` - (Required, Forces new resource) Type of OAuth 2.0 grant. Valid values are `authorization_code`, `refresh_token`, `urn:ietf:params:oauth:grant-type:jwt-bearer`, and `urn:ietf:params:oauth:grant-type:token-exchange`.
* `grant` - (Required) Configuration block for the grant. See [`grant`](#grant-argument-reference) below.

### `grant` Argument Reference

Exactly one of the following sub-blocks must be specified, matching the `grant_type`:

* `authorization_code` - (Optional) Configuration for the `authorization_code` grant type. See [`authorization_code`](#authorization_code-argument-reference) below.
* `jwt_bearer` - (Optional) Configuration for the `urn:ietf:params:oauth:grant-type:jwt-bearer` grant type. See [`jwt_bearer`](#jwt_bearer-argument-reference) below.
* `refresh_token` - (Optional) Configuration for the `refresh_token` grant type. This block has no attributes.
* `token_exchange` - (Optional) Configuration for the `urn:ietf:params:oauth:grant-type:token-exchange` grant type. This block has no attributes.

### `authorization_code` Argument Reference

* `redirect_uris` - (Required, Forces new resource) List of URIs that are valid redirect destinations after a user authorizes the application. For loopback (127.0.0.1) URIs, IAM Identity Center ignores the port number per [RFC 8252](https://www.rfc-editor.org/rfc/rfc8252).

### `jwt_bearer` Argument Reference

* `authorized_token_issuers` - (Required) One or more trusted token issuers that are authorized to exchange JWTs for tokens. See [`authorized_token_issuers`](#authorized_token_issuers-argument-reference) below.

### `authorized_token_issuers` Argument Reference

* `trusted_token_issuer_arn` - (Optional) ARN of the trusted token issuer. Must reference an `aws_ssoadmin_trusted_token_issuer` resource.
* `authorized_audiences` - (Optional) List of application client IDs that are permitted to use the JWT issued by this trusted token issuer.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - A comma-delimited string combining `application_arn` and `grant_type`.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import SSO Admin Application Grant using the `id`. For example:

```terraform
import {
  to = aws_ssoadmin_application_grant.example
  id = "arn:aws:sso::123456789012:application/ssoins-1234567890abcdef/apl-1234567890abcdef,authorization_code"
}
```

Using `terraform import`, import SSO Admin Application Grant using the `id`. For example:

```console
% terraform import aws_ssoadmin_application_grant.example "arn:aws:sso::123456789012:application/ssoins-1234567890abcdef/apl-1234567890abcdef,authorization_code"
```
