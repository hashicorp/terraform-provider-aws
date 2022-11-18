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

### Universal Target

```terraform
resource "aws_sqs_queue" "example" {}

resource "aws_scheduler_schedule" "example" {
  name = "my-schedule"

  flexible_time_window {
    mode = "OFF"
  }

  schedule_expression = "rate(1 hour)"

  target {
    arn      = "arn:aws:scheduler:::aws-sdk:sqs:sendMessage"
    role_arn = aws_iam_role.example.arn

    input = jsonencode({
      MessageBody = "Greetings, programs!"
      QueueUrl    = aws_sqs_queue.example.url
    })
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
* `end_date` - (Optional) The date, in UTC, before which the schedule can invoke its target. Depending on the schedule's recurrence expression, invocations might stop on, or before, the end date you specify. EventBridge Scheduler ignores the end date for one-time schedules. Example: `2030-01-01T01:00:00Z`.
* `group_name` - (Optional, Forces new resource) Name of the schedule group to associate with this schedule. When omitted, the default schedule group is used.
* `kms_key_arn` - (Optional) ARN for the customer managed KMS key that EventBridge Scheduler will use to encrypt and decrypt your data. 
* `name` - (Optional, Forces new resource) Name of the schedule. If omitted, Terraform will assign a random, unique name. Conflicts with `name_prefix`.
* `name_prefix` - (Optional, Forces new resource) Creates a unique name beginning with the specified prefix. Conflicts with `name`.
* `schedule_expression_timezone` - (Optional) Timezone in which the scheduling expression is evaluated. Defaults to `UTC`. Example: `Australia/Sydney`.
* `start_date` - (Optional) The date, in UTC, after which the schedule can begin invoking its target. Depending on the schedule's recurrence expression, invocations might occur on, or after, the start date you specify. EventBridge Scheduler ignores the start date for one-time schedules. Example: `2030-01-01T01:00:00Z`.
* `state` - (Optional) Specifies whether the schedule is enabled or disabled. Defaults to `ENABLED`. Valid values are `ENABLED`, `DISABLED`.

### flexible_time_window Configuration Block

* `maximum_window_in_minutes` - (Optional) Maximum time window during which a schedule can be invoked. Between 1 and 1440 minutes.
* `mode` - (Required) Determines whether the schedule is invoked within a flexible time window. Valid values: `OFF`, `FLEXIBLE`.

### target Configuration Block

The following arguments are required:

* `arn` - (Required) ARN of the target of this schedule.
* `role_arn` - (Required) ARN of the IAM role that EventBridge Scheduler will use for this target when the schedule is invoked. Read more in [Set up the execution role](https://docs.aws.amazon.com/scheduler/latest/UserGuide/setting-up.html#setting-up-execution-role).

The following arguments are optional:

* `dead_letter_config` - (Optional) Information about an Amazon SQS queue that EventBridge Scheduler uses as a dead-letter queue for your schedule. If specified, EventBridge Scheduler delivers failed events that could not be successfully delivered to a target to the queue. Detailed below.
* `input` - (Optional) Text, or well-formed JSON, passed to the target. Read more in [Universal target](https://docs.aws.amazon.com/scheduler/latest/UserGuide/managing-targets-universal.html).
* `retry_policy` - (Optional) Information about the retry policy settings. Detailed below.
* `sqs_parameters` - (Optional) The templated target type for the Amazon SQS `SendMessage` API operation. Detailed below.

#### dead_letter_config Configuration Block

* `arn` - (Optional) ARN of the SQS queue specified as the destination for the dead-letter queue.

#### retry_policy Configuration Block

* `maximum_event_age_in_seconds` - (Optional) Maximum amount of time, in seconds, to continue to make retry attempts. Between `60` and `86400` (default).
* `maximum_retry_attempts` - (Optional) Maximum number of retry attempts to make before the request fails. Between `0` and `185` (default).

#### sqs_parameters Configuration Block

* `message_group_id` - (Optional) FIFO message group ID to use as the target.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Name of the schedule.
* `arn` - ARN of the schedule.

## Import

Schedules can be imported using the combination `group_name/name`. For example:

```
$ terraform import aws_scheduler_schedule.example my-schedule-group/my-schedule
```
