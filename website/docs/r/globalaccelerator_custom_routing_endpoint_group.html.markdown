---
subcategory: "Global Accelerator"
layout: "aws"
page_title: "AWS: aws_globalaccelerator_custom_routing_endpoint_group"
description: |-
  Provides a Global Accelerator custom routing endpoint group.
---

# Resource: aws_globalaccelerator_custom_routing_endpoint_group

Provides a Global Accelerator custom routing endpoint group.

## Example Usage

```terraform
resource "aws_globalaccelerator_custom_routing_endpoint_group" "example" {
  listener_arn = aws_globalaccelerator_custom_routing_listener.example.id

  destination_configuration {
    from_port = 80
    to_port   = 8080
    protocols = ["TCP"]
  }

  endpoint_configuration {
    endpoint_id = aws_subnet.example.id
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `listener_arn` - (Required) The Amazon Resource Name (ARN) of the custom routing listener.
* `destination_configuration` - (Required) The port ranges and protocols for all endpoints in a custom routing endpoint group to accept client traffic on. Fields documented below.
* `endpoint_configuration` - (Optional) The list of endpoint objects. Fields documented below.
* `endpoint_group_region` (Optional) - The name of the AWS Region where the custom routing endpoint group is located.

`destination_configuration` supports the following arguments:

* `from_port` - (Required) The first port, inclusive, in the range of ports for the endpoint group that is associated with a custom routing accelerator.
* `protocols` - (Required) The protocol for the endpoint group that is associated with a custom routing accelerator. The protocol can be either `"TCP"` or `"UDP"`.
* `to_port` - (Required) The last port, inclusive, in the range of ports for the endpoint group that is associated with a custom routing accelerator.

`endpoint_configuration` supports the following arguments:

* `endpoint_id` - (Optional) An ID for the endpoint. For custom routing accelerators, this is the virtual private cloud (VPC) subnet ID.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The Amazon Resource Name (ARN) of the custom routing endpoint group.
* `arn` - The Amazon Resource Name (ARN) of the custom routing endpoint group.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Global Accelerator custom routing endpoint groups using the `id`. For example:

```terraform
import {
  to = aws_globalaccelerator_custom_routing_endpoint_group.example
  id = "arn:aws:globalaccelerator::111111111111:accelerator/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx/listener/xxxxxxx/endpoint-group/xxxxxxxx"
}
```

Using `terraform import`, import Global Accelerator custom routing endpoint groups using the `id`. For example:

```console
% terraform import aws_globalaccelerator_custom_routing_endpoint_group.example arn:aws:globalaccelerator::111111111111:accelerator/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx/listener/xxxxxxx/endpoint-group/xxxxxxxx
```
