---
subcategory: "Network Manager"
layout: "aws"
page_title: "AWS: aws_networkmanager_customer_gateway_association"
description: |-
  Associates a customer gateway with a device and optionally, with a link.
---

# Resource: aws_networkmanager_customer_gateway_association

Associates a customer gateway with a device and optionally, with a link.
If you specify a link, it must be associated with the specified device.

## Example Usage

```terraform
resource "aws_networkmanager_global_network" "example" {
  description = "example"
}

resource "aws_networkmanager_site" "example" {
  global_network_id = aws_networkmanager_global_network.example.id
}

resource "aws_networkmanager_device" "example" {
  global_network_id = aws_networkmanager_global_network.example.id
  site_id           = aws_networkmanager_site.example.id
}

resource "aws_customer_gateway" "example" {
  bgp_asn    = 65000
  ip_address = "172.83.124.10"
  type       = "ipsec.1"
}

resource "aws_ec2_transit_gateway" "example" {}

resource "aws_vpn_connection" "example" {
  customer_gateway_id = aws_customer_gateway.example.id
  transit_gateway_id  = aws_ec2_transit_gateway.example.id
  type                = aws_customer_gateway.example.type
  static_routes_only  = true
}

resource "aws_networkmanager_transit_gateway_registration" "example" {
  global_network_id   = aws_networkmanager_global_network.example.id
  transit_gateway_arn = aws_ec2_transit_gateway.example.arn

  depends_on = [aws_vpn_connection.example]
}

resource "aws_networkmanager_customer_gateway_association" "example" {
  global_network_id    = aws_networkmanager_global_network.example.id
  customer_gateway_arn = aws_customer_gateway.example.arn
  device_id            = aws_networkmanager_device.example.id

  depends_on = [aws_networkmanager_transit_gateway_registration.example]
}
```

## Argument Reference

This resource supports the following arguments:

* `customer_gateway_arn` - (Required) The Amazon Resource Name (ARN) of the customer gateway.
* `device_id` - (Required) The ID of the device.
* `global_network_id` - (Required) The ID of the global network.
* `link_id` - (Optional) The ID of the link.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_networkmanager_customer_gateway_association` using the global network ID and customer gateway ARN. For example:

```terraform
import {
  to = aws_networkmanager_customer_gateway_association.example
  id = "global-network-0d47f6t230mz46dy4,arn:aws:ec2:us-west-2:123456789012:customer-gateway/cgw-123abc05e04123abc"
}
```

Using `terraform import`, import `aws_networkmanager_customer_gateway_association` using the global network ID and customer gateway ARN. For example:

```console
% terraform import aws_networkmanager_customer_gateway_association.example global-network-0d47f6t230mz46dy4,arn:aws:ec2:us-west-2:123456789012:customer-gateway/cgw-123abc05e04123abc
```
