---
layout: "aws"
page_title: "AWS: aws_vpc_ids"
sidebar_current: "docs-aws-datasource-vpc-ids"
description: |-
    Provides a list of VPC Ids in a region
---

# Data Source: aws_vpc_ids

`aws_vpc_ids` provides a list of vpc ids

This resource can be useful for getting back a list of vpc ids for a region.

The following example retrieves a list of VPC Ids with a custom tag of `service` set to a value of "production".

## Example Usage

The following shows outputing all vpc ids.

```hcl
data "aws_vpc_ids" "foo" {
  tags {
    service = "production"
  }
}

output "foo" {
  value = "${data.aws_vpc_ids.foo.ids}"
}
```

The primary use case would be interpolate the vpc_ids output into `count` of an aws_flow_log resource.

```hcl
data "aws_vpc_ids" "foo" {}

resource "aws_flow_log" "test_flow_log" {
  count = "${length(data.aws_vpc_ids.foo.ids)}"
  ...
  vpc_id = "${element(data.aws_vpc_ids.foo.ids, count.index)}"
  ...
}

output "foo" {
  value = "${data.aws_vpc_ids.foo.ids}"
}
```

## Argument Reference

* `tags` - (Optional) A mapping of tags, each pair of which must exactly match
  a pair on the desired vpcs.

## Attributes Reference

* `ids` - A list of all the vpc ids found. This data source will fail if none are found.
