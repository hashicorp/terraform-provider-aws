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

The following arguments are supported:

* `transitGatewayAttachmentId` - (Required) Identifier of EC2 Transit Gateway Attachment.
* `transitGatewayRouteTableId` - (Required) Identifier of EC2 Transit Gateway Route Table.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - EC2 Transit Gateway Route Table identifier combined with EC2 Transit Gateway Attachment identifier
* `resourceId` - Identifier of the resource
* `resourceType` - Type of the resource

## Import

`awsEc2TransitGatewayRouteTableAssociation` can be imported by using the EC2 Transit Gateway Route Table identifier, an underscore, and the EC2 Transit Gateway Attachment identifier, e.g.,

```
$ terraform import aws_ec2_transit_gateway_route_table_association.example tgw-rtb-12345678_tgw-attach-87654321
```

<!-- cache-key: cdktf-0.17.0-pre.15 input-e8bd264e123bc7d73d8d6836f2d3010e1b9d33e09d6d8cab85d400cbc4c35f21 -->