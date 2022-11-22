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
* `ecs_parameters` - (Optional) Templated target type for the Amazon ECS [`RunTask`](https://docs.aws.amazon.com/AmazonECS/latest/APIReference/API_RunTask.html) API operation. Detailed below.
* `eventbridge_parameters` - (Optional) Templated target type for the EventBridge [`PutEvents`](https://docs.aws.amazon.com/eventbridge/latest/APIReference/API_PutEvents.html) API operation. Detailed below.
* `input` - (Optional) Text, or well-formed JSON, passed to the target. Read more in [Universal target](https://docs.aws.amazon.com/scheduler/latest/UserGuide/managing-targets-universal.html).
* `kinesis_parameters` - (Optional) Templated target type for the Amazon Kinesis [`PutRecord`](https://docs.aws.amazon.com/scheduler/latest/APIReference/kinesis/latest/APIReference/API_PutRecord.html) API operation. Detailed below.
* `retry_policy` - (Optional) Information about the retry policy settings. Detailed below.
* `sagemaker_pipeline_parameters` - (Optional) Templated target type for the Amazon SageMaker [`StartPipelineExecution`](https://docs.aws.amazon.com/sagemaker/latest/APIReference/API_StartPipelineExecution.html) API operation. Detailed below.
* `sqs_parameters` - (Optional) The templated target type for the Amazon SQS [`SendMessage`](https://docs.aws.amazon.com/AWSSimpleQueueService/latest/APIReference/API_SendMessage.html) API operation. Detailed below.

#### dead_letter_config Configuration Block

* `arn` - (Optional) ARN of the SQS queue specified as the destination for the dead-letter queue.

#### ecs_parameters Configuration Block

The following arguments are required:

* `task_definition_arn` - (Required) ARN of the task definition to use.

The following arguments are optional:

* `capacity_provider_strategy` - (Optional) Capacity provider strategy to use for the task. Up to `6` items. Detailed below.
* `enable_ecs_managed_tags` - (Optional) Specifies whether to enable Amazon ECS managed tags for the task. For more information, see [Tagging Your Amazon ECS Resources](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/ecs-using-tags.html) in the Amazon ECS Developer Guide.
* `enable_execute_command` - (Optional) Specifies whether to enable the execute command functionality for the containers in this task.
* `group` - (Optional) Specifies an ECS task group for the task. Up to `255` characters.
* `launch_type` - (Optional) Specifies the launch type on which your task is running. The launch type that you specify here must match one of the launch type (compatibilities) of the target task. One of `EC2`, `FARGATE`, `EXTERNAL`.
* `network_configuration` - (Optional) Detailed below.
* `placement_constraints` - (Optional) An array of placement constraint objects to use for the task. You can specify up to `10` constraints per task (including constraints in the task definition and those specified at runtime). Detailed below.
* `placement_strategy` - (Optional) Set of `0` to `5` placement strategies. Detailed below.
* `platform_version` - (Optional) Specifies the platform version for the task. Specify only the numeric portion of the platform version, such as `1.1.0`.
* `propagate_tags` - (Optional) Specifies whether to propagate the tags from the task definition to the task. The only valid value is `TASK_DEFINITION`.
* `reference_id` - (Optional) Reference ID to use for the task.
* `tags` - (Optional) The metadata that you apply to the task. Each tag consists of a key and an optional value. For more information, see [`RunTask`](https://docs.aws.amazon.com/AmazonECS/latest/APIReference/API_RunTask.html) in the Amazon ECS API Reference.
* `task_count` - (Optional) The number of tasks to create. Between `1` (default) and `10`.

##### capacity_provider_strategy Configuration Block

* `base` - (Optional) How many tasks, at a minimum, to run on the specified capacity provider. Only one capacity provider in a capacity provider strategy can have a base defined. Can range from `0` (default) to `100000`.
* `capacity_provider` - (Required) Short name of the capacity provider.
* `weight` - (Optional) Designates the relative percentage of the total number of tasks launched that should use the specified capacity provider. The weight value is taken into consideration after the base value, if defined, is satisfied. Can range from `0` to `1000`.

##### network_configuration Configuration Block

* `awsvpc_configuration` - (Optional) Detailed below.

##### awsvpc_configuration Configuration Block

* `assign_public_ip` - (Optional) Specifies whether the task's elastic network interface receives a public IP address. You can specify `ENABLED` only when the `launch_type` is set to `FARGATE`. One of `ENABLED`, `DISABLED`.
* `security_groups` - (Optional) Set of 1 to 5 Security Group ID-s to be associated with the task. These security groups must all be in the same VPC.
* `subnets` - (Optional) Set of 1 to 16 subnets to be associated with the task. These subnets must all be in the same VPC.

##### placement_constraints Configuration Block

* `expression` - (Optional) A cluster query language expression to apply to the constraint. You cannot specify an expression if the constraint type is `distinctInstance`. For more information, see [Cluster query language](https://docs.aws.amazon.com/latest/developerguide/cluster-query-language.html) in the Amazon ECS Developer Guide.
* `type` - (Required) The type of constraint. One of `distinctInstance`, `memberOf`.

##### placement_strategy Configuration Block

* `field` - (Optional) The field to apply the placement strategy against.
* `type` - (Required) The type of placement strategy. One of `random`, `spread`, `binpack`.

#### eventbridge_parameters Configuration Block

* `detail_type` - (Required) Free-form string used to decide what fields to expect in the event detail. At most `128` characters.
* `source` - (Required) Source of the event.

#### kinesis_parameters Configuration Block

* `partition_key` - (Required) Specifies the shard to which EventBridge Scheduler sends the event. At most `256` characters.

#### retry_policy Configuration Block

* `maximum_event_age_in_seconds` - (Optional) Maximum amount of time, in seconds, to continue to make retry attempts. Between `60` and `86400` (default).
* `maximum_retry_attempts` - (Optional) Maximum number of retry attempts to make before the request fails. Between `0` and `185` (default).

#### sagemaker_pipeline_parameters Configuration Block

* `pipeline_parameter` - (Optional) Set of parameter names and values to use when executing the SageMaker Model Building Pipeline. Up to `200` parameters. Detailed below.

##### pipeline_parameter Configuration Block

* `name` - (Required) Name of parameter to start execution of a SageMaker Model Building Pipeline.
* `value` - (Required) Value of parameter to start execution of a SageMaker Model Building Pipeline.

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
