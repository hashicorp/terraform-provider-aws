---
subcategory: "Bedrock AgentCore"
layout: "aws"
page_title: "AWS: aws_bedrockagentcore_payment_connector"
description: |-
  Manages an AWS Bedrock AgentCore Payment Connector.
---

# Resource: aws_bedrockagentcore_payment_connector

Manages an AWS Bedrock AgentCore Payment Connector. A Payment Connector links a [Payment Manager](bedrockagentcore_payment_manager.html.markdown) to one or more [Payment Credential Providers](bedrockagentcore_payment_credential_provider.html.markdown) for a specific payment provider type.

## Example Usage

### Basic Usage

```terraform
resource "aws_bedrockagentcore_payment_connector" "example" {
  name               = "example_connector"
  payment_manager_id = aws_bedrockagentcore_payment_manager.example.payment_manager_id
  type               = "StripePrivy"

  credential_provider_configuration {
    stripe_privy {
      credential_provider_arn = aws_bedrockagentcore_payment_credential_provider.example.credential_provider_arn
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `credential_provider_configuration` - (Required) One or more credential provider configurations. See [`credential_provider_configuration`](#credential_provider_configuration) below.
* `name` - (Required, Forces new resource) Name of the payment connector. Must start with a letter and contain only letters, numbers, and underscores.
* `payment_manager_id` - (Required, Forces new resource) Identifier of the Payment Manager that owns this connector.
* `type` - (Required, Forces new resource) Payment connector type. Valid values: `CoinbaseCDP`, `StripePrivy`.

The following arguments are optional:

* `description` - (Optional) Description of the payment connector.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

### `credential_provider_configuration`

The `credential_provider_configuration` block must contain exactly one of the following:

* `coinbase_cdp` - (Optional) Coinbase CDP credential provider reference. See [`credential provider reference`](#credential-provider-reference) below.
* `stripe_privy` - (Optional) Stripe Privy credential provider reference. See [`credential provider reference`](#credential-provider-reference) below.

### Credential Provider Reference

The `coinbase_cdp` and `stripe_privy` blocks support the following:

* `credential_provider_arn` - (Required) ARN of the payment credential provider that stores the authentication credentials for this payment provider.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `payment_connector_id` - Unique identifier of the Payment Connector.
* `status` - Status of the Payment Connector.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.12.0 and later, you can use an [`import` block](https://developer.hashicorp.com/terraform/language/import) with the `identity` attribute. For example:

```terraform
import {
  to = aws_bedrockagentcore_payment_connector.example
  identity = {
    payment_manager_id   = "payment-manager-id-12345678"
    payment_connector_id = "payment-connector-id-12345678"
  }
}

resource "aws_bedrockagentcore_payment_connector" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

- `payment_manager_id` (String) Payment manager ID.
- `payment_connector_id` (String) Payment connector ID.

#### Optional

* `account_id` (String) AWS account ID for this resource.
* `region` (String) AWS Region for this resource.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import a Bedrock AgentCore Payment Connector using the `payment_manager_id` and `payment_connector_id` separated by a comma. For example:

```terraform
import {
  to = aws_bedrockagentcore_payment_connector.example
  id = "payment-manager-id-12345678,payment-connector-id-12345678"
}
```

Using `terraform import`, import a Bedrock AgentCore Payment Connector using the `payment_manager_id` and `payment_connector_id` separated by a comma. For example:

```console
% terraform import aws_bedrockagentcore_payment_connector.example payment-manager-id-12345678,payment-connector-id-12345678
```
