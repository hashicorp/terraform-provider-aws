---
subcategory: "Bedrock"
layout: "aws"
page_title: "AWS: aws_bedrock_foundation_model_agreement"
description: |-
  Manages an AWS Bedrock Foundation Model Agreement.
---

# Resource: aws_bedrock_foundation_model_agreement

Manages an AWS Bedrock Foundation Model Agreement.

## Example Usage

### Basic Usage

```terraform
data "aws_bedrock_foundation_model_agreement_offers" "example" {
  model_id   = "eu.anthropic.claude-opus-4-5-20251101-v1:0"
  offer_type = "PUBLIC"
}

resource "aws_bedrock_foundation_model_agreement" "example" {
  model_id    = data.aws_bedrock_foundation_model_agreement_offers.example.model_id
  offer_token = data.aws_bedrock_foundation_model_agreement_offers.example.offers[0].offer_token

  lifecycle {
    ignore_changes = [offer_token]
  }
}
```

## Argument Reference

The following arguments are required:

* `model_id` - (Required) Model ID for the access request.
* `offer_token` - (Required) Offer token encapsulates information for an offer.

The following arguments are optional:

* `region` - (Optional) Region where this action should be [run](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_bedrock_foundation_model_agreement.example
  identity = {
    model_id = "eu.anthropic.claude-opus-4-5-20251101-v1:0"
  }
}

resource "aws_bedrock_foundation_model_agreement" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `model_id` - Model ID argument of the Foundation Model Agreement.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Bedrock Foundation Model Agreement using the `model_id`. For example:

```terraform
import {
  to = aws_bedrock_foundation_model_agreement.example
  id = "eu.anthropic.claude-opus-4-5-20251101-v1:0"
}
```

Using `terraform import`, import Bedrock Foundation Model Agreement using the `model_id`. For example:

```console
% terraform import aws_bedrock_foundation_model_agreement.example eu.anthropic.claude-opus-4-5-20251101-v1:0
```
