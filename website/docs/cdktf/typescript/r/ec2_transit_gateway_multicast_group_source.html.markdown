---
subcategory: "Transit Gateway"
layout: "aws"
page_title: "AWS: aws_ec2_transit_gateway_multicast_group_source"
description: |-
  Manages an EC2 Transit Gateway Multicast Group Source
---

# Resource: aws_ec2_transit_gateway_multicast_group_source

Registers sources (network interfaces) with the transit gateway multicast group.
A multicast source is a network interface attached to a supported instance that sends multicast traffic.

## Example Usage

```terraform
resource "aws_ec2_transit_gateway_multicast_group_source" "example" {
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

<!-- cache-key: cdktf-0.17.0-pre.15 input-65215e11717df1869097637d2ac9989417c165158f1d4b33781b1f0782026494 -->