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

* `name` - (Required) A unique name to assign to the inference component.
* `endpoint_name` - (Required) The name of an existing endpoint where you host the inference component.
* `variant_name` - (Optional) The name of an existing production variant where you host the inference component. Required for a standard inference component; must be omitted for an adapter inference component (one that sets `base_inference_component_name`).
* `specification` - (Optional, conflicts with `specifications`) Details about the resources to deploy with this inference component. See [Specification](#specification).
* `specifications` - (Optional, conflicts with `specification`) A list of specification objects for the inference component, one per instance type. See [Specification](#specification).
* `runtime_config` - (Optional) Runtime settings for a model that is deployed with an inference component. See [Runtime Config](#runtime-config).
* `deployment_config` - (Optional) The deployment configuration for the inference component, which controls the rollout strategy and rollback behavior. This directive applies only when the inference component is **updated** — the `CreateInferenceComponent` API does not accept it, so it has no effect on initial creation. It is not returned by the API and is therefore not refreshed into state or verified on import. See [Deployment Config](#deployment-config).
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### Specification

* `model_name` - (Optional) The name of an existing SageMaker AI model object in your account that you want to deploy with the inference component.
* `base_inference_component_name` - (Optional) The name of an existing inference component that is to contain the inference component that you're creating. Used for adapter inference components.
* `instance_type` - (Optional) The ML compute instance type for this specification. Required when using the `specifications` parameter with multiple entries.
* `compute_resource_requirements` - (Optional) The compute resources allocated to run a model. See [Compute Resource Requirements](#compute-resource-requirements).
* `container` - (Optional) Defines a container that provides the runtime environment for a model. See [Container](#container).
* `data_cache_config` - (Optional) Settings that affect how the inference component caches data. See [Data Cache Config](#data-cache-config).
* `scheduling_config` - (Optional) The scheduling configuration for placing inference component copies. See [Scheduling Config](#scheduling-config).
* `startup_parameters` - (Optional) Settings that take effect while the model container starts up. See [Startup Parameters](#startup-parameters).

### Compute Resource Requirements

* `min_memory_required_in_mb` - (Required) The minimum MB of memory to allocate to run a model. Minimum value of `128`.
* `max_memory_required_in_mb` - (Optional) The maximum MB of memory to allocate to run a model. Minimum value of `128`.
* `number_of_accelerator_devices_required` - (Optional) The number of accelerators (GPUs or AWS Inferentia) to allocate. Minimum value of `1`.
* `number_of_cpu_cores_required` - (Optional) The number of CPU cores to allocate. Minimum value of `0.25`.

### Container

* `image` - (Optional) The Amazon ECR path where the Docker image for the model is stored.
* `artifact_url` - (Optional) The Amazon S3 path where the model artifacts are stored (.tar.gz).
* `environment` - (Optional) The environment variables to set in the Docker container.
* `container_metrics_config` - (Optional) Configuration for container metrics scraping. See [Container Metrics Config](#container-metrics-config).

### Container Metrics Config

* `metrics_endpoint` - (Optional) The metrics endpoint configuration. See [Metrics Endpoint](#metrics-endpoint).

### Metrics Endpoint

* `metrics_endpoint_path` - (Required) The path to the metrics endpoint. Must start with `/`. Maximum 256 characters.
* `metric_publish_frequency_in_seconds` - (Optional) The interval in seconds at which metrics are published. Valid values: `10`, `30`, `60`, `120`, `180`, `240`, `300`. Defaults to `60`.

### Data Cache Config

* `enable_caching` - (Required) Whether the endpoint caches model artifacts and container images.

### Scheduling Config

* `placement_strategy` - (Required) The strategy for placing inference component copies. Valid values: `SPREAD`, `BINPACK`.
* `availability_zone_balance` - (Optional) Configuration for balancing copies across Availability Zones. See [Availability Zone Balance](#availability-zone-balance).

### Availability Zone Balance

* `enforcement_mode` - (Required) How strictly the AZ balance constraint is enforced. Valid values: `PERMISSIVE`.
* `max_imbalance` - (Optional) The maximum allowed difference in copy count between any two Availability Zones. Default: `0`.

### Startup Parameters

* `container_startup_health_check_timeout_in_seconds` - (Optional) The timeout in seconds for the inference container to pass health check. Valid values: `60`-`3600`.
* `model_data_download_timeout_in_seconds` - (Optional) The timeout in seconds to download and extract the model. Valid values: `60`-`3600`.

### Runtime Config

* `copy_count` - (Required) The number of runtime copies of the model container to deploy.

### Deployment Config

* `rolling_update_policy` - (Required) Specifies a rolling deployment strategy. See [Rolling Update Policy](#rolling-update-policy).
* `auto_rollback_configuration` - (Optional) Automatic rollback configuration. See [Auto Rollback Configuration](#auto-rollback-configuration).

### Rolling Update Policy

* `maximum_batch_size` - (Required) The batch size for each rolling step. See [Capacity Size](#capacity-size).
* `wait_interval_in_seconds` - (Required) The length of the baking period in seconds. Valid values: `0`-`3600`.
* `maximum_execution_timeout_in_seconds` - (Optional) The time limit for the total deployment. Valid values are between `600` and `28800`.
* `rollback_maximum_batch_size` - (Optional) The batch size for a rollback. See [Capacity Size](#capacity-size).

### Capacity Size

* `type` - (Required) The capacity size type. Valid values: `COPY_COUNT`, `CAPACITY_PERCENT`.
* `value` - (Required) The capacity size value.

### Auto Rollback Configuration

* `alarms` - (Optional) List of CloudWatch alarms that trigger rollback. See [Alarms](#alarms).

### Alarms

* `alarm_name` - (Required) The name of a CloudWatch alarm.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The Amazon Resource Name (ARN) of the inference component.
* `endpoint_arn` - The Amazon Resource Name (ARN) of the endpoint that hosts the inference component.
* `status` - The status of the inference component (e.g., `InService`, `Creating`, `Updating`, `Failed`, `Deleting`).
* `failure_reason` - If the inference component status is `Failed`, the reason for the failure.
* `creation_time` - The time when the inference component was created.
* `last_modified_time` - The time when the inference component was last updated.
* `runtime_config.0.current_copy_count` - The number of runtime copies currently deployed.
* `runtime_config.0.desired_copy_count` - The number of runtime copies requested.
* `runtime_config.0.placement_status` - The placement status of the inference component copies across instance types. Each block contains `current_copy_count` (the number of copies placed on this instance type) and `instance_type`.
* `specification.0.container.0.resolved_image` - The specific digest path of the container image that was resolved and deployed.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

SageMaker AI Inference Components can be imported using the `name`, e.g.,

```
$ terraform import aws_sagemaker_inference_component.example my-inference-component
```
