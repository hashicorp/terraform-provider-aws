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
* `base_capacity` - (Required) Number of machines allocated to the ﬂeet.
* `compute_type` - (Required) Compute resources the compute fleet uses. See [compute types](https://docs.aws.amazon.com/codebuild/latest/userguide/build-env-ref-compute-types.html#environment.types) for more information and valid values.
* `environment_type` - (Required) Environment type of the compute fleet. See [environment types](https://docs.aws.amazon.com/codebuild/latest/userguide/build-env-ref-compute-types.html#environment.types) for more information and valid values.

The following arguments are optional:

* `compute_configuration` - (Optional) The compute configuration of the compute fleet. This is only required if `compute_type` is set to `ATTRIBUTE_BASED_COMPUTE`. See [`compute_configuration`](#compute_configuration) below.
* `fleet_service_role` - (Optional) The service role associated with the compute fleet.
* `image_id` - (Optional) The Amazon Machine Image (AMI) of the compute fleet.
* `overflow_behavior` - (Optional) Overflow behavior for compute fleet. Valid values: `ON_DEMAND`, `QUEUE`.
* `proxy_configuration` - (Optional) Configuration block for proxy settings that CodeBuild fleet uses to access external resources. See [`proxy_configuration`](#proxy_configuration) below.
* `scaling_configuration` - (Optional) Configuration block. This option is only valid when your overflow behavior is `QUEUE`. See [`scaling_configuration`](#scaling_configuration) below.
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `vpc_config` - (Optional) Configuration block. See [`vpc_config`](#vpc_config) below.

### compute_configuration

* `disk` - (Optional) Amount of disk space of the instance type included in the fleet.
* `machine_type` - (Optional) Machine type of the instance type included in the fleet. Valid values: `GENERAL`, `NVME`.
* `memory` - (Optional) Amount of memory of the instance type included in the fleet.
* `vcpu` - (Optional) Number of vCPUs of the instance type included in the fleet.

### scaling_configuration

* `max_capacity` - (Optional) Maximum number of instances in the ﬂeet when auto-scaling.
* `scaling_type` - (Optional) Scaling type for a compute fleet. Valid value: `TARGET_TRACKING_SCALING`.
* `target_tracking_scaling_configs` - (Optional) Configuration block. Detailed below.

#### scaling_configuration: target_tracking_scaling_configs

* `metric_type` - (Optional) Metric type to determine auto-scaling. Valid value: `FLEET_UTILIZATION_RATE`.
* `target_value` - (Optional) Value of metricType when to start scaling.

### vpc_config

* `security_group_ids` - (Required) A list of one or more security groups IDs in your Amazon VPC.
* `subnets` - (Required) A list of one or more subnet IDs in your Amazon VPC.
* `vpc_id` - (Required) The ID of the Amazon VPC.

### proxy_configuration

* `default_behavior` - (Required) The default behavior of the proxy. Valid values: `ALLOW_ALL`, `DENY_ALL`.
* `ordered_proxy_rules` - (Optional) Configuration block for a list of proxy rules. Detailed below.

#### proxy_configuration: ordered_proxy_rules

* `effect` - (Required) The effect of the proxy rule. Valid values: `ALLOW`, `DENY`.
* `entities` - (Required) A list of entities to apply the proxy rule to. For `DOMAIN` type, enter fully qualified domain names such as `example.com` or patterns such as `*.example.org`. For `IP` type, enter IP addresses or CIDR blocks such as `192.168.1.0/24`.
* `type` - (Required) The type of the proxy rule. Valid values: `DOMAIN`, `IP`.

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
