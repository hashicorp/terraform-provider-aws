---
subcategory: "Auto Scaling"
layout: "aws"
page_title: "AWS: aws_autoscaling_schedule"
description: |-
  Provides an AutoScaling Schedule resource.
---

# Resource: aws_autoscaling_schedule

Provides an AutoScaling Schedule resource.

## Example Usage

```terraform
resource "aws_autoscaling_group" "foobar" {
  availability_zones        = ["us-west-2a"]
  name                      = "terraform-test-foobar5"
  max_size                  = 1
  min_size                  = 1
  health_check_grace_period = 300
  health_check_type         = "ELB"
  force_delete              = true
  termination_policies      = ["OldestInstance"]
}

resource "aws_autoscaling_schedule" "foobar" {
  scheduled_action_name  = "foobar"
  min_size               = 0
  max_size               = 1
  desired_capacity       = 0
  start_time             = "2016-12-11T18:00:00Z"
  end_time               = "2016-12-12T06:00:00Z"
  autoscaling_group_name = aws_autoscaling_group.foobar.name
}
```

## Argument Reference

The following arguments are required:

* `autoscaling_group_name` - (Required) The name of the Auto Scaling group.
* `scheduled_action_name` - (Required) The name of this scaling action.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `desired_capacity` - (Optional) The initial capacity of the Auto Scaling group after the scheduled action runs and the capacity it attempts to maintain. Set to `-1` if you don't want to change the desired capacity at the scheduled time. Defaults to `0`.
* `end_time` - (Optional) The date and time for the recurring schedule to end, in UTC with the format `"YYYY-MM-DDThh:mm:ssZ"` (e.g. `"2021-06-01T00:00:00Z"`).
* `max_size` - (Optional) The maximum size of the Auto Scaling group. Set to `-1` if you don't want to change the maximum size at the scheduled time. Defaults to `0`.
* `min_size` - (Optional) The minimum size of the Auto Scaling group. Set to `-1` if you don't want to change the minimum size at the scheduled time. Defaults to `0`.
* `recurrence` - (Optional) The recurring schedule for this action specified using the Unix cron syntax format.
* `start_time` - (Optional) The date and time for the recurring schedule to start, in UTC with the format `"YYYY-MM-DDThh:mm:ssZ"` (e.g. `"2021-06-01T00:00:00Z"`).
* `time_zone` - (Optional)  Specifies the time zone for a cron expression. Valid values are the canonical names of the IANA time zones (such as `Etc/GMT+9` or `Pacific/Tahiti`).

~> **NOTE:** When `start_time` and `end_time` are specified with `recurrence` , they form the boundaries of when the recurring action will start and stop.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN assigned by AWS to the autoscaling schedule.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import AutoScaling ScheduledAction using the `auto-scaling-group-name` and `scheduled-action-name`. For example:

```terraform
import {
  to = aws_autoscaling_schedule.resource-name
  id = "auto-scaling-group-name/scheduled-action-name"
}
```

Using `terraform import`, import AutoScaling ScheduledAction using the `auto-scaling-group-name` and `scheduled-action-name`. For example:

```console
% terraform import aws_autoscaling_schedule.resource-name auto-scaling-group-name/scheduled-action-name
```
