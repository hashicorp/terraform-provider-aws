---
subcategory: "Bedrock AgentCore"
layout: "aws"
page_title: "AWS: aws_bedrockagentcore_payment_credential_provider"
description: |-
  Manages an AWS Bedrock AgentCore Payment Credential Provider.
---

# Resource: aws_bedrockagentcore_payment_credential_provider

Manages an AWS Bedrock AgentCore Payment Credential Provider. A Payment Credential Provider stores the authentication credentials for a payment provider (Coinbase CDP or Stripe Privy) that a [Payment Connector](bedrockagentcore_payment_connector.html.markdown) can reference.

## Example Usage

### Stripe Privy

```terraform
resource "aws_bedrockagentcore_payment_credential_provider" "example" {
  name                       = "example_provider"
  credential_provider_vendor = "StripePrivy"

  provider_configuration {
    stripe_privy_configuration {
      app_id                    = "app_example"
      app_secret                = "sk_example"
      authorization_id          = "auth_example"
      authorization_private_key = "base64-encoded-ec-p256-private-key"
    }
  }
}
```

### Coinbase CDP

```terraform
resource "aws_bedrockagentcore_payment_credential_provider" "example" {
  name                       = "example_provider"
  credential_provider_vendor = "CoinbaseCDP"

  provider_configuration {
    coinbase_cdp_configuration {
      api_key_id    = "api_key_example"
      api_key_secret = "secret_example"
      wallet_secret = "wallet_example"
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `credential_provider_vendor` - (Required, Forces new resource) Vendor of the payment credential provider. Valid values: `CoinbaseCDP`, `StripePrivy`.
* `name` - (Required, Forces new resource) Name of the payment credential provider.
* `provider_configuration` - (Required) Provider configuration. Must contain exactly one of `coinbase_cdp_configuration` or `stripe_privy_configuration`. See [`provider_configuration`](#provider_configuration) below.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `provider_configuration`

The `provider_configuration` block must contain exactly one of the following:

* `coinbase_cdp_configuration` - (Optional) Coinbase CDP configuration. See [`coinbase_cdp_configuration`](#coinbase_cdp_configuration) below.
* `stripe_privy_configuration` - (Optional) Stripe Privy configuration. See [`stripe_privy_configuration`](#stripe_privy_configuration) below.

### `coinbase_cdp_configuration`

The `coinbase_cdp_configuration` block supports the following:

* `api_key_id` - (Optional) Coinbase CDP API key ID.
* `api_key_secret` - (Optional) Coinbase CDP API key secret. Write-only; not returned on read.
* `api_key_secret_config` - (Optional) Reference to an AWS Secrets Manager secret holding the API key secret. See [`secret_config`](#secret_config) below.
* `api_key_secret_source` - (Optional) Source of the API key secret. Valid values: `MANAGED`, `EXTERNAL`.
* `wallet_secret` - (Optional) Coinbase CDP wallet secret. Write-only; not returned on read.
* `wallet_secret_config` - (Optional) Reference to an AWS Secrets Manager secret holding the wallet secret. See [`secret_config`](#secret_config) below.
* `wallet_secret_source` - (Optional) Source of the wallet secret. Valid values: `MANAGED`, `EXTERNAL`.

### `stripe_privy_configuration`

The `stripe_privy_configuration` block supports the following:

* `app_id` - (Optional) Stripe Privy application ID.
* `app_secret` - (Optional) Stripe Privy application secret. Write-only; not returned on read.
* `app_secret_config` - (Optional) Reference to an AWS Secrets Manager secret holding the app secret. See [`secret_config`](#secret_config) below.
* `app_secret_source` - (Optional) Source of the app secret. Valid values: `MANAGED`, `EXTERNAL`.
* `authorization_id` - (Optional) Stripe Privy authorization ID.
* `authorization_private_key` - (Optional) Base64-encoded EC P-256 private key. Write-only; not returned on read.
* `authorization_private_key_config` - (Optional) Reference to an AWS Secrets Manager secret holding the authorization private key. See [`secret_config`](#secret_config) below.
* `authorization_private_key_source` - (Optional) Source of the authorization private key. Valid values: `MANAGED`, `EXTERNAL`.

### `secret_config`

The `*_secret_config` blocks support the following:

* `secret_id` - (Required) ID of the AWS Secrets Manager secret that stores the value.
* `json_key` - (Required) JSON key used to extract the value from the Secrets Manager secret.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `credential_provider_arn` - ARN of the Payment Credential Provider.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import a Bedrock AgentCore Payment Credential Provider using the `name`. For example:

```terraform
import {
  to = aws_bedrockagentcore_payment_credential_provider.example
  id = "example_provider"
}
```

Using `terraform import`, import a Bedrock AgentCore Payment Credential Provider using the `name`. For example:

```console
% terraform import aws_bedrockagentcore_payment_credential_provider.example example_provider
```
