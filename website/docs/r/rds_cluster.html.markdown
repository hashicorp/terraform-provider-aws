---
subcategory: "RDS (Relational Database)"
layout: "aws"
page_title: "AWS: aws_rds_cluster"
description: |-
  Manages an RDS Aurora Cluster or a RDS Multi-AZ DB Cluster
---

# Resource: aws_rds_cluster

Manages a [RDS Aurora Cluster][2] or a [RDS Multi-AZ DB Cluster](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/multi-az-db-clusters-concepts.html). To manage cluster instances that inherit configuration from the cluster (when not running the cluster in `serverless` engine mode), see the [`aws_rds_cluster_instance` resource](/docs/providers/aws/r/rds_cluster_instance.html). To manage non-Aurora DB instances (e.g., MySQL, PostgreSQL, SQL Server, etc.), see the [`aws_db_instance` resource](/docs/providers/aws/r/db_instance.html).

For information on the difference between the available Aurora MySQL engines see [Comparison between Aurora MySQL 1 and Aurora MySQL 2](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/AuroraMySQL.Updates.20180206.html) in the Amazon RDS User Guide.

Changes to an RDS Cluster can occur when you manually change a parameter, such as `port`, and are reflected in the next maintenance window. Because of this, Terraform may report a difference in its planning phase because a modification has not yet taken place. You can use the `apply_immediately` flag to instruct the service to apply the change immediately (see documentation below).

~> **Note:** Multi-AZ DB clusters are supported only for the MySQL and PostgreSQL DB engines.

~> **Note:** `ca_certificate_identifier` is only supported for Multi-AZ DB clusters.

~> **Note:** using `apply_immediately` can result in a brief downtime as the server reboots. See the AWS Docs on [RDS Maintenance][4] for more information.

~> **Note:** All arguments including the username and password will be stored in the raw state as plain-text.
[Read more about sensitive data in state](https://www.terraform.io/docs/state/sensitive-data.html).

~> **NOTE on RDS Clusters and RDS Cluster Role Associations:** Terraform provides both a standalone [RDS Cluster Role Association](rds_cluster_role_association.html) - (an association between an RDS Cluster and a single IAM Role) and an RDS Cluster resource with `iam_roles` attributes. Use one resource or the other to associate IAM Roles and RDS Clusters. Not doing so will cause a conflict of associations and will result in the association being overwritten.

## Example Usage

### Aurora MySQL 2.x (MySQL 5.7)

```terraform
resource "aws_rds_cluster" "default" {
  cluster_identifier      = "aurora-cluster-demo"
  engine                  = "aurora-mysql"
  engine_version          = "5.7.mysql_aurora.2.03.2"
  availability_zones      = ["us-west-2a", "us-west-2b", "us-west-2c"]
  database_name           = "mydb"
  master_username         = "foo"
  master_password         = "must_be_eight_characters"
  backup_retention_period = 5
  preferred_backup_window = "07:00-09:00"
}
```

### Aurora MySQL 1.x (MySQL 5.6)

```terraform
resource "aws_rds_cluster" "default" {
  cluster_identifier      = "aurora-cluster-demo"
  availability_zones      = ["us-west-2a", "us-west-2b", "us-west-2c"]
  database_name           = "mydb"
  master_username         = "foo"
  master_password         = "must_be_eight_characters"
  backup_retention_period = 5
  preferred_backup_window = "07:00-09:00"
}
```

### Aurora with PostgreSQL engine

```terraform
resource "aws_rds_cluster" "postgresql" {
  cluster_identifier      = "aurora-cluster-demo"
  engine                  = "aurora-postgresql"
  availability_zones      = ["us-west-2a", "us-west-2b", "us-west-2c"]
  database_name           = "mydb"
  master_username         = "foo"
  master_password         = "must_be_eight_characters"
  backup_retention_period = 5
  preferred_backup_window = "07:00-09:00"
}
```

### RDS Multi-AZ Cluster

-> More information about RDS Multi-AZ Clusters can be found in the [RDS User Guide](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/multi-az-db-clusters-concepts.html).

To create a Multi-AZ RDS cluster, you must additionally specify the `engine`, `storage_type`, `allocated_storage`, `iops` and `db_cluster_instance_class` attributes.

```terraform
resource "aws_rds_cluster" "example" {
  cluster_identifier        = "example"
  availability_zones        = ["us-west-2a", "us-west-2b", "us-west-2c"]
  engine                    = "mysql"
  db_cluster_instance_class = "db.r6gd.xlarge"
  storage_type              = "io1"
  allocated_storage         = 100
  iops                      = 1000
  master_username           = "test"
  master_password           = "mustbeeightcharaters"
}
```

### RDS Serverless v2 Cluster

-> More information about RDS Serverless v2 Clusters can be found in the [RDS User Guide](https://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/aurora-serverless-v2.html).

~> **Note:** Unlike Serverless v1, in Serverless v2 the `storage_encrypted` value is set to `false` by default.
This is because Serverless v1 uses the `serverless` `engine_mode`, but Serverless v2 uses the `provisioned` `engine_mode`.

To create a Serverless v2 RDS cluster, you must additionally specify the `engine_mode` and `serverlessv2_scaling_configuration` attributes. An `aws_rds_cluster_instance` resource must also be added to the cluster with the `instance_class` attribute specified.

```terraform
resource "aws_rds_cluster" "example" {
  cluster_identifier = "example"
  engine             = "aurora-postgresql"
  engine_mode        = "provisioned"
  engine_version     = "13.6"
  database_name      = "test"
  master_username    = "test"
  master_password    = "must_be_eight_characters"
  storage_encrypted  = true

  serverlessv2_scaling_configuration {
    max_capacity = 1.0
    min_capacity = 0.5
  }
}

resource "aws_rds_cluster_instance" "example" {
  cluster_identifier = aws_rds_cluster.example.id
  instance_class     = "db.serverless"
  engine             = aws_rds_cluster.example.engine
  engine_version     = aws_rds_cluster.example.engine_version
}
```

### RDS/Aurora Managed Master Passwords via Secrets Manager, default KMS Key

-> More information about RDS/Aurora Aurora integrates with Secrets Manager to manage master user passwords for your DB clusters can be found in the [RDS User Guide](https://aws.amazon.com/about-aws/whats-new/2022/12/amazon-rds-integration-aws-secrets-manager/) and [Aurora User Guide](https://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/rds-secrets-manager.html).

You can specify the `manage_master_user_password` attribute to enable managing the master password with Secrets Manager. You can also update an existing cluster to use Secrets Manager by specify the `manage_master_user_password` attribute and removing the `master_password` attribute (removal is required).

```terraform
resource "aws_rds_cluster" "test" {
  cluster_identifier          = "example"
  database_name               = "test"
  manage_master_user_password = true
  master_username             = "test"
}
```

### RDS/Aurora Managed Master Passwords via Secrets Manager, specific KMS Key

-> More information about RDS/Aurora Aurora integrates with Secrets Manager to manage master user passwords for your DB clusters can be found in the [RDS User Guide](https://aws.amazon.com/about-aws/whats-new/2022/12/amazon-rds-integration-aws-secrets-manager/) and [Aurora User Guide](https://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/rds-secrets-manager.html).

You can specify the `master_user_secret_kms_key_id` attribute to specify a specific KMS Key.

```terraform
resource "aws_kms_key" "example" {
  description = "Example KMS Key"
}

resource "aws_rds_cluster" "test" {
  cluster_identifier            = "example"
  database_name                 = "test"
  manage_master_user_password   = true
  master_username               = "test"
  master_user_secret_kms_key_id = aws_kms_key.example.key_id
}
```

### Global Cluster Restored From Snapshot

```terraform
data "aws_db_cluster_snapshot" "example" {
  db_cluster_identifier = "example-original-cluster"
  most_recent           = true
}

resource "aws_rds_cluster" "example" {
  # Because the global cluster is sourced from this cluster, the initial 
  # engine and engine_version values are defined here and automatically
  # inherited by the global cluster.
  engine         = "aurora"
  engine_version = "5.6.mysql_aurora.1.22.4"

  cluster_identifier  = "example"
  snapshot_identifier = data.aws_db_cluster_snapshot.example.id

  lifecycle {
    ignore_changes = [snapshot_identifier, global_cluster_identifier]
  }
}

resource "aws_rds_global_cluster" "example" {
  global_cluster_identifier    = "example"
  source_db_cluster_identifier = aws_rds_cluster.example.arn
  force_destroy                = true
}
```

## Argument Reference

For more detailed documentation about each argument, refer to
the AWS official documentation :

* [create-db-cluster](https://docs.aws.amazon.com/cli/latest/reference/rds/create-db-cluster.html)
* [modify-db-cluster](https://docs.aws.amazon.com/cli/latest/reference/rds/modify-db-cluster.html)

This resource supports the following arguments:

* `allocated_storage` - (Optional, Required for Multi-AZ DB cluster) The amount of storage in gibibytes (GiB) to allocate to each DB instance in the Multi-AZ DB cluster.
* `allow_major_version_upgrade` - (Optional) Enable to allow major engine version upgrades when changing engine versions. Defaults to `false`.
* `apply_immediately` - (Optional) Specifies whether any cluster modifications are applied immediately, or during the next maintenance window. Default is `false`. See [Amazon RDS Documentation for more information.](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Overview.DBInstance.Modifying.html)
* `availability_zones` - (Optional) List of EC2 Availability Zones for the DB cluster storage where DB cluster instances can be created.
  RDS automatically assigns 3 AZs if less than 3 AZs are configured, which will show as a difference requiring resource recreation next Terraform apply.
  We recommend specifying 3 AZs or using [the `lifecycle` configuration block `ignore_changes` argument](https://www.terraform.io/docs/configuration/meta-arguments/lifecycle.html#ignore_changes) if necessary.
  A maximum of 3 AZs can be configured.
* `backtrack_window` - (Optional) Target backtrack window, in seconds. Only available for `aurora` and `aurora-mysql` engines currently. To disable backtracking, set this value to `0`. Defaults to `0`. Must be between `0` and `259200` (72 hours)
* `backup_retention_period` - (Optional) Days to retain backups for. Default `1`
* `ca_certificate_identifier` - (Optional) The CA certificate identifier to use for the DB cluster's server certificate.
* `cluster_identifier_prefix` - (Optional, Forces new resource) Creates a unique cluster identifier beginning with the specified prefix. Conflicts with `cluster_identifier`.
* `cluster_identifier` - (Optional, Forces new resources) The cluster identifier. If omitted, Terraform will assign a random, unique identifier.
* `copy_tags_to_snapshot` – (Optional, boolean) Copy all Cluster `tags` to snapshots. Default is `false`.
* `database_name` - (Optional) Name for an automatically created database on cluster creation. There are different naming restrictions per database engine: [RDS Naming Constraints][5]
* `db_cluster_instance_class` - (Optional, Required for Multi-AZ DB cluster) The compute and memory capacity of each DB instance in the Multi-AZ DB cluster, for example `db.m6g.xlarge`. Not all DB instance classes are available in all AWS Regions, or for all database engines. For the full list of DB instance classes and availability for your engine, see [DB instance class](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Concepts.DBInstanceClass.html) in the Amazon RDS User Guide.
* `db_cluster_parameter_group_name` - (Optional) A cluster parameter group to associate with the cluster.
* `db_instance_parameter_group_name` - (Optional) Instance parameter group to associate with all instances of the DB cluster. The `db_instance_parameter_group_name` parameter is only valid in combination with the `allow_major_version_upgrade` parameter.
* `db_subnet_group_name` - (Optional) DB subnet group to associate with this DB cluster.
  **NOTE:** This must match the `db_subnet_group_name` specified on every [`aws_rds_cluster_instance`](/docs/providers/aws/r/rds_cluster_instance.html) in the cluster.
* `db_system_id` - (Optional) For use with RDS Custom.
* `delete_automated_backups` - (Optional) Specifies whether to remove automated backups immediately after the DB cluster is deleted. Default is `true`.
* `deletion_protection` - (Optional) If the DB cluster should have deletion protection enabled.
  The database can't be deleted when this value is set to `true`.
  The default is `false`.
* `domain` - (Optional) The ID of the Directory Service Active Directory domain to create the cluster in.
* `domain_iam_role_name` - (Optional, but required if `domain` is provided) The name of the IAM role to be used when making API calls to the Directory Service.
* `enable_global_write_forwarding` - (Optional) Whether cluster should forward writes to an associated global cluster. Applied to secondary clusters to enable them to forward writes to an [`aws_rds_global_cluster`](/docs/providers/aws/r/rds_global_cluster.html)'s primary cluster. See the [User Guide for Aurora](https://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/aurora-global-database-write-forwarding.html) for more information.
* `enable_http_endpoint` - (Optional) Enable HTTP endpoint (data API). Only valid for some combinations of `engine_mode`, `engine` and `engine_version` and only available in some regions. See the [Region and version availability](https://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/data-api.html#data-api.regions) section of the documentation. This option also does not work with any of these options specified: `snapshot_identifier`, `replication_source_identifier`, `s3_import`.
* `enable_local_write_forwarding` - (Optional) Whether read replicas can forward write operations to the writer DB instance in the DB cluster. By default, write operations aren't allowed on reader DB instances.. See the [User Guide for Aurora](https://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/aurora-mysql-write-forwarding.html) for more information. **NOTE:** Local write forwarding requires Aurora MySQL version 3.04 or higher.
* `enabled_cloudwatch_logs_exports` - (Optional) Set of log types to export to cloudwatch. If omitted, no logs will be exported. The following log types are supported: `audit`, `error`, `general`, `slowquery`, `postgresql` (PostgreSQL).
* `engine_mode` - (Optional) Database engine mode. Valid values: `global` (only valid for Aurora MySQL 1.21 and earlier), `parallelquery`, `provisioned`, `serverless`. Defaults to: `provisioned`. See the [RDS User Guide](https://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/aurora-serverless.html) for limitations when using `serverless`.
* `engine_lifecycle_support` - (Optional) The life cycle type for this DB instance. This setting is valid for cluster types Aurora DB clusters and Multi-AZ DB clusters. Valid values are `open-source-rds-extended-support`, `open-source-rds-extended-support-disabled`. Default value is `open-source-rds-extended-support`. [Using Amazon RDS Extended Support]: https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/extended-support.html
* `engine_version` - (Optional) Database engine version. Updating this argument results in an outage. See the [Aurora MySQL](https://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/AuroraMySQL.Updates.html) and [Aurora Postgres](https://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/AuroraPostgreSQL.Updates.html) documentation for your configured engine to determine this value, or by running `aws rds describe-db-engine-versions`. For example with Aurora MySQL 2, a potential value for this argument is `5.7.mysql_aurora.2.03.2`. The value can contain a partial version where supported by the API. The actual engine version used is returned in the attribute `engine_version_actual`, , see [Attribute Reference](#attribute-reference) below.
* `engine` - (Required) Name of the database engine to be used for this DB cluster. Valid Values: `aurora-mysql`, `aurora-postgresql`, `mysql`, `postgres`. (Note that `mysql` and `postgres` are Multi-AZ RDS clusters).
* `final_snapshot_identifier` - (Optional) Name of your final DB snapshot when this DB cluster is deleted. If omitted, no final snapshot will be made.
* `global_cluster_identifier` - (Optional) Global cluster identifier specified on [`aws_rds_global_cluster`](/docs/providers/aws/r/rds_global_cluster.html).
* `iam_database_authentication_enabled` - (Optional) Specifies whether or not mappings of AWS Identity and Access Management (IAM) accounts to database accounts is enabled. Please see [AWS Documentation](https://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/UsingWithRDS.IAMDBAuth.html) for availability and limitations.
* `iam_roles` - (Optional) List of ARNs for the IAM roles to associate to the RDS Cluster.
* `iops` - (Optional) Amount of Provisioned IOPS (input/output operations per second) to be initially allocated for each DB instance in the Multi-AZ DB cluster. For information about valid Iops values, see [Amazon RDS Provisioned IOPS storage to improve performance](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/CHAP_Storage.html#USER_PIOPS) in the Amazon RDS User Guide. (This setting is required to create a Multi-AZ DB cluster). Must be a multiple between .5 and 50 of the storage amount for the DB cluster.
* `kms_key_id` - (Optional) ARN for the KMS encryption key. When specifying `kms_key_id`, `storage_encrypted` needs to be set to true.
* `manage_master_user_password` - (Optional) Set to true to allow RDS to manage the master user password in Secrets Manager. Cannot be set if `master_password` is provided.
* `master_password` - (Required unless `manage_master_user_password` is set to true or unless a `snapshot_identifier` or `replication_source_identifier` is provided or unless a `global_cluster_identifier` is provided when the cluster is the "secondary" cluster of a global database) Password for the master DB user. Note that this may show up in logs, and it will be stored in the state file. Please refer to the [RDS Naming Constraints][5]. Cannot be set if `manage_master_user_password` is set to `true`.
* `master_user_secret_kms_key_id` - (Optional) Amazon Web Services KMS key identifier is the key ARN, key ID, alias ARN, or alias name for the KMS key. To use a KMS key in a different Amazon Web Services account, specify the key ARN or alias ARN. If not specified, the default KMS key for your Amazon Web Services account is used.
* `master_username` - (Required unless a `snapshot_identifier` or `replication_source_identifier` is provided or unless a `global_cluster_identifier` is provided when the cluster is the "secondary" cluster of a global database) Username for the master DB user. Please refer to the [RDS Naming Constraints][5]. This argument does not support in-place updates and cannot be changed during a restore from snapshot.
* `network_type` - (Optional) Network type of the cluster. Valid values: `IPV4`, `DUAL`.
* `performance_insights_enabled` - (Optional) Valid only for Non-Aurora Multi-AZ DB Clusters. Enables Performance Insights for the RDS Cluster
* `performance_insights_kms_key_id` - (Optional) Valid only for Non-Aurora Multi-AZ DB Clusters. Specifies the KMS Key ID to encrypt Performance Insights data. If not specified, the default RDS KMS key will be used (`aws/rds`).
* `performance_insights_retention_period` - (Optional) Valid only for Non-Aurora Multi-AZ DB Clusters. Specifies the amount of time to retain performance insights data for. Defaults to 7 days if Performance Insights are enabled. Valid values are `7`, `month * 31` (where month is a number of months from 1-23), and `731`. See [here](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_PerfInsights.Overview.cost.html) for more information on retention periods.
* `port` - (Optional) Port on which the DB accepts connections.
* `preferred_backup_window` - (Optional) Daily time range during which automated backups are created if automated backups are enabled using the BackupRetentionPeriod parameter.Time in UTC. Default: A 30-minute window selected at random from an 8-hour block of time per region, e.g. `04:00-09:00`.
* `preferred_maintenance_window` - (Optional) Weekly time range during which system maintenance can occur, in (UTC) e.g., `wed:04:00-wed:04:30`
* `replication_source_identifier` - (Optional) ARN of a source DB cluster or DB instance if this DB cluster is to be created as a Read Replica. If DB Cluster is part of a Global Cluster, use the [`lifecycle` configuration block `ignore_changes` argument](https://www.terraform.io/docs/configuration/meta-arguments/lifecycle.html#ignore_changes) to prevent Terraform from showing differences for this argument instead of configuring this value.
* `restore_to_point_in_time` - (Optional) Nested attribute for [point in time restore](https://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/aurora-pitr.html). More details below.
* `scaling_configuration` - (Optional) Nested attribute with scaling properties. Only valid when `engine_mode` is set to `serverless`. More details below.
* `serverlessv2_scaling_configuration`- (Optional) Nested attribute with scaling properties for ServerlessV2. Only valid when `engine_mode` is set to `provisioned`. More details below.
* `skip_final_snapshot` - (Optional) Determines whether a final DB snapshot is created before the DB cluster is deleted. If true is specified, no DB snapshot is created. If false is specified, a DB snapshot is created before the DB cluster is deleted, using the value from `final_snapshot_identifier`. Default is `false`.
* `snapshot_identifier` - (Optional) Specifies whether or not to create this cluster from a snapshot. You can use either the name or ARN when specifying a DB cluster snapshot, or the ARN when specifying a DB snapshot. Conflicts with `global_cluster_identifier`. Clusters cannot be restored from snapshot **and** joined to an existing global cluster in a single operation. See the [AWS documentation](https://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/aurora-global-database-getting-started.html#aurora-global-database.use-snapshot) or the [Global Cluster Restored From Snapshot example](#global-cluster-restored-from-snapshot) for instructions on building a global cluster starting with a snapshot.
* `source_region` - (Optional) The source region for an encrypted replica DB cluster.
* `storage_encrypted` - (Optional) Specifies whether the DB cluster is encrypted. The default is `false` for `provisioned` `engine_mode` and `true` for `serverless` `engine_mode`. When restoring an unencrypted `snapshot_identifier`, the `kms_key_id` argument must be provided to encrypt the restored cluster. Terraform will only perform drift detection if a configuration value is provided.
* `storage_type` - (Optional, Required for Multi-AZ DB cluster) (Forces new for Multi-AZ DB clusters) Specifies the storage type to be associated with the DB cluster. For Aurora DB clusters, `storage_type` modifications can be done in-place. For Multi-AZ DB Clusters, the `iops` argument must also be set. Valid values are: `""`, `aurora-iopt1` (Aurora DB Clusters); `io1`, `io2` (Multi-AZ DB Clusters). Default: `""` (Aurora DB Clusters); `io1` (Multi-AZ DB Clusters).
* `tags` - (Optional) A map of tags to assign to the DB cluster. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `vpc_security_group_ids` - (Optional) List of VPC security groups to associate with the Cluster

### S3 Import Options

Full details on the core parameters and impacts are in the API Docs: [RestoreDBClusterFromS3](https://docs.aws.amazon.com/AmazonRDS/latest/APIReference/API_RestoreDBClusterFromS3.html). Requires that the S3 bucket be in the same region as the RDS cluster you're trying to create. Sample:

~> **NOTE:** RDS Aurora Serverless does not support loading data from S3, so its not possible to directly use `engine_mode` set to `serverless` with `s3_import`.

```terraform
resource "aws_rds_cluster" "db" {
  engine = "aurora"

  s3_import {
    source_engine         = "mysql"
    source_engine_version = "5.6"
    bucket_name           = "mybucket"
    bucket_prefix         = "backups"
    ingestion_role        = "arn:aws:iam::1234567890:role/role-xtrabackup-rds-restore"
  }
}
```

* `bucket_name` - (Required) Bucket name where your backup is stored
* `bucket_prefix` - (Optional) Can be blank, but is the path to your backup
* `ingestion_role` - (Required) Role applied to load the data.
* `source_engine` - (Required) Source engine for the backup
* `source_engine_version` - (Required) Version of the source engine used to make the backup

This will not recreate the resource if the S3 object changes in some way. It's only used to initialize the database. This only works currently with the aurora engine. See AWS for currently supported engines and options. See [Aurora S3 Migration Docs](https://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/AuroraMySQL.Migrating.ExtMySQL.html#AuroraMySQL.Migrating.ExtMySQL.S3).

### restore_to_point_in_time Argument Reference

~> **NOTE:**  The DB cluster is created from the source DB cluster with the same configuration as the original DB cluster, except that the new DB cluster is created with the default DB security group. Thus, the following arguments should only be specified with the source DB cluster's respective values: `database_name`, `master_username`, `storage_encrypted`, `replication_source_identifier`, and `source_region`.

~> **NOTE:**  One of `source_cluster_identifier` or `source_cluster_resource_id` must be specified.

Example:

```terraform
resource "aws_rds_cluster" "example-clone" {
  # ... other configuration ...

  restore_to_point_in_time {
    source_cluster_identifier  = "example"
    restore_type               = "copy-on-write"
    use_latest_restorable_time = true
  }
}
```

* `source_cluster_identifier` - (Optional) Identifier of the source database cluster from which to restore. When restoring from a cluster in another AWS account, the identifier is the ARN of that cluster.
* `source_cluster_resource_id` - (Optional) Cluster resource ID of the source database cluster from which to restore. To be used for restoring a deleted cluster in the same account which still has a retained automatic backup available.
* `restore_type` - (Optional) Type of restore to be performed.
   Valid options are `full-copy` (default) and `copy-on-write`.
* `use_latest_restorable_time` - (Optional) Set to true to restore the database cluster to the latest restorable backup time. Defaults to false. Conflicts with `restore_to_time`.
* `restore_to_time` - (Optional) Date and time in UTC format to restore the database cluster to. Conflicts with `use_latest_restorable_time`.

### scaling_configuration Argument Reference

~> **NOTE:** `scaling_configuration` configuration is only valid when `engine_mode` is set to `serverless`.

Example:

```terraform
resource "aws_rds_cluster" "example" {
  # ... other configuration ...

  engine_mode = "serverless"

  scaling_configuration {
    auto_pause               = true
    max_capacity             = 256
    min_capacity             = 2
    seconds_before_timeout   = 360
    seconds_until_auto_pause = 300
    timeout_action           = "ForceApplyCapacityChange"
  }
}
```

* `auto_pause` - (Optional) Whether to enable automatic pause. A DB cluster can be paused only when it's idle (it has no connections). If a DB cluster is paused for more than seven days, the DB cluster might be backed up with a snapshot. In this case, the DB cluster is restored when there is a request to connect to it. Defaults to `true`.
* `max_capacity` - (Optional) Maximum capacity for an Aurora DB cluster in `serverless` DB engine mode. The maximum capacity must be greater than or equal to the minimum capacity. Valid Aurora MySQL capacity values are `1`, `2`, `4`, `8`, `16`, `32`, `64`, `128`, `256`. Valid Aurora PostgreSQL capacity values are (`2`, `4`, `8`, `16`, `32`, `64`, `192`, and `384`). Defaults to `16`.
* `min_capacity` - (Optional) Minimum capacity for an Aurora DB cluster in `serverless` DB engine mode. The minimum capacity must be lesser than or equal to the maximum capacity. Valid Aurora MySQL capacity values are `1`, `2`, `4`, `8`, `16`, `32`, `64`, `128`, `256`. Valid Aurora PostgreSQL capacity values are (`2`, `4`, `8`, `16`, `32`, `64`, `192`, and `384`). Defaults to `1`.
* `seconds_before_timeout` - (Optional) Amount of time, in seconds, that Aurora Serverless v1 tries to find a scaling point to perform seamless scaling before enforcing the timeout action. Valid values are `60` through `600`. Defaults to `300`.
* `seconds_until_auto_pause` - (Optional) Time, in seconds, before an Aurora DB cluster in serverless mode is paused. Valid values are `300` through `86400`. Defaults to `300`.
* `timeout_action` - (Optional) Action to take when the timeout is reached. Valid values: `ForceApplyCapacityChange`, `RollbackCapacityChange`. Defaults to `RollbackCapacityChange`. See [documentation](https://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/aurora-serverless-v1.how-it-works.html#aurora-serverless.how-it-works.timeout-action).

### serverlessv2_scaling_configuration Argument Reference

~> **NOTE:** serverlessv2_scaling_configuration configuration is only valid when engine_mode is set to provisioned

Example:

```terraform
resource "aws_rds_cluster" "example" {
  # ... other configuration ...

  serverlessv2_scaling_configuration {
    max_capacity = 128.0
    min_capacity = 0.5
  }
}
```

* `max_capacity` - (Required) Maximum capacity for an Aurora DB cluster in `provisioned` DB engine mode. The maximum capacity must be greater than or equal to the minimum capacity. Valid capacity values are in a range of `0.5` up to `128` in steps of `0.5`.
* `min_capacity` - (Required) Minimum capacity for an Aurora DB cluster in `provisioned` DB engine mode. The minimum capacity must be lesser than or equal to the maximum capacity. Valid capacity values are in a range of `0.5` up to `128` in steps of `0.5`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Amazon Resource Name (ARN) of cluster
* `id` - RDS Cluster Identifier
* `cluster_identifier` - RDS Cluster Identifier
* `cluster_resource_id` - RDS Cluster Resource ID
* `cluster_members` – List of RDS Instances that are a part of this cluster
* `availability_zones` - Availability zone of the instance
* `backup_retention_period` - Backup retention period
* `ca_certificate_identifier` - CA identifier of the CA certificate used for the DB instance's server certificate
* `ca_certificate_valid_till` - Expiration date of the DB instance’s server certificate
* `preferred_backup_window` - Daily time range during which the backups happen
* `preferred_maintenance_window` - Maintenance window
* `endpoint` - DNS address of the RDS instance
* `reader_endpoint` - Read-only endpoint for the Aurora cluster, automatically
load-balanced across replicas
* `engine` - Database engine
* `engine_version_actual` - Running version of the database.
* `database_name` - Database name
* `port` - Database port
* `master_username` - Master username for the database
* `master_user_secret` - Block that specifies the master user secret. Only available when `manage_master_user_password` is set to true. [Documented below](#master_user_secret).
* `storage_encrypted` - Specifies whether the DB cluster is encrypted
* `replication_source_identifier` - ARN of the source DB cluster or DB instance if this DB cluster is created as a Read Replica.
* `hosted_zone_id` - Route53 Hosted Zone ID of the endpoint
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

[1]: https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_ReadRepl.html
[2]: https://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/CHAP_Aurora.html
[3]: /docs/providers/aws/r/rds_cluster_instance.html
[4]: https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_UpgradeDBInstance.Maintenance.html
[5]: https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/CHAP_Limits.html#RDS_Limits.Constraints

### master_user_secret

The `master_user_secret` configuration block supports the following attributes:

* `kms_key_id` - Amazon Web Services KMS key identifier that is used to encrypt the secret.
* `secret_arn` - Amazon Resource Name (ARN) of the secret.
* `secret_status` - Status of the secret. Valid Values: `creating` | `active` | `rotating` | `impaired`.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `120m`)
- `update` - (Default `120m`)
- `delete` - (Default `120m`)
any cleanup task during the destroying process.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import RDS Clusters using the `cluster_identifier`. For example:

```terraform
import {
  to = aws_rds_cluster.aurora_cluster
  id = "aurora-prod-cluster"
}
```

Using `terraform import`, import RDS Clusters using the `cluster_identifier`. For example:

```console
% terraform import aws_rds_cluster.aurora_cluster aurora-prod-cluster
```
