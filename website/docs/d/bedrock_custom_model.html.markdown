---
subcategory: "Bedrock"
layout: "aws"
page_title: "AWS: aws_bedrock_custom_model"
description: |-
  Returns properties of a specific Amazon Bedrock custom model.
---

# Data Source: aws_bedrock_custom_model

Returns properties of a specific Amazon Bedrock custom model.

## Example Usage

```terraform
data "aws_bedrock_custom_model" "test" {
  model_id = "arn:aws:bedrock:us-west-2:123456789012:custom-model/amazon.titan-text-express-v1:0:8k/ly16hhi765j4 "
}
```

## Argument Reference

* `model_id` – (Required) Name or ARN of the custom model.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `base_model_arn` - ARN of the base model.
* `creation_time` - Creation time of the model.
* `hyperparameters` - Hyperparameter values associated with this model.
* `job_arn` - Job ARN associated with this model.
* `job_name` - Job name associated with this model.
* `job_tags` - Key-value mapping of tags for the fine-tuning job.
* `model_arn` - ARN associated with this model.
* `model_kms_key_arn` - The custom model is encrypted at rest using this key.
* `model_name` - Model name associated with this model.
* `model_tags` - Key-value mapping of tags for the model.
* `output_data_config` - Output data configuration associated with this custom model.
    * `s3_uri` - The S3 URI where the output data is stored.
* `training_data_config` - Information about the training dataset.
    * `s3_uri` - The S3 URI where the training data is stored.
* `training_metrics` - Metrics associated with the customization job.
    * `training_loss` - Loss metric associated with the customization job.
* `validation_data_config` - Information about the validation dataset.
    * `validator` - Information about the validators.
        * `s3_uri` - The S3 URI where the validation data is stored..
* `validation_metrics` - The loss metric for each validator that you provided.
    * `validation_loss` - The validation loss associated with the validator.
  