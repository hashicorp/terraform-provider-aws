---
subcategory: "Amazon Bedrock"
layout: "aws"
page_title: "AWS: aws_bedrock_custom_model"
description: |-
  Get the properties associated with a Bedrock custom model that you have created
---

# Data Source: aws_bedrock_custom_model

Get the properties associated with a Bedrock custom model that you have created.

## Example Usage

```terraform
data "aws_bedrock_custom_model" "test" {
  model_arn = "<arn of model>"
}
```

## Argument Reference

* `model_arn` – (Required) Name or ARN of the custom model.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `base_model_arn` - ARN of the base model.
* `creation_time` - Creation time of the model.
* `hyper_parameters` - Hyperparameter values associated with this model.
* `job_arn` - Job ARN associated with this model.
* `job_name` - Job name associated with this model.
* `model_kms_key_arn` - The custom model is encrypted at rest using this key.
* `model_name` - Model name associated with this model.
* `output_data_config` - Output data configuration associated with this custom model.
* `training_data_config` - Information about the training dataset.
* `training_metrics` - The training metrics from the job creation.
* `validation_data_config` - Array of up to 10 validators.
* `validation_metrics` - The validation metrics from the job creation.
* `job_tags` - Key-value mapping of tags for the fine-tuning job.
* `tags` - Key-value mapping of resource tags.
  