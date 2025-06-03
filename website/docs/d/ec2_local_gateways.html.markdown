---
subcategory: "Outposts (EC2)"
layout: "aws"
page_title: "AWS: aws_ec2_local_gateways"
description: |-
    Provides information for multiple EC2 Local Gateways
---

# Data Source: aws_ec2_local_gateways

Provides information for multiple EC2 Local Gateways, such as their identifiers.

## Example Usage

The following example retrieves Local Gateways with a resource tag of `service` set to `production`.

```terraform
data "aws_ec2_local_gateways" "foo" {
  tags = {
    service = "production"
  }
}

output "foo" {
  value = data.aws_ec2_local_gateways.foo.ids
}
```

## Argument Reference

This data source supports the following arguments:

* `tags` - (Optional) Mapping of tags, each pair of which must exactly match
  a pair on the desired local_gateways.
* `filter` - (Optional) Custom filter block as described below.

More complex filters can be expressed using one or more `filter` sub-blocks,
which take the following arguments:

* `name` - (Required) Name of the field to filter by, as defined by
  [the underlying AWS API](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeLocalGateways.html).

* `values` - (Required) Set of values that are accepted for the given field.
  A Local Gateway will be selected if any one of the given values matches.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - AWS Region.
* `ids` - Set of all the Local Gateway identifiers

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `read` - (Default `20m`)
