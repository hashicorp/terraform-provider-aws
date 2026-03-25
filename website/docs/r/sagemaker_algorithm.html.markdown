---
subcategory: "SageMaker AI"
layout: "aws"
page_title: "AWS: aws_sagemaker_algorithm"
description: |-
  Manages an AWS SageMaker AI Algorithm.
---

# Resource: aws_sagemaker_algorithm

Manages an AWS SageMaker AI Algorithm.

## Example Usage

### Basic Usage

```terraform
resource "aws_sagemaker_algorithm" "example" {
  algorithm_name = "example-algorithm"

  training_specification {
    supported_training_instance_types = ["ml.m5.large"]
    training_image                    = "123456789012.dkr.ecr.us-west-2.amazonaws.com/example-training:latest"

    training_channels {
      name                    = "train"
      supported_content_types = ["text/csv"]
      supported_input_modes   = ["File"]
    }
  }

  tags = {
    Environment = "test"
  }
}
```

### Training Specification

```terraform
data "aws_sagemaker_prebuilt_ecr_image" "example" {
  repository_name = "linear-learner"
  image_tag       = "1"
}

resource "aws_sagemaker_algorithm" "example" {
  algorithm_name = "example-training-algorithm"

  training_specification {
    supported_training_instance_types = ["ml.m5.large", "ml.c5.xlarge"]
    supports_distributed_training     = true
    training_image                    = data.aws_sagemaker_prebuilt_ecr_image.example.registry_path

    metric_definitions {
      name  = "train:loss"
      regex = "loss=(.*?);"
    }

    supported_hyper_parameters {
      default_value = "0.5"
      description   = "Continuous learning rate"
      is_required   = true
      is_tunable    = true
      name          = "eta"
      type          = "Continuous"

      range {
        continuous_parameter_range_specification {
          min_value = "0.1"
          max_value = "0.9"
        }
      }
    }

    supported_hyper_parameters {
      default_value = "5"
      description   = "Maximum tree depth"
      is_required   = false
      is_tunable    = true
      name          = "max_depth"
      type          = "Integer"

      range {
        integer_parameter_range_specification {
          min_value = "1"
          max_value = "10"
        }
      }
    }

    supported_hyper_parameters {
      default_value = "reg:squarederror"
      description   = "Objective function"
      is_required   = false
      is_tunable    = false
      name          = "objective"
      type          = "Categorical"

      range {
        categorical_parameter_range_specification {
          values = ["reg:squarederror", "binary:logistic"]
        }
      }
    }

    supported_tuning_job_objective_metrics {
      metric_name = "train:loss"
      type        = "Minimize"
    }

    training_channels {
      description                 = "Training data channel"
      is_required                 = true
      name                        = "train"
      supported_compression_types = ["None", "Gzip"]
      supported_content_types     = ["text/csv"]
      supported_input_modes       = ["File"]
    }

    training_channels {
      name                    = "validation"
      supported_content_types = ["application/json"]
      supported_input_modes   = ["Pipe"]
    }
  }
}
```

### Inference Specification

```terraform
data "aws_sagemaker_prebuilt_ecr_image" "example" {
  repository_name = "linear-learner"
  image_tag       = "1"
}

resource "aws_sagemaker_algorithm" "example" {
  algorithm_name = "example-inference-algorithm"

  training_specification {
    supported_training_instance_types = ["ml.m5.large"]
    training_image                    = data.aws_sagemaker_prebuilt_ecr_image.example.registry_path

    training_channels {
      name                    = "train"
      supported_content_types = ["text/csv"]
      supported_input_modes   = ["File"]
    }
  }

  inference_specification {
    supported_content_types                     = ["text/csv"]
    supported_realtime_inference_instance_types = ["ml.m5.large"]
    supported_response_mime_types               = ["text/csv"]
    supported_transform_instance_types          = ["ml.m5.large"]

    containers {
      container_hostname = "test-host"
      environment = {
        TEST = "value"
      }
      framework          = "XGBOOST"
      framework_version  = "1.5-1"
      image              = data.aws_sagemaker_prebuilt_ecr_image.example.registry_path
      is_checkpoint      = true
      nearest_model_name = "nearest-model"

      base_model {
        hub_content_name    = "basemodel"
        hub_content_version = "1.0.0"
        recipe_name         = "recipe"
      }

      model_input {
        data_input_config = "{}"
      }
    }
  }
}
```

### Validation Specification

```terraform
data "aws_partition" "current" {}

data "aws_sagemaker_prebuilt_ecr_image" "example" {
  repository_name = "linear-learner"
  image_tag       = "1"
}

data "aws_iam_policy_document" "assume_role" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["sagemaker.${data.aws_partition.current.dns_suffix}"]
    }
  }
}

resource "aws_iam_role" "example" {
  name               = "example-sagemaker-algorithm-role"
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

resource "aws_iam_role_policy_attachment" "example" {
  role       = aws_iam_role.example.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonSageMakerFullAccess"
}

resource "aws_s3_bucket" "example" {
  bucket        = "example-sagemaker-algorithm-validation-bucket"
  force_destroy = true
}

data "aws_iam_policy_document" "s3_access" {
  statement {
    effect = "Allow"

    actions = [
      "s3:GetBucketLocation",
      "s3:ListBucket",
      "s3:GetObject",
      "s3:PutObject",
    ]

    resources = [
      aws_s3_bucket.example.arn,
      "${aws_s3_bucket.example.arn}/*",
    ]
  }
}

resource "aws_iam_role_policy" "example" {
  role   = aws_iam_role.example.name
  policy = data.aws_iam_policy_document.s3_access.json
}

resource "aws_s3_object" "training" {
  bucket  = aws_s3_bucket.example.bucket
  key     = "algorithm/training/data.csv"
  content = <<-EOT
1,1.0,0.0
0,0.0,1.0
1,1.0,1.0
0,0.0,0.0
EOT
}

resource "aws_s3_object" "transform" {
  bucket  = aws_s3_bucket.example.bucket
  key     = "algorithm/transform/input.csv"
  content = <<-EOT
1.0,0.0
0.0,1.0
EOT
}

resource "aws_sagemaker_algorithm" "example" {
  algorithm_name = "example-validation-algorithm"

  depends_on = [
    aws_iam_role_policy_attachment.example,
    aws_iam_role_policy.example,
    aws_s3_object.training,
    aws_s3_object.transform,
  ]

  training_specification {
    training_image                    = data.aws_sagemaker_prebuilt_ecr_image.example.registry_path
    supported_training_instance_types = ["ml.m5.large"]

    supported_hyper_parameters {
      default_value = "2"
      description   = "Feature dimension"
      is_required   = true
      is_tunable    = false
      name          = "feature_dim"
      type          = "Integer"

      range {
        integer_parameter_range_specification {
          min_value = "2"
          max_value = "2"
        }
      }
    }

    supported_hyper_parameters {
      default_value = "4"
      description   = "Mini batch size"
      is_required   = true
      is_tunable    = false
      name          = "mini_batch_size"
      type          = "Integer"

      range {
        integer_parameter_range_specification {
          min_value = "4"
          max_value = "4"
        }
      }
    }

    supported_hyper_parameters {
      default_value = "binary_classifier"
      description   = "Predictor type"
      is_required   = true
      is_tunable    = false
      name          = "predictor_type"
      type          = "Categorical"

      range {
        categorical_parameter_range_specification {
          values = ["binary_classifier"]
        }
      }
    }

    training_channels {
      name                    = "train"
      supported_content_types = ["text/csv"]
      supported_input_modes   = ["File"]
    }
  }

  inference_specification {
    supported_content_types            = ["text/csv"]
    supported_response_mime_types      = ["text/csv"]
    supported_transform_instance_types = ["ml.m5.large"]

    containers {
      image = data.aws_sagemaker_prebuilt_ecr_image.example.registry_path
    }
  }

  validation_specification {
    validation_role = aws_iam_role.example.arn

    validation_profiles {
      profile_name = "validation-profile"

      training_job_definition {
        hyper_parameters = {
          feature_dim     = "2"
          mini_batch_size = "4"
          predictor_type  = "binary_classifier"
        }
        training_input_mode = "File"

        input_data_config {
          channel_name        = "train"
          compression_type    = "None"
          content_type        = "text/csv"
          input_mode          = "File"
          record_wrapper_type = "None"

          shuffle_config {
            seed = 1
          }

          data_source {
            s3_data_source {
              attribute_names           = ["label"]
              s3_data_distribution_type = "ShardedByS3Key"
              s3_data_type              = "S3Prefix"
              s3_uri                    = "s3://${aws_s3_bucket.example.bucket}/algorithm/training/"
            }
          }
        }

        output_data_config {
          compression_type = "GZIP"
          s3_output_path   = "s3://${aws_s3_bucket.example.bucket}/algorithm/output"
        }

        resource_config {
          instance_count               = 1
          instance_type                = "ml.m5.large"
          keep_alive_period_in_seconds = 60
          volume_size_in_gb            = 30
        }

        stopping_condition {
          max_pending_time_in_seconds = 7200
          max_runtime_in_seconds      = 1800
          max_wait_time_in_seconds    = 3600
        }
      }

      transform_job_definition {
        batch_strategy = "MultiRecord"
        environment = {
          Te = "enabled"
        }
        max_concurrent_transforms = 1
        max_payload_in_mb         = 6

        transform_input {
          compression_type = "None"
          content_type     = "text/csv"
          split_type       = "Line"

          data_source {
            s3_data_source {
              s3_data_type = "S3Prefix"
              s3_uri       = "s3://${aws_s3_bucket.example.bucket}/algorithm/transform/"
            }
          }
        }

        transform_output {
          accept         = "text/csv"
          assemble_with  = "Line"
          s3_output_path = "s3://${aws_s3_bucket.example.bucket}/algorithm/transform-output"
        }

        transform_resources {
          instance_count = 1
          instance_type  = "ml.m5.large"
        }
      }
    }
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `algorithm_description` - (Optional) Description of the algorithm.
* `algorithm_name` - (Required) Name of the algorithm.
* `certify_for_marketplace` - (Optional) Whether to certify the algorithm for AWS Marketplace.
* `inference_specification` - (Optional) Configuration for inference jobs that use this algorithm. See [Inference Specification](#inference-specification).
* `region` - (Optional) Region where this resource is managed. Defaults to the Region set in the provider configuration.
* `tags` - (Optional) Map of tags to assign to the resource.
* `training_specification` - (Required) Configuration for training jobs that use this algorithm. See [Training Specification](#training-specification).
* `validation_specification` - (Optional) Configuration used to validate the algorithm. See [Validation Specification](#validation-specification).

### Inference Specification

* `containers` - (Required) List of container definitions for inference.
* `supported_content_types` - (Optional) Supported MIME types for inference requests.
* `supported_realtime_inference_instance_types` - (Optional) Instance types supported for real-time inference.
* `supported_response_mime_types` - (Optional) Supported MIME types for inference responses.
* `supported_transform_instance_types` - (Optional) Instance types supported for batch transform.

#### Containers

* `additional_s3_data_source` - (Optional) Additional model data to make available to the container. See [Additional S3 Data Source](#additional-s3-data-source).
* `base_model` - (Optional) Base model information for the container. See [Base Model](#base-model).
* `container_hostname` - (Optional) DNS host name for the container.
* `environment` - (Optional) Environment variables to pass to the container.
* `framework` - (Optional) Machine learning framework in the container image.
* `framework_version` - (Optional) Framework version in the container image.
* `image` - (Optional) Container image URI.
* `image_digest` - (Optional) Digest of the container image.
* `is_checkpoint` - (Optional) Whether the container is used as a checkpoint container.
* `model_data_etag` - (Optional) ETag for `model_data_url`.
* `model_data_source` - (Optional) Source of model data for the container. See [Model Data Source](#model-data-source).
* `model_data_url` - (Optional) S3 or HTTPS URL of the model artifacts.
* `model_input` - (Optional) Additional model input configuration. See [Model Input](#model-input).
* `nearest_model_name` - (Optional) Name of a pre-existing model nearest to the one being created.
* `product_id` - (Optional) AWS Marketplace product ID.

### Additional S3 Data Source

* `compression_type` - (Optional) Compression type for the data. Allowed values are: `None` and `Gzip`.
* `etag` - (Optional) ETag of the S3 object.
* `s3_data_type` - (Required) Type of additional S3 data.
* `s3_uri` - (Required) S3 or HTTPS URI for the additional data.

### Base Model

* `hub_content_name` - (Optional) Name of the SageMaker AI Hub content.
* `hub_content_version` - (Optional) Version of the SageMaker AI Hub content.
* `recipe_name` - (Optional) Recipe name associated with the base model.

### Model Data Source

* `s3_data_source` - (Required) S3-backed model data source. See [Model Data Source S3 Data Source](#model-data-source-s3-data-source).

#### Model Data Source S3 Data Source

* `compression_type` - (Required) Compression type for the model data. Allowed values are: `None` and `Gzip`.
* `etag` - (Optional) ETag of the model data object.
* `hub_access_config` - (Optional) SageMaker AI Hub access settings. See [Hub Access Config](#hub-access-config).
* `manifest_etag` - (Optional) ETag of the manifest file.
* `manifest_s3_uri` - (Optional) S3 or HTTPS URI of the manifest file.
* `model_access_config` - (Optional) Model access settings. See [Model Access Config](#model-access-config).
* `s3_data_type` - (Required) Type of model data stored in S3.
* `s3_uri` - (Required) S3 or HTTPS URI of the model data.

### Hub Access Config

* `hub_content_arn` - (Optional) ARN of the SageMaker AI Hub content.

### Model Access Config

* `accept_eula` - (Optional) Whether to accept the model end-user license agreement.

### Model Input

* `data_input_config` - (Optional) Input configuration for the model.

### Training Specification

* `additional_s3_data_source` - (Optional) Additional training data to make available to the algorithm. See [Additional S3 Data Source](#additional-s3-data-source).
* `metric_definitions` - (Optional) List of metric definitions used to parse training logs. See [Metric Definitions](#metric-definitions).
* `supported_hyper_parameters` - (Optional) Hyperparameter definitions supported by the algorithm. See [Supported Hyper Parameters](#supported-hyper-parameters).
* `supported_training_instance_types` - (Required) Instance types supported for training.
* `supported_tuning_job_objective_metrics` - (Optional) Objective metrics supported for hyperparameter tuning jobs. See [Supported Tuning Job Objective Metrics](#supported-tuning-job-objective-metrics).
* `supports_distributed_training` - (Optional) Whether the algorithm supports distributed training.
* `training_channels` - (Required) List of channel definitions supported for training. See [Training Channels](#training-channels).
* `training_image` - (Required) Training image URI.
* `training_image_digest` - (Optional) Digest of the training image.

### Metric Definitions

* `name` - (Required) Metric name.
* `regex` - (Required) Regular expression used to extract the metric from logs.

### Supported Hyper Parameters

* `default_value` - (Optional) Default value for the hyperparameter.
* `description` - (Optional) Description of the hyperparameter.
* `is_required` - (Optional) Whether the hyperparameter is required.
* `is_tunable` - (Optional) Whether the hyperparameter can be tuned.
* `name` - (Required) Hyperparameter name.
* `range` - (Optional) Allowed value range for the hyperparameter. See [Parameter Range](#parameter-range).
* `type` - (Required) Hyperparameter type. Allowed values are: `Integer`, `Continuous`, `Categorical`, and `FreeText`.

### Parameter Range

* `categorical_parameter_range_specification` - (Optional) Categorical range definition. See [Categorical Parameter Range Specification](#categorical-parameter-range-specification).
* `continuous_parameter_range_specification` - (Optional) Continuous range definition. See [Continuous Parameter Range Specification](#continuous-parameter-range-specification).
* `integer_parameter_range_specification` - (Optional) Integer range definition. See [Integer Parameter Range Specification](#integer-parameter-range-specification).

#### Categorical Parameter Range Specification

* `values` - (Required) Allowed categorical values.

#### Continuous Parameter Range Specification

* `max_value` - (Required) Maximum allowed value.
* `min_value` - (Required) Minimum allowed value.

#### Integer Parameter Range Specification

* `max_value` - (Required) Maximum allowed value.
* `min_value` - (Required) Minimum allowed value.

### Supported Tuning Job Objective Metrics

* `metric_name` - (Required) Metric name.
* `type` - (Required) Objective type. Allowed values are: `Minimize` and `Maximize`.

### Training Channels

* `description` - (Optional) Description of the channel.
* `is_required` - (Optional) Whether the channel is required.
* `name` - (Required) Channel name.
* `supported_compression_types` - (Optional) Supported compression types. Allowed values are: `None` and `Gzip`.
* `supported_content_types` - (Required) Supported input content types.
* `supported_input_modes` - (Required) Supported training input modes.

### Validation Specification

* `validation_profiles` - (Required) Validation profiles for the algorithm. See [Validation Profiles](#validation-profiles).
* `validation_role` - (Required) IAM role ARN used for validation.

### Validation Profiles

* `profile_name` - (Required) Profile name.
* `training_job_definition` - (Required) Training job definition used during validation. See [Training Job Definition](#training-job-definition).
* `transform_job_definition` - (Optional) Transform job definition used during validation. See [Transform Job Definition](#transform-job-definition).

### Training Job Definition

* `hyper_parameters` - (Optional) Hyperparameters to pass to the training job.
* `input_data_config` - (Required) Input channel configuration for the validation training job. See [Input Data Config](#input-data-config).
* `output_data_config` - (Required) Output configuration for the validation training job. See [Output Data Config](#output-data-config).
* `resource_config` - (Required) Resource configuration for the validation training job. See [Resource Config](#resource-config).
* `stopping_condition` - (Required) Stopping condition for the validation training job. See [Stopping Condition](#stopping-condition).
* `training_input_mode` - (Required) Input mode for the validation training job. Allowed values are: `Pipe`, `File`, and `FastFile`.

### Input Data Config

* `channel_name` - (Required) Name of the channel.
* `compression_type` - (Optional) Compression type of the input data. Allowed values are: `None` and `Gzip`.
* `content_type` - (Optional) MIME type of the input data.
* `data_source` - (Required) Source of the input data. See [Data Source](#data-source).
* `input_mode` - (Optional) Training input mode for the channel. Allowed values are: `Pipe`, `File`, and `FastFile`.
* `record_wrapper_type` - (Optional) Record wrapper type. Allowed values are: `None` and `RecordIO`.
* `shuffle_config` - (Optional) Shuffle configuration for the channel. See [Shuffle Config](#shuffle-config).

### Data Source

* `file_system_data_source` - (Optional) File system-backed data source. See [File System Data Source](#file-system-data-source).
* `s3_data_source` - (Optional) S3-backed training data source. See [Training S3 Data Source](#training-s3-data-source).

### File System Data Source

* `directory_path` - (Required) Path to the directory in the mounted file system.
* `file_system_access_mode` - (Required) File system access mode.
* `file_system_id` - (Required) ID of the file system.
* `file_system_type` - (Required) File system type.

### Training S3 Data Source

* `attribute_names` - (Optional) List of JSON attribute names to select from the input data.
* `hub_access_config` - (Optional) SageMaker AI Hub access settings. See [Hub Access Config](#hub-access-config).
* `instance_group_names` - (Optional) Instance group names associated with the data source.
* `model_access_config` - (Optional) Model access settings. See [Model Access Config](#model-access-config).
* `s3_data_distribution_type` - (Optional) Distribution type for S3 data. Allowed values are: `FullyReplicated` and `ShardedByS3Key`.
* `s3_data_type` - (Required) Type of S3 training data.
* `s3_uri` - (Required) S3 or HTTPS URI of the training data.

### Shuffle Config

* `seed` - (Required) Shuffle seed.

### Output Data Config

* `compression_type` - (Optional) Compression type for the output data. Allowed values are: `None` and `GZIP`.
* `kms_key_id` - (Optional) KMS key ID used to encrypt output data.
* `s3_output_path` - (Required) S3 or HTTPS URI where output data is stored.

### Resource Config

* `instance_count` - (Optional) Number of training instances.
* `instance_groups` - (Optional) Instance group definitions for the training job. See [Instance Groups](#instance-groups).
* `instance_placement_config` - (Optional) Placement configuration for the training job. See [Instance Placement Config](#instance-placement-config).
* `instance_type` - (Optional) Training instance type.
* `keep_alive_period_in_seconds` - (Optional) Warm pool keep-alive period in seconds.
* `training_plan_arn` - (Optional) ARN of the SageMaker AI training plan.
* `volume_kms_key_id` - (Optional) KMS key ID used to encrypt the training volume.
* `volume_size_in_gb` - (Optional) Size of the training volume in GiB.

### Instance Groups

* `instance_count` - (Required) Number of instances in the group.
* `instance_group_name` - (Required) Name of the instance group.
* `instance_type` - (Required) Instance type for the group.

### Instance Placement Config

* `enable_multiple_jobs` - (Optional) Whether multiple jobs can share the placement configuration.
* `placement_specifications` - (Optional) Placement specifications for ultra servers. See [Placement Specifications](#placement-specifications).

### Placement Specifications

* `instance_count` - (Required) Number of instances for the placement specification.
* `ultra_server_id` - (Optional) Ultra server ID.

### Stopping Condition

* `max_pending_time_in_seconds` - (Optional) Maximum time, in seconds, a job can remain pending.
* `max_runtime_in_seconds` - (Optional) Maximum runtime, in seconds, for the training job.
* `max_wait_time_in_seconds` - (Optional) Maximum wait time, in seconds, including spot interruptions.

### Transform Job Definition

* `batch_strategy` - (Optional) Batch strategy for the transform job. Allowed values are: `MultiRecord` and `SingleRecord`.
* `environment` - (Optional) Environment variables to pass to the transform container.
* `max_concurrent_transforms` - (Optional) Maximum number of parallel transform requests.
* `max_payload_in_mb` - (Optional) Maximum payload size, in MiB, for transform requests.
* `transform_input` - (Required) Input configuration for the transform job. See [Transform Input](#transform-input).
* `transform_output` - (Required) Output configuration for the transform job. See [Transform Output](#transform-output).
* `transform_resources` - (Required) Compute resources for the transform job. See [Transform Resources](#transform-resources).

### Transform Input

* `compression_type` - (Optional) Compression type of the input data. Allowed values are: `None` and `Gzip`.
* `content_type` - (Optional) MIME type of the input data.
* `data_source` - (Required) Data source for the transform job. See [Transform Job Data Source](#transform-job-data-source).
* `split_type` - (Optional) Method used to split the transform input. Allowed values are: `None`, `Line`, `RecordIO`, and `TFRecord`.

### Transform Job Data Source

* `s3_data_source` - (Required) S3-backed data source for the transform job. See [Transform S3 Data Source](#transform-s3-data-source).

### Transform S3 Data Source

* `s3_data_type` - (Required) Type of S3 data for the transform input.
* `s3_uri` - (Required) S3 or HTTPS URI of the transform input data.

### Transform Output

* `accept` - (Optional) MIME type of the transform output.
* `assemble_with` - (Optional) Method used to assemble the transform output. Allowed values are: `None` and `Line`.
* `kms_key_id` - (Optional) KMS key ID used to encrypt transform output.
* `s3_output_path` - (Required) S3 or HTTPS URI where transform output is stored.

### Transform Resources

* `instance_count` - (Required) Number of transform instances.
* `instance_type` - (Required) Transform instance type.
* `transform_ami_version` - (Optional) Transform AMI version.
* `volume_kms_key_id` - (Optional) KMS key ID used to encrypt the transform volume.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `algorithm_status` - Status of the algorithm.
* `arn` - ARN of the algorithm.
* `creation_time` - Time when the algorithm was created, in RFC3339 format.
* `product_id` - AWS Marketplace product ID associated with the algorithm.
* `tags_all` - Map of tags assigned to the resource, including tags inherited from the provider `default_tags` configuration block.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `5m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_sagemaker_algorithm.example
  identity = {
    algorithm_name = "example-algorithm"
  }
}

resource "aws_sagemaker_algorithm" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `algorithm_name` - (String) Name of the algorithm.

#### Optional

* `account_id` - (String) AWS account where this resource is managed.
* `region` - (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import SageMaker AI Algorithms using `algorithm_name`. For example:

```terraform
import {
  to = aws_sagemaker_algorithm.example
  id = "example-algorithm"
}
```

Using `terraform import`, import SageMaker AI Algorithms using `algorithm_name`. For example:

```console
% terraform import aws_sagemaker_algorithm.example example-algorithm
```
