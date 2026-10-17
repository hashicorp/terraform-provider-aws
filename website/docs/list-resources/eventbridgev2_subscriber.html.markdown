---
subcategory: "EventBridge V2"
layout: "aws"
page_title: "AWS: aws_eventbridgev2_subscriber"
description: |-
  Lists EventBridge V2 subscribers.
---

# List Resource: aws_eventbridgev2_subscriber

Lists EventBridge V2 subscribers. Only subscribers owned by the caller's account are listed.

## Example Usage

```terraform
list "aws_eventbridgev2_subscriber" "example" {
  provider = aws

  config {
    event_bus_arn = aws_eventbridgev2_event_bus.example.arn
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `event_bus_arn` - (Optional) ARN of the event bus whose subscribers are listed. If omitted, subscribers of all event buses are listed.
* `name_prefix` - (Optional) Name prefix used to filter subscribers.
* `region` - (Optional) Region to query. Defaults to provider region.
