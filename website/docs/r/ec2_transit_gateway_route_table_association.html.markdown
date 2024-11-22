---
subcategory: "Transit Gateway"
layout: "aws"
page_title: "AWS: aws_ec2_transit_gateway_route_table_association"
description: |-
  Manages an EC2 Transit Gateway Route Table association
---

# Resource: aws_ec2_transit_gateway_route_table_association

Manages an EC2 Transit Gateway Route Table association.

## Example Usage

```terraform
resource "aws_ec2_transit_gateway_route_table_association" "example" {
  transit_gateway_attachment_id  = aws_ec2_transit_gateway_vpc_attachment.example.id
  transit_gateway_route_table_id = aws_ec2_transit_gateway_route_table.example.id
}
```

## Argument Reference

This resource supports the following arguments:

* `transit_gateway_attachment_id` - (Required) Identifier of EC2 Transit Gateway Attachment.
* `transit_gateway_route_table_id` - (Required) Identifier of EC2 Transit Gateway Route Table.
* `replace_existing_association` - (Optional) Boolean whether the Gateway Attachment should remove any current Route Table association before associating with the specified Route Table. Default value: `false`. This argument is intended for use with EC2 Transit Gateways shared into the current account, otherwise the `transit_gateway_default_route_table_association` argument of the `aws_ec2_transit_gateway_vpc_attachment` resource should be used.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - EC2 Transit Gateway Route Table identifier combined with EC2 Transit Gateway Attachment identifier
* `resource_id` - Identifier of the resource
* `resource_type` - Type of the resource

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_ec2_transit_gateway_route_table_association` using the EC2 Transit Gateway Route Table identifier, an underscore, and the EC2 Transit Gateway Attachment identifier. For example:

```terraform
import {
  to = aws_ec2_transit_gateway_route_table_association.example
  id = "tgw-rtb-12345678_tgw-attach-87654321"
}
```

Using `terraform import`, import `aws_ec2_transit_gateway_route_table_association` using the EC2 Transit Gateway Route Table identifier, an underscore, and the EC2 Transit Gateway Attachment identifier. For example:

```console
% terraform import aws_ec2_transit_gateway_route_table_association.example tgw-rtb-12345678_tgw-attach-87654321
```
