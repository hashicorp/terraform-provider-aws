---
subcategory: "Direct Connect"
layout: "aws"
page_title: "AWS: aws_dx_transit_virtual_interface"
description: |-
  Provides a Direct Connect transit virtual interface resource.
---

# Resource: aws_dx_transit_virtual_interface

Provides a Direct Connect transit virtual interface resource.
A transit virtual interface is a VLAN that transports traffic from a [Direct Connect gateway](dx_gateway.html) to one or more [transit gateways](ec2_transit_gateway.html).

## Example Usage

```terraform
resource "aws_dx_gateway" "example" {
  name            = "tf-dxg-example"
  amazon_side_asn = 64512
}

resource "aws_dx_transit_virtual_interface" "example" {
  connection_id = aws_dx_connection.example.id

  dx_gateway_id  = aws_dx_gateway.example.id
  name           = "tf-transit-vif-example"
  vlan           = 4094
  address_family = "ipv4"
  bgp_asn        = 65352
}
```

## Argument Reference

This resource supports the following arguments:

* `address_family` - (Required) The address family for the BGP peer. `ipv4 ` or `ipv6`.
* `bgp_asn` - (Required) The autonomous system (AS) number for Border Gateway Protocol (BGP) configuration.
* `connection_id` - (Required) The ID of the Direct Connect connection (or LAG) on which to create the virtual interface.
* `dx_gateway_id` - (Required) The ID of the Direct Connect gateway to which to connect the virtual interface.
* `name` - (Required) The name for the virtual interface.
* `vlan` - (Required) The VLAN ID.
* `amazon_address` - (Optional) The IPv4 CIDR address to use to send traffic to Amazon. Required for IPv4 BGP peers.
* `bgp_auth_key` - (Optional) The authentication key for BGP configuration.
* `customer_address` - (Optional) The IPv4 CIDR destination address to which Amazon should send traffic. Required for IPv4 BGP peers.
* `mtu` - (Optional) The maximum transmission unit (MTU) is the size, in bytes, of the largest permissible packet that can be passed over the connection.
The MTU of a virtual transit interface can be either `1500` or `8500` (jumbo frames). Default is `1500`.
* `sitelink_enabled` - (Optional) Indicates whether to enable or disable SiteLink.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The ID of the virtual interface.
* `arn` - The ARN of the virtual interface.
* `aws_device` - The Direct Connect endpoint on which the virtual interface terminates.
* `jumbo_frame_capable` - Indicates whether jumbo frames (8500 MTU) are supported.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `10m`)
- `update` - (Default `10m`)
- `delete` - (Default `10m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Direct Connect transit virtual interfaces using the VIF `id`. For example:

```terraform
import {
  to = aws_dx_transit_virtual_interface.test
  id = "dxvif-33cc44dd"
}
```

Using `terraform import`, import Direct Connect transit virtual interfaces using the VIF `id`. For example:

```console
% terraform import aws_dx_transit_virtual_interface.test dxvif-33cc44dd
```
