---
subcategory: "EC2 (Elastic Compute Cloud)"
layout: "aws"
page_title: "AWS: aws_ec2_transit_gateway_route_table_propagation"
description: |-
  Lists EC2 (Elastic Compute Cloud) Transit Gateway Route Table Propagation resources.
---

# List Resource: aws_ec2_transit_gateway_route_table_propagation

Lists EC2 (Elastic Compute Cloud) Transit Gateway Route Table Propagation resources.

## Example Usage

```terraform
list "aws_ec2_transit_gateway_route_table_propagation" "example" {
  provider = aws

  config {
    transit_gateway_route_table_id = aws_ec2_transit_gateway_route_table.example.id
  }
}
```

## Argument Reference

The following arguments are required:

* `transit_gateway_route_table_id` - (Required) ID of the Transit Gateway route table to list propagations for.

The following arguments are optional:

* `region` - (Optional) Region to query. Defaults to provider region.
