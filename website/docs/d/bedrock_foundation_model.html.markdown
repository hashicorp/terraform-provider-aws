---
subcategory: "Amazon Bedrock"
layout: "aws"
page_title: "AWS: aws_bedrock_foundation_model"
description: |-
  Get details about a Bedrock foundation model.
---

# Data Source: aws_bedrock_foundation_model

Get details about a Bedrock foundation model.

## Example Usage

```terraform
data "aws_bedrock_foundation_models" "test" {

}

data "aws_bedrock_foundation_model" "test" {
  model_id = data.aws_bedrock_foundation_models.test.model_summaries[0].model_id
}
```

## Argument Reference

* `model_id` – (Required) The model identifier. Length Constraints: Minimum length of 1. Maximum length of 2048. Pattern: ^arn:aws(-[^:]+)?:bedrock:[a-z0-9-]{1,20}:(([0-9]{12}:custom-model/[a-z0-9-]{1,63}[.]{1}[a-z0-9-]{1,63}(([:][a-z0-9-]{1,63}){0,2})?/[a-z0-9]{12})|(:foundation-model/([a-z0-9-]{1,63}[.]{1}[a-z0-9-]{1,63}([.]?[a-z0-9-]{1,63})([:][a-z0-9-]{1,63}){0,2})))|(([a-z0-9-]{1,63}[.]{1}[a-z0-9-]{1,63}([.]?[a-z0-9-]{1,63})([:][a-z0-9-]{1,63}){0,2}))|(([0-9a-zA-Z][_-]?)+)$

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `model_arn` - The model ARN. Type: String. Pattern: ^arn:aws(-[^:]+)?:bedrock:[a-z0-9-]{1,20}::foundation-model/[a-z0-9-]{1,63}[.]{1}([a-z0-9-]{1,63}[.]){0,2}[a-z0-9-]{1,63}([:][a-z0-9-]{1,63}){0,2}$. Required: Yes.
* `model_id` - The model identifier. Type: String. Length Constraints: Minimum length of 0. Maximum length of 140. Pattern: ^[a-z0-9-]{1,63}[.]{1}[a-z0-9-]{1,63}([a-z0-9-]{1,63}[.]){0,2}[a-z0-9-]{1,63}([:][a-z0-9-]{1,63}){0,2}(/[a-z0-9]{12}|)$. Required: Yes.
* `customizations_supported` - The customization that the model supports. Type: Array of strings. Valid Values FINE_TUNING. Required: No.
* `inference_types_supported` - The inference types that the model supports. Type: Array of strings. Valid Values: ON_DEMAND | PROVISIONED. Required: No.
* `input_modalities` - The input modalities that the model supports. Type: Array of strings. Valid Values: TEXT | IMAGE | EMBEDDING. Required: No.
* `model_name` - The model name. Type: String. Length Constraints: Minimum length of 1. Maximum length of 20. Pattern: ^.*$. Required: No.
* `output_modalities` - The output modalities that the model supports. Type: Array of strings. Valid Values: TEXT | IMAGE | EMBEDDING. Required: No.
* `provider_name` - The model's provider name. Type: String. Length Constraints: Minimum length of 1. Maximum length of 20. Pattern: ^.*$. Required: No
* `response_streaming_supported` - Indicates whether the model supports streaming. Type: Boolean. Required: No.
