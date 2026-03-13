---
subcategory: "SageMaker AI"
layout: "aws"
page_title: "AWS: aws_sagemaker_training_job"
description: |-
  Manages an AWS SageMaker AI Training Job.
---

# Resource: aws_sagemaker_training_job

Manages an AWS SageMaker AI Training Job.

## Example Usage

### Basic Usage

```terraform
resource "aws_sagemaker_training_job" "example" {
  training_job_name = "example"
  role_arn          = aws_iam_role.example.arn

  algorithm_specification {
    training_input_mode = "File"
    training_image      = data.aws_sagemaker_prebuilt_ecr_image.example.registry_path
  }

  output_data_config {
    s3_output_path = "s3://${aws_s3_bucket.example.bucket}/output/"
  }

  resource_config {
    instance_type     = "ml.m5.large"
    instance_count    = 1
    volume_size_in_gb = 30
  }

  stopping_condition {
    max_runtime_in_seconds = 3600
  }
}
```

### With VPC Configuration

```terraform
resource "aws_sagemaker_training_job" "example" {
  training_job_name = "example"
  role_arn          = aws_iam_role.example.arn

  algorithm_specification {
    training_input_mode = "File"
    training_image      = data.aws_sagemaker_prebuilt_ecr_image.example.registry_path
  }

  output_data_config {
    s3_output_path = "s3://${aws_s3_bucket.example.bucket}/output/"
  }

  resource_config {
    instance_type     = "ml.m5.large"
    instance_count    = 1
    volume_size_in_gb = 30
  }

  stopping_condition {
    max_runtime_in_seconds = 3600
  }

  vpc_config {
    security_group_ids = [aws_security_group.example.id]
    subnets            = [aws_subnet.example.id]
  }
}
```

### With Input Data and Hyperparameters

```terraform
resource "aws_sagemaker_training_job" "example" {
  training_job_name = "example"
  role_arn          = aws_iam_role.example.arn

  algorithm_specification {
    training_input_mode                  = "File"
    training_image                       = data.aws_sagemaker_prebuilt_ecr_image.example.registry_path
    enable_sagemaker_metrics_time_series = true
  }

  hyper_parameters = {
    "mini_batch_size" = "200"
    "epochs"          = "10"
  }

  input_data_config {
    channel_name = "train"

    data_source {
      s3_data_source {
        s3_data_type = "S3Prefix"
        s3_uri       = "s3://${aws_s3_bucket.example.bucket}/train/"
      }
    }
  }

  output_data_config {
    s3_output_path = "s3://${aws_s3_bucket.example.bucket}/output/"
  }

  resource_config {
    instance_type     = "ml.m5.large"
    instance_count    = 1
    volume_size_in_gb = 30
  }

  stopping_condition {
    max_runtime_in_seconds = 3600
  }
}
```

## Argument Reference

The following arguments are required:

* `role_arn` - (Required) ARN of the IAM role that SageMaker AI assumes to perform tasks on your behalf during training.
* `stopping_condition` - (Required) Time limit for how long the training job can run. See [`stopping_condition`](#stopping_condition) below.
* `training_job_name` - (Required) Name of the training job. Must be between 1 and 63 characters, start with a letter or number, and contain only letters, numbers, and hyphens.

The following arguments are optional:

* `algorithm_specification` - (Optional) Algorithm-related parameters of the training job. See [`algorithm_specification`](#algorithm_specification) below. Conflicts with `serverless_job_config`.
* `checkpoint_config` - (Optional) Location of checkpoints during training. See [`checkpoint_config`](#checkpoint_config) below. Conflicts with `serverless_job_config`.
* `debug_hook_config` - (Optional) Configuration for debugging rules. See [`debug_hook_config`](#debug_hook_config) below. Conflicts with `serverless_job_config`.
* `debug_rule_configurations` - (Optional) List of debug rule configurations. Maximum of 20. See [`debug_rule_configurations`](#debug_rule_configurations) below.
* `enable_inter_container_traffic_encryption` - (Optional) Whether to encrypt inter-container traffic. When enabled, communications between containers are encrypted.
* `enable_managed_spot_training` - (Optional) Whether to use managed spot training. Optimizes the cost of training by using Amazon EC2 Spot Instances. Conflicts with `serverless_job_config`.
* `enable_network_isolation` - (Optional) Whether to isolate the training container from the network. No inbound or outbound network calls can be made.
* `environment` - (Optional) Map of environment variables to set in the training container. Maximum of 100 entries.  Conflicts with `serverless_job_config`.
* `experiment_config` - (Optional) Associates a SageMaker AI Experiment or Trial to the training job. See [`experiment_config`](#experiment_config) below. Conflicts with `serverless_job_config`.
* `hyper_parameters` - (Optional) Map of hyperparameters for the training algorithm. Maximum of 100 entries.
* `infra_check_config` - (Optional) Infrastructure health check configuration. See [`infra_check_config`](#infra_check_config) below.
* `input_data_config` - (Optional) List of input data channel configurations for the training job. Maximum of 20. See [`input_data_config`](#input_data_config) below.
* `mlflow_config` - (Optional) MLflow integration configuration. See [`mlflow_config`](#mlflow_config) below.
* `model_package_config` - (Optional) Model package configuration. Requires `serverless_job_config`. See [`model_package_config`](#model_package_config) below.
* `output_data_config` - (Optional) Location of the output data from the training job. See [`output_data_config`](#output_data_config) below.
* `profiler_config` - (Optional) Configuration for the profiler. See [`profiler_config`](#profiler_config) below. Conflicts with `serverless_job_config`.
* `profiler_rule_configurations` - (Optional) List of profiler rule configurations. Maximum of 20. See [`profiler_rule_configurations`](#profiler_rule_configurations) below. Conflicts with `serverless_job_config`.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `remote_debug_config` - (Optional) Configuration for remote debugging. See [`remote_debug_config`](#remote_debug_config) below.
* `resource_config` - (Optional) Resources for the training job, including compute instances and storage volumes. See [`resource_config`](#resource_config) below.
* `retry_strategy` - (Optional) Number of times to retry the job if it fails. See [`retry_strategy`](#retry_strategy) below. Conflicts with `serverless_job_config`.
* `serverless_job_config` - (Optional) Configuration for serverless training jobs using foundation models. Conflicts with `algorithm_specification`, `enable_managed_spot_training`, `environment`, `retry_strategy`, `checkpoint_config`, `debug_hook_config`, `experiment_config`, `profiler_config`, `profiler_rule_configurations`, and `tensor_board_output_config`. See [`serverless_job_config`](#serverless_job_config) below.
* `session_chaining_config` - (Optional) Configuration for session tag chaining. See [`session_chaining_config`](#session_chaining_config) below.
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `tensor_board_output_config` - (Optional) Configuration for TensorBoard output. See [`tensor_board_output_config`](#tensor_board_output_config) below. Conflicts with `serverless_job_config`.
* `vpc_config` - (Optional) VPC configuration for the training job. See [`vpc_config`](#vpc_config) below.

### `algorithm_specification`

* `algorithm_name` - (Optional) Name or ARN of the algorithm resource to use for the training job.
* `container_arguments` - (Optional) List of arguments for the container entrypoint. Maximum of 100 entries.
* `container_entrypoint` - (Optional) List of entrypoint commands for the container. Maximum of 100 entries.
* `enable_sagemaker_metrics_time_series` - (Optional) Whether to enable SageMaker AI metrics time series collection.
* `metric_definitions` - (Optional) List of metric definitions for the training job. Maximum of 40. Use this to extract custom metrics from your own training container logs. SageMaker can still publish built-in metrics for built-in algorithms and supported prebuilt images when this block is omitted. See [`metric_definitions`](#metric_definitions) below.
* `training_image` - (Optional) Registry path of the Docker image that contains the training algorithm.
* `training_image_config` - (Optional) Training image configuration. See [`training_image_config`](#training_image_config) below.
* `training_input_mode` - (Optional) Input mode for the training data. Valid values: `File`, `Pipe`, `FastFile`.

### `metric_definitions`

Use `metric_definitions` when you need to parse custom metrics from your training container logs. For SageMaker built-in algorithms and supported prebuilt images, SageMaker can populate default metrics without this block.

* `name` - (Required) Name of the metric.
* `regex` - (Required) Regular expression that searches the output of the training job and captures the value of the metric.

### `training_image_config`

* `training_repository_access_mode` - (Optional) Access mode for the training image repository.
* `training_repository_auth_config` - (Optional) Authentication configuration for the training image repository. See [`training_repository_auth_config`](#training_repository_auth_config) below.

### `training_repository_auth_config`

* `training_repository_credentials_provider_arn` - (Optional) ARN of the Lambda function that provides credentials to authenticate to the private Docker registry.

### `checkpoint_config`

* `local_path` - (Optional) Local path where checkpoints are written.
* `s3_uri` - (Required) S3 URI where checkpoints are stored.

### `debug_hook_config`

* `collection_configurations` - (Optional) List of tensor collections to configure for the debug hook. Maximum of 20. See [`collection_configurations`](#collection_configurations) below.
* `hook_parameters` - (Optional) Map of parameters for the debug hook. Maximum of 20 entries.
* `local_path` - (Optional) Local path where debug output is written.
* `s3_output_path` - (Required) S3 URI where debug output is stored.

### `collection_configurations`

* `collection_name` - (Optional) Name of the tensor collection.
* `collection_parameters` - (Optional) Map of parameters for the tensor collection.

### `debug_rule_configurations`

* `instance_type` - (Optional) Instance type to deploy for the debug rule evaluation. Valid values are SageMaker AI processing instance types.
* `local_path` - (Optional) Local path where debug rule output is written.
* `rule_configuration_name` - (Required) Name of the rule configuration. Must be between 1 and 256 characters.
* `rule_evaluator_image` - (Required) Docker image URI for the rule evaluator.
* `rule_parameters` - (Optional) Map of parameters for the rule configuration. Maximum of 100 entries.
* `s3_output_path` - (Optional) S3 URI where rule output is stored.
* `volume_size_in_gb` - (Optional) Size of the storage volume for the rule evaluator, in GB.

### `experiment_config`

* `experiment_name` - (Optional) Name of the SageMaker AI Experiment to associate with.
* `run_name` - (Optional) Name of the Experiment Run to associate with.
* `trial_component_display_name` - (Optional) Display name for the trial component.
* `trial_name` - (Optional) Name of the SageMaker AI Trial to associate with.

### `infra_check_config`

* `enable_infra_check` - (Optional) Whether to enable infrastructure health checks before training.

### `input_data_config`

* `channel_name` - (Required) Name of the channel. Must be between 1 and 64 characters.
* `compression_type` - (Optional) Compression type for the input data. Valid values: `None`, `Gzip`.
* `content_type` - (Optional) MIME type of the input data.
* `data_source` - (Required) Location of the channel data. See [`data_source`](#data_source) below.
* `input_mode` - (Optional) Input mode for the channel data. Valid values: `File`, `Pipe`, `FastFile`.
* `record_wrapper_type` - (Optional) Record wrapper type. Valid values: `None`, `RecordIO`.
* `shuffle_config` - (Optional) Configuration for shuffling data in the channel. See [`shuffle_config`](#shuffle_config) below.

### `data_source`

* `file_system_data_source` - (Optional) File system data source. See [`file_system_data_source`](#file_system_data_source) below.
* `s3_data_source` - (Optional) S3 data source. See [`s3_data_source`](#s3_data_source) below.

### `file_system_data_source`

* `directory_path` - (Required) Full path to the directory on the file system.
* `file_system_access_mode` - (Required) Access mode for the file system. Valid values: `ro`, `rw`.
* `file_system_id` - (Required) File system ID.
* `file_system_type` - (Required) File system type. Valid values: `EFS`, `FSxLustre`.

### `s3_data_source`

* `attribute_names` - (Optional) List of attribute names to include in the training dataset. Maximum of 16.
* `hub_access_config` - (Optional) SageMaker AI Hub access configuration. See [`hub_access_config`](#hub_access_config) below.
* `instance_group_names` - (Optional) List of instance group names for the training data distribution. Maximum of 5.
* `model_access_config` - (Optional) Model access configuration. See [`model_access_config`](#model_access_config) below.
* `s3_data_distribution_type` - (Optional) Distribution type for S3 data. Valid values: `FullyReplicated`, `ShardedByS3Key`.
* `s3_data_type` - (Required) S3 data type. Valid values: `ManifestFile`, `S3Prefix`, `AugmentedManifestFile`.
* `s3_uri` - (Required) S3 URI of the data.

### `hub_access_config`

* `hub_content_arn` - (Required) ARN of the hub content.

### `model_access_config`

* `accept_eula` - (Required) Whether to accept the model EULA.

### `shuffle_config`

* `seed` - (Optional) Seed value used to shuffle the training data.

### `mlflow_config`

* `mlflow_experiment_name` - (Optional) Name of the MLflow experiment.
* `mlflow_resource_arn` - (Required) ARN of the MLflow tracking server.
* `mlflow_run_name` - (Optional) Name of the MLflow run.

### `model_package_config`

* `model_package_group_arn` - (Required) ARN of the model package group.
* `source_model_package_arn` - (Optional) ARN of the source model package.

### `output_data_config`

* `compression_type` - (Optional) Output compression type. Valid values: `GZIP`, `NONE`.
* `kms_key_id` - (Optional) KMS key ID used to encrypt the output data.
* `s3_output_path` - (Required) S3 URI where output data is stored.

### `profiler_config`

* `disable_profiler` - (Optional) Whether to disable the profiler.
* `profiling_interval_in_milliseconds` - (Optional) Time interval in milliseconds for capturing system metrics. Valid values: `100`, `200`, `500`, `1000`, `5000`, `60000`.
* `profiling_parameters` - (Optional) Map of profiling parameters. Maximum of 20 entries.
* `s3_output_path` - (Optional) S3 URI where profiler output is stored.

### `profiler_rule_configurations`

* `instance_type` - (Optional) Instance type to deploy for the profiler rule evaluation. Valid values are SageMaker AI processing instance types.
* `local_path` - (Optional) Local path where profiler rule output is written.
* `rule_configuration_name` - (Required) Name of the profiler rule configuration. Must be between 1 and 256 characters.
* `rule_evaluator_image` - (Required) Docker image URI for the profiler rule evaluator.
* `rule_parameters` - (Optional) Map of parameters for the profiler rule. Maximum of 100 entries.
* `s3_output_path` - (Optional) S3 URI where profiler rule output is stored.
* `volume_size_in_gb` - (Optional) Size of the storage volume for the profiler rule evaluator, in GB.

### `remote_debug_config`

* `enable_remote_debug` - (Optional) Whether to enable remote debugging for the training job.

### `resource_config`

* `instance_count` - (Optional) Number of ML compute instances to use. Conflicts with `instance_groups`.
* `instance_groups` - (Optional) List of instance groups for heterogeneous cluster training. Maximum of 5. Conflicts with `instance_count`, `instance_type`, and `keep_alive_period_in_seconds`. See [`instance_groups`](#instance_groups) below.
* `instance_placement_config` - (Optional) Instance placement configuration. See [`instance_placement_config`](#instance_placement_config) below.
* `instance_type` - (Optional) ML compute instance type. Conflicts with `instance_groups`.
* `keep_alive_period_in_seconds` - (Optional) Time in seconds to keep instances alive after training completes, for warm pool reuse. Valid values: 0–3600. Conflicts with `instance_groups`.
* `training_plan_arn` - (Optional) ARN of the training plan to use.
* `volume_kms_key_id` - (Optional) KMS key ID used to encrypt data on the storage volume.
* `volume_size_in_gb` - (Optional) Size of the storage volume attached to each instance, in GB.

### `instance_groups`

* `instance_count` - (Optional) Number of instances in the group.
* `instance_group_name` - (Optional) Name of the instance group.
* `instance_type` - (Optional) ML compute instance type for the group.

### `instance_placement_config`

* `enable_multiple_jobs` - (Optional) Whether to enable multiple jobs on the same instance.
* `placement_specifications` - (Optional) Placement specifications for instance placement. See [`placement_specifications`](#placement_specifications) below.

### `placement_specifications`

* `instance_count` - (Optional) Number of instances in the placement.
* `ultra_server_id` - (Optional) Ultra server ID for the placement.

### `retry_strategy`

* `maximum_retry_attempts` - (Required) Maximum number of retry attempts. Valid values: 1–30.

### `serverless_job_config`

* `accept_eula` - (Optional) Whether to accept the model EULA.
* `base_model_arn` - (Required) ARN of the base foundation model from the SageMaker AI Public Hub.
* `customization_technique` - (Optional) Customization technique to apply. Valid values: `FINE_TUNING`, `DOMAIN_ADAPTION`.
* `evaluation_type` - (Optional) Evaluation type. Valid values: `AUTOMATIC`, `HUMAN`, `NONE`.
* `evaluator_arn` - (Optional) ARN of the evaluator.
* `job_type` - (Required) Serverless job type. Valid values: `FINE_TUNING`, `EVALUATION`, `DISTILLATION`.
* `peft` - (Optional) Parameter-Efficient Fine-Tuning (PEFT) method. Valid values: `LORA`.

### `session_chaining_config`

* `enable_session_tag_chaining` - (Optional) Whether to enable session tag chaining for the training job.

### `stopping_condition`

* `max_pending_time_in_seconds` - (Optional) Maximum time in seconds a training job can be pending before it is stopped. Valid values: 7200–2419200.
* `max_runtime_in_seconds` - (Optional) Maximum time in seconds the training job can run before it is stopped.
* `max_wait_time_in_seconds` - (Optional) Maximum time in seconds to wait for a managed spot training job to complete.

### `tensor_board_output_config`

* `local_path` - (Optional) Local path where TensorBoard output is written.
* `s3_output_path` - (Required) S3 URI where TensorBoard output is stored.

### `vpc_config`

* `security_group_ids` - (Required) List of VPC security group IDs. Maximum of 5.
* `subnets` - (Required) List of subnet IDs. Maximum of 16.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Training Job.
* `id` - Name of the Training Job.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `60m`)
* `update` - (Default `180m`)
* `delete` - (Default `90m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_sagemaker_training_job.example
  identity = {
    training_job_name = "my-training-job"
  }
}

resource "aws_sagemaker_training_job" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `training_job_name` - (String) Name of the Training Job.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import SageMaker AI Training Job using the `training_job_name`. For example:

```terraform
import {
  to = aws_sagemaker_training_job.example
  id = "my-training-job"
}
```

Using `terraform import`, import SageMaker AI Training Job using the `training_job_name`. For example:

```console
% terraform import aws_sagemaker_training_job.example my-training-job
```
