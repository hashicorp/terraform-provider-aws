---
subcategory: "Application Auto Scaling"
layout: "aws"
page_title: "AWS: aws_appautoscaling_policy"
description: |-
  Provides an Application AutoScaling Policy resource.
---

# Resource: aws_appautoscaling_policy

Provides an Application AutoScaling Policy resource.

## Example Usage

### DynamoDB Table Autoscaling

```terraform
resource "aws_appautoscaling_target" "dynamodb_table_read_target" {
  max_capacity       = 100
  min_capacity       = 5
  resource_id        = "table/tableName"
  scalable_dimension = "dynamodb:table:ReadCapacityUnits"
  service_namespace  = "dynamodb"
}

resource "aws_appautoscaling_policy" "dynamodb_table_read_policy" {
  name               = "DynamoDBReadCapacityUtilization:${aws_appautoscaling_target.dynamodb_table_read_target.resource_id}"
  policy_type        = "TargetTrackingScaling"
  resource_id        = aws_appautoscaling_target.dynamodb_table_read_target.resource_id
  scalable_dimension = aws_appautoscaling_target.dynamodb_table_read_target.scalable_dimension
  service_namespace  = aws_appautoscaling_target.dynamodb_table_read_target.service_namespace

  target_tracking_scaling_policy_configuration {
    predefined_metric_specification {
      predefined_metric_type = "DynamoDBReadCapacityUtilization"
    }

    target_value = 70
  }
}
```

### ECS Service Autoscaling

```terraform
resource "aws_appautoscaling_target" "ecs_target" {
  max_capacity       = 4
  min_capacity       = 1
  resource_id        = "service/clusterName/serviceName"
  scalable_dimension = "ecs:service:DesiredCount"
  service_namespace  = "ecs"
}

resource "aws_appautoscaling_policy" "ecs_policy" {
  name               = "scale-down"
  policy_type        = "StepScaling"
  resource_id        = aws_appautoscaling_target.ecs_target.resource_id
  scalable_dimension = aws_appautoscaling_target.ecs_target.scalable_dimension
  service_namespace  = aws_appautoscaling_target.ecs_target.service_namespace

  step_scaling_policy_configuration {
    adjustment_type         = "ChangeInCapacity"
    cooldown                = 60
    metric_aggregation_type = "Maximum"

    step_adjustment {
      metric_interval_upper_bound = 0
      scaling_adjustment          = -1
    }
  }
}
```

### Preserve desired count when updating an autoscaled ECS Service

```terraform
resource "aws_ecs_service" "ecs_service" {
  name            = "serviceName"
  cluster         = "clusterName"
  task_definition = "taskDefinitionFamily:1"
  desired_count   = 2

  lifecycle {
    ignore_changes = [desired_count]
  }
}
```

### Aurora Read Replica Autoscaling

```terraform
resource "aws_appautoscaling_target" "replicas" {
  service_namespace  = "rds"
  scalable_dimension = "rds:cluster:ReadReplicaCount"
  resource_id        = "cluster:${aws_rds_cluster.example.id}"
  min_capacity       = 1
  max_capacity       = 15
}

resource "aws_appautoscaling_policy" "replicas" {
  name               = "cpu-auto-scaling"
  service_namespace  = aws_appautoscaling_target.replicas.service_namespace
  scalable_dimension = aws_appautoscaling_target.replicas.scalable_dimension
  resource_id        = aws_appautoscaling_target.replicas.resource_id
  policy_type        = "TargetTrackingScaling"

  target_tracking_scaling_policy_configuration {
    predefined_metric_specification {
      predefined_metric_type = "RDSReaderAverageCPUUtilization"
    }

    target_value       = 75
    scale_in_cooldown  = 300
    scale_out_cooldown = 300
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the policy. Must be between 1 and 255 characters in length.
* `policy_type` - (Optional) Policy type. Valid values are `StepScaling` and `TargetTrackingScaling`. Defaults to `StepScaling`. Certain services only support only one policy type. For more information see the [Target Tracking Scaling Policies](https://docs.aws.amazon.com/autoscaling/application/userguide/application-auto-scaling-target-tracking.html) and [Step Scaling Policies](https://docs.aws.amazon.com/autoscaling/application/userguide/application-auto-scaling-step-scaling-policies.html) documentation.
* `resource_id` - (Required) Resource type and unique identifier string for the resource associated with the scaling policy. Documentation can be found in the `ResourceId` parameter at: [AWS Application Auto Scaling API Reference](https://docs.aws.amazon.com/autoscaling/application/APIReference/API_RegisterScalableTarget.html)
* `scalable_dimension` - (Required) Scalable dimension of the scalable target. Documentation can be found in the `ScalableDimension` parameter at: [AWS Application Auto Scaling API Reference](https://docs.aws.amazon.com/autoscaling/application/APIReference/API_RegisterScalableTarget.html)
* `service_namespace` - (Required) AWS service namespace of the scalable target. Documentation can be found in the `ServiceNamespace` parameter at: [AWS Application Auto Scaling API Reference](https://docs.aws.amazon.com/autoscaling/application/APIReference/API_RegisterScalableTarget.html)
* `step_scaling_policy_configuration` - (Optional) Step scaling policy configuration, requires `policy_type = "StepScaling"` (default). See supported fields below.
* `target_tracking_scaling_policy_configuration` - (Optional) Target tracking policy, requires `policy_type = "TargetTrackingScaling"`. See supported fields below.

### step_scaling_policy_configuration

The `step_scaling_policy_configuration` configuration block supports the following arguments:

* `adjustment_type` - (Required) Whether the adjustment is an absolute number or a percentage of the current capacity. Valid values are `ChangeInCapacity`, `ExactCapacity`, and `PercentChangeInCapacity`.
* `cooldown` - (Required) Amount of time, in seconds, after a scaling activity completes and before the next scaling activity can start.
* `metric_aggregation_type` - (Optional) Aggregation type for the policy's metrics. Valid values are "Minimum", "Maximum", and "Average". Without a value, AWS will treat the aggregation type as "Average".
* `min_adjustment_magnitude` - (Optional) Minimum number to adjust your scalable dimension as a result of a scaling activity. If the adjustment type is PercentChangeInCapacity, the scaling policy changes the scalable dimension of the scalable target by this amount.
* `step_adjustment` - (Optional) Set of adjustments that manage scaling. These have the following structure:

 ```terraform
resource "aws_appautoscaling_policy" "ecs_policy" {
  # ...

  step_scaling_policy_configuration {
    # insert config here

    step_adjustment {
      metric_interval_lower_bound = 1.0
      metric_interval_upper_bound = 2.0
      scaling_adjustment          = -1
    }

    step_adjustment {
      metric_interval_lower_bound = 2.0
      metric_interval_upper_bound = 3.0
      scaling_adjustment          = 1
    }
  }
}
```

* `metric_interval_lower_bound` - (Optional) Lower bound for the difference between the alarm threshold and the CloudWatch metric. Without a value, AWS will treat this bound as negative infinity.
* `metric_interval_upper_bound` - (Optional) Upper bound for the difference between the alarm threshold and the CloudWatch metric. Without a value, AWS will treat this bound as infinity. The upper bound must be greater than the lower bound.
* `scaling_adjustment` - (Required) Number of members by which to scale, when the adjustment bounds are breached. A positive value scales up. A negative value scales down.

### target_tracking_scaling_policy_configuration

The `target_tracking_scaling_policy_configuration` configuration block supports the following arguments:

* `target_value` - (Required) Target value for the metric.
* `disable_scale_in` - (Optional) Whether scale in by the target tracking policy is disabled. If the value is true, scale in is disabled and the target tracking policy won't remove capacity from the scalable resource. Otherwise, scale in is enabled and the target tracking policy can remove capacity from the scalable resource. The default value is `false`.
* `scale_in_cooldown` - (Optional) Amount of time, in seconds, after a scale in activity completes before another scale in activity can start.
* `scale_out_cooldown` - (Optional) Amount of time, in seconds, after a scale out activity completes before another scale out activity can start.
* `customized_metric_specification` - (Optional) Custom CloudWatch metric. Documentation can be found  at: [AWS Customized Metric Specification](https://docs.aws.amazon.com/autoscaling/ec2/APIReference/API_CustomizedMetricSpecification.html). See supported fields below.
* `predefined_metric_specification` - (Optional) Predefined metric. See supported fields below.

### target_tracking_scaling_policy_configuration customized_metric_specification

Example usage:

```terraform
resource "aws_appautoscaling_policy" "example" {
  policy_type = "TargetTrackingScaling"

  # ... other configuration ...

  target_tracking_scaling_policy_configuration {
    target_value = 40

    # ... potentially other configuration ...

    customized_metric_specification {
      metric_name = "MyUtilizationMetric"
      namespace   = "MyNamespace"
      statistic   = "Average"
      unit        = "Percent"

      dimensions {
        name  = "MyOptionalMetricDimensionName"
        value = "MyOptionalMetricDimensionValue"
      }
    }
  }
}
```

The `target_tracking_scaling_policy_configuration` `customized_metric_specification` configuration block supports the following arguments:

* `dimensions` - (Optional) Configuration block(s) with the dimensions of the metric if the metric was published with dimensions. Detailed below.
* `metric_name` - (Required) Name of the metric.
* `namespace` - (Required) Namespace of the metric.
* `statistic` - (Required) Statistic of the metric. Valid values: `Average`, `Minimum`, `Maximum`, `SampleCount`, and `Sum`.
* `unit` - (Optional) Unit of the metric.

### target_tracking_scaling_policy_configuration customized_metric_specification dimensions

The `target_tracking_scaling_policy_configuration` `customized_metric_specification` `dimensions` configuration block supports the following arguments:

* `name` - (Required) Name of the dimension.
* `value` - (Required) Value of the dimension.

### target_tracking_scaling_policy_configuration predefined_metric_specification

The `target_tracking_scaling_policy_configuration` `predefined_metric_specification` configuration block supports the following arguments:

* `predefined_metric_type` - (Required) Metric type.
* `resource_label` - (Optional) Reserved for future use if the `predefined_metric_type` is not `ALBRequestCountPerTarget`. If the `predefined_metric_type` is `ALBRequestCountPerTarget`, you must specify this argument. Documentation can be found at: [AWS Predefined Scaling Metric Specification](https://docs.aws.amazon.com/autoscaling/plans/APIReference/API_PredefinedScalingMetricSpecification.html). Must be less than or equal to 1023 characters in length.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `alarm_arns` - List of CloudWatch alarm ARNs associated with the scaling policy.
* `arn` - ARN assigned by AWS to the scaling policy.
* `name` - Scaling policy's name.
* `policy_type` - Scaling policy's type.

## Import

Application AutoScaling Policy can be imported using the `service-namespace` , `resource-id`, `scalable-dimension` and `policy-name` separated by `/`.

```
$ terraform import aws_appautoscaling_policy.test-policy service-namespace/resource-id/scalable-dimension/policy-name
```
