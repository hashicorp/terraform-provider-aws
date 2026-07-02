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
* `explainer_config` - (Optional) Configuration for Clarify real-time explainability. [See below](#explainer_config).
* `kms_key_arn` - (Optional) ARN of a AWS KMS key that SageMaker AI uses to encrypt data on the storage volume attached to the ML compute instance that hosts the endpoint.
* `name_prefix` - (Optional) Unique endpoint configuration name beginning with the specified prefix. Conflicts with `name`.
* `name` - (Optional) Name of the endpoint configuration. If omitted, Terraform will assign a random, unique name. Conflicts with `name_prefix`.
* `production_variants` - (Required) List each model that you want to host at this endpoint. [See below](#production_variants).
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `shadow_production_variants` - (Optional) Models that you want to host at this endpoint in shadow mode with production traffic replicated from the model specified on `production_variants`. If you use this field, you can only specify one variant for `production_variants` and one variant for `shadow_production_variants`. [See below](#production_variants) (same arguments as `production_variants`).
* `tags` - (Optional) Mapping of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `vpc_config` - (Optional) VPC configuration for the endpoint. Cannot be used when `model_name` is specified in `production_variants`. [See below](#vpc_config).

### production_variants

* `accelerator_type` - (Optional) Size of the Elastic Inference (EI) instance to use for the production variant.
* `capacity_reservation_config` - (Optional) Settings for the capacity reservation for the compute instances that SageMaker AI reserves for an endpoint. See [capacity_reservation_config](#capacity_reservation_config) below.
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

#### capacity_reservation_config

* `capacity_reservation_preference` - (Optional) Capacity reservation preference. Valid value is `capacity-reservations-only`. When set to `capacity-reservations-only`, SageMaker AI launches instances only into an ML capacity reservation; if no capacity is available, the instances fail to launch.
* `ml_reservation_arn` - (Optional) The Amazon Resource Name (ARN) that uniquely identifies the ML capacity reservation that SageMaker AI applies when it deploys the endpoint.

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

### explainer_config

* `clarify_explainer_config` - (Required) Configuration for Clarify explainer. [See below](#clarify_explainer_config).

#### clarify_explainer_config

* `enable_explanations` - (Optional) A JMESPath boolean expression used to filter which records to explain. Refer to the [SageMaker AI documentation](https://docs.aws.amazon.com/sagemaker/latest/dg/clarify-online-explainability-create-endpoint.html#clarify-online-explainability-create-endpoint-enable) for more information.
* `inference_config` - (Optional) The inference configuration parameter for the model container. [See below](#inference_config).
* `shap_config` - (Required) The SHAP baseline configuration. [See below](#shap_config).

#### shap_config

* `number_of_samples` - (Optional) The number of samples to be used for analysis by the Kernal SHAP algorithm.
* `seed` - (Optional) The starting value used to initialize the random number generator in the explainer.
* `shap_baseline_config` - (Required) The configuration for the SHAP baseline. [See below](#shap_baseline_config).
* `text_config` - (Optional) Configuration for text explainability. [See below](#text_config).
* `use_logit` - (Optional) A Boolean toggle to indicate if you want to use the logit function (true) or log-odds units (false).

##### shap_baseline_config

* `mime_type` - (Optional) The MIME type of the baseline data. Choose from `text/csv` or `application/jsonlines`.
* `shap_baseline` - (Optional) The inline SHAP baseline data in string format.
* `shap_baseline_uri` - (Optional) The S3 URI where the SHAP baseline file is stored.

##### text_config

* `granularity` - (Required) The unit of granularity for the analysis of text features. Valid values are `token`, `sentence`, and `paragraph`.
* `language` - (Required) The language of the text features. Valid values include language codes such as `en`, `fr`, `de`, `es`, etc.

#### inference_config

* `content_template` - (Optional) A template string used to format a JSON record into an acceptable model container input.
* `feature_headers` - (Optional) The names of the features.
* `feature_types` - (Optional) A list of data types of the features. Valid values are `numerical`, `categorical`, and `text`.
* `features_attribute` - (Optional) Provides the JMESPath expression to extract the features from a model container input in JSON Lines format.
* `label_attribute` - (Optional) Provides the JMESPath expression to extract the label from a model container output in JSON Lines format.
* `label_headers` - (Optional) The names of the label classes.
* `label_index` - (Optional) Zero-based index used to extract a label header from the model container output in CSV format.
* `max_payload_in_mb` - (Optional) The maximum payload size (MB) allowed of a request from the explainer to the model container.
* `max_record_count` - (Optional) The maximum number of records in a request that the model container can process when querying the model container for the predictions of a synthetic dataset.
* `probability_attribute` - (Optional) Provides the JMESPath expression to extract the probability (or score) from the model container output if the model container is in JSON Lines format.
* `probability_index` - (Optional) Zero-based index used to extract a probability value (score) from the model container output in CSV format.

### vpc_config

* `security_group_ids` - (Required) Set of security group IDs.
* `subnet_ids` - (Required) Set of subnet IDs.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN assigned by AWS to this endpoint configuration.
* `name` - Name of the endpoint configuration.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_sagemaker_endpoint_configuration.example
  identity = {
    name = "example-endpoint-config"
  }
}

resource "aws_sagemaker_endpoint_configuration" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `name` (String) Name of the endpoint configuration.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Endpoint Configurations using `name`. For example:

```terraform
import {
  to = aws_sagemaker_endpoint_configuration.example
  id = "example-endpoint-config"
}
```

Using `terraform import`, import Endpoint Configurations using `name`. For example:

```console
% terraform import aws_sagemaker_endpoint_configuration.example example-endpoint-config
```
