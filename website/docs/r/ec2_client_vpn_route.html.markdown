---
subcategory: "VPN (Client)"
layout: "aws"
page_title: "AWS: aws_ec2_client_vpn_route"
description: |-
  Provides additional routes for AWS Client VPN endpoints.
---

# Resource: aws_ec2_client_vpn_route

Provides additional routes for AWS Client VPN endpoints. For more information on usage, please see the
[AWS Client VPN Administrator's Guide](https://docs.aws.amazon.com/vpn/latest/clientvpn-admin/what-is.html).

## Example Usage

```terraform
resource "aws_ec2_client_vpn_route" "example" {
  client_vpn_endpoint_id = aws_ec2_client_vpn_endpoint.example.id
  destination_cidr_block = "0.0.0.0/0"
  target_vpc_subnet_id   = aws_ec2_client_vpn_network_association.example.subnet_id
}

resource "aws_ec2_client_vpn_network_association" "example" {
  client_vpn_endpoint_id = aws_ec2_client_vpn_endpoint.example.id
  subnet_id              = aws_subnet.example.id
}

resource "aws_ec2_client_vpn_endpoint" "example" {
  description            = "Example Client VPN endpoint"
  server_certificate_arn = aws_acm_certificate.example.arn
  client_cidr_block      = "10.0.0.0/16"

  authentication_options {
    type                       = "certificate-authentication"
    root_certificate_chain_arn = aws_acm_certificate.example.arn
  }

  connection_log_options {
    enabled = false
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `client_vpn_endpoint_id` - (Required) The ID of the Client VPN endpoint.
* `destination_cidr_block` - (Required) The IPv4 address range, in CIDR notation, of the route destination.
* `description` - (Optional) A brief description of the route.
* `target_vpc_subnet_id` - (Required) The ID of the Subnet to route the traffic through. It must already be attached to the Client VPN.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The ID of the Client VPN endpoint.
* `origin` - Indicates how the Client VPN route was added. Will be `add-route` for routes created by this resource.
* `type` - The type of the route.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `4m`)
- `delete` - (Default `4m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import AWS Client VPN routes using the endpoint ID, target subnet ID, and destination CIDR block. All values are separated by a `,`. For example:

```terraform
import {
  to = aws_ec2_client_vpn_route.example
  id = "cvpn-endpoint-1234567890abcdef,subnet-9876543210fedcba,10.1.0.0/24"
}
```

Using `terraform import`, import AWS Client VPN routes using the endpoint ID, target subnet ID, and destination CIDR block. All values are separated by a `,`. For example:

```console
% terraform import aws_ec2_client_vpn_route.example cvpn-endpoint-1234567890abcdef,subnet-9876543210fedcba,10.1.0.0/24
```
