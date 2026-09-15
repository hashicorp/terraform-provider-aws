---
subcategory: "EventBridge"
layout: "aws"
page_title: "AWS: aws_cloudwatch_event_target"
description: |-
  Lists EventBridge Target resources.
---

# List Resource: aws_cloudwatch_event_target

Lists EventBridge Target resources for a specific rule.

## Example Usage

```terraform
list "aws_cloudwatch_event_target" "example" {
  provider = aws
  config {
    event_bus_name = "default"
    rule           = "my-rule"
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `event_bus_name` - (Required) Name or ARN of the event bus associated with the rule.
* `region` - (Optional) Region to query. Defaults to provider region.
* `rule` - (Required) Name of the rule.
