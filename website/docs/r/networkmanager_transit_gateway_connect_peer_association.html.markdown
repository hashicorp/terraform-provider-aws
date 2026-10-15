---
subcategory: "Network Manager"
layout: "aws"
page_title: "AWS: aws_networkmanager_transit_gateway_connect_peer_association"
description: |-
  Manages a Network Manager transit gateway Connect peer association.
---

# Resource: aws_networkmanager_transit_gateway_connect_peer_association

Manages a Network Manager transit gateway Connect peer association. Associates a transit gateway Connect peer with a device, and optionally, with a link. If you specify a link, it must be associated with the specified device.

## Example Usage

```terraform
resource "aws_networkmanager_transit_gateway_connect_peer_association" "example" {
  global_network_id                = aws_networkmanager_global_network.example.id
  device_id                        = aws_networkmanager_device.example.id
  transit_gateway_connect_peer_arn = aws_ec2_transit_gateway_connect_peer.example.arn
}
```

## Argument Reference

The following arguments are required:

* `device_id` - (Required) ID of the device.
* `global_network_id` - (Required) ID of the global network.
* `transit_gateway_connect_peer_arn` - (Required) ARN of the Connect peer.

The following arguments are optional:

* `link_id` - (Optional) ID of the link.

## Attribute Reference

This resource exports no additional attributes.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `10m`)
* `delete` - (Default `10m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_networkmanager_transit_gateway_connect_peer_association` using the global network ID and Connect peer ARN. For example:

```terraform
import {
  to = aws_networkmanager_transit_gateway_connect_peer_association.example
  id = "global-network-0d47f6t230mz46dy4,arn:aws:ec2:us-west-2:123456789012:transit-gateway-connect-peer/tgw-connect-peer-12345678"
}
```

Using `terraform import`, import `aws_networkmanager_transit_gateway_connect_peer_association` using the global network ID and Connect peer ARN. For example:

```console
% terraform import aws_networkmanager_transit_gateway_connect_peer_association.example global-network-0d47f6t230mz46dy4,arn:aws:ec2:us-west-2:123456789012:transit-gateway-connect-peer/tgw-connect-peer-12345678
```
