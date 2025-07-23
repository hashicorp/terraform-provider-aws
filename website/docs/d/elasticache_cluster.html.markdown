---
subcategory: "ElastiCache"
layout: "aws"
page_title: "AWS: aws_elasticache_cluster"
description: |-
  Get information on an ElastiCache Cluster resource.
---

# Data Source: aws_elasticache_cluster

Use this data source to get information about an ElastiCache Cluster

## Example Usage

```terraform
data "aws_elasticache_cluster" "my_cluster" {
  cluster_id = "my-cluster-id"
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `cluster_id` - (Required) Group identifier.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `node_type` - The cluster node type.
* `num_cache_nodes` - The number of cache nodes that the cache cluster has.
* `engine` - Name of the cache engine.
* `engine_version` - Version number of the cache engine.
* `ip_discovery` - The IP version advertised in the discovery protocol.
* `network_type` - The IP versions for cache cluster connections.
* `subnet_group_name` - Name of the subnet group associated to the cache cluster.
* `security_group_ids` - List VPC security groups associated with the cache cluster.
* `parameter_group_name` - Name of the parameter group associated with this cache cluster.
* `replication_group_id` - The replication group to which this cache cluster belongs.
* `log_delivery_configuration` - Redis [SLOWLOG](https://redis.io/commands/slowlog) or Redis [Engine Log](https://docs.aws.amazon.com/AmazonElastiCache/latest/red-ug/Log_Delivery.html#Log_contents-engine-log) delivery settings.
* `maintenance_window` - Specifies the weekly time range for when maintenance
on the cache cluster is performed.
* `snapshot_window` - Daily time range (in UTC) during which ElastiCache will
begin taking a daily snapshot of the cache cluster.
* `snapshot_retention_limit` - The number of days for which ElastiCache will
retain automatic cache cluster snapshots before deleting them.
* `availability_zone` - Availability Zone for the cache cluster.
* `notification_topic_arn` - An ARN of an
SNS topic that ElastiCache notifications get sent to.
* `port` - The port number on which each of the cache nodes will
accept connections.
* `configuration_endpoint` - (Memcached only) Configuration endpoint to allow host discovery.
* `cluster_address` - (Memcached only) DNS name of the cache cluster without the port appended.
* `preferred_outpost_arn` - The outpost ARN in which the cache cluster was created if created in outpost.
* `cache_nodes` - List of node objects including `id`, `address`, `port`, `availability_zone` and `outpost_arn`.
   Referenceable e.g., as `${data.aws_elasticache_cluster.bar.cache_nodes.0.address}`
* `tags` - Tags assigned to the resource
