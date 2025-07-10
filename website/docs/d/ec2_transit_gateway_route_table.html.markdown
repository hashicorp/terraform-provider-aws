---
subcategory: "Transit Gateway"
layout: "aws"
page_title: "AWS: aws_ec2_transit_gateway_route_table"
description: |-
  Get information on an EC2 Transit Gateway Route Table
---

# Data Source: aws_ec2_transit_gateway_route_table

Get information on an EC2 Transit Gateway Route Table.

## Example Usage

### By Filter

```terraform
data "aws_ec2_transit_gateway_route_table" "example" {
  filter {
    name   = "default-association-route-table"
    values = ["true"]
  }

  filter {
    name   = "transit-gateway-id"
    values = ["tgw-12345678"]
  }
}
```

### By Identifier

```terraform
data "aws_ec2_transit_gateway_route_table" "example" {
  id = "tgw-rtb-12345678"
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `filter` - (Optional) One or more configuration blocks containing name-values filters. Detailed below.
* `id` - (Optional) Identifier of the EC2 Transit Gateway Route Table.

### filter Argument Reference

* `name` - (Required) Name of the filter.
* `values` - (Required) List of one or more values for the filter.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - EC2 Transit Gateway Route Table ARN.
* `default_association_route_table` - Boolean whether this is the default association route table for the EC2 Transit Gateway
* `default_propagation_route_table` - Boolean whether this is the default propagation route table for the EC2 Transit Gateway
* `id` - EC2 Transit Gateway Route Table identifier
* `transit_gateway_id` - EC2 Transit Gateway identifier
* `tags` - Key-value tags for the EC2 Transit Gateway Route Table

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `read` - (Default `20m`)
