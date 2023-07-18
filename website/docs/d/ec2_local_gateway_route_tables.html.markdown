---
subcategory: "Outposts (EC2)"
layout: "aws"
page_title: "AWS: aws_ec2_local_gateway_route_tables"
description: |-
    Provides information for multiple EC2 Local Gateway Route Tables
---

# Data Source: aws_ec2_local_gateway_route_tables

Provides information for multiple EC2 Local Gateway Route Tables, such as their identifiers.

## Example Usage

The following shows outputting all Local Gateway Route Table Ids.

```terraform
data "aws_ec2_local_gateway_route_tables" "foo" {}

output "foo" {
  value = data.aws_ec2_local_gateway_route_tables.foo.ids
}
```

## Argument Reference

* `tags` - (Optional) Mapping of tags, each pair of which must exactly match
  a pair on the desired local gateway route table.

* `filter` - (Optional) Custom filter block as described below.

More complex filters can be expressed using one or more `filter` sub-blocks,
which take the following arguments:

* `name` - (Required) Name of the field to filter by, as defined by
  [the underlying AWS API](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeLocalGatewayRouteTables.html).

* `values` - (Required) Set of values that are accepted for the given field.
  A Local Gateway Route Table will be selected if any one of the given values matches.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - AWS Region.
* `ids` - Set of Local Gateway Route Table identifiers

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `read` - (Default `20m`)
