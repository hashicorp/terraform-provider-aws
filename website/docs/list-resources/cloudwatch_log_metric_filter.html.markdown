---
subcategory: "CloudWatch Logs"
layout: "aws"
page_title: "AWS: aws_cloudwatch_log_metric_filter"
description: |-
  Lists CloudWatch Logs Metric Filter resources.
---

# List Resource: aws_cloudwatch_log_metric_filter

Lists CloudWatch Logs Metric Filter resources.

## Example Usage

```terraform
list "aws_cloudwatch_log_metric_filter" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
