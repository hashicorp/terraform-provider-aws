---
layout: "aws"
page_title: "AWS: aws_autoscaling_policy"
sidebar_current: "docs-aws-resource-autoscaling-policy"
description: |-
  Provides an AutoScaling Scaling Group resource.
---

# aws_autoscaling_policy

Provides an AutoScaling Scaling Policy resource.

~> **NOTE:** You may want to omit `desired_capacity` attribute from attached `aws_autoscaling_group`
when using autoscaling policies. It's good practice to pick either
[manual](https://docs.aws.amazon.com/AutoScaling/latest/DeveloperGuide/as-manual-scaling.html)
or [dynamic](https://docs.aws.amazon.com/AutoScaling/latest/DeveloperGuide/as-scale-based-on-demand.html)
(policy-based) scaling.

## Example Usage

### Simple Scaling

```hcl
resource "aws_autoscaling_policy" "bat" {
  name                   = "foobar3-terraform-test"
  scaling_adjustment     = 4
  adjustment_type        = "ChangeInCapacity"
  cooldown               = 300
  autoscaling_group_name = "${aws_autoscaling_group.bar.name}"
}

resource "aws_autoscaling_group" "bar" {
  availability_zones        = ["us-east-1a"]
  name                      = "foobar3-terraform-test"
  max_size                  = 5
  min_size                  = 2
  health_check_grace_period = 300
  health_check_type         = "ELB"
  force_delete              = true
  launch_configuration      = "${aws_launch_configuration.foo.name}"
}
```

### Step Scaling

```hcl
resource "aws_autoscaling_policy" "foobar_step" {
  name = "foobar_step"
  adjustment_type = "ChangeInCapacity"
  policy_type = "StepScaling"
  estimated_instance_warmup = 200
  metric_aggregation_type = "Minimum"
  step_adjustment {
    scaling_adjustment = 1
    metric_interval_lower_bound = 2.0
  }
  autoscaling_group_name = "${aws_autoscaling_group.bar.name}"
}
```

### Target Tracking Scaling

```hcl
resource "aws_autoscaling_policy" "foobar_targettracking" {
  name = "foobar_targettracking"
  policy_type = "TargetTrackingScaling"
  estimated_instance_warmup = 200
  target_tracking_configuration {
    predefined_metric_specification {
      predefined_metric_type = "ASGAverageCPUUtilization"
    }
    target_value = 20
  }
  autoscaling_group_name = "${aws_autoscaling_group.bar.name}"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the policy.
* `autoscaling_group_name` - (Required) The name of the autoscaling group.
* `policy_type` - (Optional) The policy type, either "SimpleScaling", "StepScaling" or "TargetTrackingScaling". If this value isn't provided, AWS will default to "SimpleScaling."

The following arguments are available to "SimpleScaling" type policies:

* `adjustment_type` - (Required) Specifies whether the adjustment is an absolute number or a percentage of the current capacity. Valid values are `ChangeInCapacity`, `ExactCapacity`, and `PercentChangeInCapacity`.
* `cooldown` - (Optional) The amount of time, in seconds, after a scaling activity completes and before the next scaling activity can start.
* `scaling_adjustment` - (Optional) The number of instances by which to scale. `adjustment_type` determines the interpretation of this number (e.g., as an absolute number or as a percentage of the existing Auto Scaling group size). A positive increment adds to the current capacity and a negative value removes from the current capacity.

The following arguments are available to "StepScaling" type policies:

* `adjustment_type` - (Required) Specifies whether the adjustment is an absolute number or a percentage of the current capacity. Valid values are `ChangeInCapacity`, `ExactCapacity`, and `PercentChangeInCapacity`.
* `metric_aggregation_type` - (Optional) The aggregation type for the policy's metrics. Valid values are "Minimum", "Maximum", and "Average". Without a value, AWS will treat the aggregation type as "Average".
* `estimated_instance_warmup` - (Optional) The estimated time, in seconds, until a newly launched instance will contribute CloudWatch metrics. Without a value, AWS will default to the group's specified cooldown period.
* `step_adjustments` - (Optional) A set of adjustments that manage
group scaling. These have the following structure:

```hcl
step_adjustment {
  scaling_adjustment = -1
  metric_interval_lower_bound = 1.0
  metric_interval_upper_bound = 2.0
}
step_adjustment {
  scaling_adjustment = 1
  metric_interval_lower_bound = 2.0
  metric_interval_upper_bound = 3.0
}
```

The following fields are available in step adjustments:

* `scaling_adjustment` - (Required) The number of members by which to
scale, when the adjustment bounds are breached. A positive value scales
up. A negative value scales down.
* `metric_interval_lower_bound` - (Optional) The lower bound for the
difference between the alarm threshold and the CloudWatch metric.
Without a value, AWS will treat this bound as infinity.
* `metric_interval_upper_bound` - (Optional) The upper bound for the
difference between the alarm threshold and the CloudWatch metric.
Without a value, AWS will treat this bound as infinity. The upper bound
must be greater than the lower bound.

The following arguments are supported for backwards compatibility but should not be used:

* `min_adjustment_step` - (Optional) Use `min_adjustment_magnitude` instead.

The following arguments are available to "TargetTrackingScaling" type policies:

* `target_tracking_configuration` - (Required) A target tracking policy, requires `policy_type = "TargetTrackingScaling"`.

`target_tracking_configuration` supported fields below.

* `target_value` - (Optional) The target value for the metric.
* `disable_scale_in` - (Optional) Indicates whether scale in by the target tracking policy is disabled. If the value is true, scale in is disabled and the target tracking policy won't remove capacity from the scalable resource. Otherwise, scale in is enabled and the target tracking policy can remove capacity from the scalable resource. The default value is `false`.
* `customized_metric_specification` - (Optional) Reserved for future use. See supported fields below.
* `predefined_metric_specification` - (Optional) A predefined metric. See supported fields below.

`customized_metric_specification` supported fields below.

* `dimensions` - (Optional) The dimensions of the metric.
* `metric_name` - (Optional) The name of the metric.
* `namespace` - (Optional) The namespace of the metric.
* `statistic` - (Optional) The statistic of the metric. May be one of `"Average"`, `"Minimum"`, `"Maximum"`, `"SampleCount"` or `"Sum"`.
* `unit` - (Optional) The unit of the metric.

`predefined_metric_specification` supported fields below.

* `predefined_metric_type` - (Required) The metric type. May be one of `"ASGAverageCPUUtilization"`, `"ASGAverageNetworkIn"`, `"ASGAverageNetworkOut"` or `"ALBRequestCountPerTarget"`.
* `resource_label` - (Optional) Reserved for future use.

## Attribute Reference
* `arn` - The ARN assigned by AWS to the scaling policy.
* `name` - The scaling policy's name.
* `autoscaling_group_name` - The scaling policy's assigned autoscaling group.
* `adjustment_type` - The scaling policy's adjustment type.
* `policy_type` - The scaling policy's type.
