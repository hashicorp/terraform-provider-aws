---
subcategory: "Global Accelerator"
layout: "aws"
page_title: "AWS: aws_globalaccelerator_custom_routing_listener"
description: |-
  Provides a Global Accelerator custom routing listener.
---

# Resource: aws_globalaccelerator_custom_routing_listener

Provides a Global Accelerator custom routing listener.

## Example Usage

```terraform
resource "aws_globalaccelerator_custom_routing_accelerator" "example" {
  name            = "Example"
  ip_address_type = "IPV4"
  enabled         = true

  attributes {
    flow_logs_enabled   = true
    flow_logs_s3_bucket = "example-bucket"
    flow_logs_s3_prefix = "flow-logs/"
  }
}

resource "aws_globalaccelerator_custom_routing_listener" "example" {
  accelerator_arn = aws_globalaccelerator_custom_routing_accelerator.example.id

  port_range {
    from_port = 80
    to_port   = 80
  }
}
```

## Argument Reference

The following arguments are supported:

* `accelerator_arn` - (Required) The Amazon Resource Name (ARN) of a custom routing accelerator.
* `port_range` - (Optional) The list of port ranges for the connections from clients to the accelerator. Fields documented below.

**port_range** supports the following attributes:

* `from_port` - (Optional) The first port in the range of ports, inclusive.
* `to_port` - (Optional) The last port in the range of ports, inclusive.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The Amazon Resource Name (ARN) of the custom routing listener.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

Global Accelerator custom routing listeners can be imported using the `id`, e.g.,

```
$ terraform import aws_globalaccelerator_custom_routing_listener.example arn:aws:globalaccelerator::111111111111:accelerator/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx/listener/xxxxxxxx
```
