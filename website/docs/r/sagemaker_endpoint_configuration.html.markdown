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

The following arguments are supported:

* `production_variants` - (Required) Fields are documented below.
* `kms_key_arn` - (Optional) Amazon Resource Name (ARN) of a AWS Key Management Service key that Amazon SageMaker uses to encrypt data on the storage volume attached to the ML compute instance that hosts the endpoint.
* `name` - (Optional) The name of the endpoint configuration. If omitted, Terraform will assign a random, unique name.
* `tags` - (Optional) A mapping of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `data_capture_config` - (Optional) Specifies the parameters to capture input/output of SageMaker models endpoints. Fields are documented below.
* `async_inference_config` - (Optional) Specifies configuration for how an endpoint performs asynchronous inference.

The `production_variants` block supports:

* `initial_instance_count` - (Required) Initial number of instances used for auto-scaling.
* `instance_type` (Required) - The type of instance to start.
* `accelerator_type` (Optional) - The size of the Elastic Inference (EI) instance to use for the production variant.
* `initial_variant_weight` (Optional) - Determines initial traffic distribution among all of the models that you specify in the endpoint configuration. If unspecified, it defaults to 1.0.
* `model_name` - (Required) The name of the model to use.
* `variant_name` - (Optional) The name of the variant. If omitted, Terraform will assign a random, unique name.

The `data_capture_config` block supports:

* `initial_sampling_percentage` - (Required) Portion of data to capture. Should be between 0 and 100.
* `destination_s3_uri` - (Required) The URL for S3 location where the captured data is stored.
* `capture_options` - (Required) Specifies what data to capture. Fields are documented below.
* `kms_key_id` - (Optional) Amazon Resource Name (ARN) of a AWS Key Management Service key that Amazon SageMaker uses to encrypt the captured data on Amazon S3.
* `enable_capture` - (Optional) Flag to enable data capture. Defaults to `false`.
* `capture_content_type_header` - (Optional) The content type headers to capture. Fields are documented below.

The `capture_options` block supports:

* `capture_mode` - (Required) Specifies the data to be captured. Should be one of `Input` or `Output`.

The `capture_content_type_header` block supports:

* `csv_content_types` - (Optional) The CSV content type headers to capture.
* `json_content_types` - (Optional) The JSON content type headers to capture.

The `async_inference_config` block supports:

* `output_config` - (Required) Specifies the configuration for asynchronous inference invocation outputs.
* `client_config` - (Optional) Configures the behavior of the client used by Amazon SageMaker to interact with the model container during asynchronous inference.

The `client_config` block supports:

* `max_concurrent_invocations_per_instance` - (Optional) The maximum number of concurrent requests sent by the SageMaker client to the model container. If no value is provided, Amazon SageMaker will choose an optimal value for you.

The `output_config` block supports:

* `s3_output_path` - (Required) The Amazon S3 location to upload inference responses to.
* `kms_key_id` - (Optional) The Amazon Web Services Key Management Service (Amazon Web Services KMS) key that Amazon SageMaker uses to encrypt the asynchronous inference output in Amazon S3.
* `notification_config` - (Optional) Specifies the configuration for notifications of inference results for asynchronous inference.

The `notification_config` block supports:

* `error_topic` - (Optional) Amazon SNS topic to post a notification to when inference fails. If no topic is provided, no notification is sent on failure.
* `success_topic` - (Optional) Amazon SNS topic to post a notification to when inference completes successfully. If no topic is provided, no notification is sent on success.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The Amazon Resource Name (ARN) assigned by AWS to this endpoint configuration.
* `name` - The name of the endpoint configuration.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

Endpoint configurations can be imported using the `name`, e.g.,

```
$ terraform import aws_sagemaker_endpoint_configuration.test_endpoint_config endpoint-config-foo
```
