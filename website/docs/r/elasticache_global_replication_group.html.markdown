---
subcategory: "ElastiCache"
layout: "aws"
page_title: "AWS: aws_elasticache_global_replication_group"
description: |-
  Provides an ElastiCache Global Replication Group resource.
---

# Resource: aws_elasticache_global_replication_group

Provides an ElastiCache Global Replication Group resource, which manage a replication between 2 or more redis replication group in different region.

## Example Usage

### Global replication group with a single instance redis replication group

To create a single shard primary with single read replica:

```hcl
resource "aws_elasticache_global_replication_group" "replication_group" {
  global_replication_group_id_suffix = "example"
  primary_replication_group_id       = aws_elasticache_replication_group.primary.id
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
* `primary_replication_group_id` – (Required) The name of the primary cluster that accepts writes and will replicate updates to the secondary cluster.
* `global_replication_group_description` – (Optional) A user-created description for the global replication group.
* `retain_primary_replication_group` - (Optional) Whether to retain the primary replication group when the global replication group is deleted.
* `apply_immediately` - (Required) This parameter causes the modifications in this request and any pending modifications to be applied, asynchronously and as soon as possible. Modifications to Global Replication Groups cannot be requested to be applied in PreferredMaintenceWindow.
* `automatic_failover_enabled` - (Optional) Determines whether a read replica is automatically promoted to read/write primary if the existing primary encounters a failure.
* `cache_node_type` - (Optional) A valid cache node type that you want to scale this Global Datastore to.
* `engine_version` - (Optional) The upgraded version of the cache engine to be run on the clusters in the Global Datastore.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the ElastiCache Global Replication Group.
* `arn` - The ARN of the ElastiCache Global Replication Group.
* `at_rest_encryption_enabled` - A flag that indicate whether the encryption at rest is enabled.
* `auth_token_enabled` - A flag that indicate whether AuthToken (password) is enabled.
* `cluster_enabled` - A flag that indicates whether the Global Datastore is cluster enabled.
* `engine` - The Elasticache engine. For redis only
* `global_replication_group_members` - The identifiers of all the replication group members that are part of this global replication group.
    * `replication_group_id` - The replication group id of the Global Datastore member
    * `replication_group_region` - The AWS region of the Global Datastore member
    * `role` - Indicates the role of the replication group, primary or secondary
* `transit_encryption_enabled` - A flag that indicates whether the encryption in transit is enabled.

## Import

ElastiCache Global Replication Groups can be imported using the `global_replication_group_id`, e.g.

```
$ terraform import aws_elasticache_global_replication_group.my_global_replication_group okuqm-global-replication-group-1
```
