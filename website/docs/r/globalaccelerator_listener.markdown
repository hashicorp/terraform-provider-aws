---
layout: "aws"
page_title: "AWS: aws_globalaccelerator_listener"
sidebar_current: "docs-aws-resource-globalaccelerator-listener"
description: |-
  Provides a Global Accelerator listener.
---

# aws_globalaccelerator_listener

Provides a Global Accelerator listener.

## Example Usage

```hcl
resource "aws_globalaccelerator_accelerator" "example" {
  name            = "Example"
  ip_address_type = "IPV4"
  enabled         = true

  attributes {
    flow_logs_enabled   = true
    flow_logs_s3_bucket = "example-bucket"
    flow_logs_s3_prefix = "flow-logs/"
  }
}

resource "aws_globalaccelerator_listener" "example" {
  accelerator_arn = "${aws_globalaccelerator_accelerator.example.id}"
  client_affinity = "SOURCE_IP"
  protocol        = "TCP"

  port_range {
    from_port = 80
    to_port   = 80
  }
}
```

## Argument Reference

The following arguments are supported:

* `accelerator_arn` - (Required) The Amazon Resource Name (ARN) of your accelerator.
* `client_affinity` - (Optional) The value for the address type must be `IPV4`. Valid values are `NONE`, `SOURCE_IP`.
* `protocol` - (Optional) The protocol for the connections from clients to the accelerator. Valid values are `TCP`, `UDP`.
* `port_range` - (Optional) The list of port ranges for the connections from clients to the accelerator. Fields documented below.

**port_range** supports the following attributes:

* `from_port` - (Optional) The first port in the range of ports, inclusive.
* `to_port` - (Optional) The last port in the range of ports, inclusive.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The Amazon Resource Name (ARN) of the listener.

## Import

Global Accelerator listeners can be imported using the `id`, e.g.

```
$ terraform import aws_globalaccelerator_listener.example arn:aws:globalaccelerator::111111111111:listener/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
```
