---
subcategory: "Transit Gateway"
layout: "aws"
page_title: "AWS: aws_ec2_transit_gateway_connect_peer"
description: |-
  Get information on an EC2 Transit Gateway Connect Peer
---

# Data Source: aws_ec2_transit_gateway_connect_peer

Get information on an EC2 Transit Gateway Connect Peer.

## Example Usage

### By Filter

```terraform
data "aws_ec2_transit_gateway_connect_peer" "example" {
  filter {
    name   = "transit-gateway-attachment-id"
    values = ["tgw-attach-12345678"]
  }
}
```

### By Identifier

```terraform
data "aws_ec2_transit_gateway_connect_peer" "example" {
  transit_gateway_connect_peer_id = "tgw-connect-peer-12345678"
}
```

## Argument Reference

This data source supports the following arguments:

* `filter` - (Optional) One or more configuration blocks containing name-values filters. Detailed below.
* `transit_gateway_connect_peer_id` - (Optional) Identifier of the EC2 Transit Gateway Connect Peer.

### filter Argument Reference

* `name` - (Required) Name of the filter.
* `values` - (Required) List of one or more values for the filter.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - EC2 Transit Gateway Connect Peer ARN
* `bgp_asn` - BGP ASN number assigned customer device
* `bgp_peer_address` - The IP address assigned to customer device, which is used as BGP IP address.
* `bgp_transit_gateway_addresses` - The IP addresses assigned to Transit Gateway, which are used as BGP IP addresses.
* `inside_cidr_blocks` - CIDR blocks that will be used for addressing within the tunnel.
* `peer_address` - IP addressed assigned to customer device, which is used as tunnel endpoint
* `tags` - Key-value tags for the EC2 Transit Gateway Connect Peer
* `transit_gateway_address` - The IP address assigned to Transit Gateway, which is used as tunnel endpoint.
* `transit_gateway_attachment_id` - The Transit Gateway Connect

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `read` - (Default `20m`)
