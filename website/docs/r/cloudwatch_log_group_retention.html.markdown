---
subcategory: "CloudWatch Logs"
layout: "aws"
page_title: "AWS: aws_cloudwatch_log_group_retention"
description: |-
  Manages a CloudWatch Log Group retention policy.
---

# Resource: aws_cloudwatch_log_group_retention

Manages a CloudWatch Log Group retention policy. This resource allows you to set or manage the retention policy for an existing log group without affecting other properties of the log group.

~> **Note:** When this resource is deleted, the retention policy is removed from the log group, and log events will be retained indefinitely.

~> **Warning:** Do not use this resource together with `aws_cloudwatch_log_group` resource for the same log group. Both resources manage the same retention policy, which will cause configuration drift and unpredictable behavior. Use either the `aws_cloudwatch_log_group` resource with `retention_in_days` configured, or use this `aws_cloudwatch_log_group_retention` resource, but not both.

## Example Usage

```terraform
data "aws_cloudwatch_log_group" "example" {
  name = "example"
}

resource "aws_cloudwatch_log_group_retention" "example" {
  log_group_name    = data.aws_cloudwatch_log_group.example.name
  retention_in_days = 7
}
```

## Argument Reference

This resource supports the following arguments:

* `log_group_name` - (Required) The name of the log group. If the log group does not exist, an error will be returned.
* `retention_in_days` - (Required) Specifies the number of days you want to retain log events in the specified log group. Possible values are: 1, 3, 5, 7, 14, 30, 60, 90, 120, 150, 180, 365, 400, 545, 731, 1096, 1827, 2192, 2557, 2922, 3288, 3653.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The name of the log group.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import CloudWatch Logs Log Group retention policies using the log group name. For example:

```terraform
import {
  to = aws_cloudwatch_log_group_retention.example
  id = "example"
}
```

Using `terraform import`, import CloudWatch Logs Log Group retention policies using the log group name. For example:

```console
% terraform import aws_cloudwatch_log_group_retention.example example
```
