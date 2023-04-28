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

The following arguments are supported:

* `replication_group_id` – (Required) Identifier for the replication group.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `description` - Description of the replication group.
* `arn` - ARN of the created ElastiCache Replication Group.
* `auth_token_enabled` - Whether an AuthToken (password) is enabled.
* `automatic_failover_enabled` - A flag whether a read-only replica will be automatically promoted to read/write primary if the existing primary fails.
* `node_type` – The cluster node type.
* `num_cache_clusters` – The number of cache clusters that the replication group has.
* `num_node_groups` - Number of node groups (shards) for the replication group.
* `number_cache_clusters` – (**Deprecated** use `num_cache_clusters` instead) Number of cache clusters that the replication group has.
* `member_clusters` - Identifiers of all the nodes that are part of this replication group.
* `multi_az_enabled` - Whether Multi-AZ Support is enabled for the replication group.
* `replicas_per_node_group` - Number of replica nodes in each node group.
* `replication_group_description` - (**Deprecated** use `description` instead) Description of the replication group.
* `log_delivery_configuration` - Redis [SLOWLOG](https://redis.io/commands/slowlog) or Redis [Engine Log](https://docs.aws.amazon.com/AmazonElastiCache/latest/red-ug/Log_Delivery.html#Log_contents-engine-log) delivery settings.
* `snapshot_window` - Daily time range (in UTC) during which ElastiCache begins taking a daily snapshot of your node group (shard).
* `snapshot_retention_limit` - The number of days for which ElastiCache retains automatic cache cluster snapshots before deleting them.
* `port` – The port number on which the configuration endpoint will accept connections.
* `configuration_endpoint_address` - The configuration endpoint address to allow host discovery.
* `primary_endpoint_address` - The endpoint of the primary node in this node group (shard).
* `reader_endpoint_address` - The endpoint of the reader node in this node group (shard).
