---
subcategory: "ElastiCache"
layout: "aws"
page_title: "AWS: aws_elasticache_global_replication_group"
description: |-
  Provides an ElastiCache Global Replication Group resource.
---

# Resource: aws_elasticache_global_replication_group

Provides an ElastiCache Global Replication Group resource, which manages replication between two or more Replication Groups in different regions. For more information, see the [ElastiCache User Guide](https://docs.aws.amazon.com/AmazonElastiCache/latest/red-ug/Redis-Global-Datastore.html).

## Example Usage

### Global replication group with one secondary replication group

The global replication group depends on the primary group existing. Secondary replication groups depend on the global replication group. Terraform dependency management will handle this transparently using resource value references.

```terraform
resource "aws_elasticache_global_replication_group" "example" {
  global_replication_group_id_suffix = "example"
  primary_replication_group_id       = aws_elasticache_replication_group.primary.id
}

resource "aws_elasticache_replication_group" "primary" {
  replication_group_id = "example-primary"
  description          = "primary replication group"

  engine         = "redis"
  engine_version = "5.0.6"
  node_type      = "cache.m5.large"

  num_cache_clusters = 1
}

resource "aws_elasticache_replication_group" "secondary" {
  provider = aws.other_region

  replication_group_id        = "example-secondary"
  description                 = "secondary replication group"
  global_replication_group_id = aws_elasticache_global_replication_group.example.global_replication_group_id

  num_cache_clusters = 1
}
```

### Managing Redis OOS/Valkey Engine Versions

The initial Redis version is determined by the version set on the primary replication group.
However, once it is part of a Global Replication Group,
the Global Replication Group manages the version of all member replication groups.

The member replication groups must have [`lifecycle.ignore_changes[engine_version]`](https://www.terraform.io/language/meta-arguments/lifecycle) set,
or Terraform will always return a diff.

In this example,
the primary replication group will be created with Redis 6.0,
and then upgraded to Redis 6.2 once added to the Global Replication Group.
The secondary replication group will be created with Redis 6.2.

```terraform
resource "aws_elasticache_global_replication_group" "example" {
  global_replication_group_id_suffix = "example"
  primary_replication_group_id       = aws_elasticache_replication_group.primary.id

  engine_version = "6.2"
}

resource "aws_elasticache_replication_group" "primary" {
  replication_group_id = "example-primary"
  description          = "primary replication group"

  engine         = "redis"
  engine_version = "6.0"
  node_type      = "cache.m5.large"

  num_cache_clusters = 1

  lifecycle {
    ignore_changes = [engine_version]
  }
}

resource "aws_elasticache_replication_group" "secondary" {
  provider = aws.other_region

  replication_group_id        = "example-secondary"
  description                 = "secondary replication group"
  global_replication_group_id = aws_elasticache_global_replication_group.example.global_replication_group_id

  num_cache_clusters = 1

  lifecycle {
    ignore_changes = [engine_version]
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `automatic_failover_enabled` - (Optional) Specifies whether read-only replicas will be automatically promoted to read/write primary if the existing primary fails.
  When creating, by default the Global Replication Group inherits the automatic failover setting of the primary replication group.
* `cache_node_type` - (Optional) The instance class used.
  See AWS documentation for information on [supported node types](https://docs.aws.amazon.com/AmazonElastiCache/latest/red-ug/CacheNodes.SupportedTypes.html)
  and [guidance on selecting node types](https://docs.aws.amazon.com/AmazonElastiCache/latest/red-ug/nodes-select-size.html).
  When creating, by default the Global Replication Group inherits the node type of the primary replication group.
* `engine_version` - (Optional) Redis version to use for the Global Replication Group.
  When creating, by default the Global Replication Group inherits the version of the primary replication group.
  If a version is specified, the Global Replication Group and all member replication groups will be upgraded to this version.
  Cannot be downgraded without replacing the Global Replication Group and all member replication groups.
  When the version is 7 or higher, the major and minor version should be set, e.g., `7.2`.
  When the version is 6, the major and minor version can be set, e.g., `6.2`,
  or the minor version can be unspecified which will use the latest version at creation time, e.g., `6.x`.
  The actual engine version used is returned in the attribute `engine_version_actual`, see [Attribute Reference](#attribute-reference) below.
* `global_replication_group_id_suffix` – (Required) The suffix name of a Global Datastore. If `global_replication_group_id_suffix` is changed, creates a new resource.
* `primary_replication_group_id` – (Required) The ID of the primary cluster that accepts writes and will replicate updates to the secondary cluster. If `primary_replication_group_id` is changed, creates a new resource.
* `global_replication_group_description` – (Optional) A user-created description for the global replication group.
* `num_node_groups` - (Optional) The number of node groups (shards) on the global replication group.
* `parameter_group_name` - (Optional) An ElastiCache Parameter Group to use for the Global Replication Group.
  Required when upgrading a major engine version, but will be ignored if left configured after the upgrade is complete.
  Specifying without a major version upgrade will fail.
  Note that ElastiCache creates a copy of this parameter group for each member replication group.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The ID of the ElastiCache Global Replication Group.
* `arn` - The ARN of the ElastiCache Global Replication Group.
* `engine_version_actual` - The full version number of the cache engine running on the members of this global replication group.
* `at_rest_encryption_enabled` - A flag that indicate whether the encryption at rest is enabled.
* `auth_token_enabled` - A flag that indicate whether AuthToken (password) is enabled.
* `cluster_enabled` - Indicates whether the Global Datastore is cluster enabled.
* `engine` - The name of the cache engine to be used for the clusters in this global replication group.
* `global_replication_group_id` - The full ID of the global replication group.
* `global_node_groups` - Set of node groups (shards) on the global replication group.
  Has the values:
    * `global_node_group_id` - The ID of the global node group.
    * `slots` - The keyspace for this node group.
* `transit_encryption_enabled` - A flag that indicates whether the encryption in transit is enabled.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `60m`)
* `update` - (Default `60m`)
* `delete` - (Default `20m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import ElastiCache Global Replication Groups using the `global_replication_group_id`. For example:

```terraform
import {
  to = aws_elasticache_global_replication_group.my_global_replication_group
  id = "okuqm-global-replication-group-1"
}
```

Using `terraform import`, import ElastiCache Global Replication Groups using the `global_replication_group_id`. For example:

```console
% terraform import aws_elasticache_global_replication_group.my_global_replication_group okuqm-global-replication-group-1
```
