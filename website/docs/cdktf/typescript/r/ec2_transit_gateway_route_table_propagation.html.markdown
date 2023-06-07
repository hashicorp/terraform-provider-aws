---
subcategory: "Transit Gateway"
layout: "aws"
page_title: "AWS: aws_ec2_transit_gateway_route_table_propagation"
description: |-
  Manages an EC2 Transit Gateway Route Table propagation
---

# Resource: aws_ec2_transit_gateway_route_table_propagation

Manages an EC2 Transit Gateway Route Table propagation.

## Example Usage

```terraform
resource "aws_ec2_transit_gateway_route_table_propagation" "example" {
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

`awsEc2TransitGatewayRouteTablePropagation` can be imported by using the EC2 Transit Gateway Route Table identifier, an underscore, and the EC2 Transit Gateway Attachment identifier, e.g.,

```
$ terraform import aws_ec2_transit_gateway_route_table_propagation.example tgw-rtb-12345678_tgw-attach-87654321
```

<!-- cache-key: cdktf-0.17.0-pre.15 input-3f7fc604735b965a0bdab490984fe7efcec8e99742613ec9237376463dd81611 -->