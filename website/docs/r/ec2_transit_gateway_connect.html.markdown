---
subcategory: "Transit Gateway"
layout: "aws"
page_title: "AWS: aws_ec2_transit_gateway_connect"
description: |-
  Manages an EC2 Transit Gateway Connect
---

# Resource: aws_ec2_transit_gateway_connect

Manages an EC2 Transit Gateway Connect.

## Example Usage

```terraform
resource "aws_ec2_transit_gateway_vpc_attachment" "example" {
  subnet_ids         = [aws_subnet.example.id]
  transit_gateway_id = aws_ec2_transit_gateway.example.id
  vpc_id             = aws_vpc.example.id
}

resource "aws_ec2_transit_gateway_connect" "attachment" {
  transport_attachment_id = aws_ec2_transit_gateway_vpc_attachment.example.id
  transit_gateway_id      = aws_ec2_transit_gateway.example.id
}
```

## Argument Reference

The following arguments are supported:

* `protocol` - (Optional) The tunnel protocol. Valida values: `gre`. Default is `gre`.
* `tags` - (Optional) Key-value tags for the EC2 Transit Gateway Connect. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `transit_gateway_default_route_table_association` - (Optional) Boolean whether the Connect should be associated with the EC2 Transit Gateway association default route table. This cannot be configured or perform drift detection with Resource Access Manager shared EC2 Transit Gateways. Default value: `true`.
* `transit_gateway_default_route_table_propagation` - (Optional) Boolean whether the Connect should propagate routes with the EC2 Transit Gateway propagation default route table. This cannot be configured or perform drift detection with Resource Access Manager shared EC2 Transit Gateways. Default value: `true`.
* `transit_gateway_id` - (Required) Identifier of EC2 Transit Gateway.
* `transport_attachment_id` - (Required) The underlaying VPC attachment

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - EC2 Transit Gateway Attachment identifier
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `10m`)
- `update` - (Default `10m`)
- `delete` - (Default `10m`)

## Import

`aws_ec2_transit_gateway_connect` can be imported by using the EC2 Transit Gateway Connect identifier, e.g.,

```
$ terraform import aws_ec2_transit_gateway_connect.example tgw-attach-12345678
```
