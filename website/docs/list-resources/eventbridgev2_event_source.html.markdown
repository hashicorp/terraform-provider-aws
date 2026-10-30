---
subcategory: "EventBridge V2"
layout: "aws"
page_title: "AWS: aws_eventbridgev2_event_source"
description: |-
  Lists EventBridge V2 event sources.
---

# List Resource: aws_eventbridgev2_event_source

Lists EventBridge V2 event sources. Only event sources owned by the caller's account are listed.

## Example Usage

```terraform
list "aws_eventbridgev2_event_source" "example" {
  provider = aws

  config {
    event_bus_arn = aws_eventbridgev2_event_bus.example.arn
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `event_bus_arn` - (Optional) ARN of the event bus whose event sources are listed. If omitted, event sources of all event buses are listed.
* `name_prefix` - (Optional) Name prefix used to filter event sources.
* `region` - (Optional) Region to query. Defaults to provider region.
