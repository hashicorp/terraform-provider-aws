---
subcategory: "ElastiCache"
layout: "aws"
page_title: "AWS: aws_elasticache_replication_group"
description: |-
  Get information on an ElastiCache Replication Group resource.
---

# Data Source: aws_elasticache_replication_group

Use this data source to get information about an ElastiCache Replication Group.

## Example Usage

```terraform
data "aws_elasticache_replication_group" "bar" {
  replication_group_id = "example"
}
```

## Argument Reference

This data source supports the following arguments:

* `replication_group_id` – (Required) Identifier for the replication group.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `description` - Description of the replication group.
* `arn` - ARN of the created ElastiCache Replication Group.
* `auth_token_enabled` - Whether an AuthToken (password) is enabled.
* `automatic_failover_enabled` - A flag whether a read-only replica will be automatically promoted to read/write primary if the existing primary fails.
* `cluster_mode` - Whether cluster mode is enabled or disabled.
* `node_type` – The cluster node type.
* `num_cache_clusters` – The number of cache clusters that the replication group has.
* `num_node_groups` - Number of node groups (shards) for the replication group.
* `member_clusters` - Identifiers of all the nodes that are part of this replication group.
* `multi_az_enabled` - Whether Multi-AZ Support is enabled for the replication group.
* `replicas_per_node_group` - Number of replica nodes in each node group.
* `log_delivery_configuration` - Redis [SLOWLOG](https://redis.io/commands/slowlog) or Redis [Engine Log](https://docs.aws.amazon.com/AmazonElastiCache/latest/red-ug/Log_Delivery.html#Log_contents-engine-log) delivery settings.
* `snapshot_window` - Daily time range (in UTC) during which ElastiCache begins taking a daily snapshot of your node group (shard).
* `snapshot_retention_limit` - The number of days for which ElastiCache retains automatic cache cluster snapshots before deleting them.
* `port` – The port number on which the configuration endpoint will accept connections.
* `configuration_endpoint_address` - The configuration endpoint address to allow host discovery.
* `primary_endpoint_address` - The endpoint of the primary node in this node group (shard).
* `reader_endpoint_address` - The endpoint of the reader node in this node group (shard).
