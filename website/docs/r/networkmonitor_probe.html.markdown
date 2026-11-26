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

This resource supports the following arguments:

- `destination` - (Required) Destination IP address. This must be either IPV4 or IPV6.
- `destination_port` - (Optional) Port associated with the destination. This is required only if the protocol is TCP and must be a number between 1 and 65536.
- `monitor_name` - (Required) Name of the monitor.
- `packet_size` - (Optional) Size of the packets sent between the source and destination. This must be a number between 56 and 8500.
- `protocol` - (Required) Protocol used for the network traffic between the source and destination. This must be either TCP or ICMP.
- `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
- `source_arn` - (Required) ARN of the subnet.
- `tags` - (Optional) Key-value tags for the monitor. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

- `address_family` - Address family of the probe. Valid values are `IPV4` and `IPV6`.
- `arn` - ARN of the attachment.
- `probe_id` - ID of the probe.
- `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
- `vpc_id` - ID of the VPC associated with the probe's subnet.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_networkmonitor_probe` using the monitor name and probe ID. For example:

```terraform
import {
  to = aws_networkmonitor_probe.example
  id = "monitor-7786087912324693644,probe-3qm8p693i4fi1h8lqylzkbp42e"
}
```

Using `terraform import`, import `aws_networkmonitor_probe` using the monitor name and probe ID. For example:

```console
% terraform import aws_networkmonitor_probe.example monitor-7786087912324693644,probe-3qm8p693i4fi1h8lqylzkbp42e
```
