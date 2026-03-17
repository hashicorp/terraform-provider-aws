---
subcategory: "VPN (Site-to-Site)"
layout: "aws"
page_title: "AWS: aws_vpn_connection"
description: |-
  Fetches details of a Site-to-Site VPN connection. A Site-to-Site VPN connection is an Internet Protocol security (IPsec) VPN connection between a VPC and an on-premises network.
---

# Data Source: aws_vpn_connection

Fetches details of a Site-to-Site VPN connection. A Site-to-Site VPN connection is an Internet Protocol security (IPsec) VPN connection between a VPC and an on-premises network.

## Example Usage

### Basic Usage

```terraform
data "aws_vpn_connection" "example" {
  filter {
    name   = "customer-gateway-id"
    values = ["cgw-1234567890"]
  }
}

output "vpn_connection_id" {
  value = data.aws_vpn_connection.example.vpn_connection_id
}
```

### Find by VPN Connection ID

```terraform
data "aws_vpn_connection" "example" {
  vpn_connection_id = "vpn-abcd1234567890"
}

output "gateway_association_state" {
  value = data.aws_vpn_connection.example.gateway_association_state
}
```

## Argument Reference

This data source supports the following arguments:

* `vpn_connection_id` - (Optional) Identifier of the EC2 VPN Connection.
* `filter` - (Optional) Configuration block(s) for filtering. Detailed below.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

### Filter Configuration Block

The `filter` configuration block supports the following arguments:

* `name` - (Required) Name of the filter field. Valid values can be found in the [EC2 `DescribeVPNConnections` API Reference](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeVpnConnections.html).
* `values` - (Required) Set of values that are accepted for the given filter field. Results will be selected if any given value matches.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `category` - Category of the VPN connection. A value of VPN indicates an AWS VPN connection. A value of VPN-Classic indicates an AWS Classic VPN connection.
* `core_network_arn` - ARN of the core network.
* `core_network_attachment_arn` - ARN of the core network attachment.
* `customer_gateway_configuration` - Configuration information for the VPN connection's customer gateway (in the native XML format).
* `customer_gateway_id` - ID of the customer gateway at your end of the VPN connection.
* `gateway_association_state` - Current state of the gateway association.
* `pre_shared_key_arn` - (ARN) of the Secrets Manager secret storing the pre-shared key(s) for the VPN connection.
* `routes` - List of static routes associated with the VPN connection.
* `state` - Current state of the VPN connection.
* `tags` - Tags associated to the VPN Connection.
* `transit_gateway_id` - ID of a transit gateway associated with the VPN connection.
* `type` - Type of VPN connection. Currently the only supported type is ipsec.1.
* `vgw_telemetries` - List of objects containing information about the VPN tunnel.
* `vpn_concentrator_id` - ID of a VPN concentrator associated with the VPN connection.
* `vpn_gateway_id` - ID of a virtual private gateway associated with the VPN connection.
