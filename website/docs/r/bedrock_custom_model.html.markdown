---
subcategory: "Bedrock"
layout: "aws"
page_title: "AWS: aws_bedrock_custom_model"
description: |-
  Manages an Amazon Bedrock custom model.
---

# Resource: aws_bedrock_custom_model

Manages an Amazon Bedrock custom model.
Model customization is the process of providing training data to a base model in order to improve its performance for specific use-cases.

This Terraform resource interacts with two Amazon Bedrock entities:

1. A Continued Pre-training or Fine-tuning job which is started when the Terraform resource is created. The customization job can take several hours to run to completion. The duration of the job depends on the size of the training data (number of records, input tokens, and output tokens), and [hyperparameters](https://docs.aws.amazon.com/bedrock/latest/userguide/custom-models-hp.html) (number of epochs, and batch size).
2. The custom model output on successful completion of the customization job.

This resource's [behaviors](https://developer.hashicorp.com/terraform/language/resources/behavior) correspond to operations on these Amazon Bedrock entities:

* [_Create_](https://developer.hashicorp.com/terraform/plugin/framework/resources/create) starts the customization job and immediately returns.
* [_Read_](https://developer.hashicorp.com/terraform/plugin/framework/resources/read) returns the status and results of the customization job. If the customization job has completed, the output model's properties are returned.
* [_Update_](https://developer.hashicorp.com/terraform/plugin/framework/resources/update) updates the customization job's [tags](https://docs.aws.amazon.com/bedrock/latest/userguide/tagging.html).
* [_Delete_](https://developer.hashicorp.com/terraform/plugin/framework/resources/delete) stops the customization job if it is still active. If the customization job has completed, the custom model output by the job is deleted.

## Example Usage

```terraform
data "aws_bedrock_foundation_model" "example" {
  model_id = "amazon.titan-text-express-v1"
}

resource "aws_bedrock_custom_model" "example" {
  custom_model_name     = "example-model"
  job_name              = "example-job-1"
  base_model_identifier = data.aws_bedrock_foundation_model.example.model_arn
  role_arn              = aws_iam_role.example.arn

  hyperparameters = {
    "epochCount"              = "1"
    "batchSize"               = "1"
    "learningRate"            = "0.005"
    "learningRateWarmupSteps" = "0"
  }

  output_data_config {
    s3_uri = "s3://${aws_s3_bucket.output.id}/data/"
  }

  training_data_config {
    s3_uri = "s3://${aws_s3_bucket.training.id}/data/train.jsonl"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `base_model_identifier` - (Required) The Amazon Resource Name (ARN) of the base model.
* `custom_model_kms_key_id` - (Optional) The custom model is encrypted at rest using this key. Specify the key ARN.
* `custom_model_name` - (Required) Name for the custom model.
* `customization_type` -(Optional) The customization type. Valid values: `FINE_TUNING`, `CONTINUED_PRE_TRAINING`.
* `hyperparameters` - (Required) [Parameters](https://docs.aws.amazon.com/bedrock/latest/userguide/custom-models-hp.html) related to tuning the model.
* `job_name` - (Required) A name for the customization job.
* `output_data_config` - (Required) S3 location for the output data.
    * `s3_uri` - (Required) The S3 URI where the output data is stored.
* `role_arn` - (Required) The Amazon Resource Name (ARN) of an IAM role that Bedrock can assume to perform tasks on your behalf.
* `tags` - (Optional) A map of tags to assign to the customization job and custom model. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `training_data_config` - (Required) Information about the training dataset.
    * `s3_uri` - (Required) The S3 URI where the training data is stored.
* `validation_data_config` - (Optional) Information about the validation dataset.
    * `validator` - (Required) Information about the validators.
        * `s3_uri` - (Required) The S3 URI where the validation data is stored.
* `vpc_config` - (Optional) Configuration parameters for the private Virtual Private Cloud (VPC) that contains the resources you are using for this job.
    * `security_group_ids` – (Required) VPC configuration security group IDs.
    * `subnet_ids` – (Required) VPC configuration subnets.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `custom_model_arn` - The ARN of the output model.
* `job_arn` - The ARN of the customization job.
* `job_status` - The status of the customization job. A successful job transitions from `InProgress` to `Completed` when the output model is ready to use.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `training_metrics` - Metrics associated with the customization job.
    * `training_loss` - Loss metric associated with the customization job.
* `validation_metrics` - The loss metric for each validator that you provided.
    * `validation_loss` - The validation loss associated with the validator.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `delete` - (Default `120m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Bedrock Custom Model using the `job_arn`. For example:

```terraform
import {
  to       = aws_bedrock_custom_model.example
  model_id = "arn:aws:bedrock:us-west-2:123456789012:model-customization-job/amazon.titan-text-express-v1:0:8k/1y5n57gh5y2e"
}
```

Using `terraform import`, import Bedrock custom model using the `job_arn`. For example:

```console
% terraform import aws_bedrock_custom_model.example arn:aws:bedrock:us-west-2:123456789012:model-customization-job/amazon.titan-text-express-v1:0:8k/1y5n57gh5y2e
```
