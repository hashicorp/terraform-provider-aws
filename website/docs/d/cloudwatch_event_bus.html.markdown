---
subcategory: "EventBridge"
layout: "aws"
page_title: "AWS: aws_cloudwatch_event_bus"
description: |-
  Get information on an EventBridge (Cloudwatch) Event Bus.
---

# Data Source: aws_cloudwatch_event_bus

This data source can be used to fetch information about a specific
EventBridge event bus. Use this data source to compute the ARN of
an event bus, given the name of the bus.

## Example Usage

```terraform
data "aws_cloudwatch_event_bus" "example" {
  name = "example-bus-name"
}
```

## Argument Reference

* `name` - (Required) The friendly EventBridge event bus name.

## Attributes Reference

* `arn` - The Amazon Resource Name (ARN) specifying the role.
