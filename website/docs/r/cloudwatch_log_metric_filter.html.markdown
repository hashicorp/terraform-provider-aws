---
subcategory: "CloudWatch Logs"
layout: "aws"
page_title: "AWS: aws_cloudwatch_log_metric_filter"
description: |-
  Provides a CloudWatch Log Metric Filter resource.
---

# Resource: aws_cloudwatch_log_metric_filter

Provides a CloudWatch Log Metric Filter resource.

## Example Usage

```terraform
resource "aws_cloudwatch_log_metric_filter" "yada" {
  name           = "MyAppAccessCount"
  pattern        = ""
  log_group_name = aws_cloudwatch_log_group.dada.name

  metric_transformation {
    name      = "EventCount"
    namespace = "YourNamespace"
    value     = "1"
  }
}

resource "aws_cloudwatch_log_group" "dada" {
  name = "MyApp/access.log"
}
```

## Argument Reference

This resource supports the following arguments:

* `name` - (Required) A name for the metric filter.
* `pattern` - (Required) A valid [CloudWatch Logs filter pattern](https://docs.aws.amazon.com/AmazonCloudWatch/latest/DeveloperGuide/FilterAndPatternSyntax.html)
  for extracting metric data out of ingested log events.
* `log_group_name` - (Required) The name of the log group to associate the metric filter with.
* `metric_transformation` - (Required) A block defining collection of information needed to define how metric data gets emitted. See below.

The `metric_transformation` block supports the following arguments:

* `name` - (Required) The name of the CloudWatch metric to which the monitored log information should be published (e.g., `ErrorCount`)
* `namespace` - (Required) The destination namespace of the CloudWatch metric.
* `value` - (Required) What to publish to the metric. For example, if you're counting the occurrences of a particular term like "Error", the value will be "1" for each occurrence. If you're counting the bytes transferred the published value will be the value in the log event.
* `default_value` - (Optional) The value to emit when a filter pattern does not match a log event. Conflicts with `dimensions`.
* `dimensions` - (Optional) Map of fields to use as dimensions for the metric. Up to 3 dimensions are allowed. Conflicts with `default_value`.
* `unit` - (Optional) The unit to assign to the metric. If you omit this, the unit is set as `None`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The name of the metric filter.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import CloudWatch Log Metric Filter using the `log_group_name:name`. For example:

```terraform
import {
  to = aws_cloudwatch_log_metric_filter.test
  id = "/aws/lambda/function:test"
}
```

Using `terraform import`, import CloudWatch Log Metric Filter using the `log_group_name:name`. For example:

```console
% terraform import aws_cloudwatch_log_metric_filter.test /aws/lambda/function:test
```
