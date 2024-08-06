---
subcategory: "SageMaker"
layout: "aws"
page_title: "AWS: aws_sagemaker_endpoint_configuration"
description: |-
  Provides a SageMaker Endpoint Configuration resource.
---

# Resource: aws_sagemaker_endpoint_configuration

Provides a SageMaker endpoint configuration resource.

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

* `production_variants` - (Required) An list of ProductionVariant objects, one for each model that you want to host at this endpoint. Fields are documented below.
* `kms_key_arn` - (Optional) Amazon Resource Name (ARN) of a AWS Key Management Service key that Amazon SageMaker uses to encrypt data on the storage volume attached to the ML compute instance that hosts the endpoint.
* `name` - (Optional) The name of the endpoint configuration. If omitted, Terraform will assign a random, unique name. Conflicts with `name_prefix`.
* `name_prefix` - (Optional) Creates a unique endpoint configuration name beginning with the specified prefix. Conflicts with `name`.
* `tags` - (Optional) A mapping of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `data_capture_config` - (Optional) Specifies the parameters to capture input/output of SageMaker models endpoints. Fields are documented below.
* `async_inference_config` - (Optional) Specifies configuration for how an endpoint performs asynchronous inference.
* `shadow_production_variants` - (Optional) Array of ProductionVariant objects. There is one for each model that you want to host at this endpoint in shadow mode with production traffic replicated from the model specified on ProductionVariants. If you use this field, you can only specify one variant for ProductionVariants and one variant for ShadowProductionVariants. Fields are documented below.

### production_variants

* `accelerator_type` - (Optional) The size of the Elastic Inference (EI) instance to use for the production variant.
* `container_startup_health_check_timeout_in_seconds` - (Optional) The timeout value, in seconds, for your inference container to pass health check by SageMaker Hosting. For more information about health check, see [How Your Container Should Respond to Health Check (Ping) Requests](https://docs.aws.amazon.com/sagemaker/latest/dg/your-algorithms-inference-code.html#your-algorithms-inference-algo-ping-requests). Valid values between `60` and `3600`.
* `core_dump_config` - (Optional) Specifies configuration for a core dump from the model container when the process crashes. Fields are documented below.
* `enable_ssm_access` - (Optional) You can use this parameter to turn on native Amazon Web Services Systems Manager (SSM) access for a production variant behind an endpoint. By default, SSM access is disabled for all production variants behind an endpoints.
* `inference_ami_version` - (Optional) Specifies an option from a collection of preconfigured Amazon Machine Image (AMI) images. Each image is configured by Amazon Web Services with a set of software and driver versions. Amazon Web Services optimizes these configurations for different machine learning workloads.
* `initial_instance_count` - (Optional) Initial number of instances used for auto-scaling.
* `instance_type` - (Optional)  The type of instance to start.
* `initial_variant_weight` - (Optional) Determines initial traffic distribution among all of the models that you specify in the endpoint configuration. If unspecified, it defaults to `1.0`.
* `model_data_download_timeout_in_seconds` - (Optional) The timeout value, in seconds, to download and extract the model that you want to host from Amazon S3 to the individual inference instance associated with this production variant. Valid values between `60` and `3600`.
* `model_name` - (Required) The name of the model to use.
* `routing_config` - (Optional) Sets how the endpoint routes incoming traffic. See [routing_config](#routing_config) below.
* `serverless_config` - (Optional) Specifies configuration for how an endpoint performs asynchronous inference.
* `variant_name` - (Optional) The name of the variant. If omitted, Terraform will assign a random, unique name.
* `volume_size_in_gb` - (Optional) The size, in GB, of the ML storage volume attached to individual inference instance associated with the production variant. Valid values between `1` and `512`.

#### core_dump_config

* `destination_s3_uri` - (Required) The Amazon S3 bucket to send the core dump to.
* `kms_key_id` - (Required) The Amazon Web Services Key Management Service (Amazon Web Services KMS) key that SageMaker uses to encrypt the core dump data at rest using Amazon S3 server-side encryption.

#### routing_config

* `routing_strategy` - (Required) Sets how the endpoint routes incoming traffic. Valid values are `LEAST_OUTSTANDING_REQUESTS` and `RANDOM`. `LEAST_OUTSTANDING_REQUESTS` routes requests to the specific instances that have more capacity to process them. `RANDOM` routes each request to a randomly chosen instance.

#### serverless_config

* `max_concurrency` - (Required) The maximum number of concurrent invocations your serverless endpoint can process. Valid values are between `1` and `200`.
* `memory_size_in_mb` - (Required) The memory size of your serverless endpoint. Valid values are in 1 GB increments: `1024` MB, `2048` MB, `3072` MB, `4096` MB, `5120` MB, or `6144` MB.
* `provisioned_concurrency` - The amount of provisioned concurrency to allocate for the serverless endpoint. Should be less than or equal to `max_concurrency`. Valid values are between `1` and `200`.

### data_capture_config

* `initial_sampling_percentage` - (Required) Portion of data to capture. Should be between 0 and 100.
* `destination_s3_uri` - (Required) The URL for S3 location where the captured data is stored.
* `capture_options` - (Required) Specifies what data to capture. Fields are documented below.
* `kms_key_id` - (Optional) Amazon Resource Name (ARN) of a AWS Key Management Service key that Amazon SageMaker uses to encrypt the captured data on Amazon S3.
* `enable_capture` - (Optional) Flag to enable data capture. Defaults to `false`.
* `capture_content_type_header` - (Optional) The content type headers to capture. Fields are documented below.

#### capture_options

* `capture_mode` - (Required) Specifies the data to be captured. Should be one of `Input`, `Output` or `InputAndOutput`.

#### capture_content_type_header

* `csv_content_types` - (Optional) The CSV content type headers to capture.
* `json_content_types` - (Optional) The JSON content type headers to capture.

### async_inference_config

* `output_config` - (Required) Specifies the configuration for asynchronous inference invocation outputs.
* `client_config` - (Optional) Configures the behavior of the client used by Amazon SageMaker to interact with the model container during asynchronous inference.

#### client_config

* `max_concurrent_invocations_per_instance` - (Optional) The maximum number of concurrent requests sent by the SageMaker client to the model container. If no value is provided, Amazon SageMaker will choose an optimal value for you.

#### output_config

* `s3_output_path` - (Required) The Amazon S3 location to upload inference responses to.
* `s3_failure_path` - (Optional) The Amazon S3 location to upload failure inference responses to.
* `kms_key_id` - (Optional) The Amazon Web Services Key Management Service (Amazon Web Services KMS) key that Amazon SageMaker uses to encrypt the asynchronous inference output in Amazon S3.
* `notification_config` - (Optional) Specifies the configuration for notifications of inference results for asynchronous inference.

##### notification_config

* `include_inference_response_in` - (Optional) The Amazon SNS topics where you want the inference response to be included. Valid values are `SUCCESS_NOTIFICATION_TOPIC` and `ERROR_NOTIFICATION_TOPIC`.
* `error_topic` - (Optional) Amazon SNS topic to post a notification to when inference fails. If no topic is provided, no notification is sent on failure.
* `success_topic` - (Optional) Amazon SNS topic to post a notification to when inference completes successfully. If no topic is provided, no notification is sent on success.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The Amazon Resource Name (ARN) assigned by AWS to this endpoint configuration.
* `name` - The name of the endpoint configuration.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

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
