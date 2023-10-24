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

This resource supports the following arguments:

* `group_ip_address` - (Required) The IP address assigned to the transit gateway multicast group.
* `network_interface_id` - (Required) The group members' network interface ID to register with the transit gateway multicast group.
* `transit_gateway_multicast_domain_id` - (Required) The ID of the transit gateway multicast domain.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - EC2 Transit Gateway Multicast Group Member identifier.
