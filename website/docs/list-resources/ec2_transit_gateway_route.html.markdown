---
subcategory: "Transit Gateway"
layout: "aws"
page_title: "AWS: aws_ec2_transit_gateway_route"
description: |-
  Lists EC2 Transit Gateway Route resources.
---

# List Resource: aws_ec2_transit_gateway_route

Lists EC2 Transit Gateway Route resources.

## Example Usage

```terraform
list "aws_ec2_transit_gateway_route" "example" {
  provider = aws

  config {
    transit_gateway_route_table_id = aws_ec2_transit_gateway.example.association_default_route_table_id
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
* `transit_gateway_route_table_id` - (Required) ID of the transit gateway route table.
