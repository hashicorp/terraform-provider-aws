---
subcategory: "Application Auto Scaling"
layout: "aws"
page_title: "AWS: aws_appautoscaling_scheduled_action"
description: |-
  Provides an Application AutoScaling ScheduledAction resource.
---

# Resource: aws_appautoscaling_scheduled_action

Provides an Application AutoScaling ScheduledAction resource.

## Example Usage

### DynamoDB Table Autoscaling

```terraform
resource "aws_appautoscaling_target" "dynamodb" {
  max_capacity       = 100
  min_capacity       = 5
  resource_id        = "table/tableName"
  scalable_dimension = "dynamodb:table:ReadCapacityUnits"
  service_namespace  = "dynamodb"
}

resource "aws_appautoscaling_scheduled_action" "dynamodb" {
  name               = "dynamodb"
  service_namespace  = aws_appautoscaling_target.dynamodb.service_namespace
  resource_id        = aws_appautoscaling_target.dynamodb.resource_id
  scalable_dimension = aws_appautoscaling_target.dynamodb.scalable_dimension
  schedule           = "at(2006-01-02T15:04:05)"

  scalable_target_action {
    min_capacity = 1
    max_capacity = 200
  }
}
```

### ECS Service Autoscaling

```terraform
resource "aws_appautoscaling_target" "ecs" {
  max_capacity       = 4
  min_capacity       = 1
  resource_id        = "service/clusterName/serviceName"
  scalable_dimension = "ecs:service:DesiredCount"
  service_namespace  = "ecs"
}

resource "aws_appautoscaling_scheduled_action" "ecs" {
  name               = "ecs"
  service_namespace  = aws_appautoscaling_target.ecs.service_namespace
  resource_id        = aws_appautoscaling_target.ecs.resource_id
  scalable_dimension = aws_appautoscaling_target.ecs.scalable_dimension
  schedule           = "at(2006-01-02T15:04:05)"

  scalable_target_action {
    min_capacity = 1
    max_capacity = 10
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `name` - (Required) Name of the scheduled action.
* `service_namespace` - (Required) Namespace of the AWS service. Documentation can be found in the `ServiceNamespace` parameter at: [AWS Application Auto Scaling API Reference](https://docs.aws.amazon.com/autoscaling/application/APIReference/API_PutScheduledAction.html) Example: ecs
* `resource_id` - (Required) Identifier of the resource associated with the scheduled action. Documentation can be found in the `ResourceId` parameter at: [AWS Application Auto Scaling API Reference](https://docs.aws.amazon.com/autoscaling/application/APIReference/API_PutScheduledAction.html)
* `scalable_dimension` - (Required) Scalable dimension. Documentation can be found in the `ScalableDimension` parameter at: [AWS Application Auto Scaling API Reference](https://docs.aws.amazon.com/autoscaling/application/APIReference/API_PutScheduledAction.html) Example: ecs:service:DesiredCount
* `scalable_target_action` - (Required) New minimum and maximum capacity. You can set both values or just one. See [below](#scalable-target-action-arguments)
* `schedule` - (Required) Schedule for this action. The following formats are supported: At expressions - at(yyyy-mm-ddThh:mm:ss), Rate expressions - rate(valueunit), Cron expressions - cron(fields). Times for at expressions and cron expressions are evaluated using the time zone configured in `timezone`. Documentation can be found in the `Timezone` parameter at: [AWS Application Auto Scaling API Reference](https://docs.aws.amazon.com/autoscaling/application/APIReference/API_PutScheduledAction.html)
* `start_time` - (Optional) Date and time for the scheduled action to start in RFC 3339 format. The timezone is not affected by the setting of `timezone`.
* `end_time` - (Optional) Date and time for the scheduled action to end in RFC 3339 format. The timezone is not affected by the setting of `timezone`.
* `timezone` - (Optional) Time zone used when setting a scheduled action by using an at or cron expression. Does not affect timezone for `start_time` and `end_time`. Valid values are the [canonical names of the IANA time zones supported by Joda-Time](https://www.joda.org/joda-time/timezones.html), such as `Etc/GMT+9` or `Pacific/Tahiti`. Default is `UTC`.

### Scalable Target Action Arguments

* `max_capacity` - (Optional) Maximum capacity. At least one of `max_capacity` or `min_capacity` must be set.
* `min_capacity` - (Optional) Minimum capacity. At least one of `min_capacity` or `max_capacity` must be set.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the scheduled action.
