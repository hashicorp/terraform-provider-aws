---
subcategory: "SNS (Simple Notification)"
layout: "aws"
page_title: "AWS: aws_sns_topic_subscription"
description: |-
  Lists SNS (Simple Notification) Topic Subscription resources.
---

# List Resource: aws_sns_topic_subscription

Lists SNS (Simple Notification) Topic Subscription resources.

## Example Usage

```terraform
list "aws_sns_topic_subscription" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
