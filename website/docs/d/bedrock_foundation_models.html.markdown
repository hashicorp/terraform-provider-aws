---
subcategory: "Amazon Bedrock"
layout: "aws"
page_title: "AWS: aws_bedrock_foundation_models"
description: |-
  List of Bedrock foundation models that you can use.
---

# Data Source: aws_bedrock_foundation_models

List of Bedrock foundation models that you can use.

## Example Usage

```terraform
data "aws_bedrock_foundation_models" "test" {}
```

## Argument Reference

NONE

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `model_summaries` - array of FoundationModelSummary objects

### FoundationModelSummary

Attributes of the FoundationModelSummary object.

* `model_arn` - The ARN of the foundation model.
* `model_id` - The model Id of the foundation model.
* `customizations_supported` - Whether the model supports fine-tuning or continual pre-training.
* `inference_types_supported` - he inference types that the model supports.
* `input_modalities` - The input modalities that the model supports.
* `model_name` - The model name.
* `output_modalities` - The output modalities that the model supports.
* `provider_name` - The model's provider name.
* `response_streaming_supported` - Indicates whether the model supports streaming.
