---
subcategory: "SageMaker AI"
layout: "aws"
page_title: "AWS: aws_sagemaker_monitoring_schedule"
description: |-
  Provides a SageMaker AI Monitoring Schedule resource.
---

# Resource: aws_sagemaker_monitoring_schedule

Provides a SageMaker AI monitoring schedule resource.

## Example Usage

Basic usage:

```terraform
resource "aws_sagemaker_monitoring_schedule" "test" {
  name = "my-monitoring-schedule"

  monitoring_schedule_config {
    monitoring_job_definition_name = aws_sagemaker_data_quality_job_definition.test.name
    monitoring_type                = "DataQuality"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `monitoring_schedule_config` - (Required) The configuration object that specifies the monitoring schedule and defines the monitoring job. Fields are documented below.
* `name` - (Optional) The name of the monitoring schedule. The name must be unique within an AWS Region within an AWS account. If omitted, Terraform will assign a random, unique name.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) A mapping of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### monitoring_schedule_config

* `monitoring_job_definition` - (Optional) Defines the monitoring job. Fields are documented below.
* `monitoring_job_definition_name` - (Required) The name of the monitoring job definition to schedule.
* `monitoring_type` - (Required) The type of the monitoring job definition to schedule. Valid values are `DataQuality`, `ModelQuality`, `ModelBias` or `ModelExplainability`
* `schedule_config` - (Optional) Configures the monitoring schedule. Fields are documented below.

### monitoring_job_definition

* `baseline` - (Optional) Baseline configuration used to validate that the data conforms to the specified constraints and statistics. Fields are documented below.
* `environment` - (Optional) Map of environment variables in the Docker container.
* `monitoring_app_specification` - (Required) Configures the monitoring job to run a specified Docker container image. Fields are documented below.
* `monitoring_input` - (Required) Inputs for the monitoring job. Fields are documented below.
* `monitoring_output_config` - (Required) Outputs from the monitoring job to be uploaded to Amazon S3. Fields are documented below.
* `monitoring_resources` - (Required) Identifies the resources, ML compute instances, and ML storage volumes to deploy for a monitoring job. Fields are documented below.
* `network_config` - (Optional) Networking options for the monitoring job. Fields are documented below.
* `role_arn` - (Required) ARN of an IAM role that Amazon SageMaker AI can assume to perform tasks on your behalf.
* `stopping_condition` - (Optional) Networking options for the monitoring job. Fields are documented below.

### monitoring_app_specification

* `container_arguments` - (Optional) List of arguments for the container used to run the monitoring job.
* `container_entrypoint` - (Optional) Entrypoint for the container used to run the monitoring job.
* `image_uri` - (Required) Container image to be run by the monitoring job.
* `post_analytics_processor_source_uri` - (Optional) Script that is called after analysis has been performed.
* `record_preprocessor_source_uri` - (Optional) Script that is called per row prior to running analysis.

### monitoring_input

* `batch_transform_input` - (Optional) Input object for the batch transform job. Fields are documented below.
* `endpoint_input` - (Optional) Endpoint for a monitoring job. Fields are documented below.

### batch_transform_input

* `data_captured_destination_s3_uri` - (Required) Amazon S3 location being used to capture the data.
* `dataset_format` - (Required) Dataset format for the batch transform job. Fields are documented below.
* `end_time_offset` - (Optional) Monitoring jobs subtract this time from the end time.
* `exclude_features_attribute` - (Optional) Attributes of the input data to exclude from the analysis.
* `features_attribute` - (Optional) Attributes of the input data that are the input features.
* `inference_attribute` - (Optional) Attribute of the input data that represents the ground truth label.
* `local_path` - (Required) Path to the filesystem where the batch transform data is available to the container.
* `probability_attribute` - (Optional) In a classification problem, the attribute that represents the class probability.
* `probability_threshold_attribute` - (Optional) Threshold for the class probability to be evaluated as a positive result.
* `s3_data_distribution_type` - (Optional) Whether input data distributed in Amazon S3 is fully replicated or sharded by an S3 key. Valid values: `FullyReplicated`, `ShardedByS3Key`.
* `s3_input_mode` - (Optional) Input mode for transferring data for the monitoring job. Valid values: `Pipe`, `File`.
* `start_time_offset` - (Optional) Monitoring jobs subtract this time from the start time.

### dataset_format

* `csv` - (Optional) CSV dataset used in the monitoring job. Fields are documented below.
* `json` - (Optional) JSON dataset used in the monitoring job. Fields are documented below.

### csv

* `header` - (Optional) Indicates if the CSV data has a header.

### json

* `line` - (Optional) Indicates if the file should be read as a JSON object per line.

### endpoint_input

* `end_time_offset` - (Optional) Monitoring jobs subtract this time from the end time.
* `endpoint_name` - (Required) Endpoint in customer's account which has enabled `DataCaptureConfig`.
* `exclude_features_attribute` - (Optional) Attributes of the input data to exclude from the analysis.
* `features_attribute` - (Optional) Attributes of the input data that are the input features.
* `inference_attribute` - (Optional) Attribute of the input data that represents the ground truth label.
* `local_path` - (Required) Path to the filesystem where the endpoint data is available to the container.
* `probability_attribute` - (Optional) In a classification problem, the attribute that represents the class probability.
* `probability_threshold_attribute` - (Optional) Threshold for the class probability to be evaluated as a positive result.
* `s3_data_distribution_type` - (Optional) Whether input data distributed in Amazon S3 is fully replicated or sharded by an S3 key. Valid values: `FullyReplicated`, `ShardedByS3Key`.
* `s3_input_mode` - (Optional) Input mode for transferring data for the monitoring job. Valid values: `Pipe`, `File`.
* `start_time_offset` - (Optional) Monitoring jobs subtract this time from the start time.

### monitoring_output_config

* `kms_key_id` - (Optional) AWS KMS key that Amazon SageMaker AI uses to encrypt the model artifacts at rest using Amazon S3 server-side encryption.
* `monitoring_output` - (Required) Monitoring outputs for monitoring jobs. Fields are documented below.

### monitoring_output

* `s3_output` - (Required) Amazon S3 storage location where the results of a monitoring job are saved. Fields are documented below.

### s3_output

* `local_path` - (Required) Local path to the Amazon S3 storage location where Amazon SageMaker AI saves the results of a monitoring job.
* `s3_uri` - (Required) URI that identifies the Amazon S3 storage location where Amazon SageMaker AI saves the results of a monitoring job.
* `s3_upload_mode` - (Optional) Whether to upload the results of the monitoring job continuously or after the job completes. Valid values: `Continuous`, `EndOfJob`.

### monitoring_resources

* `cluster_config` - (Required) Configuration for the cluster resources used to run the processing job. Fields are documented below.

### cluster_config

* `instance_count` - (Required) Number of ML compute instances to use in the model monitoring job.
* `instance_type` - (Required) ML compute instance type for the processing job.
* `volume_size_in_gb` - (Required) size of the ML storage volume, in gigabytes, to provision.
* `volume_kms_key_id` - (Optional) AWS KMS key that Amazon SageMaker AI uses to encrypt data on the storage volume attached to the ML compute instance(s) that run the model monitoring job.

### network_config

* `enable_inter_container_traffic_encryption` - (Optional) Whether to encrypt all communications between distributed processing jobs.
* `enable_network_isolation` - (Optional) Whether to allow inbound and outbound network calls to and from the containers used for the processing job.
* `vpc_config` - (Optional) VPC that SageMaker jobs, hosted models, and compute resources have access to. Fields are documented below.

### vpc_config

* `security_group_ids` - (Required) VPC security group IDs.
* `subnets` - (Required) Subnet IDs.

### stopping_condition

* `max_runtime_in_seconds` - (Required) Maximum runtime allowed in seconds.

#### schedule_config

* `schedule_expression` - (Required) A cron expression that describes details about the monitoring schedule. For example, and hourly schedule would be `cron(0 * ? * * *)`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The Amazon Resource Name (ARN) assigned by AWS to this monitoring schedule.
* `name` - The name of the monitoring schedule.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import monitoring schedules using the `name`. For example:

```terraform
import {
  to = aws_sagemaker_monitoring_schedule.test_monitoring_schedule
  id = "monitoring-schedule-foo"
}
```

Using `terraform import`, import monitoring schedules using the `name`. For example:

```console
% terraform import aws_sagemaker_monitoring_schedule.test_monitoring_schedule monitoring-schedule-foo
```
