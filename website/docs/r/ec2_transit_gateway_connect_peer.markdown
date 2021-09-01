---
subcategory: "EC2"
layout: "aws"
page_title: "AWS: aws_ec2_transit_gateway_connect_peer"
description: |-
  Manages a Connect peer for a specified transit gateway Connect attachment between a transit gateway and an appliance.
---

# Resource: aws_ec2_transit_gateway_connect_peer

Manages a Connect peer for a specified transit gateway Connect attachment between a transit gateway and an appliance.

## Example Usage

```terraform
resource "aws_ec2_transit_gateway_connect_peer" "test" {
  transit_gateway_attachment_id = aws_ec2_transit_gateway_connect.test.id
  inside_cidr_blocks = ["169.254.10.0/29"]
  peer_address       = "10.0.0.4"
}
```

## Argument Reference

The following arguments are supported:

* `transit_gateway_attachment_id` - (Required) The ID of the Connect attachment.
* `inside_cidr_blocks` - (Required) The range of inside IP addresses that are used for BGP peering. You must specify a size /29 IPv4 CIDR block from the 169.254.0.0/16 range.
* `peer_address` - (Required) The peer IP address (GRE outer IP address) on the appliance side of the Connect peer.
* `peer_asn` - (Optional) The peer Autonomous System Number (ASN).
* `transit_gateway_address` - (Optional) The peer IP address (GRE outer IP address) on the transit gateway side of the Connect peer, which must be specified from a transit gateway CIDR block.
* `tags` - (Optional) Key-value tags for the EC2 Transit Gateway Connect Attachment. If configured with a provider [`default_tags` configuration block](https://www.terraform.io/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - EC2 Transit Gateway Attachment identifier
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://www.terraform.io/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

`aws_ec2_transit_gateway_connect_peer` can be imported by using the EC2 Transit Gateway Connect peer identifier, e.g.

```
$ terraform import aws_ec2_transit_gateway_connect_peer.example tgw-connect-peer-1234567890123c567
```
