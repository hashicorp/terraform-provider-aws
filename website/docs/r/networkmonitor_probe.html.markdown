---
subcategory: "CloudWatch Network Monitor"
layout: "aws"
page_title: "AWS: aws_networkmonitor_probe"
description: |-
  Terraform resource for managing an Amazon Network Monitor Probe.
---

# Resource: aws_networkmonitor_probe

Terraform resource for managing an AWS Network Monitor Probe.

## Example Usage

### Basic Usage

```terraform
resource "aws_networkmonitor_monitor" "example" {
  aggregation_period = 30
  monitor_name       = "example"
}

resource "aws_networkmonitor_probe" "example" {
  monitor_name     = aws_networkmonitor_monitor.example.monitor_name
  destination      = "127.0.0.1"
  destination_port = 80
  protocol         = "TCP"
  source_arn       = aws_subnet.example.arn
  packet_size      = 200
}
```

## Argument Reference

The following arguments are required:

- `destination` - (Required) The destination IP address. This must be either IPV4 or IPV6.
- `destination_port` - (Optional) The port associated with the destination. This is required only if the protocol is TCP and must be a number between 1 and 65536.
- `monitor_name` - (Required) The name of the monitor.
- `protocol` - (Required) The protocol used for the network traffic between the source and destination. This must be either TCP or ICMP.
- `source_arn` - (Required) The ARN of the subnet.
- `packet_size` - (Optional) The size of the packets sent between the source and destination. This must be a number between 56 and 8500.

The following arguments are optional:

- `tags` - (Optional) Key-value tags for the monitor. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

- `arn` - The ARN of the attachment.
- `source_arn` - The ARN of the subnet.
- `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_networkmonitor_probe` using the monitor name and probe id. For example:

```terraform
import {
  to = aws_networkmonitor_probe.example
  id = "monitor-7786087912324693644,probe-3qm8p693i4fi1h8lqylzkbp42e"
}
```

Using `terraform import`, import `aws_networkmonitor_probe` using the monitor name and probe id. For example:

```console
% terraform import aws_networkmonitor_probe.example monitor-7786087912324693644,probe-3qm8p693i4fi1h8lqylzkbp42e
```
