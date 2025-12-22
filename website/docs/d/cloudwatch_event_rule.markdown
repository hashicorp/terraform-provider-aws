---
subcategory: "EventBridge"
layout: "aws"
page_title: "AWS: aws_cloudwatch_event_bus"
description: |-
    Get information about an EventBridge (Cloudwatch) event rule
---

# Data Source: aws_events_rule

  A data source to get information about an EventBridge event rule.

~> **Note:** EventBridge was formerly known as CloudWatch Events. The functionality is identical.

## Example Usage

```terraform
data "aws_cloudwatch_event_rule" "example" {
    name = "example-rule"
    event_bus_name = "example-bus" # omitting uses the "default" bus
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the EventBridge rule.

* `event_bus_name` - (Optional) Name of the EventBridge event bus associated with the rule. The "default" bus is used if omitted.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Rule.

* `event_pattern` - Rule pattern

* `schedule_expression` - schedule expression

* `state` - State of the event rule. One of `DISABLED`, `ENABLED`, and `ENABLED_WITH_ALL_CLOUDTRAIL_MANAGEMENT_EVENTS`
