---
subcategory: "Auto Scaling"
layout: "aws"
page_title: "AWS: aws_autoscaling_policy"
description: |-
  Provides an AutoScaling Scaling Group resource.
---

# Resource: aws_autoscaling_policy

Provides an AutoScaling Scaling Policy resource.

~> **NOTE:** You may want to omit `desired_capacity` attribute from attached `aws_autoscaling_group`
when using autoscaling policies. It's good practice to pick either
[manual](https://docs.aws.amazon.com/AutoScaling/latest/DeveloperGuide/as-manual-scaling.html)
or [dynamic](https://docs.aws.amazon.com/AutoScaling/latest/DeveloperGuide/as-scale-based-on-demand.html)
(policy-based) scaling.

> **Hands-on:** Try the [Manage AWS Auto Scaling Groups](https://learn.hashicorp.com/tutorials/terraform/aws-asg?utm_source=WEBSITE&utm_medium=WEB_IO&utm_offer=ARTICLE_PAGE&utm_content=DOCS) tutorial on HashiCorp Learn.

## Example Usage

```terraform
resource "aws_autoscaling_policy" "bat" {
  name                   = "foobar3-terraform-test"
  scaling_adjustment     = 4
  adjustment_type        = "ChangeInCapacity"
  cooldown               = 300
  autoscaling_group_name = aws_autoscaling_group.bar.name
}

resource "aws_autoscaling_group" "bar" {
  availability_zones        = ["us-east-1a"]
  name                      = "foobar3-terraform-test"
  max_size                  = 5
  min_size                  = 2
  health_check_grace_period = 300
  health_check_type         = "ELB"
  force_delete              = true
  launch_configuration      = aws_launch_configuration.foo.name
}
```

### Create target tracking scaling policy using metric math

```terraform
resource "aws_autoscaling_policy" "example" {
  autoscaling_group_name = "my-test-asg"
  name                   = "foo"
  policy_type            = "TargetTrackingScaling"
  target_tracking_configuration {
    target_value = 100
    customized_metric_specification {
      metrics {
        label = "Get the queue size (the number of messages waiting to be processed)"
        id    = "m1"
        metric_stat {
          metric {
            namespace   = "AWS/SQS"
            metric_name = "ApproximateNumberOfMessagesVisible"
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
        label = "Get the group size (the number of InService instances)"
        id    = "m2"
        metric_stat {
          metric {
            namespace   = "AWS/AutoScaling"
            metric_name = "GroupInServiceInstances"
            dimensions {
              name  = "AutoScalingGroupName"
              value = "my-asg"
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

### Create predictive scaling policy using customized metrics

```terraform
resource "aws_autoscaling_policy" "example" {
  autoscaling_group_name = "my-test-asg"
  name                   = "foo"
  policy_type            = "PredictiveScaling"
  predictive_scaling_configuration {
    metric_specification {
      target_value = 10
      customized_load_metric_specification {
        metric_data_queries {
          id         = "load_sum"
          expression = "SUM(SEARCH('{AWS/EC2,AutoScalingGroupName} MetricName=\"CPUUtilization\" my-test-asg', 'Sum', 3600))"
        }
      }
      customized_capacity_metric_specification {
        metric_data_queries {
          id         = "capacity_sum"
          expression = "SUM(SEARCH('{AWS/AutoScaling,AutoScalingGroupName} MetricName=\"GroupInServiceIntances\" my-test-asg', 'Average', 300))"
        }
      }
      customized_scaling_metric_specification {
        metric_data_queries {
          id          = "capacity_sum"
          expression  = "SUM(SEARCH('{AWS/AutoScaling,AutoScalingGroupName} MetricName=\"GroupInServiceIntances\" my-test-asg', 'Average', 300))"
          return_data = false
        }
        metric_data_queries {
          id          = "load_sum"
          expression  = "SUM(SEARCH('{AWS/EC2,AutoScalingGroupName} MetricName=\"CPUUtilization\" my-test-asg', 'Sum', 300))"
          return_data = false
        }
        metric_data_queries {
          id         = "weighted_average"
          expression = "load_sum / (capacity_sum * PERIOD(capacity_sum) / 60)"
        }
      }
    }
  }
}
```

### Create predictive scaling policy using customized scaling and predefined load metric

```terraform
resource "aws_autoscaling_policy" "example" {
  autoscaling_group_name = "my-test-asg"
  name                   = "foo"
  policy_type            = "PredictiveScaling"
  predictive_scaling_configuration {
    metric_specification {
      target_value = 10
      predefined_load_metric_specification {
        predefined_metric_type = "ASGTotalCPUUtilization"
        resource_label         = "app/my-alb/778d41231b141a0f/targetgroup/my-alb-target-group/943f017f100becff"
      }
      customized_scaling_metric_specification {
        metric_data_queries {
          id = "scaling"
          metric_stat {
            metric {
              metric_name = "CPUUtilization"
              namespace   = "AWS/EC2"
              dimensions {
                name  = "AutoScalingGroupName"
                value = "my-test-asg"
              }
            }
            stat = "Average"
          }
        }
      }
    }
  }
}
```

## Argument Reference

* `name` - (Required) Name of the policy.
* `autoscaling_group_name` - (Required) Name of the autoscaling group.
* `adjustment_type` - (Optional) Whether the adjustment is an absolute number or a percentage of the current capacity. Valid values are `ChangeInCapacity`, `ExactCapacity`, and `PercentChangeInCapacity`.
* `policy_type` - (Optional) Policy type, either "SimpleScaling", "StepScaling", "TargetTrackingScaling", or "PredictiveScaling". If this value isn't provided, AWS will default to "SimpleScaling."
* `predictive_scaling_configuration` - (Optional) Predictive scaling policy configuration to use with Amazon EC2 Auto Scaling.
* `estimated_instance_warmup` - (Optional) Estimated time, in seconds, until a newly launched instance will contribute CloudWatch metrics. Without a value, AWS will default to the group's specified cooldown period.
* `enabled` - (Optional) Whether the scaling policy is enabled or disabled. Default: `true`.

The following argument is only available to "SimpleScaling" and "StepScaling" type policies:

* `min_adjustment_magnitude` - (Optional) Minimum value to scale by when `adjustment_type` is set to `PercentChangeInCapacity`.

The following arguments are only available to "SimpleScaling" type policies:

* `cooldown` - (Optional) Amount of time, in seconds, after a scaling activity completes and before the next scaling activity can start.
* `scaling_adjustment` - (Optional) Number of instances by which to scale. `adjustment_type` determines the interpretation of this number (e.g., as an absolute number or as a percentage of the existing Auto Scaling group size). A positive increment adds to the current capacity and a negative value removes from the current capacity.

The following arguments are only available to "StepScaling" type policies:

* `metric_aggregation_type` - (Optional) Aggregation type for the policy's metrics. Valid values are "Minimum", "Maximum", and "Average". Without a value, AWS will treat the aggregation type as "Average".
* `step_adjustment` - (Optional) Set of adjustments that manage
group scaling. These have the following structure:

```terraform
resource "aws_autoscaling_policy" "example" {
  # ... other configuration ...

  step_adjustment {
    scaling_adjustment          = -1
    metric_interval_lower_bound = 1.0
    metric_interval_upper_bound = 2.0
  }

  step_adjustment {
    scaling_adjustment          = 1
    metric_interval_lower_bound = 2.0
    metric_interval_upper_bound = 3.0
  }
}
```

The following fields are available in step adjustments:

* `scaling_adjustment` - (Required) Number of members by which to
scale, when the adjustment bounds are breached. A positive value scales
up. A negative value scales down.
* `metric_interval_lower_bound` - (Optional) Lower bound for the
difference between the alarm threshold and the CloudWatch metric.
Without a value, AWS will treat this bound as negative infinity.
* `metric_interval_upper_bound` - (Optional) Upper bound for the
difference between the alarm threshold and the CloudWatch metric.
Without a value, AWS will treat this bound as positive infinity. The upper bound
must be greater than the lower bound.

Notice the bounds are **relative** to the alarm threshold, meaning that the starting point is not 0%, but the alarm threshold. Check the official [docs](https://docs.aws.amazon.com/autoscaling/ec2/userguide/as-scaling-simple-step.html#as-scaling-steps) for a detailed example.

The following arguments are only available to "TargetTrackingScaling" type policies:

* `target_tracking_configuration` - (Optional) Target tracking policy. These have the following structure:

```terraform
resource "aws_autoscaling_policy" "example" {
  # ... other configuration ...

  target_tracking_configuration {
    predefined_metric_specification {
      predefined_metric_type = "ASGAverageCPUUtilization"
    }

    target_value = 40.0
  }
}
```

The following fields are available in target tracking configuration:

* `predefined_metric_specification` - (Optional) Predefined metric. Conflicts with `customized_metric_specification`.
* `customized_metric_specification` - (Optional) Customized metric. Conflicts with `predefined_metric_specification`.
* `target_value` - (Required) Target value for the metric.
* `disable_scale_in` - (Optional, Default: false) Whether scale in by the target tracking policy is disabled.

### predefined_metric_specification

This configuration block supports the following arguments:

* `predefined_metric_type` - (Required) Metric type.
* `resource_label` - (Optional) Identifies the resource associated with the metric type.

### customized_metric_specification

This configuration block supports the following arguments:

* `metric_dimension` - (Optional) Dimensions of the metric.
* `metric_name` - (Optional) Name of the metric.
* `namespace` - (Optional) Namespace of the metric.
* `statistic` - (Optional) Statistic of the metric.
* `unit` - (Optional) Unit of the metric.
* `metrics` - (Optional) Metrics to include, as a metric data query.

#### metric_dimension

This configuration block supports the following arguments:

* `name` - (Required) Name of the dimension.
* `value` - (Required) Value of the dimension.

#### metrics

This configuration block supports the following arguments:

* `expression` - (Optional) Math expression used on the returned metric. You must specify either `expression` or `metric_stat`, but not both.
* `id` - (Required) Short name for the metric used in target tracking scaling policy.
* `label` - (Optional) Human-readable label for this metric or expression.
* `metric_stat` - (Optional) Structure that defines CloudWatch metric to be used in target tracking scaling policy. You must specify either `expression` or `metric_stat`, but not both.
* `return_data` - (Optional) Boolean that indicates whether to return the timestamps and raw data values of this metric, the default is true

##### metric_stat

This configuration block supports the following arguments:

* `metric` - (Required) Structure that defines the CloudWatch metric to return, including the metric name, namespace, and dimensions.
* `stat` - (Required) Statistic of the metrics to return.
* `unit` - (Optional) Unit of the metrics to return.

##### metric

This configuration block supports the following arguments:

* `dimensions` - (Optional) Dimensions of the metric.
* `metric_name` - (Required) Name of the metric.
* `namespace` - (Required) Namespace of the metric.

###### dimensions

This configuration block supports the following arguments:

* `name` - (Required) Name of the dimension.
* `value` - (Required) Value of the dimension.

### predictive_scaling_configuration

This configuration block supports the following arguments:

* `max_capacity_breach_behavior` - (Optional) Defines the behavior that should be applied if the forecast capacity approaches or exceeds the maximum capacity of the Auto Scaling group. Valid values are `HonorMaxCapacity` or `IncreaseMaxCapacity`. Default is `HonorMaxCapacity`.
* `max_capacity_buffer` - (Optional) Size of the capacity buffer to use when the forecast capacity is close to or exceeds the maximum capacity. Valid range is `0` to `100`. If set to `0`, Amazon EC2 Auto Scaling may scale capacity higher than the maximum capacity to equal but not exceed forecast capacity.
* `metric_specification` - (Required) This structure includes the metrics and target utilization to use for predictive scaling.
* `mode` - (Optional) Predictive scaling mode. Valid values are `ForecastAndScale` and `ForecastOnly`. Default is `ForecastOnly`.
* `scheduling_buffer_time` - (Optional) Amount of time, in seconds, by which the instance launch time can be advanced. Minimum is `0`.

#### metric_specification

This configuration block supports the following arguments:

* `customized_capacity_metric_specification` - (Optional) Customized capacity metric specification. The field is only valid when you use `customized_load_metric_specification`
* `customized_load_metric_specification` - (Optional) Customized load metric specification.
* `customized_scaling_metric_specification` - (Optional) Customized scaling metric specification.
* `predefined_load_metric_specification` - (Optional) Predefined load metric specification.
* `predefined_metric_pair_specification` - (Optional) Metric pair specification from which Amazon EC2 Auto Scaling determines the appropriate scaling metric and load metric to use.
* `predefined_scaling_metric_specification` - (Optional) Predefined scaling metric specification.

##### predefined_load_metric_specification

This configuration block supports the following arguments:

* `predefined_metric_type` - (Required) Metric type. Valid values are `ASGTotalCPUUtilization`, `ASGTotalNetworkIn`, `ASGTotalNetworkOut`, or `ALBTargetGroupRequestCount`.
* `resource_label` - (Required) Label that uniquely identifies a specific Application Load Balancer target group from which to determine the request count served by your Auto Scaling group. You create the resource label by appending the final portion of the load balancer ARN and the final portion of the target group ARN into a single value, separated by a forward slash (/). Refer to [PredefinedMetricSpecification](https://docs.aws.amazon.com/autoscaling/ec2/APIReference/API_PredefinedMetricSpecification.html) for more information.

##### predefined_metric_pair_specification

This configuration block supports the following arguments:

* `predefined_metric_type` - (Required) Which metrics to use. There are two different types of metrics for each metric type: one is a load metric and one is a scaling metric. For example, if the metric type is `ASGCPUUtilization`, the Auto Scaling group's total CPU metric is used as the load metric, and the average CPU metric is used for the scaling metric. Valid values are `ASGCPUUtilization`, `ASGNetworkIn`, `ASGNetworkOut`, or `ALBRequestCount`.
* `resource_label` - (Required) Label that uniquely identifies a specific Application Load Balancer target group from which to determine the request count served by your Auto Scaling group. You create the resource label by appending the final portion of the load balancer ARN and the final portion of the target group ARN into a single value, separated by a forward slash (/). Refer to [PredefinedMetricSpecification](https://docs.aws.amazon.com/autoscaling/ec2/APIReference/API_PredefinedMetricSpecification.html) for more information.

##### predefined_scaling_metric_specification

This configuration block supports the following arguments:

* `predefined_metric_type` - (Required) Describes a scaling metric for a predictive scaling policy. Valid values are `ASGAverageCPUUtilization`, `ASGAverageNetworkIn`, `ASGAverageNetworkOut`, or `ALBRequestCountPerTarget`.
* `resource_label` - (Required) Label that uniquely identifies a specific Application Load Balancer target group from which to determine the request count served by your Auto Scaling group. You create the resource label by appending the final portion of the load balancer ARN and the final portion of the target group ARN into a single value, separated by a forward slash (/). Refer to [PredefinedMetricSpecification](https://docs.aws.amazon.com/autoscaling/ec2/APIReference/API_PredefinedMetricSpecification.html) for more information.

##### customized_scaling_metric_specification

This configuration block supports the following arguments:

* `metric_data_queries` - (Required) List of up to 10 structures that defines custom scaling metric in predictive scaling policy

##### customized_load_metric_specification

This configuration block supports the following arguments:

* `metric_data_queries` - (Required) List of up to 10 structures that defines custom load metric in predictive scaling policy

##### customized_capacity_metric_specification

This configuration block supports the following arguments:

* `metric_data_queries` - (Required) List of up to 10 structures that defines custom capacity metric in predictive scaling policy

##### metric_data_queries

This configuration block supports the following arguments:

* `expression` - (Optional) Math expression used on the returned metric. You must specify either `expression` or `metric_stat`, but not both.
* `id` - (Required) Short name for the metric used in predictive scaling policy.
* `label` - (Optional) Human-readable label for this metric or expression.
* `metric_stat` - (Optional) Structure that defines CloudWatch metric to be used in predictive scaling policy. You must specify either `expression` or `metric_stat`, but not both.
* `return_data` - (Optional) Boolean that indicates whether to return the timestamps and raw data values of this metric, the default is true

##### metric_stat

This configuration block supports the following arguments:

* `metric` - (Required) Structure that defines the CloudWatch metric to return, including the metric name, namespace, and dimensions.
* `stat` - (Required) Statistic of the metrics to return.
* `unit` - (Optional) Unit of the metrics to return.

##### metric

This configuration block supports the following arguments:

* `dimensions` - (Optional) Dimensions of the metric.
* `metric_name` - (Required) Name of the metric.
* `namespace` - (Required) Namespace of the metric.

##### dimensions

This configuration block supports the following arguments:

* `name` - (Required) Name of the dimension.
* `value` - (Required) Value of the dimension.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN assigned by AWS to the scaling policy.
* `name` - Scaling policy's name.
* `autoscaling_group_name` - The scaling policy's assigned autoscaling group.
* `adjustment_type` - Scaling policy's adjustment type.
* `policy_type` - Scaling policy's type.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import AutoScaling scaling policy using the role autoscaling_group_name and name separated by `/`. For example:

```terraform
import {
  to = aws_autoscaling_policy.test-policy
  id = "asg-name/policy-name"
}
```

Using `terraform import`, import AutoScaling scaling policy using the role autoscaling_group_name and name separated by `/`. For example:

```console
% terraform import aws_autoscaling_policy.test-policy asg-name/policy-name
```
