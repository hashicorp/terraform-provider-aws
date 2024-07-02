---
subcategory: "Bedrock"
layout: "aws"
page_title: "AWS: aws_bedrock_foundation_models"
description: |-
  Terraform data source for managing AWS Bedrock Foundation Models.
---

# Data Source: aws_bedrock_foundation_models

Terraform data source for managing AWS Bedrock Foundation Models.

## Example Usage

### Basic Usage

```terraform
data "aws_bedrock_foundation_models" "test" {}
```

### Filter by Inference Type

```terraform
data "aws_bedrock_foundation_models" "test" {
  by_inference_type = "ON_DEMAND"
}
```

## Argument Reference

The following arguments are optional:

* `by_customization_type` - (Optional) Customization type to filter on. Valid values are `FINE_TUNING`.
* `by_inference_type` - (Optional) Inference type to filter on. Valid values are `ON_DEMAND` and `PROVISIONED`.
* `by_output_modality` - (Optional) Output modality to filter on. Valid values are `TEXT`, `IMAGE`, and `EMBEDDING`.
* `by_provider` - (Optional) Model provider to filter on.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - AWS region.
* `model_summaries` - List of model summary objects. See [`model_summaries`](#model_summaries).

### `model_summaries`

* `customizations_supported` - Customizations that the model supports.
* `inference_types_supported` - Inference types that the model supports.
* `input_modalities` - Input modalities that the model supports.
* `model_arn` - Model ARN.
* `model_id` - Model identifier.
* `model_name` - Model name.
* `output_modalities` - Output modalities that the model supports.
* `provider_name` - Model provider name.
* `response_streaming_supported` - Indicates whether the model supports streaming.
