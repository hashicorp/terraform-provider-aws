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

### Create target tracking scaling policy using metric math

```terraform
resource "aws_appautoscaling_target" "ecs_target" {
  max_capacity       = 4
  min_capacity       = 1
  resource_id        = "service/clusterName/serviceName"
  scalable_dimension = "ecs:service:DesiredCount"
  service_namespace  = "ecs"
}

resource "aws_appautoscaling_policy" "example" {
  name               = "foo"
  policy_type        = "TargetTrackingScaling"
  resource_id        = aws_appautoscaling_target.ecs_target.resource_id
  scalable_dimension = aws_appautoscaling_target.ecs_target.scalable_dimension
  service_namespace  = aws_appautoscaling_target.ecs_target.service_namespace

  target_tracking_scaling_policy_configuration {
    target_value = 100

    customized_metric_specification {
      metrics {
        label = "Get the queue size (the number of messages waiting to be processed)"
        id    = "m1"

        metric_stat {
          metric {
            metric_name = "ApproximateNumberOfMessagesVisible"
            namespace   = "AWS/SQS"

            dimensions {
              name  = "QueueName"
              value = "my-queue"
            }
          }

          stat = "Sum"
        }

        return_data = false
      }

      metrics {
        label = "Get the ECS running task count (the number of currently running tasks)"
        id    = "m2"

        metric_stat {
          metric {
            metric_name = "RunningTaskCount"
            namespace   = "ECS/ContainerInsights"

            dimensions {
              name  = "ClusterName"
              value = "default"
            }

            dimensions {
              name  = "ServiceName"
              value = "web-app"
            }
          }

          stat = "Average"
        }

        return_data = false
      }

      metrics {
        label       = "Calculate the backlog per instance"
        id          = "e1"
        expression  = "m1 / m2"
        return_data = true
      }
    }
  }
}
```

### Predictive Scaling

```terraform
resource "aws_appautoscaling_policy" "example" {
  name               = "example-policy"
  resource_id        = aws_appautoscaling_target.example.resource_id
  scalable_dimension = aws_appautoscaling_target.example.scalable_dimension
  service_namespace  = aws_appautoscaling_target.example.service_namespace
  policy_type        = "PredictiveScaling"

  predictive_scaling_policy_configuration {
    metric_specification {
      target_value = 40

      predefined_metric_pair_specification {
        predefined_metric_type = "ECSServiceMemoryUtilization"
      }
    }
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `name` - (Required) Name of the policy. Must be between 1 and 255 characters in length.
* `policy_type` - (Optional) Policy type. Valid values are `StepScaling`, `TargetTrackingScaling`, and `PredictiveScaling`. Defaults to `StepScaling`. Certain services only support only one policy type. For more information see the [Target Tracking Scaling Policies](https://docs.aws.amazon.com/autoscaling/application/userguide/application-auto-scaling-target-tracking.html), [Step Scaling Policies](https://docs.aws.amazon.com/autoscaling/application/userguide/application-auto-scaling-step-scaling-policies.html), and [Predictive Scaling](https://docs.aws.amazon.com/autoscaling/application/userguide/application-auto-scaling-predictive-scaling.html) documentation.
* `predictive_scaling_policy_configuration` - (Optional) Predictive scaling policy configuration, requires `policy_type = "PredictiveScaling"`. See supported fields below.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `resource_id` - (Required) Resource type and unique identifier string for the resource associated with the scaling policy. Documentation can be found in the `ResourceId` parameter at: [AWS Application Auto Scaling API Reference](https://docs.aws.amazon.com/autoscaling/application/APIReference/API_RegisterScalableTarget.html)
* `scalable_dimension` - (Required) Scalable dimension of the scalable target. Documentation can be found in the `ScalableDimension` parameter at: [AWS Application Auto Scaling API Reference](https://docs.aws.amazon.com/autoscaling/application/APIReference/API_RegisterScalableTarget.html)
* `service_namespace` - (Required) AWS service namespace of the scalable target. Documentation can be found in the `ServiceNamespace` parameter at: [AWS Application Auto Scaling API Reference](https://docs.aws.amazon.com/autoscaling/application/APIReference/API_RegisterScalableTarget.html)
* `step_scaling_policy_configuration` - (Optional) Step scaling policy configuration, requires `policy_type = "StepScaling"` (default). See supported fields below.
* `target_tracking_scaling_policy_configuration` - (Optional) Target tracking policy configuration, requires `policy_type = "TargetTrackingScaling"`. See supported fields below.

### predictive_scaling_policy_configuration

The `predictive_scaling_policy_configuration` configuration block supports the following arguments:

* `max_capacity_breach_behavior` - (Optional) The behavior that should be applied if the forecast capacity approaches or exceeds the maximum capacity. Valid values are `HonorMaxCapacity` and `IncreaseMaxCapacity`.
* `max_capacity_buffer` - (Optional) Size of the capacity buffer to use when the forecast capacity is close to or exceeds the maximum capacity. The value is specified as a percentage relative to the forecast capacity. Required if the `max_capacity_breach_behavior` argument is set to `IncreaseMaxCapacity`, and cannot be used otherwise.
* `metric_specification` - (Required) Metrics and target utilization to use for predictive scaling. See supported fields below.
* `mode` - (Optional) Predictive scaling mode. Valid values are `ForecastOnly` and `ForecastAndScale`.
* `scheduling_buffer_time` - (Optional) Amount of time, in seconds, that the start time can be advanced.

### predictive_scaling_policy_configuration metric_specification

The `predictive_scaling_policy_configuration` `metric_specification` configuration block supports the following arguments:

* `customized_capacity_metric_specification` - (Optional) Customized capacity metric specification. See supported fields below.
* `customized_load_metric_specification` - (Optional) Customized load metric specification. See supported fields below.
* `customized_scaling_metric_specification` - (Optional) Customized scaling metric specification. See supported fields below.
* `predefined_load_metric_specification` - (Optional) Predefined load metric specification. See supported fields below.
* `predefined_metric_pair_specification` - (Optional) Predefined metric pair specification that determines the appropriate scaling metric and load metric to use. See supported fields below.
* `predefined_scaling_metric_specification` - (Optional) Predefined scaling metric specification. See supported fields below.
* `target_value` - (Required) Target utilization.

### predictive_scaling_policy_configuration metric_specification customized_capacity_metric_specification, customized_load_metric_specification and customized_scaling_metric_specification

The `predictive_scaling_policy_configuration` `metric_specification` `customized_capacity_metric_specification`, `customized_load_metric_specification`, and `customized_scaling_metric_specification` configuration blocks supports the following arguments:

* `metric_data_query` - (Required) One or more metric data queries to provide data points for a metric specification. See supported fields below.

### predictive_scaling_policy_configuration metric_specification customized_capacity_metric_specification metric_data_query

The `predictive_scaling_policy_configuration` `metric_specification` `customized_capacity_metric_specification` `metric_data_query` configuration block supports the following arguments:

* `expression` - (Optional) Math expression to perform on the returned data, if this object is performing a math expression.
* `id` - (Required) Short name that identifies the object's results in the response.
* `label` - (Optional) Human-readable label for this metric or expression.
* `metric_stat` - (Optional) Information about the metric data to return. See supported fields below.
* `return_data` - (Optional) Whether to return the timestamps and raw data values of this metric.

### predictive_scaling_policy_configuration metric_specification customized_capacity_metric_specification metric_data_query metric_stat

The `predictive_scaling_policy_configuration` `metric_specification` `customized_capacity_metric_specification` `metric_data_query` `metric_stat` configuration block supports the following arguments:

* `metric` - (Required) CloudWatch metric to return, including the metric name, namespace, and dimensions. See supported fields below.
* `stat` - (Required) Statistic to return.
* `unit` - (Optional) Unit to use for the returned data points.

### predictive_scaling_policy_configuration metric_specification customized_capacity_metric_specification metric_data_query metric_stat metric

The `predictive_scaling_policy_configuration` `metric_specification` `customized_capacity_metric_specification` `metric_data_query` `metric_stat` `metric` configuration block supports the following arguments:

* `dimension` - (Optional) Dimensions of the metric. See supported fields below.
* `metric_name` - (Optional) Name of the metric.
* `namespace` - (Optional) Namespace of the metric.

### predictive_scaling_policy_configuration metric_specification customized_capacity_metric_specification metric_data_query metric_stat metric dimension

The `predictive_scaling_policy_configuration` `metric_specification` `customized_capacity_metric_specification` `metric_data_query` `metric_stat` `metric` `dimension` configuration block supports the following arguments:

* `name` - (Optional) Name of the dimension.
* `value` - (Optional) Value of the dimension.

### predictive_scaling_policy_configuration metric_specification predefined_load_metric_specification

The `predictive_scaling_policy_configuration` `metric_specification` `predefined_load_metric_specification` configuration block supports the following arguments:

* `predefined_metric_type` - (Required) Metric type.
* `resource_label` - (Optional) Label that uniquely identifies a target group.

### predictive_scaling_policy_configuration metric_specification predefined_metric_pair_specification

The `predictive_scaling_policy_configuration` `metric_specification` `predefined_metric_pair_specification` configuration block supports the following arguments:

* `predefined_metric_type` - (Required) Which metrics to use. There are two different types of metrics for each metric type: one is a load metric and one is a scaling metric.
* `resource_label` - (Optional) Label that uniquely identifies a specific target group from which to determine the total and average request count.

### predictive_scaling_policy_configuration metric_specification predefined_scaling_metric_specification

The `predictive_scaling_policy_configuration` `metric_specification` `predefined_scaling_metric_specification` configuration block supports the following arguments:

* `predefined_metric_type` - (Required) Metric type.
* `resource_label` - (Optional) Label that uniquely identifies a specific target group from which to determine the average request count.

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
* `metric_name` - (Optional) Name of the metric.
* `namespace` - (Optional) Namespace of the metric.
* `statistic` - (Optional) Statistic of the metric. Valid values: `Average`, `Minimum`, `Maximum`, `SampleCount`, and `Sum`.
* `unit` - (Optional) Unit of the metric.
* `metrics` - (Optional) Metrics to include, as a metric data query.

### target_tracking_scaling_policy_configuration customized_metric_specification dimensions

The `target_tracking_scaling_policy_configuration` `customized_metric_specification` `dimensions` configuration block supports the following arguments:

* `name` - (Required) Name of the dimension.
* `value` - (Required) Value of the dimension.

### target_tracking_scaling_policy_configuration customized_metric_specification metrics

The `target_tracking_scaling_policy_configuration` `customized_metric_specification` `metrics` configuration block supports the following arguments:

* `expression` - (Optional) Math expression used on the returned metric. You must specify either `expression` or `metric_stat`, but not both.
* `id` - (Required) Short name for the metric used in target tracking scaling policy.
* `label` - (Optional) Human-readable label for this metric or expression.
* `metric_stat` - (Optional) Structure that defines CloudWatch metric to be used in target tracking scaling policy. You must specify either `expression` or `metric_stat`, but not both.
* `return_data` - (Optional) Boolean that indicates whether to return the timestamps and raw data values of this metric, the default is true

### target_tracking_scaling_policy_configuration customized_metric_specification metrics metric_stat

The `target_tracking_scaling_policy_configuration` `customized_metric_specification` `metrics` `metric_stat` configuration block supports the following arguments:

* `metric` - (Required) Structure that defines the CloudWatch metric to return, including the metric name, namespace, and dimensions.
* `stat` - (Required) Statistic of the metrics to return.
* `unit` - (Optional) Unit of the metrics to return.

### target_tracking_scaling_policy_configuration customized_metric_specification metrics metric

The `target_tracking_scaling_policy_configuration` `customized_metric_specification` `metrics` `metric` configuration block supports the following arguments:

* `dimensions` - (Optional) Dimensions of the metric.
* `metric_name` - (Required) Name of the metric.
* `namespace` - (Required) Namespace of the metric.

### target_tracking_scaling_policy_configuration customized_metric_specification metrics dimensions

The `target_tracking_scaling_policy_configuration` `customized_metric_specification` `metrics` `dimensions` configuration block supports the following arguments:

* `name` - (Required) Name of the dimension.
* `value` - (Required) Value of the dimension.

### target_tracking_scaling_policy_configuration predefined_metric_specification

The `target_tracking_scaling_policy_configuration` `predefined_metric_specification` configuration block supports the following arguments:

* `predefined_metric_type` - (Required) Metric type.
* `resource_label` - (Optional) Reserved for future use if the `predefined_metric_type` is not `ALBRequestCountPerTarget`. If the `predefined_metric_type` is `ALBRequestCountPerTarget`, you must specify this argument. Documentation can be found at: [AWS Predefined Scaling Metric Specification](https://docs.aws.amazon.com/autoscaling/plans/APIReference/API_PredefinedScalingMetricSpecification.html). Must be less than or equal to 1023 characters in length.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `alarm_arns` - List of CloudWatch alarm ARNs associated with the scaling policy.
* `arn` - ARN assigned by AWS to the scaling policy.
* `name` - Scaling policy's name.
* `policy_type` - Scaling policy's type.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Application AutoScaling Policy using the `service-namespace` , `resource-id`, `scalable-dimension` and `policy-name` separated by `/`. For example:

```terraform
import {
  to = aws_appautoscaling_policy.test-policy
  id = "service-namespace/resource-id/scalable-dimension/policy-name"
}
```

Using `terraform import`, import Application AutoScaling Policy using the `service-namespace` , `resource-id`, `scalable-dimension` and `policy-name` separated by `/`. For example:

```console
% terraform import aws_appautoscaling_policy.test-policy service-namespace/resource-id/scalable-dimension/policy-name
```
