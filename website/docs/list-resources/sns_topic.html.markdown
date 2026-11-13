---
subcategory: "SNS (Simple Notification)"
layout: "aws"
page_title: "AWS: aws_sns_topic"
description: |-
  Lists SNS (Simple Notification) Topic resources.
---

# List Resource: aws_sns_topic

Lists SNS (Simple Notification) Topic resources.

## Example Usage

```terraform
list "aws_sns_topic" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
