---
subcategory: "ElastiCache"
layout: "aws"
page_title: "AWS: aws_elasticache_snapshot"
description: |-
  Terraform data source for managing an AWS ElastiCache Snapshot.
---

# Data Source: aws_elasticache_snapshot

Terraform data source for managing an AWS ElastiCache Snapshot.

## Example Usage

### Basic Usage

```terraform
data "aws_elasticache_snapshot" "example" {
  cluster_id  = "example-cluster"
  most_recent = true
}
```

## Argument Reference

The following arguments are required:

* `snapshot_name` - (Required unless one of `cluster_id` or `replication_group_id` is provided) Name of the Elasticache snapshot. Conflicts with `cluster_id` and ` replication_group_id`.
* `cluster_id` - (Required unless one of `snapshot_name` or `replication_group_id` is provided) Cache cluster ID used as source of the snapshot. Conflicts with `snapshot_name` and `replication_group_id`.
* `replication_group_id` - (Required unless one of `cluster_id` or `replication_group_id` is provided) Replication group ID used as source of the snapshot. Conflicts with `snapshot_name` and `replication_group_id`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of the Snapshot.
* `automatic_failover` - Status of automatic failover for the source Redis replication group.
* `auto_minor_version_upgrade` - Status of automatic version upgrades for source Redis cluster or replication group
* `cluster_create_time` - Date and time when the source cluster was created, in UTC [RFC3339](https://tools.ietf.org/html/rfc3339#section-5.8) format(for example, YYYY-MM-DDTHH:MM:SSZ)
* `engine` - Elasticache engine used by the source cluster.
* `engine_version` - Version number of the cache engine used by the source cluster.
* `kms_key_id` - KMS key used for the snapshot.
* `node_type` - Node type used for the source cluster.
* `num_cache_nodes` - Number of nodes in the source cluster.
* `parameter_group_name` - Cache parameter group that is associated with the source cluster.
* `port` - Port number of the source cluster.
* `subnet_group_name` - Subnet group of the source cluster.
* `replication_group_description` - Replication group description of the source cluster.
* `snapshot_source` - Source of the snapshot, can be `automated` or `manual`.
* `snapshot_status` - Status of the snapshot, can be `creating`, `available`, `restoring`, `copying` or `deleting`
* `vpc_id` - The Amazon Virtual Private Cloud identifier (VPC ID) of the cache subnet group for the source cluster.
