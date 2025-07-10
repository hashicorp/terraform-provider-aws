---
subcategory: "Cognito Identity"
layout: "aws"
page_title: "AWS: aws_cognito_identity_pool"
description: |-
  Provides an AWS Cognito Identity Pool.
---

# Resource: aws_cognito_identity_pool

Provides an AWS Cognito Identity Pool.

## Example Usage

```terraform
resource "aws_iam_saml_provider" "default" {
  name                   = "my-saml-provider"
  saml_metadata_document = file("saml-metadata.xml")
}

resource "aws_cognito_identity_pool" "main" {
  identity_pool_name               = "identity pool"
  allow_unauthenticated_identities = false
  allow_classic_flow               = false

  cognito_identity_providers {
    client_id               = "6lhlkkfbfb4q5kpp90urffae"
    provider_name           = "cognito-idp.us-east-1.amazonaws.com/us-east-1_Tv0493apJ"
    server_side_token_check = false
  }

  cognito_identity_providers {
    client_id               = "7kodkvfqfb4qfkp39eurffae"
    provider_name           = "cognito-idp.us-east-1.amazonaws.com/eu-west-1_Zr231apJu"
    server_side_token_check = false
  }

  supported_login_providers = {
    "graph.facebook.com"  = "7346241598935552"
    "accounts.google.com" = "123456789012.apps.googleusercontent.com"
  }

  saml_provider_arns           = [aws_iam_saml_provider.default.arn]
  openid_connect_provider_arns = ["arn:aws:iam::123456789012:oidc-provider/id.example.com"]
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `identity_pool_name` (Required) - The Cognito Identity Pool name.
* `allow_unauthenticated_identities` (Required) - Whether the identity pool supports unauthenticated logins or not.
* `allow_classic_flow` (Optional) - Enables or disables the classic / basic authentication flow. Default is `false`.
* `developer_provider_name` (Optional) - The "domain" by which Cognito will refer to your users. This name acts as a placeholder that allows your
backend and the Cognito service to communicate about the developer provider.
* `cognito_identity_providers` (Optional) - An array of [Amazon Cognito Identity user pools](#cognito-identity-providers) and their client IDs.
* `openid_connect_provider_arns` (Optional) - Set of OpendID Connect provider ARNs.
* `saml_provider_arns` (Optional) - An array of Amazon Resource Names (ARNs) of the SAML provider for your identity.
* `supported_login_providers` (Optional) - Key-Value pairs mapping provider names to provider app IDs.
* `tags` - (Optional) A map of tags to assign to the Identity Pool. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

#### Cognito Identity Providers

* `client_id` (Optional) - The client ID for the Amazon Cognito Identity User Pool.
* `provider_name` (Optional) - The provider name for an Amazon Cognito Identity User Pool.
* `server_side_token_check` (Optional) - Whether server-side token validation is enabled for the identity provider’s token or not.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - An identity pool ID, e.g. `us-west-2:1a234567-8901-234b-5cde-f6789g01h2i3`.
* `arn` - The ARN of the identity pool.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Cognito Identity Pool using its ID. For example:

```terraform
import {
  to = aws_cognito_identity_pool.mypool
  id = "us-west-2:1a234567-8901-234b-5cde-f6789g01h2i3"
}
```

Using `terraform import`, import Cognito Identity Pool using its ID. For example:

```console
% terraform import aws_cognito_identity_pool.mypool us-west-2:1a234567-8901-234b-5cde-f6789g01h2i3
```
