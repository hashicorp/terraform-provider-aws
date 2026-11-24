---
subcategory: "RDS (Relational Database)"
layout: "aws"
page_title: "AWS: aws_rds_cluster"
description: |-
  Provides an RDS cluster data source.
---

# Data Source: aws_rds_cluster

Provides information about an RDS cluster.

## Example Usage

```terraform
data "aws_rds_cluster" "clusterName" {
  cluster_identifier = "clusterName"
}
```

## Argument Reference

This data source supports the following arguments:

* `cluster_identifier` - (Required) Cluster identifier of the RDS cluster.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the cluster.
* `availability_zones` - Availability Zones of the RDS cluster.
* `backtrack_window` - Target backtrack window, in seconds.
* `backup_retention_period` - Days to retain backups for.
* `cluster_members` - List of RDS Instances that are a part of this cluster.
* `cluster_resource_id` - RDS Cluster Resource ID.
* `cluster_scalability_type` - Scalability mode of the cluster.
* `database_insights_mode` - Mode of Database Insights that is enabled for the cluster.
* `database_name` - Name for an automatically created database on cluster creation.
* `db_cluster_parameter_group_name` - Cluster parameter group associated with the cluster.
* `db_subnet_group_name` - DB subnet group associated with the cluster.
* `db_system_id` - System ID of the cluster.
* `deletion_protection` - Whether the cluster has deletion protection enabled.
* `enabled_cloudwatch_logs_exports` - List of log types exported to CloudWatch Logs.
* `endpoint` - DNS address of the RDS instance.
* `engine` - Database engine.
* `engine_mode` - Database engine mode.
* `engine_version` - Database engine version.
* `final_snapshot_identifier` - Name of the final snapshot taken when the cluster is deleted.
* `hosted_zone_id` - Route53 Hosted Zone ID of the endpoint.
* `iam_database_authentication_enabled` - Whether mapping of AWS Identity and Access Management (IAM) accounts to database accounts is enabled.
* `iam_roles` - IAM roles associated with the cluster.
* `kms_key_id` - ARN for the KMS encryption key.
* `master_user_secret` - Block that specifies the master user secret. Only available when `manage_master_user_password` is set to `true`. [Documented below](#master_user_secret-block).
* `master_username` - Master username for the database.
* `monitoring_interval` - Interval, in seconds, between points when Enhanced Monitoring metrics are collected for the cluster.
* `monitoring_role_arn` - ARN of the IAM role used by RDS to send Enhanced Monitoring metrics to CloudWatch Logs.
* `network_type` - Network type of the cluster.
* `port` - Port on which the DB accepts connections.
* `preferred_backup_window` - Daily time range during which automated backups are created.
* `preferred_maintenance_window` - Weekly time range during which system maintenance can occur.
* `reader_endpoint` - Read-only endpoint for the cluster, automatically load-balanced across replicas.
* `replication_source_identifier` - ARN of the source DB cluster or DB instance if this DB cluster is created as a read replica.
* `storage_encrypted` - Whether the DB cluster is encrypted.
* `tags` - Map of tags assigned to the resource.
* `upgrade_rollout_order` - Order in which minor and major version upgrades are applied to the cluster.
* `vpc_security_group_ids` - VPC security groups the cluster belongs to.

### `master_user_secret` Block

* `kms_key_id` - Amazon Web Services KMS key identifier that is used to encrypt the secret.
* `secret_arn` - ARN of the secret.
* `secret_status` - Status of the secret.
