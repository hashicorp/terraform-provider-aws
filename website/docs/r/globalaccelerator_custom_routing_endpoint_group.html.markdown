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
  listener_arn = aws_globalaccelerator_custom_routing_listener.example.arn

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

* `destination_configuration` - (Required) Port ranges and protocols for all endpoints in a custom routing endpoint group to accept client traffic on. See [`destination_configuration`](#destination_configuration-block) below.
* `endpoint_configuration` - (Optional) Endpoint objects. See [`endpoint_configuration`](#endpoint_configuration-block) below.
* `endpoint_group_region` - (Optional) AWS Region where the custom routing endpoint group is located.
* `listener_arn` - (Required) ARN of the custom routing listener.

### `destination_configuration` Block

* `from_port` - (Required) First port, inclusive, in the range of ports for the endpoint group that is associated with a custom routing accelerator.
* `protocols` - (Required) Protocol for the endpoint group that is associated with a custom routing accelerator. The protocol can be either `"TCP"` or `"UDP"`.
* `to_port` - (Required) Last port, inclusive, in the range of ports for the endpoint group that is associated with a custom routing accelerator.

### `endpoint_configuration` Block

* `endpoint_id` - (Optional) ID for the endpoint. For custom routing accelerators, this is the VPC subnet ID.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the custom routing endpoint group.
* `id` - ARN of the custom routing endpoint group.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_globalaccelerator_custom_routing_endpoint_group.example
  identity = {
    "arn" = "arn:aws:globalaccelerator::123456789012:accelerator/1234abcd-abcd-1234-abcd-1234abcdefgh/listener/0123vxyz/endpoint-group/098765zyxwvu"
  }
}

resource "aws_globalaccelerator_custom_routing_endpoint_group" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

- `arn` (String) ARN of the Global Accelerator custom routing endpoint group.

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
