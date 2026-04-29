---
subcategory: "SNS (Simple Notification)"
layout: "aws"
page_title: "AWS: aws_sns_topic_policy"
description: |-
  Lists SNS (Simple Notification) Topic Policy resources.
---

# List Resource: aws_sns_topic_policy

Lists SNS (Simple Notification) Topic Policy resources.

## Example Usage

```terraform
list "aws_sns_topic_policy" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
