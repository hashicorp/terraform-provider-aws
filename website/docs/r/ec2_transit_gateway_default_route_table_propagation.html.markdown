---
subcategory: "Transit Gateway"
layout: "aws"
page_title: "AWS: aws_ec2_transit_gateway_default_route_table_propagation"
description: |-
  Terraform resource for managing an AWS EC2 (Elastic Compute Cloud) Transit Gateway Default Route Table Propagation.
---
# Resource: aws_ec2_transit_gateway_default_route_table_propagation

Terraform resource for managing an AWS EC2 (Elastic Compute Cloud) Transit Gateway Default Route Table Propagation.

## Example Usage

### Basic Usage

```terraform
resource "aws_ec2_transit_gateway_default_route_table_propagation" "example" {
  transit_gateway_id             = aws_ec2_transit_gateway.example.id
  transit_gateway_route_table_id = aws_ec2_transit_gateway_route_table.example.id
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `transit_gateway_id` - (Required) ID of the Transit Gateway to change the default association route table on.
* `transit_gateway_route_table_id` - (Required) ID of the Transit Gateway Route Table to be made the default association route table.

## Attribute Reference

This resource exports no additional attributes.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `5m`)
* `update` - (Default `5m`)
* `delete` - (Default `5m`)
