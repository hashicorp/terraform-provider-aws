---
subcategory: "SSO Admin"
layout: "aws"
page_title: "AWS: aws_ssoadmin_trusted_token_issuer"
description: |-
  Terraform resource for managing an AWS SSO Admin Trusted Token Issuer.
---
# Resource: aws_ssoadmin_trusted_token_issuer

Terraform resource for managing an AWS SSO Admin Trusted Token Issuer.

## Example Usage

### Basic Usage

```terraform
data "aws_ssoadmin_instances" "example" {}

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
```

## Argument Reference

The following arguments are required:

* `instance_arn` - (Required) ARN of the instance of IAM Identity Center.
* `name` - (Required) Name of the trusted token issuer.
* `trusted_token_issuer_configuration` - (Required) A block that specifies settings that apply to the trusted token issuer, these change depending on the type you specify in `trusted_token_issuer_type`. [Documented below](#trusted_token_issuer_configuration-argument-reference).
* `trusted_token_issuer_type` - (Required) Specifies the type of the trusted token issuer. Valid values are `OIDC_JWT`

The following arguments are optional:

* `client_token` - (Optional) A unique, case-sensitive ID that you provide to ensure the idempotency of the request. AWS generates a random value when not provided.
* `tags` - (Optional) Key-value mapping of resource tags. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `trusted_token_issuer_configuration` Argument Reference

* `oidc_jwt_configuration` - (Optional) A block that describes the settings for a trusted token issuer that works with OpenID Connect (OIDC) by using JSON Web Tokens (JWT). See [Documented below](#oidc_jwt_configuration-argument-reference) below.

### `oidc_jwt_configuration` Argument Reference

* `claim_attribute_path` - (Required) Specifies the path of the source attribute in the JWT from the trusted token issuer.
* `identity_store_attribute_path` - (Required) Specifies path of the destination attribute in a JWT from IAM Identity Center. The attribute mapped by this JMESPath expression is compared against the attribute mapped by `claim_attribute_path` when a trusted token issuer token is exchanged for an IAM Identity Center token.
* `issuer_url` - (Required) Specifies the URL that IAM Identity Center uses for OpenID Discovery. OpenID Discovery is used to obtain the information required to verify the tokens that the trusted token issuer generates.
* `jwks_retrieval_option` - (Required) The method that the trusted token issuer can use to retrieve the JSON Web Key Set used to verify a JWT. Valid values are `OPEN_ID_DISCOVERY`

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the trusted token issuer.
* `id` - ARN of the trusted token issuer.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import SSO Admin Trusted Token Issuer using the `id`. For example:

```terraform
import {
  to = aws_ssoadmin_trusted_token_issuer.example
  id = "arn:aws:sso::123456789012:trustedTokenIssuer/ssoins-lu1ye3gew4mbc7ju/tti-2657c556-9707-11ee-b9d1-0242ac120002"
}
```

Using `terraform import`, import SSO Admin Trusted Token Issuer using the `id`. For example:

```console
% terraform import aws_ssoadmin_trusted_token_issuer.example arn:aws:sso::123456789012:trustedTokenIssuer/ssoins-lu1ye3gew4mbc7ju/tti-2657c556-9707-11ee-b9d1-0242ac120002
```
