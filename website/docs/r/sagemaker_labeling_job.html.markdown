---
subcategory: "SageMaker AI"
layout: "aws"
page_title: "AWS: aws_sagemaker_labeling_job"
description: |-
  Manage an Amazon SageMaker labeling job.
---

# Resource: aws_sagemaker_labeling_job

Manage an Amazon SageMaker labeling job.

## Example Usage

Basic usage:

```terraform
# https://docs.aws.amazon.com/sagemaker/latest/dg/sms-named-entity-recg.html#sms-creating-ner-api.
resource "aws_sagemaker_labeling_job" "test" {
  label_attribute_name = "label1"
  labeling_job_name    = "my-labeling-job"
  role_arn             = aws_iam_role.example.arn

  label_category_config_s3_uri = "s3://${aws_s3_bucket.example.bucket}/${aws_s3_object.example.key}"

  human_task_config {
    number_of_human_workers_per_data_object = 1
    task_description                        = "Apply the labels provided to specific words or phrases within the larger text block."
    task_title                              = "Named entity Recognition task"
    task_time_limit_in_seconds              = 28800
    workteam_arn                            = aws_sagemaker_workteam.example.arn

    ui_config {
      human_task_ui_arn = "arn:aws:sagemaker:us-west-2:394669845002:human-task-ui/NamedEntityRecognition"
    }

    pre_human_task_lambda_arn = "arn:aws:lambda:us-west-2:081040173940:function:PRE-NamedEntityRecognition"

    annotation_consolidation_config {
      annotation_consolidation_lambda_arn = "arn:aws:lambda:us-west-2:081040173940:function:ACS-NamedEntityRecognition"
    }
  }

  input_config {
    data_source {
      sns_data_source {
        sns_topic_arn = aws_sns_topic.example.arn
      }
    }
  }

  output_config {
    s3_output_path = "s3://${aws_s3_bucket.example.bucket}/"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `human_task_config` - (Required) Configuration information required for human workers to complete a labeling task. Fields are documented below.
* `input_config` - (Required) Input data for the labeling job. Fields are documented below.
* `label_attribute_name` - (Required) Attribute name to use for the label in the output manifest file.
* `label_category_config_s3_uri` - (Optional) S3 URI of the file that defines the categories used to label the data objects.
* `labeling_job_algorithms_config` - (Optional) Information required to perform automated data labeling.. Fields are documented below.
* `labeling_job_name` - (Required) Name of the labeling job
* `output_config` - (Required) Location of the output data. Fields are documented below.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `role_arn` - (Required) ARN of IAM role that Amazon SageMaker assumes to perform tasks during data labeling.
* `stopping_conditions` - (Optional) Conditions for stopping a labeling job. If any of the conditions are met, the job is automatically stopped. Fields are documented below.
* `tags` - (Optional) A mapping of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### human_task_config

* `annotation_consolidation_config` - (Optional) How labels are consolidated across human workers. Fields are documented below.
* `max_concurrent_task_count` - (Optional) Maximum number of data objects that can be labeled by human workers at the same time.
* `number_of_human_workers_per_data_object` - (Required) Number of human workers that will label an object.
* `public_workforce_task_price` - (Optional) Price to pay for each task performed by an Amazon Mechanical Turk worker. Fields are documented below.
* `pre_human_task_lambda_arn` - (Optional) ARN of a Lambda function that is run before a data object is sent to a human worker.
* `task_availability_lifetime_in_seconds` - (Optional) length of time that a task remains available for labeling by human workers.
* `task_description` - (Required) Description of the task.
* `task_keywords` - (Optional) Keywords used to describe the task.
* `task_time_limit_in_seconds` - (Required) Amount of time that a worker has to complete a task.
* `task_title` - (Required) Title for the task.
* `ui_config` - (Required) Information about the user interface that workers use to complete the labeling task. Fields are documented below.
* `workteam_arn` - (Required) ARN of the work team assigned to complete the tasks.

### annotation_consolidation_config

* `annotation_consolidation_lambda_arn` - (Required) ARN of a Lambda function that implements the logic for annotation consolidation and to process output data.

### public_workforce_task_price

* `amount_in_usd` - (Optional) Amount of money paid to an Amazon Mechanical Turk worker in United States dollars. Fields are documented below.

### amount_in_usd

* `cents` - (Optional) Fractional portion, in cents, of the amount.
* `dollars` - (Optional) Whole number of dollars in the amount.
* `tenth_fractions_of_a_cent` - (Optional) Fractions of a cent, in tenths.

### ui_config

* `human_task_ui_arn` - (Optional) ARN of the worker task template used to render the worker UI and tools for labeling job tasks.
* `ui_template_s3_uri` - (Optional) S3 bucket location of the UI template, or worker task template.

### input_config

* `data_attributes` - (Optional) Attributes of the data. Fields are documented below.
* `data_source` - (Required) Location of the input data.. Fields are documented below.

### data_attributes

* `content_classifiers` - (Optional) Declares that your content is free of personally identifiable information or adult content. Valid values: `FreeOfPersonallyIdentifiableInformation`, `FreeOfAdultContent`.

### data_source

* `s3_data_source` - (Optional) S3 location of the input data objects.. Fields are documented below.
* `sns_data_source` - (Optional) SNS data source used for streaming labeling jobs. Fields are documented below.

### s3_data_source

* `manifest_s3_uri` - (Required) S3 location of the manifest file that describes the input data objects.

### sns_data_source

* `sns_topic_arn` - (Required) SNS input topic ARN.

### labeling_job_algorithms_config

* `initial_active_learning_model_arn` - (Optional) ARN of the final model used for auto-labeling.
* `labeling_job_algorithm_specification_arn` - (Required) ARN of the algorithm used for auto-labeling.
* `labeling_job_resource_config` - (Optional) Configuration information for the labeling job. Fields are documented below.

### labeling_job_resource_config

* `volume_kms_key_id` - (Optional) ID of the key that Amazon SageMaker uses to encrypt data on the storage volume attached to the ML compute instance(s) that run the training and inference jobs used for automated data labeling.
* `vpc_config` - (Optional) VPC that SageMaker jobs, hosted models, and compute resources have access to. Fields are documented below.

### vpc_config

* `security_group_ids` - (Required) VPC security group IDs.
* `subnets` - (Required) IDs of the subnets in the VPC to which to connect the training job. Fields are documented below.

### output_config

* `kms_key_id` - (Optional) ID of the key used to encrypt the output data.
* `s3_output_path` - (Required) S3 location to write output data.
* `sns_topic_arn` - (Optional) SNS output topic ARN.

### stopping_conditions

* `max_human_labeled_object_count` - (Optional) Maximum number of objects that can be labeled by human workers.
* `max_percentage_of_input_dataset_labeled` - (Optional) Maximum number of input data objects that should be labeled.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `failure_reason` - If the job failed, the reason that it failed.
* `job_reference_code` - Unique identifier for work done as part of a labeling job.
* `label_counters` - A breakdown of the number of objects labeled.
    * `failed_non_retryable_error` - Total number of objects that could not be labeled due to an error.
    * `human_labeled` - Total number of objects labeled by a human worker.
    * `machine_labeled` - Total number of objects labeled by automated data labeling.
    * `total_labeled` - Total number of objects labeled.
    * `unlabeled` - Total number of objects not yet labeled.
* `labeling_job_arn` - ARN of the labeling job.
* `labeling_job_status` - Processing status of the labeling job.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import labeling jobs using the `labeling_job_name`. For example:

```terraform
import {
  to = aws_sagemaker_labeling_job.example
  id = "my-labeling-job"
}
```

Using `terraform import`, import labeling jobs using the `labeling_job_name`. For example:

```console
% terraform import aws_sagemaker_labeling_job.example my-labeling-job
```
