---
subcategory: "MemoryDB for Redis"
layout: "aws"
page_title: "AWS: aws_memorydb_cluster"
description: |-
  Provides information about a MemoryDB Cluster.
---

# Resource: aws_memorydb_cluster

Provides information about a MemoryDB Cluster.

## Example Usage

```terraform
data "aws_memorydb_cluster" "example" {
  name = "my-cluster"
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the cluster.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - Same as `name`.
* `arn` - ARN of the cluster.
* `acl_name` - Name of the Access Control List associated with the cluster.
* `auto_minor_version_upgrade` - True when the cluster allows automatic minor version upgrades.
* `cluster_endpoint`
    * `address` - DNS hostname of the cluster configuration endpoint.
    * `port` - Port number that the cluster configuration endpoint is listening on.
* `data_tiering` - True when data tiering is enabled.
* `description` - Description for the cluster.
* `engine_patch_version` - Patch version number of the Redis engine used by the cluster.
* `engine` - Engine that will run on cluster nodes.
* `engine_version` - Version number of the Redis engine used by the cluster.
* `final_snapshot_name` - Name of the final cluster snapshot to be created when this resource is deleted. If omitted, no final snapshot will be made.
* `kms_key_arn` - ARN of the KMS key used to encrypt the cluster at rest.
* `maintenance_window` - Weekly time range during which maintenance on the cluster is performed. Specify as a range in the format `ddd:hh24:mi-ddd:hh24:mi` (24H Clock UTC). Example: `sun:23:00-mon:01:30`.
* `node_type` - Compute and memory capacity of the nodes in the cluster.
* `num_replicas_per_shard` - The number of replicas to apply to each shard.
* `num_shards` - Number of shards in the cluster.
* `parameter_group_name` - The name of the parameter group associated with the cluster.
* `port` - Port number on which each of the nodes accepts connections.
* `security_group_ids` - Set of VPC Security Group ID-s associated with this cluster.
* `shards` - Set of shards in this cluster.
    * `name` - Name of this shard.
    * `num_nodes` - Number of individual nodes in this shard.
    * `slots` - Keyspace for this shard. Example: `0-16383`.
    * `nodes` - Set of nodes in this shard.
        * `availability_zone` - The Availability Zone in which the node resides.
        * `create_time` - The date and time when the node was created. Example: `2022-01-01T21:00:00Z`.
        * `name` - Name of this node.
        * `endpoint`
            * `address` - DNS hostname of the node.
            * `port` - Port number that this node is listening on.
* `snapshot_retention_limit` - The number of days for which MemoryDB retains automatic snapshots before deleting them. When set to `0`, automatic backups are disabled.
* `snapshot_window` - Daily time range (in UTC) during which MemoryDB begins taking a daily snapshot of your shard. Example: `05:00-09:00`.
* `sns_topic_arn` - ARN of the SNS topic to which cluster notifications are sent.
* `subnet_group_name` -The name of the subnet group used for the cluster.
* `tls_enabled` - When true, in-transit encryption is enabled for the cluster.
* `tags` - Map of tags assigned to the cluster.
