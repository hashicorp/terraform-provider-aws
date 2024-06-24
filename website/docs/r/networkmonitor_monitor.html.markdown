---
subcategory: "CloudWatch Network Monitor"
layout: "aws"
page_title: "AWS: aws_networkmonitor_monitor"
description: |-
  Terraform resource for managing an Amazon Network Monitor Monitor.
---

# Resource: aws_networkmonitor_monitor

Terraform resource for managing an AWS Network Monitor Monitor.

## Example Usage

### Basic Usage

```terraform
resource "aws_networkmonitor_monitor" "example" {
  aggregation_period = 30
  monitor_name       = "example"
}
```

## Argument Reference

The following arguments are required:

- `monitor_name` - (Required) The name of the monitor.

The following arguments are optional:

- `aggregation_period` - (Optional) The time, in seconds, that metrics are aggregated and sent to Amazon CloudWatch. Valid values are either 30 or 60.
- `tags` - (Optional) Key-value tags for the monitor. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

- `arn` - The ARN of the monitor.
- `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_networkmonitor_monitor` using the monitor name. For example:

```terraform
import {
  to = aws_networkmonitor_monitor.example
  id = "monitor-7786087912324693644"
}
```

Using `terraform import`, import `aws_networkmonitor_monitor` using the monitor name. For example:

```console
% terraform import aws_networkmonitor_monitor.example monitor-7786087912324693644
```
