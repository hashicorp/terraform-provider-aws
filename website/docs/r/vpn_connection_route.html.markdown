---
subcategory: "VPN (Site-to-Site)"
layout: "aws"
page_title: "AWS: aws_vpn_connection_route"
description: |-
  Provides a static route between a VPN connection and a customer gateway.
---

# Resource: aws_vpn_connection_route

Provides a static route between a VPN connection and a customer gateway.

## Example Usage

```terraform
resource "aws_vpc" "vpc" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_vpn_gateway" "vpn_gateway" {
  vpc_id = aws_vpc.vpc.id
}

resource "aws_customer_gateway" "customer_gateway" {
  bgp_asn    = 65000
  ip_address = "172.0.0.1"
  type       = "ipsec.1"
}

resource "aws_vpn_connection" "main" {
  vpn_gateway_id      = aws_vpn_gateway.vpn_gateway.id
  customer_gateway_id = aws_customer_gateway.customer_gateway.id
  type                = "ipsec.1"
  static_routes_only  = true
}

resource "aws_vpn_connection_route" "office" {
  destination_cidr_block = "192.168.10.0/24"
  vpn_connection_id      = aws_vpn_connection.main.id
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `destination_cidr_block` - (Required) The CIDR block associated with the local subnet of the customer network.
* `vpn_connection_id` - (Required) The ID of the VPN connection.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `destination_cidr_block` - The CIDR block associated with the local subnet of the customer network.
* `vpn_connection_id` - The ID of the VPN connection.
