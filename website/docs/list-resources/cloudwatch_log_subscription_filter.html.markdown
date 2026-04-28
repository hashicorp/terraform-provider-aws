---
subcategory: "CloudWatch Logs"
layout: "aws"
page_title: "AWS: aws_cloudwatch_log_subscription_filter"
description: |-
  Lists CloudWatch Logs Subscription Filter resources.
---

# List Resource: aws_cloudwatch_log_subscription_filter

Lists CloudWatch Logs Subscription Filter resources.

## Example Usage

```terraform
list "aws_cloudwatch_log_subscription_filter" "example" {
  provider = aws

  log_group_name = "example-group"
}
```

## Argument Reference

This list resource supports the following arguments:

* `log_group_name` - (Required) Name of the log group.
* `region` - (Optional) Region to query. Defaults to provider region.
