---
subcategory: "CloudWatch"
layout: "aws"
page_title: "AWS: aws_cloudwatch_metric_alarm"
description: |-
  Lists CloudWatch Metric Alarm resources.
---

# List Resource: aws_cloudwatch_metric_alarm

Lists CloudWatch Metric Alarm resources.

## Example Usage

```terraform
list "aws_cloudwatch_metric_alarm" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
