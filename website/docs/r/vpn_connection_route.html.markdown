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

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_vpn_connection_route.office
  identity = {
    destination_cidr_block = "192.168.10.0/24"
    vpn_connection_id      = "vpn-12345678"
  }
}

resource "aws_vpn_connection_route" "office" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `destination_cidr_block` (String) CIDR block associated with the local subnet of the customer network.
* `vpn_connection_id` (String) ID of the VPN connection.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import VPN Connection Routes using the `destination_cidr_block` and `vpn_connection_id` separated by a colon (`:`). For example:

```terraform
import {
  to = aws_vpn_connection_route.office
  id = "192.168.10.0/24:vpn-12345678"
}
```

Using `terraform import`, import VPN Connection Routes using the `destination_cidr_block` and `vpn_connection_id` separated by a colon (`:`). For example:

```console
% terraform import aws_vpn_connection_route.office 192.168.10.0/24:vpn-12345678
```
