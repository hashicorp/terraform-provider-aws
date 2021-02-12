---
subcategory: "EC2"
layout: "aws"
page_title: "AWS: aws_ec2_transit_gateway_route_tables"
description: |-
  Get information on EC2 Transit Gateway Route Tables
---

# Data Source: aws_ec2_transit_gateway_route_tables

Get information on EC2 Transit Gateway Route Tables.

## Example Usage

### By Transit Gateway Id

```hcl
data "aws_ec2_transit_gateway_route_tables" "example" {
  transit_gateway_id = "tgw-12345678"
}
```

## Argument Reference

The following arguments are supported:

* `filter` - (Optional) One or more configuration blocks containing name-values filters. Detailed below.
* `transit_gateway_id` - (Optional) Identifier of the EC2 Transit Gateway.

### filter Argument Reference

* `name` - (Required) Name of the filter.
* `values` - (Required) List of one or more values for the filter.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `ids` - A set of all the transit gateway route table ids found. This data source will fail if none are found.
