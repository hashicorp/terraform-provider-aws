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

* `descending` - (Optional) Whether to return results in descending order. Defaults to `false`.
* `log_group_name` - (Required) Name of the log group.
* `order_by` - (Optional) Method used to sort the log streams. Valid values are `LogStreamName` or `LastEventTime`. Defaults to `LogStreamName`.
* `region` - (Optional) Region to query. Defaults to provider region.
