---
subcategory: "ElastiCache"
layout: "aws"
page_title: "AWS: aws_elasticache_global_replication_group"
description: |-
  Provides an ElastiCache Global Replication Group resource.
---

# Resource: aws_elasticache_global_replication_group

Provides an ElastiCache Global Replication Group resource.

## Example Usage

### Simple redis global replication group mode cluster disabled

To create a single shard primary with single read replica:

```hcl
resource "aws_elasticache_global_replication_group" "replication_group" {
  global_replication_group_id_suffix   = "example"
  primary_replication_group_id         = aws_elasticache_replication_group.primary.id
}

resource "aws_elasticache_replication_group" "primary" {
  replication_group_id          = "example"
  replication_group_description = "test example"

  engine                = "redis"
  engine_version        = "5.0.6"
  node_type             = "cache.m5.large"
  number_cache_clusters = 1
}
```

## Argument Reference

The following arguments are supported:

* `global_replication_group_id_suffix` – (Required) The suffix name of a Global Datastore.
* `replication_group_id` – (Required) The replication group identifier. The Global Datastore will be created from this replication group.
* `global_replication_group_description` – (Optional) A user-created description for the global replication group.
* `retain_primary_replication_group` - (Optional) Whether to retain the primary replication group when the global replication group is deleted.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the ElastiCache Global Replication Group.

## Import

ElastiCache Global Replication Groups can be imported using the `global_replication_group_id`, e.g.

```
$ terraform import aws_elasticache_global_replication_group.my_global_replication_group global-replication-group-1
```
