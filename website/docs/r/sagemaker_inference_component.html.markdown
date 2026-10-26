---
subcategory: "SageMaker AI"
layout: "aws"
page_title: "AWS: aws_sagemaker_inference_component"
description: |-
  Provides a SageMaker AI Inference Component resource.
---

# Resource: aws_sagemaker_inference_component

Provides a SageMaker AI Inference Component resource. An inference component is a SageMaker AI hosting object that you can use to deploy a model to an endpoint. You can optimize resource utilization by tailoring how the required CPU cores, accelerators, and memory are allocated. You can deploy multiple inference components to an endpoint, where each inference component contains one model and the resource utilization needs for that individual model.

## Example Usage

### Basic Usage

```terraform
resource "aws_sagemaker_inference_component" "example" {
  name          = "my-inference-component"
  endpoint_name = aws_sagemaker_endpoint.example.name
  variant_name  = "variant-1"

  specification {
    model_name = aws_sagemaker_model.example.name

    compute_resource_requirements {
      min_memory_required_in_mb = 1024
    }
  }

  runtime_config {
    copy_count = 1
  }
}
```

### With Container Specification

```terraform
resource "aws_sagemaker_inference_component" "example" {
  name          = "my-inference-component"
  endpoint_name = aws_sagemaker_endpoint.example.name
  variant_name  = "variant-1"

  specification {
    container {
      image        = "123456789012.dkr.ecr.us-east-1.amazonaws.com/my-model:latest"
      artifact_url = "s3://my-bucket/model.tar.gz"

      environment = {
        MODEL_NAME = "my-model"
      }
    }

    compute_resource_requirements {
      min_memory_required_in_mb              = 2048
      max_memory_required_in_mb              = 4096
      number_of_cpu_cores_required           = 2
      number_of_accelerator_devices_required = 1
    }

    startup_parameters {
      container_startup_health_check_timeout_in_seconds = 300
      model_data_download_timeout_in_seconds            = 300
    }
  }

  runtime_config {
    copy_count = 2
  }
}
```

### With Deployment Config

`deployment_config` controls how *updates* to an existing inference component are rolled out; it has no effect on the initial create. Add it (or change it) on a subsequent apply to govern the rollout of a `specification` or `runtime_config` change.

```terraform
resource "aws_sagemaker_inference_component" "example" {
  name          = "my-inference-component"
  endpoint_name = aws_sagemaker_endpoint.example.name
  variant_name  = "variant-1"

  specification {
    model_name = aws_sagemaker_model.example.name

    compute_resource_requirements {
      min_memory_required_in_mb = 1024
    }
  }

  runtime_config {
    copy_count = 4
  }

  deployment_config {
    rolling_update_policy {
      maximum_batch_size {
        type  = "COPY_COUNT"
        value = 1
      }

      wait_interval_in_seconds = 60
    }

    auto_rollback_configuration {
      alarms {
        alarm_name = aws_cloudwatch_metric_alarm.example.alarm_name
      }
    }
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `deployment_config` - (Optional) Deployment configuration for the inference component, which controls the rollout strategy and rollback behavior. This directive applies only when the inference component is **updated** — the `CreateInferenceComponent` API does not accept it, so it has no effect on initial creation. It is not returned by the API and is therefore not refreshed into state or verified on import. See [`deployment_config` Block](#deployment_config-block).
* `endpoint_name` - (Required) Name of an existing endpoint where you host the inference component.
* `name` - (Required) Unique name to assign to the inference component.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `runtime_config` - (Optional) Runtime settings for a model that is deployed with an inference component. See [`runtime_config` Block](#runtime_config-block).
* `specification` - (Optional) Details about the resources to deploy with this inference component. Exactly one of `specification` or `specifications` must be configured. See [`specification` and `specifications` Block](#specification-and-specifications-block).
* `specifications` - (Optional) List of specification objects for the inference component, one per instance type. At least two blocks are required; configure `specification` instead to deploy to a single instance type. Exactly one of `specification` or `specifications` must be configured. See [`specification` and `specifications` Block](#specification-and-specifications-block).
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `variant_name` - (Optional) Name of an existing production variant where you host the inference component. Required for a standard inference component; must be omitted for an adapter inference component (one that sets `base_inference_component_name`).

### `specification` and `specifications` Block

* `base_inference_component_name` - (Optional) Name of an existing inference component that is to contain the inference component that you're creating. Used for adapter inference components.
* `compute_resource_requirements` - (Optional) Compute resources allocated to run a model. See [`compute_resource_requirements` Block](#compute_resource_requirements-block).
* `container` - (Optional) Container that provides the runtime environment for a model. See [`container` Block](#container-block).
* `data_cache_config` - (Optional) Settings that affect how the inference component caches data. See [`data_cache_config` Block](#data_cache_config-block).
* `instance_type` - (Optional) ML compute instance type for this specification. Required when using the `specifications` parameter with multiple entries.
* `model_name` - (Optional) Name of an existing SageMaker AI model object in your account that you want to deploy with the inference component.
* `scheduling_config` - (Optional) Scheduling configuration for placing inference component copies. See [`scheduling_config` Block](#scheduling_config-block).
* `startup_parameters` - (Optional) Settings that take effect while the model container starts up. See [`startup_parameters` Block](#startup_parameters-block).

### `compute_resource_requirements` Block

* `max_memory_required_in_mb` - (Optional) Maximum MB of memory to allocate to run a model. Minimum value of `128`.
* `min_memory_required_in_mb` - (Required) Minimum MB of memory to allocate to run a model. Minimum value of `128`.
* `number_of_accelerator_devices_required` - (Optional) Number of accelerators (GPUs or AWS Inferentia) to allocate. Minimum value of `1`.
* `number_of_cpu_cores_required` - (Optional) Number of CPU cores to allocate. Minimum value of `0.25`.

### `container` Block

* `artifact_url` - (Optional) Amazon S3 path where the model artifacts are stored (.tar.gz).
* `container_metrics_config` - (Optional) Configuration for container metrics scraping. See [`container_metrics_config` Block](#container_metrics_config-block).
* `environment` - (Optional) Environment variables to set in the Docker container.
* `image` - (Optional) Amazon ECR path where the Docker image for the model is stored.

### `container_metrics_config` Block

* `metrics_endpoint` - (Optional) Metrics endpoint configuration. See [`metrics_endpoint` Block](#metrics_endpoint-block).

### `metrics_endpoint` Block

* `metric_publish_frequency_in_seconds` - (Optional) Interval in seconds at which metrics are published. Valid values: `10`, `30`, `60`, `120`, `180`, `240`, `300`. Defaults to `60`.
* `metrics_endpoint_path` - (Required) Path to the metrics endpoint. Must start with `/`. Maximum 256 characters.

### `data_cache_config` Block

* `enable_caching` - (Required) Whether the endpoint caches model artifacts and container images.

### `scheduling_config` Block

* `availability_zone_balance` - (Optional) Configuration for balancing copies across Availability Zones. See [`availability_zone_balance` Block](#availability_zone_balance-block).
* `placement_strategy` - (Required) Strategy for placing inference component copies. Valid values: `SPREAD`, `BINPACK`.

### `availability_zone_balance` Block

* `enforcement_mode` - (Required) How strictly the AZ balance constraint is enforced. Valid values: `PERMISSIVE`.
* `max_imbalance` - (Optional) Maximum allowed difference in copy count between any two Availability Zones. Defaults to `0`, which the service applies but does not report, so the value stays absent from state unless configured.

### `startup_parameters` Block

* `container_startup_health_check_timeout_in_seconds` - (Optional) Timeout in seconds for the inference container to pass health check. Valid values: `60`-`3600`.
* `model_data_download_timeout_in_seconds` - (Optional) Timeout in seconds to download and extract the model. Valid values: `60`-`3600`.

### `runtime_config` Block

* `copy_count` - (Required) Number of runtime copies of the model container to deploy.

### `deployment_config` Block

* `auto_rollback_configuration` - (Optional) Automatic rollback configuration. See [`auto_rollback_configuration` Block](#auto_rollback_configuration-block).
* `rolling_update_policy` - (Required) Rolling deployment strategy. See [`rolling_update_policy` Block](#rolling_update_policy-block).

### `rolling_update_policy` Block

* `maximum_batch_size` - (Required) Batch size for each rolling step. See [`maximum_batch_size` and `rollback_maximum_batch_size` Block](#maximum_batch_size-and-rollback_maximum_batch_size-block).
* `maximum_execution_timeout_in_seconds` - (Optional) Time limit for the total deployment. Valid values are between `600` and `28800`.
* `rollback_maximum_batch_size` - (Optional) Batch size for a rollback. See [`maximum_batch_size` and `rollback_maximum_batch_size` Block](#maximum_batch_size-and-rollback_maximum_batch_size-block).
* `wait_interval_in_seconds` - (Required) Length of the baking period in seconds. Valid values: `0`-`3600`.

### `maximum_batch_size` and `rollback_maximum_batch_size` Block

* `type` - (Required) Capacity size type. Valid values: `COPY_COUNT`, `CAPACITY_PERCENT`.
* `value` - (Required) Capacity size value.

### `auto_rollback_configuration` Block

* `alarms` - (Optional) List of CloudWatch alarms that trigger rollback. See [`alarms` Block](#alarms-block).

### `alarms` Block

* `alarm_name` - (Required) Name of a CloudWatch alarm.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the inference component.
* `creation_time` - Time when the inference component was created.
* `endpoint_arn` - ARN of the endpoint that hosts the inference component.
* `failure_reason` - If the inference component status is `Failed`, the reason for the failure.
* `last_modified_time` - Time when the inference component was last updated.
* `runtime_config` - Runtime status of the inference component. See below.
* `specification` / `specifications` - Each `container` block exports additional attributes. See below.
* `status` - Status of the inference component (e.g., `InService`, `Creating`, `Updating`, `Failed`, `Deleting`).
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

### `runtime_config` Block

* `current_copy_count` - Number of runtime copies currently deployed.
* `desired_copy_count` - Number of runtime copies requested.
* `placement_status` - Placement status of the inference component copies across instance types. See below.

### `placement_status` Block

* `current_copy_count` - Number of copies placed on this instance type.
* `instance_type` - Instance type on which the copies are placed.

### `container` Block

* `resolved_image` - Specific digest path of the container image that was resolved and deployed.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_sagemaker_inference_component.example
  identity = {
    name = "my-inference-component"
  }
}

resource "aws_sagemaker_inference_component" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `name` (String) Name of the inference component.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `id` attribute. For example:

```terraform
import {
  to = aws_sagemaker_inference_component.example
  id = "my-inference-component"
}
```

Using `terraform import`, import SageMaker AI Inference Components using the `name`. For example:

```console
% terraform import aws_sagemaker_inference_component.example my-inference-component
```
