---
subcategory: "CodeBuild"
layout: "aws"
page_title: "AWS: aws_codebuild_fleet"
description: |-
  Provides a CodeBuild Fleet Resource.
---

# Resource: aws_codebuild_fleet

Provides a CodeBuild Fleet Resource.

## Example Usage

```terraform
resource "aws_codebuild_fleet" "test" {
  base_capacity     = 2
  compute_type      = "BUILD_GENERAL1_SMALL"
  environment_type  = "LINUX_CONTAINER"
  name              = "full-example-codebuild-fleet"
  overflow_behavior = "QUEUE"
  scaling_configuration {
    max_capacity = 5
    scaling_type = "TARGET_TRACKING_SCALING"
    target_tracking_scaling_configs {
      metric_type  = "FLEET_UTILIZATION_RATE"
      target_value = 97.5
    }
  }
}
```

### Basic Usage

```terraform
resource "aws_codebuild_fleet" "example" {
  name = "example-codebuild-fleet"
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Fleet name.

The following arguments are optional:

* `base_capacity` - (Optional) Number of machines allocated to the ﬂeet.
* `compute_type` - (Optional) Compute resources the compute fleet uses. See [compute types](https://docs.aws.amazon.com/codebuild/latest/userguide/build-env-ref-compute-types.html#environment.types) for more information and valid values.
* `environment_type` - (Optional) Environment type of the compute fleet. See [environment types](https://docs.aws.amazon.com/codebuild/latest/userguide/build-env-ref-compute-types.html#environment.types) for more information and valid values.
* `overflow_behavior` - (Optional) Overflow behavior for compute fleet. Valid values: `ON_DEMAND`, `QUEUE`.
* `scaling_configuration` - (Optional) Configuration block. Detailed below. This option is only valid when your overflow behavior is `QUEUE`.
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### scaling_configuration

* `max_capacity` - (Optional) Maximum number of instances in the ﬂeet when auto-scaling.
* `scaling_type` - (Optional) Scaling type for a compute fleet. Valid value: `TARGET_TRACKING_SCALING`.
* `target_tracking_scaling_configs` - (Optional) Configuration block. Detailed below.

#### scaling_configuration:target_tracking_scaling_configs

* `metric_type` - (Optional) Metric type to determine auto-scaling. Valid value: `FLEET_UTILIZATION_RATE`.
* `target_value` - (Optional) Value of metricType when to start scaling.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Fleet.
* `created` - Creation time of the fleet.
* `id` - ARN of the Fleet.
* `last_modified` - Last modification time of the fleet.
* `status` - Nested attribute containing information about the current status of the fleet.
    * `context` - Additional information about a compute fleet.
    * `message` - Message associated with the status of a compute fleet.
    * `status_code` - Status code of the compute fleet.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import CodeBuild Fleet using the `name` or the `arn`. For example:

```terraform
import {
  to = aws_codebuild_fleet.name
  id = "fleet-name"
}
```

Using `terraform import`, import CodeBuild Fleet using the `name`. For example:

```console
% terraform import aws_codebuild_fleet.name fleet-name
```
