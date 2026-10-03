---
subcategory: "ElastiCache"
layout: "aws"
page_title: "AWS: aws_elasticache_snapshot"
description: |-
  Provides information about an ElastiCache Snapshot for a cache cluster.
---

# Data Source: aws_elasticache_snapshot

Provides information about an ElastiCache Snapshot taken from a cache cluster.

Use this data source to discover a snapshot for a given cache cluster so it can be referenced
elsewhere in a configuration, for example to recreate an `aws_elasticache_cluster` from its most
recent snapshot. Discovery works even after the source cluster has been deleted, because snapshot
metadata persists.

~> **Note:** This data source looks up snapshots by `cluster_id` (the `CacheClusterId` filter of
the `DescribeSnapshots` API), so it returns only snapshots taken from a cache cluster. Snapshots
taken from a replication group are not returned; for those, `replication_group_id` is empty.

## Example Usage

```terraform
data "aws_elasticache_snapshot" "latest" {
  cluster_id  = "example-cluster"
  most_recent = true
}

resource "aws_elasticache_cluster" "example" {
  cluster_id      = "example-cluster"
  engine          = "redis"
  node_type       = "cache.t3.small"
  num_cache_nodes = 1
  snapshot_name   = data.aws_elasticache_snapshot.latest.snapshot_name
}
```

## Argument Reference

This data source supports the following arguments:

* `cluster_id` - (Required) User-supplied identifier of the source cache cluster whose snapshots are described.
* `most_recent` - (Optional) If more than one snapshot matches, return the most recent one. Defaults to `false`, in which case matching more than one snapshot is an error.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the snapshot.
* `auto_minor_version_upgrade` - Whether minor version engine upgrades are applied automatically to the source cluster.
* `automatic_failover` - Status of automatic failover for the source replication group.
* `cache_cluster_create_time` - Date and time when the source cluster was created.
* `data_tiering` - Whether data tiering is enabled.
* `durability` - Durability setting of the cluster when the snapshot was taken.
* `engine` - Name of the cache engine used by the source cluster.
* `engine_version` - Version of the cache engine used by the source cluster.
* `kms_key_id` - ID of the KMS key used to encrypt the snapshot.
* `maintenance_window` - Weekly time range during which maintenance on the source cluster is performed.
* `node_snapshots` - List of the cache nodes in the source cluster. Each element exports the following:
    * `cache_cluster_id` - Unique identifier for the source cluster.
    * `cache_node_create_time` - Date and time when the cache node was created in the source cluster.
    * `cache_node_id` - Cache node identifier for the node in the source cluster.
    * `cache_size` - Size of the cache on the source cache node.
    * `node_group_configuration` - Configuration for the source node group (shard). Each element exports the following:
        * `node_group_id` - Identifier for the node group (shard).
        * `primary_availability_zone` - Availability Zone where the primary node of this node group is launched.
        * `primary_outpost_arn` - Outpost ARN of the primary node.
        * `replica_availability_zones` - List of Availability Zones used for the read replicas.
        * `replica_count` - Number of read replica nodes in this node group (shard).
        * `replica_outpost_arns` - List of Outpost ARNs of the node replicas.
        * `slots` - Keyspace for this node group (shard), in the format `startkey-endkey`.
    * `node_group_id` - Unique identifier for the source node group (shard).
    * `snapshot_create_time` - Date and time when the source node's metadata and cache data set was obtained for the snapshot.
* `node_type` - Compute and memory capacity of the nodes in the source cluster.
* `num_cache_nodes` - Number of cache nodes in the source cluster.
* `num_node_groups` - Number of node groups (shards) in the snapshot.
* `parameter_group_name` - Cache parameter group associated with the source cluster.
* `port` - Port number used by each cache node in the source cluster.
* `preferred_availability_zone` - Availability Zone in which the source cluster is located.
* `preferred_outpost_arn` - ARN of the preferred Outpost.
* `replication_group_description` - Description of the source replication group. Empty for cluster snapshots.
* `replication_group_id` - Identifier of the source replication group. Empty for cluster snapshots.
* `snapshot_name` - Name of the snapshot.
* `snapshot_retention_limit` - Number of days for which ElastiCache retains automatic snapshots before deleting them.
* `snapshot_source` - Whether the snapshot is from an automatic backup (`automated`) or was created manually (`manual`).
* `snapshot_window` - Daily time range during which ElastiCache takes daily snapshots of the source cluster.
* `status` - Status of the snapshot.
* `subnet_group_name` - Cache subnet group associated with the source cluster.
* `topic_arn` - ARN of the topic used by the source cluster for publishing notifications.
* `vpc_id` - VPC in which the source cluster exists.
