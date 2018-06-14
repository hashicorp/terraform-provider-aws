---
layout: "aws"
page_title: "AWS: aws_route_table_ids"
sidebar_current: "docs-aws-datasource-route-table-ids"
description: |-
    Provides a list of route table Ids for a VPC
---

# Data Source: aws_route_table_ids

`aws_route_table_ids` provides a list of ids for a vpc_id

This resource can be useful for getting back a list of route table ids for a vpc.

## Example Usage

The following adds a route for a particular cidr block to every route table
in the vpc to use a particular vpc peering connection.

```hcl

data "aws_route_table_ids" "rts" {
  vpc_id = "${var.vpc_id}"
}

resource "aws_route" "r" {
  count                  = "${length(data.aws_route_table_ids.rts.ids)}"
  route_table_id         = "${data.aws_route_table_ids.rts.ids[count.index]}"
  destination_cidr_block = "10.0.1.0/22"
  vpc_peering_connection_id = "pcx-0e9a7a9ecd137dc54"
}

```

## Argument Reference

* `vpc_id` - (Required) The VPC ID that you want to filter from.

* `tags` - (Optional) A mapping of tags, each pair of which must exactly match
  a pair on the desired route tables.

## Attributes Reference

* `ids` - A list of all the route table ids found. This data source will fail if none are found.
