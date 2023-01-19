---
subcategory: "ElastiCache"
layout: "aws"
page_title: "AWS: aws_elasticache_snapshot"
description: |-
  Terraform resource for managing an AWS ElastiCache Snapshot.
---

# Resource: aws_elasticache_snapshot

Terraform resource for managing an AWS ElastiCache Snapshot.

## Example Usage

### Basic Usage

```terraform
resource "aws_elasticache_snapshot" "example" {
  cluster_id    = "cluster-example"
  snapshot_name = "cluster-example-snapshot"
}
```

## Argument Reference

The following arguments are required:

* `snapshot_name` - (Required) Name of the Elasticache snapshot. Changing this value will re-create the resource.
* `cluster_id` - (Required unless `replication_group_id` is provided) Cache cluster ID to snapshot. Conflicts with `replication_group_id`.
* `replication_group_id` - (Required unless `cluster_id` is provided) Replication group ID to snapshot. Conflicts with `cluster_id`.

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

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `60m`)
* `update` - (Default `180m`)
* `delete` - (Default `90m`)

## Import

ElastiCache Snapshot can be imported using the `snapshot_name`, e.g.,

```
$ terraform import aws_elasticache_snapshot.example cluster-example-snapshot
```
