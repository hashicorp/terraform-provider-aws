---
subcategory: "Neptune"
layout: "aws"
page_title: "AWS: aws_neptune_cluster"
description: |-
  Provides details about an AWS Neptune Cluster.
---

# Data Source: aws_neptune_cluster

Provides details about an AWS Neptune Cluster.

## Example Usage

```terraform
data "aws_neptune_cluster" "default" {
  identifier = "neptune-cluster-demo
}

output "cluster_id" {
  value = data.aws_neptune_cluster.default.id
}

output "writer_endpoint" {
  value = data.aws_neptune_cluster.default.endpoint
}
```


## Argument Reference

The following arguments are required:

* `identifier` (Required) — DB cluster **identifier or ARN** of the Neptune cluster.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

- `associated_roles` - List of the IAM roles that are associated with the Neptune cluster.
- `availability_zones` - List of availability zones which instances of the Neptune cluster is created.
- `arn` — The Amazon Resource Name (ARN) for the Neptune cluster.
- `backup_retention_period` — Number of days for which automatic DB snapshots are retained.
- `cluster_id` — Neptune cluster identifier.
- `cluster_parameter_group_name` — Name of the Neptune cluster parameter group for the DB cluster..
- `db_subnet_group` — Subnet group associated with the Neptune cluster.
- `deletion_protection` — Indicates whether or not the DB cluster has deletion protection enabled.
- `enabled_cloudwatch_logs_exports` — List of enabled CloudWatch log exports.
- `endpoint` — Writer/primary connection endpoint for the Neptune cluster.
- `engine` — Neptune engine name. 
- `engine_version` — Neptune engine version.
- `global_cluster_identifier` - Global Neptune cluster identifier if part of global neptune cluster.
- `iam_database_authentication_enabled` — Whether Identity and Access Management (IAM) auth is enabled for Neptune cluster.
- `kms_key_id` — Amazon KMS key ID used for storage encryption.
- `members` — List of cluster members:
  - `db_instance_identifier` — Neptune instance identifier.
  - `is_cluster_writer` — Whether if this member is the primary/writer instance of the cluster.
  - `parameter_group_status` — Status of the member’s parameter group.
  - `promotion_tier` — Failover promotion tier in which a read replica is promoted to the primary instance.
- `multi_az` - Whether the Neptune cluster has instances in multiple Availability Zones.
- `port` — Port number which the database engine is listening on.
- `preferred_backup_window` — Daily time range during which automated backups are created if automated backups are enabled.
- `preferred_maintenance_window` — weekly time range during which system maintenance can occur in UTC.
- `reader_endpoint` — Read-only endpoint for the Neptune cluster. Neptune distributes the connection requests among the Read Replicas in the DB cluster
- `resource_id` — Amazon Region-unique, immutable identifier for the Neptune cluster. **Not** the same as DB cluster identifier.
- `status` —  Current state of the Neptune cluster.
- `storage_encrypted` - Storage encryption status of Neptune cluster.
- `storage_type` — Storage type used for Neptune cluster.
- `tags` - A map of AWS tags to assigned with the Neptune cluster.
- `vpc_security_group_ids` — List of VPC security group IDs associated with the cluster.
