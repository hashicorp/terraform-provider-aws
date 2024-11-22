---
subcategory: "DocumentDB"
layout: "aws"
page_title: "AWS: aws_docdb_cluster"
description: |-
  Manages a DocumentDB Aurora Cluster
---

# Resource: aws_docdb_cluster

Manages a DocumentDB Cluster.

Changes to a DocumentDB Cluster can occur when you manually change a
parameter, such as `port`, and are reflected in the next maintenance
window. Because of this, Terraform may report a difference in its planning
phase because a modification has not yet taken place. You can use the
`apply_immediately` flag to instruct the service to apply the change immediately
(see documentation below).

~> **Note:** using `apply_immediately` can result in a brief downtime as the server reboots.
~> **Note:** All arguments including the username and password will be stored in the raw state as plain-text.
[Read more about sensitive data in state](https://www.terraform.io/docs/state/sensitive-data.html).

## Example Usage

```terraform
resource "aws_docdb_cluster" "docdb" {
  cluster_identifier      = "my-docdb-cluster"
  engine                  = "docdb"
  master_username         = "foo"
  master_password         = "mustbeeightchars"
  backup_retention_period = 5
  preferred_backup_window = "07:00-09:00"
  skip_final_snapshot     = true
}
```

## Argument Reference

For more detailed documentation about each argument, refer to
the [AWS official documentation](https://docs.aws.amazon.com/cli/latest/reference/docdb/create-db-cluster.html).

This resource supports the following arguments:

* `allow_major_version_upgrade` - (Optional) A value that indicates whether major version upgrades are allowed. Constraints: You must allow major version upgrades when specifying a value for the EngineVersion parameter that is a different major version than the DB cluster's current version.
* `apply_immediately` - (Optional) Specifies whether any cluster modifications
     are applied immediately, or during the next maintenance window. Default is
     `false`.
* `availability_zones` - (Optional) A list of EC2 Availability Zones that
  instances in the DB cluster can be created in.
* `backup_retention_period` - (Optional) The days to retain backups for. Default `1`
* `cluster_identifier_prefix` - (Optional, Forces new resource) Creates a unique cluster identifier beginning with the specified prefix. Conflicts with `cluster_identifier`.
* `cluster_identifier` - (Optional, Forces new resources) The cluster identifier. If omitted, Terraform will assign a random, unique identifier.
* `db_subnet_group_name` - (Optional) A DB subnet group to associate with this DB instance.
* `db_cluster_parameter_group_name` - (Optional) A cluster parameter group to associate with the cluster.
* `deletion_protection` - (Optional) A boolean value that indicates whether the DB cluster has deletion protection enabled. The database can't be deleted when deletion protection is enabled. Defaults to `false`.
* `enabled_cloudwatch_logs_exports` - (Optional) List of log types to export to cloudwatch. If omitted, no logs will be exported.
   The following log types are supported: `audit`, `profiler`.
* `engine_version` - (Optional) The database engine version. Updating this argument results in an outage.
* `engine` - (Optional) The name of the database engine to be used for this DB cluster. Defaults to `docdb`. Valid values: `docdb`.
* `final_snapshot_identifier` - (Optional) The name of your final DB snapshot
    when this DB cluster is deleted. If omitted, no final snapshot will be
    made.
* `global_cluster_identifier` - (Optional) The global cluster identifier specified on [`aws_docdb_global_cluster`](/docs/providers/aws/r/docdb_global_cluster.html).
* `kms_key_id` - (Optional) The ARN for the KMS encryption key. When specifying `kms_key_id`, `storage_encrypted` needs to be set to true.
* `master_password` - (Required unless a `snapshot_identifier` or unless a `global_cluster_identifier` is provided when the cluster is the "secondary" cluster of a global database) Password for the master DB user. Note that this may
    show up in logs, and it will be stored in the state file. Please refer to the DocumentDB Naming Constraints.
* `master_username` - (Required unless a `snapshot_identifier` or unless a `global_cluster_identifier` is provided when the cluster is the "secondary" cluster of a global database) Username for the master DB user.
* `port` - (Optional) The port on which the DB accepts connections
* `preferred_backup_window` - (Optional) The daily time range during which automated backups are created if automated backups are enabled using the BackupRetentionPeriod parameter.Time in UTC
Default: A 30-minute window selected at random from an 8-hour block of time per regionE.g., 04:00-09:00
* `preferred_maintenance_window` - (Optional) The weekly time range during which system maintenance can occur, in (UTC) e.g., wed:04:00-wed:04:30
* `restore_to_point_in_time` - (Optional, Forces new resource) A configuration block for restoring a DB instance to an arbitrary point in time. Requires the `identifier` argument to be set with the name of the new DB instance to be created. See [Restore To Point In Time](#restore-to-point-in-time) below for details.
* `skip_final_snapshot` - (Optional) Determines whether a final DB snapshot is created before the DB cluster is deleted. If true is specified, no DB snapshot is created. If false is specified, a DB snapshot is created before the DB cluster is deleted, using the value from `final_snapshot_identifier`. Default is `false`.
* `snapshot_identifier` - (Optional) Specifies whether or not to create this cluster from a snapshot. You can use either the name or ARN when specifying a DB cluster snapshot, or the ARN when specifying a DB snapshot. Automated snapshots **should not** be used for this attribute, unless from a different cluster. Automated snapshots are deleted as part of cluster destruction when the resource is replaced.
* `storage_encrypted` - (Optional) Specifies whether the DB cluster is encrypted. The default is `false`.
* `storage_type` - (Optional) The storage type to associate with the DB cluster. Valid values: `standard`, `iopt1`.
* `tags` - (Optional) A map of tags to assign to the DB cluster. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `vpc_security_group_ids` - (Optional) List of VPC security groups to associate
  with the Cluster

### Restore To Point In Time

The `restore_to_point_in_time` block supports the following arguments:

* `restore_to_time` - (Optional) The date and time to restore from. Value must be a time in Universal Coordinated Time (UTC) format and must be before the latest restorable time for the DB instance. Cannot be specified with `use_latest_restorable_time`.
* `restore_type` - (Optional) The type of restore to be performed. Valid values are `full-copy`, `copy-on-write`.
* `source_cluster_identifier` - (Required) The identifier of the source DB cluster from which to restore. Must match the identifier of an existing DB cluster.
* `use_latest_restorable_time` - (Optional) A boolean value that indicates whether the DB cluster is restored from the latest backup time. Defaults to `false`. Cannot be specified with `restore_to_time`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Amazon Resource Name (ARN) of cluster
* `cluster_members` – List of DocumentDB Instances that are a part of this cluster
* `cluster_resource_id` - The DocumentDB Cluster Resource ID
* `endpoint` - The DNS address of the DocumentDB instance
* `hosted_zone_id` - The Route53 Hosted Zone ID of the endpoint
* `id` - The DocumentDB Cluster Identifier
* `reader_endpoint` - A read-only endpoint for the DocumentDB cluster, automatically load-balanced across replicas
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `120m`)
- `update` - (Default `120m`)
- `delete` - (Default `120m`)
any cleanup task during the destroying process.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import DocumentDB Clusters using the `cluster_identifier`. For example:

```terraform
import {
  to = aws_docdb_cluster.docdb_cluster
  id = "docdb-prod-cluster"
}
```

Using `terraform import`, import DocumentDB Clusters using the `cluster_identifier`. For example:

```console
% terraform import aws_docdb_cluster.docdb_cluster docdb-prod-cluster
```
