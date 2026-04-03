---
subcategory: "SageMaker AI"
layout: "aws"
page_title: "AWS: aws_sagemaker_hyper_parameter_tuning_job"
description: |-
  Manages an AWS SageMaker AI Hyper Parameter Tuning Job.
---

# Resource: aws_sagemaker_hyper_parameter_tuning_job

Manages an AWS SageMaker AI Hyper Parameter Tuning Job.

## Example Usage

### Basic Usage

```terraform
resource "aws_sagemaker_hyper_parameter_tuning_job" "example" {
  name = "example"

  config {
    strategy = "Bayesian"

    objective {
      metric_name = "test:msd"
      type        = "Minimize"
    }

    parameter_ranges {
      categorical_parameter_ranges {
        name   = "init_method"
        values = ["kmeans++", "random"]
      }

      integer_parameter_ranges {
        name         = "epochs"
        min_value    = "1"
        max_value    = "10"
        scaling_type = "Auto"
      }

      integer_parameter_ranges {
        name         = "extra_center_factor"
        min_value    = "4"
        max_value    = "10"
        scaling_type = "Auto"
      }

      integer_parameter_ranges {
        name         = "mini_batch_size"
        min_value    = "3000"
        max_value    = "15000"
        scaling_type = "Auto"
      }
    }

    resource_limits {
      max_number_of_training_jobs = 2
      max_parallel_training_jobs  = 1
    }
  }

  training_job_definition {
    role_arn = "arn:aws:iam::123456789012:role/example-sagemaker-execution-role"

    algorithm_specification {
      training_image      = "174872318107.dkr.ecr.us-west-2.amazonaws.com/kmeans:1"
      training_input_mode = "File"
    }

    static_hyper_parameters = {
      feature_dim = "3"
      k           = "2"
    }

    input_data_config {
      channel_name = "train"
      content_type = "text/csv"
      input_mode   = "File"

      data_source {
        s3_data_source {
          s3_data_type = "S3Prefix"
          s3_uri       = "s3://example-bucket/input/"
        }
      }
    }

    input_data_config {
      channel_name = "test"
      content_type = "text/csv"
      input_mode   = "File"

      data_source {
        s3_data_source {
          s3_data_type = "S3Prefix"
          s3_uri       = "s3://example-bucket/input/"
        }
      }
    }

    output_data_config {
      s3_output_path = "s3://example-bucket/output/"
    }

    resource_config {
      instance_count    = 1
      instance_type     = "ml.m5.large"
      volume_size_in_gb = 30
    }

    stopping_condition {
      max_runtime_in_seconds = 3600
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the tuning job.
* `config` - (Required) Tuning job settings. See [`config`](#config).

The following arguments are optional:

* `autotune` - (Optional) Autotune settings. See [`autotune`](#autotune).
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Map of tags to assign to this resource.
* `training_job_definition` - (Optional) Single training job definition for tuning. See [`training_job_definition`](#training_job_definition-and-training_job_definitions).
* `training_job_definitions` - (Optional) Multiple training job definitions for tuning. See [`training_job_definition`](#training_job_definition-and-training_job_definitions).
* `warm_start_config` - (Optional) Warm start settings. See [`warm_start_config`](#warm_start_config).

### autotune

* `mode` - (Required) Autotune mode.

### config

* `random_seed` - (Optional) Random seed for tuning.
* `strategy` - (Required) Search strategy for tuning.
* `training_job_early_stopping_type` - (Optional) Early stopping behavior for training jobs.
* `objective` - (Optional) Objective metric used by tuning. See [`objective`](#objective).
* `parameter_ranges` - (Optional) Hyperparameter search ranges. See [`parameter_ranges`](#parameter_ranges).
* `resource_limits` - (Required) Training job limits for tuning. See [`resource_limits`](#resource_limits).
* `strategy_config` - (Optional) Extra strategy options. See [`strategy_config`](#strategy_config).
* `tuning_job_completion_criteria` - (Optional) Conditions to complete tuning. See [`tuning_job_completion_criteria`](#tuning_job_completion_criteria).

#### objective

* `metric_name` - (Required) Metric name that tuning tries to optimize.
* `type` - (Required) Optimization direction. Valid values include `Minimize` and `Maximize`.

#### parameter_ranges

* `auto_parameters` - (Optional) Parameter list for automatic range selection.
* `categorical_parameter_ranges` - (Optional) Categorical parameter ranges.
* `continuous_parameter_ranges` - (Optional) Continuous parameter ranges.
* `integer_parameter_ranges` - (Optional) Integer parameter ranges.

##### auto_parameters

* `name` - (Required) Parameter name.
* `value_hint` - (Required) Value hint for the parameter.

##### categorical_parameter_ranges

* `name` - (Required) Parameter name.
* `values` - (Required) Set of allowed values.

##### continuous_parameter_ranges

* `max_value` - (Required) Maximum value.
* `min_value` - (Required) Minimum value.
* `name` - (Required) Parameter name.
* `scaling_type` - (Optional) Scaling rule for the range.

##### integer_parameter_ranges

* `max_value` - (Required) Maximum value.
* `min_value` - (Required) Minimum value.
* `name` - (Required) Parameter name.
* `scaling_type` - (Optional) Scaling rule for the range.

#### resource_limits

* `max_number_of_training_jobs` - (Optional) Maximum total training jobs.
* `max_parallel_training_jobs` - (Required) Maximum parallel training jobs.
* `max_runtime_in_seconds` - (Optional) Maximum total runtime in seconds.

#### strategy_config

* `hyperband_strategy_config` - (Optional) Hyperband strategy settings. See [`hyperband_strategy_config`](#hyperband_strategy_config).

##### hyperband_strategy_config

* `max_resource` - (Optional) Upper bound for resource allocation.
* `min_resource` - (Optional) Lower bound for resource allocation.

#### tuning_job_completion_criteria

* `target_objective_metric_value` - (Optional) Target metric value that can stop tuning.
* `best_objective_not_improving` - (Optional) Stop condition for non-improving jobs. See [`best_objective_not_improving`](#best_objective_not_improving).
* `convergence_detected` - (Optional) Stop condition based on convergence. See [`convergence_detected`](#convergence_detected).

##### best_objective_not_improving

* `max_number_of_training_jobs_not_improving` - (Optional) Maximum training jobs without improvement before completion.

##### convergence_detected

* `complete_on_convergence` - (Optional) Whether to complete tuning when convergence is detected.

### training_job_definition and training_job_definitions

`training_job_definition` supports up to 1 block.

`training_job_definitions` supports 1 to 10 blocks.

Each block supports:

* `definition_name` - (Optional) Name for this definition.
* `enable_inter_container_traffic_encryption` - (Optional) Whether to encrypt traffic between containers.
* `enable_managed_spot_training` - (Optional) Whether to use managed spot training.
* `enable_network_isolation` - (Optional) Whether to isolate network access for containers.
* `environment` - (Optional) Map of environment variables.
* `role_arn` - (Required) IAM role ARN used by SageMaker AI.
* `static_hyper_parameters` - (Optional) Map of fixed hyperparameters.
* `algorithm_specification` - (Required) Algorithm settings. See [`algorithm_specification`](#algorithm_specification).
* `checkpoint_config` - (Optional) Checkpoint output location. See [`checkpoint_config`](#checkpoint_config).
* `hyper_parameter_tuning_resource_config` - (Optional) Tuning resource settings. See [`hyper_parameter_tuning_resource_config`](#hyper_parameter_tuning_resource_config).
* `hyper_parameter_ranges` - (Optional) Hyperparameter ranges for this definition. See [`parameter_ranges`](#parameter_ranges).
* `input_data_config` - (Optional) Input data channels. See [`input_data_config`](#input_data_config).
* `output_data_config` - (Required) Output data settings. See [`output_data_config`](#output_data_config).
* `resource_config` - (Optional) Training resources. See [`resource_config`](#resource_config).
* `retry_strategy` - (Optional) Retry settings. See [`retry_strategy`](#retry_strategy).
* `stopping_condition` - (Required) Stopping settings. See [`stopping_condition`](#stopping_condition).
* `tuning_objective` - (Optional) Objective for this training definition. See [`tuning_objective`](#tuning_objective).
* `vpc_config` - (Optional) VPC settings. See [`vpc_config`](#vpc_config).

#### algorithm_specification

* `algorithm_name` - (Optional) Built-in algorithm name or ARN.
* `training_image` - (Optional) Container image used for training.
* `training_input_mode` - (Required) Training input mode.
* `metric_definitions` - (Optional) Metric extraction rules.

Provide exactly one of `algorithm_name` or `training_image`.

##### metric_definitions

* `name` - (Required) Metric name.
* `regex` - (Required) Pattern used to extract metric values.

#### checkpoint_config

* `local_path` - (Optional) Local path for checkpoints.
* `s3_uri` - (Required) S3 or HTTPS destination for checkpoints.

#### hyper_parameter_tuning_resource_config

* `allocation_strategy` - (Optional) Allocation strategy for tuning resources.
* `instance_count` - (Optional) Number of training instances.
* `instance_type` - (Optional) Training instance type.
* `volume_kms_key_id` - (Optional) KMS key ID for volume encryption.
* `volume_size_in_gb` - (Optional) Volume size in GB.
* `instance_configs` - (Optional) Per-instance-type resource settings. See [`instance_configs`](#instance_configs).

Do not set `instance_count`, `instance_type`, or `volume_size_in_gb` when `instance_configs` is set.

##### instance_configs

* `instance_count` - (Optional) Number of instances.
* `instance_type` - (Optional) Instance type.
* `volume_size_in_gb` - (Optional) Volume size in GB.

#### input_data_config

* `channel_name` - (Required) Input channel name.
* `compression_type` - (Optional) Compression type.
* `content_type` - (Optional) Content type string.
* `input_mode` - (Optional) Input mode.
* `record_wrapper_type` - (Optional) Record wrapper format.
* `data_source` - (Required) Data source settings. See [`data_source`](#data_source).
* `shuffle_config` - (Optional) Shuffling settings. See [`shuffle_config`](#shuffle_config).

#### data_source

* `file_system_data_source` - (Optional) File system source settings. See [`file_system_data_source`](#file_system_data_source).
* `s3_data_source` - (Optional) S3 source settings. See [`s3_data_source`](#s3_data_source).

#### file_system_data_source

* `directory_path` - (Required) Directory path in the file system.
* `file_system_access_mode` - (Required) Access mode for the file system.
* `file_system_id` - (Required) File system ID.
* `file_system_type` - (Required) File system type.

#### s3_data_source

* `attribute_names` - (Optional) Attribute names for Pipe mode.
* `instance_group_names` - (Optional) Instance group names used with this channel.
* `s3_data_distribution_type` - (Optional) Distribution mode for S3 data.
* `s3_data_type` - (Required) S3 data type.
* `s3_uri` - (Required) S3 or HTTPS source URI.
* `hub_access_config` - (Optional) Hub access settings. See [`hub_access_config`](#hub_access_config).
* `model_access_config` - (Optional) Model access settings. See [`model_access_config`](#model_access_config).

##### hub_access_config

* `hub_content_arn` - (Required) Hub content ARN.

##### model_access_config

* `accept_eula` - (Required) Whether to accept model EULA. Value must be `true`.

#### shuffle_config

* `seed` - (Required) Shuffle seed.

#### output_data_config

* `compression_type` - (Optional) Compression type for output.
* `kms_key_id` - (Optional) KMS key ID for output encryption.
* `s3_output_path` - (Required) S3 or HTTPS output path.

#### resource_config

* `instance_count` - (Optional) Number of instances.
* `instance_type` - (Optional) Instance type.
* `keep_alive_period_in_seconds` - (Optional) Warm pool keep-alive period in seconds.
* `training_plan_arn` - (Optional) Training plan ARN.
* `volume_kms_key_id` - (Optional) KMS key ID for volume encryption.
* `volume_size_in_gb` - (Optional) Volume size in GB.
* `instance_groups` - (Optional) Instance group settings. See [`instance_groups`](#instance_groups).
* `instance_placement_config` - (Optional) Placement settings. See [`instance_placement_config`](#instance_placement_config).

#### instance_groups

* `instance_count` - (Required) Number of instances in the group.
* `instance_group_name` - (Required) Name of the group.
* `instance_type` - (Required) Instance type.

#### instance_placement_config

* `enable_multiple_jobs` - (Optional) Whether to run multiple jobs on shared infrastructure.
* `placement_specifications` - (Optional) Placement details. See [`placement_specifications`](#placement_specifications).

#### placement_specifications

* `instance_count` - (Required) Number of instances in this placement item.
* `ultra_server_id` - (Optional) UltraServer ID.

#### retry_strategy

* `maximum_retry_attempts` - (Optional) Maximum retry attempts.

#### stopping_condition

* `max_pending_time_in_seconds` - (Optional) Maximum pending time in seconds.
* `max_runtime_in_seconds` - (Optional) Maximum runtime in seconds.
* `max_wait_time_in_seconds` - (Optional) Maximum wait time in seconds.

#### tuning_objective

* `metric_name` - (Required) Metric name for objective.
* `type` - (Required) Optimization direction. Valid values include `Minimize` and `Maximize`.

#### vpc_config

* `security_group_ids` - (Required) Security group IDs.
* `subnets` - (Required) Subnet IDs.

### warm_start_config

* `warm_start_type` - (Optional) Warm start mode.
* `parent_hyper_parameter_tuning_jobs` - (Optional) Parent tuning jobs for warm start.

#### parent_hyper_parameter_tuning_jobs

* `name` - (Required) Parent tuning job name.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Hyper Parameter Tuning Job.
* `failure_reason` - Reason returned by SageMaker AI when a job fails.
* `status` - Current tuning job status.
* `tags_all` - Map of tags assigned to the resource, including provider default tags.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `5m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_sagemaker_hyper_parameter_tuning_job.example
  identity = {
    name = "example-hyper-parameter-tuning-job"
  }
}

resource "aws_sagemaker_hyper_parameter_tuning_job" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `name` (String) Name of the Hyper Parameter Tuning Job.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import SageMaker AI Hyper Parameter Tuning Jobs using `name`. For example:

```terraform
import {
  to = aws_sagemaker_hyper_parameter_tuning_job.example
  id = "example-hyper-parameter-tuning-job"
}
```

Using `terraform import`, import SageMaker AI Hyper Parameter Tuning Jobs using `name`. For example:

```console
% terraform import aws_sagemaker_hyper_parameter_tuning_job.example example-hyper-parameter-tuning-job
```
