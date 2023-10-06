---
subcategory: "Amazon Bedrock"
layout: "aws"
page_title: "AWS: aws_bedrock_custom_models"
description: |-
  Returns a list of the custom models that you have created with the CreateModelCustomizationJob operation.
---

# Data Source: aws_bedrock_custom_models

Returns a list of the custom models that you have created with the CreateModelCustomizationJob operation.

## Example Usage

```terraform
data "aws_bedrock_custom_models" "test" {}
```

## Argument Reference

None

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `model_summaries` - array of CustomModelSummary objects

### CustomModelSummary

Attributes of the CustomModelSummary object.

* `base_model_arn` - The base model ARN.
* `base_model_name` - The base model name.
* `creation_time` - Creation time of the model.
* `model_arn` - The ARN of the custom model.
* `model_name` - The name of the custom model.
* `model_kms_key_arn` - The custom model is encrypted at rest using this key.
* `model_name` - Model name associated with this model.
* `output_data_config` - Output data configuration associated with this custom model.
* `training_data_config` - Information about the training dataset.
* `training_metrics` - The training metrics from the job creation.
* `validation_data_config` - Array of up to 10 validators.
* `validation_metrics` - The validation metrics from the job creation.
