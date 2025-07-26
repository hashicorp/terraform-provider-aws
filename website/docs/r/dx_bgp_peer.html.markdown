---
subcategory: "Direct Connect"
layout: "aws"
page_title: "AWS: aws_dx_bgp_peer"
description: |-
  Provides a Direct Connect BGP peer resource.
---

# Resource: aws_dx_bgp_peer

Provides a Direct Connect BGP peer resource.

## Example Usage

```terraform
resource "aws_dx_bgp_peer" "peer" {
  virtual_interface_id = aws_dx_private_virtual_interface.foo.id
  address_family       = "ipv6"
  bgp_asn              = 65351
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `address_family` - (Required) The address family for the BGP peer. `ipv4 ` or `ipv6`.
* `bgp_asn` - (Required) The autonomous system (AS) number for Border Gateway Protocol (BGP) configuration.
* `virtual_interface_id` - (Required) The ID of the Direct Connect virtual interface on which to create the BGP peer.
* `amazon_address` - (Optional) The IPv4 CIDR address to use to send traffic to Amazon.
Required for IPv4 BGP peers on public virtual interfaces.
* `bgp_auth_key` - (Optional) The authentication key for BGP configuration.
* `customer_address` - (Optional) The IPv4 CIDR destination address to which Amazon should send traffic.
Required for IPv4 BGP peers on public virtual interfaces.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The ID of the BGP peer resource.
* `bgp_status` - The Up/Down state of the BGP peer.
* `bgp_peer_id` - The ID of the BGP peer.
* `aws_device` - The Direct Connect endpoint on which the BGP peer terminates.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `10m`)
- `delete` - (Default `10m`)
