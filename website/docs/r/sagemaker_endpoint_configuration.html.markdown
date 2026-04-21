---
subcategory: "SageMaker AI"
layout: "aws"
page_title: "AWS: aws_sagemaker_endpoint_configuration"
description: |-
  Provides a SageMaker AI Endpoint Configuration resource.
---

# Resource: aws_sagemaker_endpoint_configuration

Provides a SageMaker AI endpoint configuration resource.

## Example Usage

Basic usage:

```terraform
resource "aws_sagemaker_endpoint_configuration" "ec" {
  name = "my-endpoint-config"

  production_variants {
    variant_name           = "variant-1"
    model_name             = aws_sagemaker_model.m.name
    initial_instance_count = 1
    instance_type          = "ml.t2.medium"
  }

  tags = {
    Name = "foo"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `async_inference_config` - (Optional) How an endpoint performs asynchronous inference.
* `data_capture_config` - (Optional) Parameters to capture input/output of SageMaker AI models endpoints. Fields are documented below.
* `execution_role_arn` - (Optional) ARN of an IAM role that SageMaker AI can assume to perform actions on your behalf. Required when `model_name` is not specified in `production_variants` to support Inference Components.
* `kms_key_arn` - (Optional) ARN of a AWS KMS key that SageMaker AI uses to encrypt data on the storage volume attached to the ML compute instance that hosts the endpoint.
* `name_prefix` - (Optional) Unique endpoint configuration name beginning with the specified prefix. Conflicts with `name`.
* `name` - (Optional) Name of the endpoint configuration. If omitted, Terraform will assign a random, unique name. Conflicts with `name_prefix`.
* `production_variants` - (Required) List each model that you want to host at this endpoint. [See below](#production_variants).
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `shadow_production_variants` - (Optional) Models that you want to host at this endpoint in shadow mode with production traffic replicated from the model specified on `production_variants`. If you use this field, you can only specify one variant for `production_variants` and one variant for `shadow_production_variants`. [See below](#production_variants) (same arguments as `production_variants`).
* `tags` - (Optional) Mapping of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### production_variants

* `accelerator_type` - (Optional) Size of the Elastic Inference (EI) instance to use for the production variant.
* `container_startup_health_check_timeout_in_seconds` - (Optional) Timeout value, in seconds, for your inference container to pass health check by SageMaker AI Hosting. For more information about health check, see [How Your Container Should Respond to Health Check (Ping) Requests](https://docs.aws.amazon.com/sagemaker/latest/dg/your-algorithms-inference-code.html#your-algorithms-inference-algo-ping-requests). Valid values between `60` and `3600`.
* `core_dump_config` - (Optional) Core dump configuration from the model container when the process crashes. Fields are documented below.
* `enable_ssm_access` - (Optional) Whether to turn on native AWS SSM access for a production variant behind an endpoint. By default, SSM access is disabled for all production variants behind endpoints. Ignored if `model_name` is not set (Inference Components endpoint).
* `inference_ami_version` - (Optional) Option from a collection of preconfigured AMI images. Each image is configured by AWS with a set of software and driver versions. AWS optimizes these configurations for different machine learning workloads.
* `initial_instance_count` - (Optional) Initial number of instances used for auto-scaling.
* `initial_variant_weight` - (Optional) Initial traffic distribution among all of the models that you specify in the endpoint configuration. If unspecified, defaults to `1.0`. Ignored if `model_name` is not set (Inference Components endpoint).
* `instance_type` - (Optional)  Type of instance to start.
* `managed_instance_scaling` - (Optional) Control the range in the number of instances that the endpoint provisions as it scales up or down to accommodate traffic.
* `model_data_download_timeout_in_seconds` - (Optional) Timeout value, in seconds, to download and extract the model that you want to host from S3 to the individual inference instance associated with this production variant. Valid values between `60` and `3600`.
* `model_name` - (Optional) Name of the model to use. Required unless using Inference Components (in which case `execution_role_arn` must be specified at the endpoint configuration level).
* `routing_config` - (Optional) How the endpoint routes incoming traffic. See [routing_config](#routing_config) below.
* `serverless_config` - (Optional) How an endpoint performs asynchronous inference.
* `variant_name` - (Optional) Name of the variant. If omitted, Terraform will assign a random, unique name.
* `volume_size_in_gb` - (Optional) Size, in GB, of the ML storage volume attached to individual inference instance associated with the production variant. Valid values between `1` and `512`.

#### core_dump_config

* `destination_s3_uri` - (Required) S3 bucket to send the core dump to.
* `kms_key_id` - (Required) KMS key that SageMaker AI uses to encrypt the core dump data at rest using S3 server-side encryption.

#### routing_config

* `routing_strategy` - (Required) How the endpoint routes incoming traffic. Valid values are `LEAST_OUTSTANDING_REQUESTS` and `RANDOM`. `LEAST_OUTSTANDING_REQUESTS` routes requests to the specific instances that have more capacity to process them. `RANDOM` routes each request to a randomly chosen instance.

#### serverless_config

* `max_concurrency` - (Required) Maximum number of concurrent invocations your serverless endpoint can process. Valid values are between `1` and `200`.
* `memory_size_in_mb` - (Required) Memory size of your serverless endpoint. Valid values are in 1 GB increments: `1024` MB, `2048` MB, `3072` MB, `4096` MB, `5120` MB, or `6144` MB.
* `provisioned_concurrency` - Amount of provisioned concurrency to allocate for the serverless endpoint. Should be less than or equal to `max_concurrency`. Valid values are between `1` and `200`.

#### managed_instance_scaling

* `max_instance_count` - (Optional) Maximum number of instances that the endpoint can provision when it scales up to accommodate an increase in traffic.
* `min_instance_count` - (Optional) Minimum number of instances that the endpoint must retain when it scales down to accommodate a decrease in traffic.
* `status` - (Optional) Whether managed instance scaling is enabled. Valid values are `ENABLED` and `DISABLED`.

### data_capture_config

* `capture_content_type_header` - (Optional) Content type headers to capture. See [`capture_content_type_header`](#capture_content_type_header) below.
* `capture_options` - (Required) What data to capture. Fields are documented below.
* `destination_s3_uri` - (Required) URL for S3 location where the captured data is stored.
* `enable_capture` - (Optional) Flag to enable data capture. Defaults to `false`.
* `initial_sampling_percentage` - (Required) Portion of data to capture. Should be between 0 and 100.
* `kms_key_id` - (Optional) ARN of a KMS key that SageMaker AI uses to encrypt the captured data on S3.

#### capture_options

* `capture_mode` - (Required) Data to be captured. Should be one of `Input`, `Output` or `InputAndOutput`.

#### capture_content_type_header

* `csv_content_types` - (Optional) CSV content type headers to capture. One of `csv_content_types` or `json_content_types` is required.
* `json_content_types` - (Optional) The JSON content type headers to capture. One of `json_content_types` or `csv_content_types` is required.

### async_inference_config

* `client_config` - (Optional) Configures the behavior of the client used by SageMaker AI to interact with the model container during asynchronous inference.
* `output_config` - (Required) Configuration for asynchronous inference invocation outputs.

#### client_config

* `max_concurrent_invocations_per_instance` - (Optional) Maximum number of concurrent requests sent by the SageMaker AI client to the model container. If no value is provided, SageMaker AI will choose an optimal value for you.

#### output_config

* `s3_output_path` - (Required) S3 location to upload inference responses to.
* `s3_failure_path` - (Optional) S3 location to upload failure inference responses to.
* `kms_key_id` - (Optional) KMS key that SageMaker AI uses to encrypt the asynchronous inference output in S3.
* `notification_config` - (Optional) Configuration for notifications of inference results for asynchronous inference.

##### notification_config

* `error_topic` - (Optional) SNS topic to post a notification to when inference fails. If no topic is provided, no notification is sent on failure.
* `include_inference_response_in` - (Optional) SNS topics where you want the inference response to be included. Valid values are `SUCCESS_NOTIFICATION_TOPIC` and `ERROR_NOTIFICATION_TOPIC`.
* `success_topic` - (Optional) SNS topic to post a notification to when inference completes successfully. If no topic is provided, no notification is sent on success.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN assigned by AWS to this endpoint configuration.
* `name` - Name of the endpoint configuration.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import endpoint configurations using the `name`. For example:

```terraform
import {
  to = aws_sagemaker_endpoint_configuration.test_endpoint_config
  id = "endpoint-config-foo"
}
```

Using `terraform import`, import endpoint configurations using the `name`. For example:

```console
% terraform import aws_sagemaker_endpoint_configuration.test_endpoint_config endpoint-config-foo
```
