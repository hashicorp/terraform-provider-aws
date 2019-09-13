---
subcategory: "CloudWatch"
layout: "aws"
page_title: "AWS: aws_cloudwatch_event_bus"
description: |-
  Provides a CloudWatch Event Bus resource.
---

# Resource: aws_cloudwatch_event_bus

Provides a CloudWatch Event Bus resource.

## Example Usage

```hcl
resource "aws_cloudwatch_event_bus" "messenger" {
  name = "chat-messages"
}

resource "aws_cloudwatch_event_rule" "foo" {
  name           = "foo"
  event_bus_name = aws_cloudwatch_event_bus.messenger.name
  event_pattern = <<PATTERN
{
  "detail-type": [
    "foo"
  ]
}
PATTERN
}

resource "aws_cloudwatch_event_target" "foo" {
  rule           = aws_cloudwatch_event_rule.foo.name
  event_bus_name = aws_cloudwatch_event_bus.messenger.name
  arn            = "arn:aws:lambda:YOUR_REGION:YOUR_ACCOUNT:function:foo"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the new event bus.
	The names of custom event buses can't contain the / character.
	If this is a partner event bus, the name must exactly match the name of the partner event source that this event bus is matched to. This name will include the / character.
* `event_source_name` - (Optional) If you're creating a partner event bus, this specifies the partner event source that the new event bus will be matched with.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The Amazon Resource Name (ARN) of the event bus.
* `policy` - The policy that enables the external account to send events to your account.


## Import

Cloudwatch Event Buses can be imported using the `name`, e.g.

```
$ terraform import aws_cloudwatch_event_bus.messenger chat-messages
```