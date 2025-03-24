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

* `name` - (Required) Name of the event bus.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the event bus.
* `description` - Event bus description.
* `id` - Name of the event bus.
* `kms_key_identifier` - Identifier of the AWS KMS customer managed key for EventBridge to use to encrypt events on this event bus, if one has been specified.
