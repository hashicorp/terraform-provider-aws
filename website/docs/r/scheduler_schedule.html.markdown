---
subcategory: "EventBridge Scheduler"
layout: "aws"
page_title: "AWS: aws_scheduler_schedule"
description: |-
Provides an EventBridge Scheduler Schedule resource.
---

# Resource: aws_scheduler_schedule

Provides an EventBridge Scheduler Schedule resource.

You can find out more about EventBridge Scheduler in the [User Guide](https://docs.aws.amazon.com/scheduler/latest/UserGuide/what-is-scheduler.html).

~> **Note:** EventBridge was formerly known as CloudWatch Events. The functionality is identical.

## Example Usage

### Basic Usage

```terraform
resource "aws_scheduler_schedule" "example" {
  name       = "my-schedule"
  group_name = "default"

  flexible_time_window {
    mode = "OFF"
  }
  
  schedule_expression = "rate(1 hour)"
  
  target {
    arn      = aws_sqs_queue.example.arn
    role_arn = aws_iam_role.example.arn
  }
}
```

## Argument Reference

The following arguments are required:

* `flexible_time_window` - (Required) Configures a time window during which EventBridge Scheduler invokes the schedule. Detailed below.
* `schedule_expression` - (Required) Defines when the schedule runs. Read more in [Schedule types on EventBridge Scheduler](https://docs.aws.amazon.com/scheduler/latest/UserGuide/schedule-types.html).
* `target` - (Required) Detailed below.

The following arguments are optional:

* `description` - (Optional) Description specified for the schedule.
* `group_name` - (Optional, Forces new resource) Name of the schedule group to associate with this schedule. If you omit this, the default schedule group is used.
* `name` - (Optional, Forces new resource) Name of the schedule. If omitted, Terraform will assign a random, unique name. Conflicts with `name_prefix`.
* `name_prefix` - (Optional, Forces new resource) Creates a unique name beginning with the specified prefix. Conflicts with `name`.

### flexible_time_window Configuration Block

* `mode` - (Required) Determines whether the schedule is invoked within a flexible time window. Valid values: `OFF`, `FLEXIBLE`.

### target Configuration Block

* `arn` - (Required) ARN of the target of this schedule.
* `role_arn` - (Required) ARN of the IAM role that EventBridge Scheduler will use for this target when the schedule is invoked. Read more in [Set up the execution role](https://docs.aws.amazon.com/scheduler/latest/UserGuide/setting-up.html#setting-up-execution-role).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Name of the schedule.
* `arn` - ARN of the schedule.

## Import

Schedules can be imported using the combination `group_name/name`. For example:

```
$ terraform import aws_scheduler_schedule.example my-schedule-group/my-schedule
```
