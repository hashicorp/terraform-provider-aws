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

This resource supports the following arguments:

* `accelerator_arn` - (Required) The Amazon Resource Name (ARN) of a custom routing accelerator.
* `port_range` - (Optional) The list of port ranges for the connections from clients to the accelerator. Fields documented below.

`port_range` supports the following arguments:

* `from_port` - (Optional) The first port in the range of ports, inclusive.
* `to_port` - (Optional) The last port in the range of ports, inclusive.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The Amazon Resource Name (ARN) of the custom routing listener.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Global Accelerator custom routing listeners using the `id`. For example:

```terraform
import {
  to = aws_globalaccelerator_custom_routing_listener.example
  id = "arn:aws:globalaccelerator::111111111111:accelerator/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx/listener/xxxxxxxx"
}
```

Using `terraform import`, import Global Accelerator custom routing listeners using the `id`. For example:

```console
% terraform import aws_globalaccelerator_custom_routing_listener.example arn:aws:globalaccelerator::111111111111:accelerator/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx/listener/xxxxxxxx
```
