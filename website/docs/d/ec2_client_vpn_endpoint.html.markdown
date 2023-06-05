---
subcategory: "VPN (Client)"
layout: "aws"
page_title: "AWS: aws_ec2_client_vpn_endpoint"
description: |-
  Get information on an EC2 Client VPN endpoint
---

# Data Source: aws_ec2_client_vpn_endpoint

Get information on an EC2 Client VPN endpoint.

## Example Usage

### By Filter

```hcl
data "aws_ec2_client_vpn_endpoint" "example" {
  filter {
    name   = "tag:Name"
    values = ["ExampleVpn"]
  }
}
```

### By Identifier

```hcl
data "aws_ec2_client_vpn_endpoint" "example" {
  client_vpn_endpoint_id = "cvpn-endpoint-083cf50d6eb314f21"
}
```

## Argument Reference

The following arguments are supported:

* `client_vpn_endpoint_id` - (Optional) ID of the Client VPN endpoint.
* `filter` - (Optional) One or more configuration blocks containing name-values filters. Detailed below.
* `tags` - (Optional) Map of tags, each pair of which must exactly match a pair on the desired endpoint.

### filter

This block allows for complex filters. You can use one or more `filter` blocks.

The following arguments are required:

* `name` - (Required) Name of the field to filter by, as defined by [the underlying AWS API](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeClientVpnEndpoints.html).
* `values` - (Required) Set of values that are accepted for the given field. An endpoint will be selected if any one of the given values matches.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `arn` -  The ARN of the Client VPN endpoint.
* `authentication_options` - Information about the authentication method used by the Client VPN endpoint.
* `client_cidr_block` - IPv4 address range, in CIDR notation, from which client IP addresses are assigned.
* `client_connect_options` - The options for managing connection authorization for new client connections.
* `client_login_banner_options` - Options for enabling a customizable text banner that will be displayed on AWS provided clients when a VPN session is established.
* `connection_log_options` - Information about the client connection logging options for the Client VPN endpoint.
* `description` - Brief description of the endpoint.
* `dns_name` - DNS name to be used by clients when connecting to the Client VPN endpoint.
* `dns_servers` - Information about the DNS servers to be used for DNS resolution.
* `security_group_ids` - IDs of the security groups for the target network associated with the Client VPN endpoint.
* `self_service_portal` - Whether the self-service portal for the Client VPN endpoint is enabled.
* `server_certificate_arn` - The ARN of the server certificate.
* `session_timeout_hours` - The maximum VPN session duration time in hours.
* `split_tunnel` - Whether split-tunnel is enabled in the AWS Client VPN endpoint.
* `transport_protocol` - Transport protocol used by the Client VPN endpoint.
* `vpc_id` - ID of the VPC associated with the Client VPN endpoint.
* `vpn_port` - Port number for the Client VPN endpoint.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `read` - (Default `20m`)
