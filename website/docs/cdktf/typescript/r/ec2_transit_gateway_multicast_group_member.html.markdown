---
subcategory: "Transit Gateway"
layout: "aws"
page_title: "AWS: aws_ec2_transit_gateway_multicast_group_member"
description: |-
  Manages an EC2 Transit Gateway Multicast Group Member
---

# Resource: aws_ec2_transit_gateway_multicast_group_member

Registers members (network interfaces) with the transit gateway multicast group.
A member is a network interface associated with a supported EC2 instance that receives multicast traffic.

## Example Usage

```terraform
resource "aws_ec2_transit_gateway_multicast_group_member" "example" {
  group_ip_address                    = "224.0.0.1"
  network_interface_id                = aws_network_interface.example.id
  transit_gateway_multicast_domain_id = aws_ec2_transit_gateway_multicast_domain.example.id
}
```

## Argument Reference

The following arguments are supported:

* `groupIpAddress` - (Required) The IP address assigned to the transit gateway multicast group.
* `networkInterfaceId` - (Required) The group members' network interface ID to register with the transit gateway multicast group.
* `transitGatewayMulticastDomainId` - (Required) The ID of the transit gateway multicast domain.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - EC2 Transit Gateway Multicast Group Member identifier.

<!-- cache-key: cdktf-0.17.0-pre.15 input-286c39b2f59b7534bbe97baf79dd0b9741f58e4a81a46b026a37ff2930718d60 -->