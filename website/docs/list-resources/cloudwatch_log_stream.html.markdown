---
subcategory: "CloudWatch Logs"
layout: "aws"
page_title: "AWS: aws_cloudwatch_log_stream"
description: |-
  Lists CloudWatch Log Stream resources.
---

# List Resource: aws_cloudwatch_log_stream

Lists CloudWatch Log Stream resources.

## Example Usage

```terraform
list "aws_cloudwatch_log_stream" "example" {
  provider = aws

  config {
    log_group_name = "example-group"
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `log_group_name` - (Required) Name of the log group.
* `descending` - (Optional) If the value is `true`, results are returned in descending order. If the value is to `false`, results are returned in ascending order. The default value is `false`.
* `order_by` - (Optional) If the value is `LogStreamName`, the results are ordered by log stream name. If the value is `LastEventTime`, the results are ordered by the event time. The default value is `LogStreamName`.
* `region` - (Optional) Region to query. Defaults to provider region.
