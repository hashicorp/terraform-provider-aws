---
subcategory: "VPC"
layout: "aws"
page_title: "AWS: aws_local_gateways"
description: |-
    Provides a list local gateways in a region
---

# Data Source: aws_local_gateways

This resource can be useful for getting back a list of Local Gateway Ids for a region.

## Example Usage

The following example retrieves a list of Local Gateway Ids with a custom tag of `service` set to a value of "production".

```hcl
data "aws_local_gateways" "foo" {
  tags = {
    service = "production"
  }
}

output "foo" {
  value = "${data.aws_local_gateways.foo.ids}"
}
```

## Argument Reference

* `tags` - (Optional) A mapping of tags, each pair of which must exactly match
  a pair on the desired vpcs.

* `filter` - (Optional) Custom filter block as described below.

More complex filters can be expressed using one or more `filter` sub-blocks,
which take the following arguments:

* `name` - (Required) The name of the field to filter by, as defined by
  [the underlying AWS API](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeLocalGateways.html).

* `values` - (Required) Set of values that are accepted for the given field.
  A VPC will be selected if any one of the given values matches.

## Attributes Reference

* `ids` - A list of all the VPC Ids found. This data source will fail if none are found.
