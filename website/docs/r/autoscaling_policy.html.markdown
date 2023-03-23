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
      customized_capacity_metric_specification {
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
          expression = "load_sum / capacity_sum"
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
        resource_label         = "testLabel"
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

* `name` - (Required) The name of the policy.
* `autoscaling_group_name` - (Required) The name of the autoscaling group.
* `adjustment_type` - (Optional) Specifies whether the adjustment is an absolute number or a percentage of the current capacity. Valid values are `ChangeInCapacity`, `ExactCapacity`, and `PercentChangeInCapacity`.
* `policy_type` - (Optional) The policy type, either "SimpleScaling", "StepScaling", "TargetTrackingScaling", or "PredictiveScaling". If this value isn't provided, AWS will default to "SimpleScaling."
* `predictive_scaling_configuration` - (Optional) The predictive scaling policy configuration to use with Amazon EC2 Auto Scaling.
* `estimated_instance_warmup` - (Optional) The estimated time, in seconds, until a newly launched instance will contribute CloudWatch metrics. Without a value, AWS will default to the group's specified cooldown period.
* `enabled` - (Optional) Indicates whether the scaling policy is enabled or disabled. Default: `true`.

The following argument is only available to "SimpleScaling" and "StepScaling" type policies:

* `min_adjustment_magnitude` - (Optional) Minimum value to scale by when `adjustment_type` is set to `PercentChangeInCapacity`.

The following arguments are only available to "SimpleScaling" type policies:

* `cooldown` - (Optional) The amount of time, in seconds, after a scaling activity completes and before the next scaling activity can start.
* `scaling_adjustment` - (Optional) The number of instances by which to scale. `adjustment_type` determines the interpretation of this number (e.g., as an absolute number or as a percentage of the existing Auto Scaling group size). A positive increment adds to the current capacity and a negative value removes from the current capacity.

The following arguments are only available to "StepScaling" type policies:

* `metric_aggregation_type` - (Optional) The aggregation type for the policy's metrics. Valid values are "Minimum", "Maximum", and "Average". Without a value, AWS will treat the aggregation type as "Average".
* `step_adjustment` - (Optional) A set of adjustments that manage
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

* `scaling_adjustment` - (Required) The number of members by which to
scale, when the adjustment bounds are breached. A positive value scales
up. A negative value scales down.
* `metric_interval_lower_bound` - (Optional) The lower bound for the
difference between the alarm threshold and the CloudWatch metric.
Without a value, AWS will treat this bound as negative infinity.
* `metric_interval_upper_bound` - (Optional) The upper bound for the
difference between the alarm threshold and the CloudWatch metric.
Without a value, AWS will treat this bound as positive infinity. The upper bound
must be greater than the lower bound.

Notice the bounds are **relative** to the alarm threshold, meaning that the starting point is not 0%, but the alarm threshold. Check the official [docs](https://docs.aws.amazon.com/autoscaling/ec2/userguide/as-scaling-simple-step.html#as-scaling-steps) for a detailed example.

The following arguments are only available to "TargetTrackingScaling" type policies:

* `target_tracking_configuration` - (Optional) A target tracking policy. These have the following structure:

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

* `predefined_metric_specification` - (Optional) A predefined metric. Conflicts with `customized_metric_specification`.
* `customized_metric_specification` - (Optional) A customized metric. Conflicts with `predefined_metric_specification`.
* `target_value` - (Required) The target value for the metric.
* `disable_scale_in` - (Optional, Default: false) Indicates whether scale in by the target tracking policy is disabled.

### predefined_metric_specification

The following arguments are supported:

* `predefined_metric_type` - (Required) The metric type.
* `resource_label` - (Optional) Identifies the resource associated with the metric type.

### customized_metric_specification

The following arguments are supported:

* `metric_dimension` - (Optional) The dimensions of the metric.
* `metric_name` - (Required) The name of the metric.
* `namespace` - (Required) The namespace of the metric.
* `statistic` - (Required) The statistic of the metric.
* `unit` - (Optional) The unit of the metric.

#### metric_dimension

The following arguments are supported:

* `name` - (Required) The name of the dimension.
* `value` - (Required) The value of the dimension.

### predictive_scaling_configuration

The following arguments are supported:

* `max_capacity_breach_behavior` - (Optional) Defines the behavior that should be applied if the forecast capacity approaches or exceeds the maximum capacity of the Auto Scaling group. Valid values are `HonorMaxCapacity` or `IncreaseMaxCapacity`. Default is `HonorMaxCapacity`.
* `max_capacity_buffer` - (Optional) The size of the capacity buffer to use when the forecast capacity is close to or exceeds the maximum capacity. Valid range is `0` to `100`. If set to `0`, Amazon EC2 Auto Scaling may scale capacity higher than the maximum capacity to equal but not exceed forecast capacity.
* `metric_specification` - (Required) This structure includes the metrics and target utilization to use for predictive scaling.
* `mode` - (Optional) The predictive scaling mode. Valid values are `ForecastAndScale` and `ForecastOnly`. Default is `ForecastOnly`.
* `scheduling_buffer_time` - (Optional) The amount of time, in seconds, by which the instance launch time can be advanced. Minimum is `0`.

#### metric_specification

The following arguments are supported:

* `customized_capacity_metric_specification` - (Optional) The customized capacity metric specification. The field is only valid when you use `customized_load_metric_specification`
* `customized_load_metric_specification` - (Optional) The customized load metric specification.
* `customized_scaling_metric_specification` - (Optional) The customized scaling metric specification.
* `predefined_load_metric_specification` - (Optional) The predefined load metric specification.
* `predefined_metric_pair_specification` - (Optional) The metric pair specification from which Amazon EC2 Auto Scaling determines the appropriate scaling metric and load metric to use.
* `predefined_scaling_metric_specification` - (Optional) The predefined scaling metric specification.

##### predefined_load_metric_specification

The following arguments are supported:

* `predefined_metric_type` - (Required) The metric type. Valid values are `ASGTotalCPUUtilization`, `ASGTotalNetworkIn`, `ASGTotalNetworkOut`, or `ALBTargetGroupRequestCount`.
* `resource_label` - (Required) A label that uniquely identifies a specific Application Load Balancer target group from which to determine the request count served by your Auto Scaling group.

##### predefined_metric_pair_specification

The following arguments are supported:

* `predefined_metric_type` - (Required) Indicates which metrics to use. There are two different types of metrics for each metric type: one is a load metric and one is a scaling metric. For example, if the metric type is `ASGCPUUtilization`, the Auto Scaling group's total CPU metric is used as the load metric, and the average CPU metric is used for the scaling metric. Valid values are `ASGCPUUtilization`, `ASGNetworkIn`, `ASGNetworkOut`, or `ALBRequestCount`.
* `resource_label` - (Required) A label that uniquely identifies a specific Application Load Balancer target group from which to determine the request count served by your Auto Scaling group.

##### predefined_scaling_metric_specification

The following arguments are supported:

* `predefined_metric_type` - (Required) Describes a scaling metric for a predictive scaling policy. Valid values are `ASGAverageCPUUtilization`, `ASGAverageNetworkIn`, `ASGAverageNetworkOut`, or `ALBRequestCountPerTarget`.
* `resource_label` - (Required) A label that uniquely identifies a specific Application Load Balancer target group from which to determine the request count served by your Auto Scaling group.

##### customized_scaling_metric_specification
The following arguments are supported:

* `metric_data_queries` - (Required) A list of up to 10 structures that defines custom scaling metric in predictive scaling policy

##### customized_load_metric_specification
The following arguments are supported:

* `metric_data_queries` - (Required) A list of up to 10 structures that defines custom load metric in predictive scaling policy

##### customized_capacity_metric_specification
The following arguments are supported:

* `metric_data_queries` - (Required) A list of up to 10 structures that defines custom capacity metric in predictive scaling policy

##### metric_data_queries
The following arguments are supported:

* `expression` - (Optional) The math expression used on the returned metric. You must specify either `expression` or `metric_stat`, but not both.
* `id` - (Required) A short name for the metric used in predictive scaling policy.
* `metric_stat` - (Optional) A structure that defines CloudWatch metric to be used in predictive scaling policy. You must specify either `expression` or `metric_stat`, but not both.
* `label` - (Optional) A human-readable label for this metric or expression.
* `return_data` - (Optional) A boolean that indicates whether to return the timestamps and raw data values of this metric, the default it true

##### metric_stat
The following arguments are supported:

* `metric` - (Required) A structure that defines the CloudWatch metric to return, including the metric name, namespace, and dimensions.
* `stat` - (Required) The statistic of the metrics to return.
* `unit` - (Optional) The unit of the metrics to return.

##### metric
The following arguments are supported:

* `dimensions` - (Optional) The dimensions of the metric.
* `metric_name` - (Required) The name of the metric.
* `namespace` - (Required) The namespace of the metric.

##### dimensions
The following arguments are supported:

* `name` - (Required) The name of the dimension.
* `value` - (Required) The value of the dimension.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The ARN assigned by AWS to the scaling policy.
* `name` - The scaling policy's name.
* `autoscaling_group_name` - The scaling policy's assigned autoscaling group.
* `adjustment_type` - The scaling policy's adjustment type.
* `policy_type` - The scaling policy's type.

## Import

AutoScaling scaling policy can be imported using the role autoscaling_group_name and name separated by `/`.

```
$ terraform import aws_autoscaling_policy.test-policy asg-name/policy-name
```
