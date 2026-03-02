---
subcategory: "IAM (Identity & Access Management)"
layout: "aws"
page_title: "AWS: aws_iam_outbound_web_identity_federation"
description: |-
  Provides details about IAM Outbound Web Identity Federation configuration.
---

# Data Source: aws_iam_outbound_web_identity_federation

Provides details about the IAM Outbound Web Identity Federation configuration for the AWS account.

This data source can be used to check whether outbound web identity federation is enabled and retrieve the issuer identifier URL.

## Example Usage

### Basic Usage

```terraform
data "aws_iam_outbound_web_identity_federation" "example" {}

output "federation_enabled" {
  value = data.aws_iam_outbound_web_identity_federation.example.enabled
}

output "issuer_url" {
  value = data.aws_iam_outbound_web_identity_federation.example.issuer_identifier
}
```

## Argument Reference

This data source does not support any arguments.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `enabled` - Whether outbound identity federation is currently enabled for the AWS account. When `true`, IAM principals in the account can call the `GetWebIdentityToken` API to obtain JSON Web Tokens (JWTs) for authentication with external services.
* `issuer_identifier` - Unique issuer URL for the AWS account that hosts the OpenID Connect (OIDC) discovery endpoints at `/.well-known/openid-configuration` and `/.well-known/jwks.json`. The OpenID Connect (OIDC) discovery endpoints contain verification keys and metadata necessary for token verification. `null` if the feature is disabled.
