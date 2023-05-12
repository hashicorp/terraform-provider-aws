---
subcategory: "DMS (Database Migration)"
layout: "aws"
page_title: "AWS: aws_dms_replication_subnet_groups"
description: |-
  Terraform data source for managing an AWS DMS (Database Migration) Replication Subnet Groups.
---

# Data Source: aws_dms_replication_subnet_groups

This resource can be useful for getting back a set of DMS (Database Migration) Replication Subnet Group IDs.

## Example Usage

The following example outputs a set of all VPC IDs associated with a Replication Subnet Group.

```terraform
data "aws_dms_replication_subnet_groups" "example" {}

data "aws_dms_replication_subnet_group" "example" {
  for_each = toset(data.aws_dms_replication_subnet_groups.example.ids)
  replication_subnet_group_id = each.value
}

output "vpc_ids" {
	value = toset([for s in data.aws_dms_replication_subnet_group.this : s.vpc_id])
}
```

## Argument Reference

The following arguments are optional:

* `filter` - (Optional) Custom filter block as described below.

Filters can be expressed using one or more `filter` sub-blocks,
which take the following arguments:

* `name` - (Required) Name of the field to filter by, as defined by
  [the underlying AWS API](https://docs.aws.amazon.com/dms/latest/APIReference/API_DescribeReplicationSubnetGroups.html).
  Note that only `replication-subnet-group-id` is currently supported. E.G.

```terraform
data "aws_replication_subnet_groups" "example" {
  filter {
    name   = "replication-subnet-group-id"
	values = [""] # insert values here
}
```

* `values` - (Required) Set of values that are accepted for the given field. Replication Subnet Groups will be selected if any one of the given values match.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `ids` - List of all the Replication Subnet Group IDs found.
