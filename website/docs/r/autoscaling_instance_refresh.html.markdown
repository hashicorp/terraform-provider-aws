---
subcategory: "Autoscaling"
layout: "aws"
page_title: "AWS: aws_autoscaling_instance_refresh"
description: |-
  Provides an Auto-Scaling Instance Refresh resource.
---

# Resource: aws_autoscaling_instance_refresh

Provides an Auto-Scaling Instance Refresh resource. An [Instance Refresh](https://docs.aws.amazon.com/autoscaling/ec2/userguide/asg-instance-refresh.html)
can be used to replace instances in response to changes to the auto-scaling group.

~> **NOTE:** The time it takes to refresh an auto-scaling group is highly sensitive to factors such
as capacity and instance warmup time. The default timeout is
intentionally a very short duration. Ensure that you have chosen a create timeout
that appropriately estimates the expected duration of a successful rollout.

## Example Usage

```hcl
resource "aws_launch_configuration" "test" {
  image_id      = "ami-123456789"
  instance_type = "t2.micro"
}

resource "aws_autoscaling_group" "test" {
  availability_zones        = ["us-west-2"]
  min_size                  = 1
  max_size                  = 2
  launch_configuration      = aws_launch_configuration.test.name
  health_check_grace_period = 5
}

resource "aws_autoscaling_instance_refresh" "test" {
  autoscaling_group_name  = aws_autoscaling_group.test.name
  min_healthy_percentage  = 50
  instance_warmup_seconds = 5
  strategy                = "Rolling"

  triggers = {
    token = aws_autoscaling_group.test.instance_refresh_token
  }

  timeouts {
    create = "30m"
  }
}
```

## Argument Reference

The following arguments are supported:

* `autoscaling_group_name` - (Required, Forces new resource) The name of the auto-scaling group.
* `instance_warmup_seconds` - (Optional) The number of seconds until a newly launched
  instance is configured and ready to use. Default behavior (set with `-1` or `null`)
  is to match the auto-scaling group's health check grace period.
* `min_healthy_percentage` - (Optional) The percentage of capacity in the Auto Scaling group
  that must remain healthy during an instance refresh to allow the operation to continue.
  Defaults to `90`.
* `strategy` - (Required) The strategy to use for instance refresh. The only allowed
  value is `"Rolling"`. See the [`StartInstanceRefresh` API](https://docs.aws.amazon.com/autoscaling/ec2/APIReference/API_StartInstanceRefresh.html#API_StartInstanceRefresh_RequestParameters) for more information.
* `triggers` - (Optional, Forces new resource) A map of arbitrary keys and values that, when changed, will
  trigger an instance refresh. To force a new redeployment without changing these
  keys or values, use the [`terraform taint` command](/docs/commands/taint.html).
  It is recommended that, at minimum, the `instance_refresh_token` of the auto-scaling group is
  used as one of the triggers.
* `wait_for_completion` - (Optional) When true, wait for the instance refresh to complete
  before marking resource creation a success. If the instance refresh times out or
  fails, the pending instance refresh is cancelled, and creation of this resource
  fails as well. Defaults to `true`.
  
~> **NOTE:** If `wait_for_completion` is set to `false`, Terraform will never check
the outcome of the refresh. If the instance refresh fails or times out on its own,
Terraform will report success and not re-attempt unless the resource is
re-created.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The concatenation `autoscaling_group_name/instance_refresh_id`.
* `instance_refresh_id` - The ID of the instance refresh.

## Timeouts

This resource provides the following
[Timeouts](/docs/configuration/resources.html#timeouts) configuration options:

- `create` - (Default `5 minutes`) The maximum amount of time this resource will wait for the instance refresh to succeed.

~> **NOTE:** The time it takes to refresh an auto-scaling group is highly sensitive to factors such
as capacity and instance warmup time. The default timeout is
intentionally a very short duration. Ensure that you have chosen a create timeout
that appropriately estimates the expected duration of a successful rollout.

## Import

`aws_autoscaling_instance_refresh` can be imported using the concatenation
`autoscaling_group_name/instance_refresh_id`. For example:

```
$ terraform import aws_autoscaling_instance_refresh.test my-asg/6eba610f-40c5-4211-8d95-31c92c9161f0
```
