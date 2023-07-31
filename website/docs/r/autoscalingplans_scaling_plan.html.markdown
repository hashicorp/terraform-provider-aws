---
subcategory: "Auto Scaling Plans"
layout: "aws"
page_title: "AWS: aws_autoscalingplans_scaling_plan"
description: |-
  Manages an AWS Auto Scaling scaling plan.
---

# Resource: aws_autoscalingplans_scaling_plan

Manages an AWS Auto Scaling scaling plan.
More information can be found in the [AWS Auto Scaling User Guide](https://docs.aws.amazon.com/autoscaling/plans/userguide/what-is-aws-auto-scaling.html).

~> **NOTE:** The AWS Auto Scaling service uses an AWS IAM service-linked role to manage predictive scaling of Amazon EC2 Auto Scaling groups. The service attempts to automatically create this role the first time a scaling plan with predictive scaling enabled is created.
An [`aws_iam_service_linked_role`](/docs/providers/aws/r/iam_service_linked_role.html) resource can be used to manually manage this role.
See the [AWS documentation](https://docs.aws.amazon.com/autoscaling/plans/userguide/aws-auto-scaling-service-linked-roles.html#create-service-linked-role-manual) for more details.

## Example Usage

### Basic Dynamic Scaling

```terraform
data "aws_availability_zones" "available" {}

resource "aws_autoscaling_group" "example" {
  name_prefix = "example"

  launch_configuration = aws_launch_configuration.example.name
  availability_zones   = [data.aws_availability_zones.available.names[0]]

  min_size = 0
  max_size = 3

  tags = [
    {
      key                 = "application"
      value               = "example"
      propagate_at_launch = true
    },
  ]
}

resource "aws_autoscalingplans_scaling_plan" "example" {
  name = "example-dynamic-cost-optimization"

  application_source {
    tag_filter {
      key    = "application"
      values = ["example"]
    }
  }

  scaling_instruction {
    max_capacity       = 3
    min_capacity       = 0
    resource_id        = format("autoScalingGroup/%s", aws_autoscaling_group.example.name)
    scalable_dimension = "autoscaling:autoScalingGroup:DesiredCapacity"
    service_namespace  = "autoscaling"

    target_tracking_configuration {
      predefined_scaling_metric_specification {
        predefined_scaling_metric_type = "ASGAverageCPUUtilization"
      }

      target_value = 70
    }
  }
}
```

### Basic Predictive Scaling

```terraform
data "aws_availability_zones" "available" {}

resource "aws_autoscaling_group" "example" {
  name_prefix = "example"

  launch_configuration = aws_launch_configuration.example.name
  availability_zones   = [data.aws_availability_zones.available.names[0]]

  min_size = 0
  max_size = 3

  tags = [
    {
      key                 = "application"
      value               = "example"
      propagate_at_launch = true
    },
  ]
}

resource "aws_autoscalingplans_scaling_plan" "example" {
  name = "example-predictive-cost-optimization"

  application_source {
    tag_filter {
      key    = "application"
      values = ["example"]
    }
  }

  scaling_instruction {
    disable_dynamic_scaling = true

    max_capacity       = 3
    min_capacity       = 0
    resource_id        = format("autoScalingGroup/%s", aws_autoscaling_group.example.name)
    scalable_dimension = "autoscaling:autoScalingGroup:DesiredCapacity"
    service_namespace  = "autoscaling"

    target_tracking_configuration {
      predefined_scaling_metric_specification {
        predefined_scaling_metric_type = "ASGAverageCPUUtilization"
      }

      target_value = 70
    }

    predictive_scaling_max_capacity_behavior = "SetForecastCapacityToMaxCapacity"
    predictive_scaling_mode                  = "ForecastAndScale"

    predefined_load_metric_specification {
      predefined_load_metric_type = "ASGTotalCPUUtilization"
    }
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `name` - (Required) Name of the scaling plan. Names cannot contain vertical bars, colons, or forward slashes.
* `application_source` - (Required) CloudFormation stack or set of tags. You can create one scaling plan per application source.
* `scaling_instruction` - (Required) Scaling instructions. More details can be found in the [AWS Auto Scaling API Reference](https://docs.aws.amazon.com/autoscaling/plans/APIReference/API_ScalingInstruction.html).

The `application_source` object supports the following:

* `cloudformation_stack_arn` - (Optional) ARN of a AWS CloudFormation stack.
* `tag_filter` - (Optional) Set of tags.

The `tag_filter` object supports the following:

* `key` - (Required) Tag key.
* `values` - (Optional) Tag values.

The `scaling_instruction` object supports the following:

* `max_capacity` - (Required) Maximum capacity of the resource. The exception to this upper limit is if you specify a non-default setting for `predictive_scaling_max_capacity_behavior`.
* `min_capacity` - (Required) Minimum capacity of the resource.
* `resource_id` - (Required) ID of the resource. This string consists of the resource type and unique identifier.
* `scalable_dimension` - (Required) Scalable dimension associated with the resource. Valid values: `autoscaling:autoScalingGroup:DesiredCapacity`, `dynamodb:index:ReadCapacityUnits`, `dynamodb:index:WriteCapacityUnits`, `dynamodb:table:ReadCapacityUnits`, `dynamodb:table:WriteCapacityUnits`, `ecs:service:DesiredCount`, `ec2:spot-fleet-request:TargetCapacity`, `rds:cluster:ReadReplicaCount`.
* `service_namespace` - (Required) Namespace of the AWS service. Valid values: `autoscaling`, `dynamodb`, `ecs`, `ec2`, `rds`.
* `target_tracking_configuration` - (Required) Structure that defines new target tracking configurations. Each of these structures includes a specific scaling metric and a target value for the metric, along with various parameters to use with dynamic scaling.
More details can be found in the [AWS Auto Scaling API Reference](https://docs.aws.amazon.com/autoscaling/plans/APIReference/API_TargetTrackingConfiguration.html).
* `customized_load_metric_specification` - (Optional) Customized load metric to use for predictive scaling. You must specify either `customized_load_metric_specification` or `predefined_load_metric_specification` when configuring predictive scaling.
More details can be found in the [AWS Auto Scaling API Reference](https://docs.aws.amazon.com/autoscaling/plans/APIReference/API_CustomizedLoadMetricSpecification.html).
* `disable_dynamic_scaling` - (Optional) Boolean controlling whether dynamic scaling by AWS Auto Scaling is disabled. Defaults to `false`.
* `predefined_load_metric_specification` - (Optional) Predefined load metric to use for predictive scaling. You must specify either `predefined_load_metric_specification` or `customized_load_metric_specification` when configuring predictive scaling.
More details can be found in the [AWS Auto Scaling API Reference](https://docs.aws.amazon.com/autoscaling/plans/APIReference/API_PredefinedLoadMetricSpecification.html).
* `predictive_scaling_max_capacity_behavior`- (Optional) Defines the behavior that should be applied if the forecast capacity approaches or exceeds the maximum capacity specified for the resource.
Valid values: `SetForecastCapacityToMaxCapacity`, `SetMaxCapacityAboveForecastCapacity`, `SetMaxCapacityToForecastCapacity`.
* `predictive_scaling_max_capacity_buffer` - (Optional) Size of the capacity buffer to use when the forecast capacity is close to or exceeds the maximum capacity.
* `predictive_scaling_mode` - (Optional) Predictive scaling mode. Valid values: `ForecastAndScale`, `ForecastOnly`.
* `scaling_policy_update_behavior` - (Optional) Controls whether a resource's externally created scaling policies are kept or replaced. Valid values: `KeepExternalPolicies`, `ReplaceExternalPolicies`. Defaults to `KeepExternalPolicies`.
* `scheduled_action_buffer_time` - (Optional) Amount of time, in seconds, to buffer the run time of scheduled scaling actions when scaling out.

The `customized_load_metric_specification` object supports the following:

* `metric_name` - (Required) Name of the metric.
* `namespace` - (Required) Namespace of the metric.
* `statistic` - (Required) Statistic of the metric. Currently, the value must always be `Sum`.
* `dimensions` - (Optional) Dimensions of the metric.
* `unit` - (Optional) Unit of the metric.

The `predefined_load_metric_specification` object supports the following:

* `predefined_load_metric_type` - (Required) Metric type. Valid values: `ALBTargetGroupRequestCount`, `ASGTotalCPUUtilization`, `ASGTotalNetworkIn`, `ASGTotalNetworkOut`.
* `resource_label` - (Optional) Identifies the resource associated with the metric type.

The `target_tracking_configuration` object supports the following:

* `target_value` - (Required) Target value for the metric.
* `customized_scaling_metric_specification` - (Optional) Customized metric. You can specify either `customized_scaling_metric_specification` or `predefined_scaling_metric_specification`.
More details can be found in the [AWS Auto Scaling API Reference](https://docs.aws.amazon.com/autoscaling/plans/APIReference/API_CustomizedScalingMetricSpecification.html).
* `disable_scale_in` - (Optional) Boolean indicating whether scale in by the target tracking scaling policy is disabled. Defaults to `false`.
* `predefined_scaling_metric_specification` - (Optional) Predefined metric. You can specify either `predefined_scaling_metric_specification` or `customized_scaling_metric_specification`.
More details can be found in the [AWS Auto Scaling API Reference](https://docs.aws.amazon.com/autoscaling/plans/APIReference/API_PredefinedScalingMetricSpecification.html).
* `estimated_instance_warmup` - (Optional) Estimated time, in seconds, until a newly launched instance can contribute to the CloudWatch metrics.
This value is used only if the resource is an Auto Scaling group.
* `scale_in_cooldown` - (Optional) Amount of time, in seconds, after a scale in activity completes before another scale in activity can start.
This value is not used if the scalable resource is an Auto Scaling group.
* `scale_out_cooldown` - (Optional) Amount of time, in seconds, after a scale-out activity completes before another scale-out activity can start.
This value is not used if the scalable resource is an Auto Scaling group.

The `customized_scaling_metric_specification` object supports the following:

* `metric_name` - (Required) Name of the metric.
* `namespace` - (Required) Namespace of the metric.
* `statistic` - (Required) Statistic of the metric. Valid values: `Average`, `Maximum`, `Minimum`, `SampleCount`, `Sum`.
* `dimensions` - (Optional) Dimensions of the metric.
* `unit` - (Optional) Unit of the metric.

The `predefined_scaling_metric_specification` object supports the following:

* `predefined_scaling_metric_type` - (Required) Metric type. Valid values: `ALBRequestCountPerTarget`, `ASGAverageCPUUtilization`, `ASGAverageNetworkIn`, `ASGAverageNetworkOut`, `DynamoDBReadCapacityUtilization`, `DynamoDBWriteCapacityUtilization`, `ECSServiceAverageCPUUtilization`, `ECSServiceAverageMemoryUtilization`, `EC2SpotFleetRequestAverageCPUUtilization`, `EC2SpotFleetRequestAverageNetworkIn`, `EC2SpotFleetRequestAverageNetworkOut`, `RDSReaderAverageCPUUtilization`, `RDSReaderAverageDatabaseConnections`.
* `resource_label` - (Optional) Identifies the resource associated with the metric type.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Scaling plan identifier.
* `scaling_plan_version` - The version number of the scaling plan. This value is always 1.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Auto Scaling scaling plans using the `name`. For example:

```terraform
import {
  to = aws_autoscalingplans_scaling_plan.example
  id = "MyScale1"
}
```

Using `terraform import`, import Auto Scaling scaling plans using the `name`. For example:

```console
% terraform import aws_autoscalingplans_scaling_plan.example MyScale1
```
