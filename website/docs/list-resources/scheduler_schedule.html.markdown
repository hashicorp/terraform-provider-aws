---
subcategory: "EventBridge Scheduler"
layout: "aws"
page_title: "AWS: aws_scheduler_schedule"
description: |-
  Lists EventBridge Scheduler Schedule resources.
---

# List Resource: aws_scheduler_schedule

Lists EventBridge Scheduler Schedule resources.

## Example Usage

```terraform
list "aws_scheduler_schedule" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
