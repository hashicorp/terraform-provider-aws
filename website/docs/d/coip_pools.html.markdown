---
subcategory: "VPC"
layout: "aws"
page_title: "AWS: aws_coip_pools"
description: |-
    Provides a list of COIP Pool Ids in a region
---

# Data Source: aws_coip_pools

This resource can be useful for getting back a list of COIP Pool Ids for a region.

## Example Usage

The following shows outputing all COIP Pool Ids.

```hcl
data "aws_vpcs" "foo" {}

output "foo" {
  value = "${data.aws_vpcs.foo.ids}"
}
```

## Argument Reference

* `tags` - (Optional) A mapping of tags, each pair of which must exactly match
  a pair on the desired vpcs.

* `filter` - (Optional) Custom filter block as described below.

More complex filters can be expressed using one or more `filter` sub-blocks,
which take the following arguments:

* `name` - (Required) The name of the field to filter by, as defined by
  [the underlying AWS API](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeCoipPools.html).

* `values` - (Required) Set of values that are accepted for the given field.
  A COIP Pool will be selected if any one of the given values matches.

## Attributes Reference

* `ids` - A list of all the COIP Pool Ids found. This data source will fail if none are found.
