---
subcategory: "Transit Gateway"
layout: "aws"
page_title: "AWS: aws_ec2_transit_gateway_default_route_table_association"
description: |-
  Terraform resource for managing an AWS EC2 (Elastic Compute Cloud) Transit Gateway Default Route Table Association.
---
# Resource: aws_ec2_transit_gateway_default_route_table_association

Terraform resource for managing an AWS EC2 (Elastic Compute Cloud) Transit Gateway Default Route Table Association.

## Example Usage

### Basic Usage

```terraform
resource "aws_ec2_transit_gateway_default_route_table_association" "example" {
  transit_gateway_id             = aws_ec2_transit_gateway.example.id
  transit_gateway_route_table_id = aws_ec2_transit_gateway_route_table.example.id
}
```

## Argument Reference

The following arguments are required:

* `transit_gateway_id` - (Required) ID of the Transit Gateway to change the default association route table on.
* `transit_gateway_route_table_id` - (Required) ID of the Transit Gateway Route Table to be made the default association route table.

## Attribute Reference

This resource exports no additional attributes.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `5m`)
* `update` - (Default `5m`)
* `delete` - (Default `5m`)
