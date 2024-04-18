---
subcategory: "MemoryDB for Redis"
layout: "aws"
page_title: "AWS: aws_memorydb_snapshot"
description: |-
  Provides information about a MemoryDB Snapshot.
---

# Resource: aws_memorydb_snapshot

Provides information about a MemoryDB Snapshot.

## Example Usage

```terraform
data "aws_memorydb_snapshot" "example" {
  name = "my-snapshot"
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the snapshot.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - Name of the snapshot.
* `arn` - ARN of the snapshot.
* `cluster_configuration` - The configuration of the cluster from which the snapshot was taken.
    * `description` - Description for the cluster.
    * `engine_version` - Version number of the Redis engine used by the cluster.
    * `maintenance_window` - The weekly time range during which maintenance on the cluster is performed.
    * `name` - Name of the cluster.
    * `node_type` - Compute and memory capacity of the nodes in the cluster.
    * `num_shards` - Number of shards in the cluster.
    * `parameter_group_name` - Name of the parameter group associated with the cluster.
    * `port` - Port number on which the cluster accepts connections.
    * `snapshot_retention_limit` - Number of days for which MemoryDB retains automatic snapshots before deleting them.
    * `snapshot_window` - The daily time range (in UTC) during which MemoryDB begins taking a daily snapshot of the shard.
    * `subnet_group_name` - Name of the subnet group used by the cluster.
    * `topic_arn` - ARN of the SNS topic to which cluster notifications are sent.
    * `vpc_id` - The VPC in which the cluster exists.
* `cluster_name` - Name of the MemoryDB cluster that this snapshot was taken from.
* `kms_key_arn` - ARN of the KMS key used to encrypt the snapshot at rest.
* `source` - Whether the snapshot is from an automatic backup (`automated`) or was created manually (`manual`).
* `tags` - Map of tags assigned to the snapshot.
