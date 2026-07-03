---
subcategory: "ElastiCache"
layout: "aws"
page_title: "AWS: aws_elasticache_replication_groups"
description: |-
    Provides a list of ElastiCache Replication Group IDs in a Region
---

# Data Source: aws_elasticache_replication_groups

This resource can be useful for getting back a list of ElastiCache Replication Group IDs for a Region.

## Example Usage

The following example retrieves a list of all ElastiCache Replication Group IDs.

```terraform
data "aws_elasticache_replication_groups" "example" {}

output "example" {
  value = data.aws_elasticache_replication_groups.example.replication_group_ids
}
```

## Argument Reference

This data source does not support any arguments.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - AWS Region.
* `replication_group_ids` - A list of all the ElastiCache Replication Group IDs found.
