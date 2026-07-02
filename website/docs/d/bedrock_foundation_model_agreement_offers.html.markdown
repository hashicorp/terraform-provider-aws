---
subcategory: "Bedrock"
layout: "aws"
page_title: "AWS: aws_bedrock_foundation_model_agreement_offers"
description: |-
  Provides details about AWS Bedrock Foundation Model Agreement Offers.
---

# Data Source: aws_bedrock_foundation_model_agreement_offers

Provides details about AWS Bedrock Foundation Model Agreement Offers.

## Example Usage

### Basic Usage

```terraform
data "aws_bedrock_foundation_model_agreement_offers" "example" {
  model_id = data.aws_bedrock_foundation_models.example.model_summaries[0].model_id
}
```

## Argument Reference

The following arguments are required:

* `model_id` - (Required) Model ID of the foundation model

The following arguments are optional:

* `offer_type` - (Optional) Type of offer associated with the model. Valid values are `ALL` and `PUBLIC`.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `offers` - List of the offers associated with the specified model. See [`offers`](#offers).

### `offers`

* `offer_id` - Offer ID for a model offer.
* `offer_token` - Offer token.
* `term_details` - Details about the terms of the offer. See [`term_details`](#term_details).

#### `term_details`

* `legal_term` - Details about the legal terms. See [`legal_term`](#legal_term).
* `support_term` - Details about the support terms. See [`support_term`](#support_term).
* `usage_based_pricing_term` - Details about the pricing terms. See [`usage_based_pricing_term`](#usage_based_pricing_term).
* `validity_term` - Details about the validity terms. See [`validity_term`](#validity_term).

##### `legal_term`

* `url` - URL to the legal term document.

##### `support_term`

* `refund_policy_description` - Refund policy description.

##### `usage_based_pricing_term` Block

* `rate_card` - Details about a usage price for each dimension. See [`rate_card`](#rate_card).

###### `rate_card`

* `description` - Description of the price rate.
* `dimension` - Dimension for the price rate.
* `price` - Single-dimensional rate information.
* `unit` - Unit associated with the price.

##### `validity_term`

* `agreement_duration` - Duration of the agreement.
