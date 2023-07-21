---
subcategory: "Transit Gateway"
layout: "aws"
page_title: "AWS: aws_ec2_transit_gateway_route_table"
description: |-
  Manages an EC2 Transit Gateway Route Table
---

# Resource: aws_ec2_transit_gateway_route_table

Manages an EC2 Transit Gateway Route Table.

## Example Usage

```terraform
resource "aws_ec2_transit_gateway_route_table" "example" {
  transit_gateway_id = aws_ec2_transit_gateway.example.id
}
```

## Argument Reference

The following arguments are supported:

* `transitGatewayId` - (Required) Identifier of EC2 Transit Gateway.
* `tags` - (Optional) Key-value tags for the EC2 Transit Gateway Route Table. If configured with a provider [`defaultTags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - EC2 Transit Gateway Route Table Amazon Resource Name (ARN).
* `defaultAssociationRouteTable` - Boolean whether this is the default association route table for the EC2 Transit Gateway.
* `defaultPropagationRouteTable` - Boolean whether this is the default propagation route table for the EC2 Transit Gateway.
* `id` - EC2 Transit Gateway Route Table identifier
* `tagsAll` - A map of tags assigned to the resource, including those inherited from the provider [`defaultTags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

`awsEc2TransitGatewayRouteTable` can be imported by using the EC2 Transit Gateway Route Table identifier, e.g.,

```
$ terraform import aws_ec2_transit_gateway_route_table.example tgw-rtb-12345678
```

<!-- cache-key: cdktf-0.17.0-pre.15 input-092be375f2f09f63bf27cf8ba5f401de29a29a0e18c98e42df9a11ae41925104 -->