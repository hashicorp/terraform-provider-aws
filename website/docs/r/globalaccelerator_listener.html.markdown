---
subcategory: "Global Accelerator"
layout: "aws"
page_title: "AWS: aws_globalaccelerator_listener"
description: |-
  Provides a Global Accelerator listener.
---

# Resource: aws_globalaccelerator_listener

Provides a Global Accelerator listener.

## Example Usage

```terraform
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
  accelerator_arn = aws_globalaccelerator_accelerator.example.id
  client_affinity = "SOURCE_IP"
  protocol        = "TCP"

  port_range {
    from_port = 80
    to_port   = 80
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `accelerator_arn` - (Required) The Amazon Resource Name (ARN) of your accelerator.
* `client_affinity` - (Optional) Direct all requests from a user to the same endpoint. Valid values are `NONE`, `SOURCE_IP`. Default: `NONE`. If `NONE`, Global Accelerator uses the "five-tuple" properties of source IP address, source port, destination IP address, destination port, and protocol to select the hash value. If `SOURCE_IP`, Global Accelerator uses the "two-tuple" properties of source (client) IP address and destination IP address to select the hash value.
* `protocol` - (Optional) The protocol for the connections from clients to the accelerator. Valid values are `TCP`, `UDP`.
* `port_range` - (Optional) The list of port ranges for the connections from clients to the accelerator. Fields documented below.

`port_range` supports the following arguments:

* `from_port` - (Optional) The first port in the range of ports, inclusive.
* `to_port` - (Optional) The last port in the range of ports, inclusive.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The Amazon Resource Name (ARN) of the listener.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Global Accelerator listeners using the `id`. For example:

```terraform
import {
  to = aws_globalaccelerator_listener.example
  id = "arn:aws:globalaccelerator::111111111111:accelerator/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx/listener/xxxxxxxx"
}
```

Using `terraform import`, import Global Accelerator listeners using the `id`. For example:

```console
% terraform import aws_globalaccelerator_listener.example arn:aws:globalaccelerator::111111111111:accelerator/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx/listener/xxxxxxxx
```
