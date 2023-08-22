---
subcategory: "Transit Gateway"
layout: "aws"
page_title: "AWS: aws_ec2_transit_gateway_route"
description: |-
  Manages an EC2 Transit Gateway Route
---

# Resource: aws_ec2_transit_gateway_route

Manages an EC2 Transit Gateway Route.

## Example Usage

### Standard usage

```terraform
resource "aws_ec2_transit_gateway_route" "example" {
  destination_cidr_block         = "0.0.0.0/0"
  transit_gateway_attachment_id  = aws_ec2_transit_gateway_vpc_attachment.example.id
  transit_gateway_route_table_id = aws_ec2_transit_gateway.example.association_default_route_table_id
}
```

### Blackhole route

```terraform
resource "aws_ec2_transit_gateway_route" "example" {
  destination_cidr_block         = "0.0.0.0/0"
  blackhole                      = true
  transit_gateway_route_table_id = aws_ec2_transit_gateway.example.association_default_route_table_id
}
```

## Argument Reference

This resource supports the following arguments:

* `destination_cidr_block` - (Required) IPv4 or IPv6 RFC1924 CIDR used for destination matches. Routing decisions are based on the most specific match.
* `transit_gateway_attachment_id` - (Optional) Identifier of EC2 Transit Gateway Attachment (required if `blackhole` is set to false).
* `blackhole` - (Optional) Indicates whether to drop traffic that matches this route (default to `false`).
* `transit_gateway_route_table_id` - (Required) Identifier of EC2 Transit Gateway Route Table.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - EC2 Transit Gateway Route Table identifier combined with destination

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_ec2_transit_gateway_route` using the EC2 Transit Gateway Route Table, an underscore, and the destination. For example:

```terraform
import {
  to = aws_ec2_transit_gateway_route.example
  id = "tgw-rtb-12345678_0.0.0.0/0"
}
```

Using `terraform import`, import `aws_ec2_transit_gateway_route` using the EC2 Transit Gateway Route Table, an underscore, and the destination. For example:

```console
% terraform import aws_ec2_transit_gateway_route.example tgw-rtb-12345678_0.0.0.0/0
```
