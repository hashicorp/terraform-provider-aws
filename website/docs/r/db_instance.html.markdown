---
subcategory: "RDS (Relational Database)"
layout: "aws"
page_title: "AWS: aws_db_instance"
description: |-
  Provides an RDS instance resource.
---

# Resource: aws_db_instance

Provides an RDS instance resource.  A DB instance is an isolated database
environment in the cloud.  A DB instance can contain multiple user-created
databases.

Changes to a DB instance can occur when you manually change a parameter, such as
`allocated_storage`, and are reflected in the next maintenance window. Because
of this, Terraform may report a difference in its planning phase because a
modification has not yet taken place. You can use the `apply_immediately` flag
to instruct the service to apply the change immediately (see documentation
below).

When upgrading the major version of an engine, `allow_major_version_upgrade` must be set to `true`.

~> **Note:** using `apply_immediately` can result in a brief downtime as the server reboots.
See the AWS Docs on [RDS Instance Maintenance][instance-maintenance] for more information.

~> **Note:** All arguments including the username and password will be stored in the raw state as plain-text.
[Read more about sensitive data instate](https://www.terraform.io/docs/state/sensitive-data.html).

> **Hands-on:** Try the [Manage AWS RDS Instances](https://learn.hashicorp.com/tutorials/terraform/aws-rds) tutorial on HashiCorp Learn.

## RDS Instance Class Types

Amazon RDS supports instance classes for the following use cases: General-purpose, Memory-optimized, Burstable Performance, and Optimized-reads.
For more information please read the AWS RDS documentation about [DB Instance Class Types](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Concepts.DBInstanceClass.html)

## Low-Downtime Updates

By default, RDS applies updates to DB Instances in-place, which can lead to service interruptions.
Low-downtime updates minimize service interruptions by performing the updates with an [RDS Blue/Green deployment][blue-green] and switching over the instances when complete.

Low-downtime updates are only available for DB Instances using MySQL, MariaDB and PostgreSQL,
as other engines are not supported by RDS Blue/Green deployments.
They cannot be used with DB Instances with replicas.

Backups must be enabled to use low-downtime updates.

Enable low-downtime updates by setting `blue_green_update.enabled` to `true`.

## Example Usage

### Basic Usage

```terraform
resource "aws_db_instance" "default" {
  allocated_storage    = 10
  db_name              = "mydb"
  engine               = "mysql"
  engine_version       = "8.0"
  instance_class       = "db.t3.micro"
  username             = "foo"
  password             = "foobarbaz"
  parameter_group_name = "default.mysql8.0"
  skip_final_snapshot  = true
}
```

### RDS Custom for Oracle Usage with Replica

```terraform
# Lookup the available instance classes for the custom engine for the region being operated in
data "aws_rds_orderable_db_instance" "custom-oracle" {
  engine                     = "custom-oracle-ee" # CEV engine to be used
  engine_version             = "19.c.ee.002"      # CEV engine version to be used
  license_model              = "bring-your-own-license"
  storage_type               = "gp3"
  preferred_instance_classes = ["db.r5.xlarge", "db.r5.2xlarge", "db.r5.4xlarge"]
}

# The RDS instance resource requires an ARN. Look up the ARN of the KMS key associated with the CEV.
data "aws_kms_key" "by_id" {
  key_id = "example-ef278353ceba4a5a97de6784565b9f78" # KMS key associated with the CEV
}

resource "aws_db_instance" "default" {
  allocated_storage           = 50
  auto_minor_version_upgrade  = false                         # Custom for Oracle does not support minor version upgrades
  custom_iam_instance_profile = "AWSRDSCustomInstanceProfile" # Instance profile is required for Custom for Oracle. See: https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/custom-setup-orcl.html#custom-setup-orcl.iam-vpc
  backup_retention_period     = 7
  db_subnet_group_name        = local.db_subnet_group_name
  engine                      = data.aws_rds_orderable_db_instance.custom-oracle.engine
  engine_version              = data.aws_rds_orderable_db_instance.custom-oracle.engine_version
  identifier                  = "ee-instance-demo"
  instance_class              = data.aws_rds_orderable_db_instance.custom-oracle.instance_class
  kms_key_id                  = data.aws_kms_key.by_id.arn
  license_model               = data.aws_rds_orderable_db_instance.custom-oracle.license_model
  multi_az                    = false # Custom for Oracle does not support multi-az
  password                    = "avoid-plaintext-passwords"
  username                    = "test"
  storage_encrypted           = true

  timeouts {
    create = "3h"
    delete = "3h"
    update = "3h"
  }
}

resource "aws_db_instance" "test-replica" {
  replicate_source_db         = aws_db_instance.default.identifier
  replica_mode                = "mounted"
  auto_minor_version_upgrade  = false
  custom_iam_instance_profile = "AWSRDSCustomInstanceProfile" # Instance profile is required for Custom for Oracle. See: https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/custom-setup-orcl.html#custom-setup-orcl.iam-vpc
  backup_retention_period     = 7
  identifier                  = "ee-instance-replica"
  instance_class              = data.aws_rds_orderable_db_instance.custom-oracle.instance_class
  kms_key_id                  = data.aws_kms_key.by_id.arn
  multi_az                    = false # Custom for Oracle does not support multi-az
  skip_final_snapshot         = true
  storage_encrypted           = true

  timeouts {
    create = "3h"
    delete = "3h"
    update = "3h"
  }
}
```

### RDS Custom for SQL Server

```terraform
# Lookup the available instance classes for the custom engine for the region being operated in
data "aws_rds_orderable_db_instance" "custom-sqlserver" {
  engine                     = "custom-sqlserver-se" # CEV engine to be used
  engine_version             = "15.00.4249.2.v1"     # CEV engine version to be used
  storage_type               = "gp3"
  preferred_instance_classes = ["db.r5.xlarge", "db.r5.2xlarge", "db.r5.4xlarge"]
}

# The RDS instance resource requires an ARN. Look up the ARN of the KMS key.
data "aws_kms_key" "by_id" {
  key_id = "example-ef278353ceba4a5a97de6784565b9f78" # KMS key
}

resource "aws_db_instance" "example" {
  allocated_storage           = 500
  auto_minor_version_upgrade  = false                                  # Custom for SQL Server does not support minor version upgrades
  custom_iam_instance_profile = "AWSRDSCustomSQLServerInstanceProfile" # Instance profile is required for Custom for SQL Server. See: https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/custom-setup-sqlserver.html#custom-setup-sqlserver.iam
  backup_retention_period     = 7
  db_subnet_group_name        = local.db_subnet_group_name # Copy the subnet group from the RDS Console
  engine                      = data.aws_rds_orderable_db_instance.custom-sqlserver.engine
  engine_version              = data.aws_rds_orderable_db_instance.custom-sqlserver.engine_version
  identifier                  = "sql-instance-demo"
  instance_class              = data.aws_rds_orderable_db_instance.custom-sqlserver.instance_class
  kms_key_id                  = data.aws_kms_key.by_id.arn
  multi_az                    = false # Custom for SQL Server does support multi-az
  password                    = "avoid-plaintext-passwords"
  storage_encrypted           = true
  username                    = "test"

  timeouts {
    create = "3h"
    delete = "3h"
    update = "3h"
  }
}
```

### RDS Db2 Usage

```terraform
# Lookup the default version for the engine. Db2 Standard Edition is `db2-se`, Db2 Advanced Edition is `db2-ae`.
data "aws_rds_engine_version" "default" {
  engine = "db2-se" #Standard Edition
}

# Lookup the available instance classes for the engine in the region being operated in
data "aws_rds_orderable_db_instance" "example" {
  engine                     = data.aws_rds_engine_version.default.engine
  engine_version             = data.aws_rds_engine_version.default.version
  license_model              = "bring-your-own-license"
  storage_type               = "gp3"
  preferred_instance_classes = ["db.t3.small", "db.r6i.large", "db.m6i.large"]
}

# The RDS Db2 instance resource requires licensing information. Create a new parameter group using the default paramater group as a source, and set license information.
resource "aws_db_parameter_group" "example" {
  name   = "db-db2-params"
  family = data.aws_rds_engine_version.default.parameter_group_family

  parameter {
    apply_method = "immediate"
    name         = "rds.ibm_customer_id"
    value        = 0000000000
  }
  parameter {
    apply_method = "immediate"
    name         = "rds.ibm_site_id"
    value        = 0000000000
  }
}

# Create the RDS Db2 instance, use the data sources defined to set attributes
resource "aws_db_instance" "example" {
  allocated_storage       = 100
  backup_retention_period = 7
  db_name                 = "test"
  engine                  = data.aws_rds_orderable_db_instance.example.engine
  engine_version          = data.aws_rds_orderable_db_instance.example.engine_version
  identifier              = "db2-instance-demo"
  instance_class          = data.aws_rds_orderable_db_instance.example.instance_class
  parameter_group_name    = aws_db_parameter_group.example.name
  password                = "avoid-plaintext-passwords"
  username                = "test"
}
```

### Storage Autoscaling

To enable Storage Autoscaling with instances that support the feature, define the `max_allocated_storage` argument higher than the `allocated_storage` argument. Terraform will automatically hide differences with the `allocated_storage` argument value if autoscaling occurs.

```terraform
resource "aws_db_instance" "example" {
  # ... other configuration ...

  allocated_storage     = 50
  max_allocated_storage = 100
}
```

### Managed Master Passwords via Secrets Manager, default KMS Key

-> More information about RDS/Aurora Aurora integrates with Secrets Manager to manage master user passwords for your DB clusters can be found in the [RDS User Guide](https://aws.amazon.com/about-aws/whats-new/2022/12/amazon-rds-integration-aws-secrets-manager/) and [Aurora User Guide](https://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/rds-secrets-manager.html).

You can specify the `manage_master_user_password` attribute to enable managing the master password with Secrets Manager. You can also update an existing cluster to use Secrets Manager by specify the `manage_master_user_password` attribute and removing the `password` attribute (removal is required).

```terraform
resource "aws_db_instance" "default" {
  allocated_storage           = 10
  db_name                     = "mydb"
  engine                      = "mysql"
  engine_version              = "8.0"
  instance_class              = "db.t3.micro"
  manage_master_user_password = true
  username                    = "foo"
  parameter_group_name        = "default.mysql8.0"
}
```

### Managed Master Passwords via Secrets Manager, specific KMS Key

-> More information about RDS/Aurora Aurora integrates with Secrets Manager to manage master user passwords for your DB clusters can be found in the [RDS User Guide](https://aws.amazon.com/about-aws/whats-new/2022/12/amazon-rds-integration-aws-secrets-manager/) and [Aurora User Guide](https://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/rds-secrets-manager.html).

You can specify the `master_user_secret_kms_key_id` attribute to specify a specific KMS Key.

```terraform
resource "aws_kms_key" "example" {
  description = "Example KMS Key"
}

resource "aws_db_instance" "default" {
  allocated_storage             = 10
  db_name                       = "mydb"
  engine                        = "mysql"
  engine_version                = "8.0"
  instance_class                = "db.t3.micro"
  manage_master_user_password   = true
  master_user_secret_kms_key_id = aws_kms_key.example.key_id
  username                      = "foo"
  parameter_group_name          = "default.mysql8.0"
}
```

## Argument Reference

For more detailed documentation about each argument, refer to the [AWS official
documentation](http://docs.aws.amazon.com/AmazonRDS/latest/APIReference/API_CreateDBInstance.html).

This resource supports the following arguments:

* `allocated_storage` - (Required unless a `snapshot_identifier` or `replicate_source_db` is provided) The allocated storage in gibibytes. If `max_allocated_storage` is configured, this argument represents the initial storage allocation and differences from the configuration will be ignored automatically when Storage Autoscaling occurs. If `replicate_source_db` is set, the value is ignored during the creation of the instance.
* `allow_major_version_upgrade` - (Optional) Indicates that major version
upgrades are allowed. Changing this parameter does not result in an outage and
the change is asynchronously applied as soon as possible.
* `apply_immediately` - (Optional) Specifies whether any database modifications
are applied immediately, or during the next maintenance window. Default is
`false`. See [Amazon RDS Documentation for more
information.](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Overview.DBInstance.Modifying.html)
* `auto_minor_version_upgrade` - (Optional) Indicates that minor engine upgrades
will be applied automatically to the DB instance during the maintenance window.
Defaults to true.
* `availability_zone` - (Optional) The AZ for the RDS instance.
* `backup_retention_period` - (Optional) The days to retain backups for.
  Must be between `0` and `35`.
  Default is `0`.
  Must be greater than `0` if the database is used as a source for a [Read Replica][instance-replication],
  uses [low-downtime updates](#low-downtime-updates),
  or will use [RDS Blue/Green deployments][blue-green].
* `backup_target` - (Optional, Forces new resource) Specifies where automated backups and manual snapshots are stored. Possible values are `region` (default) and `outposts`. See [Working with Amazon RDS on AWS Outposts](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/rds-on-outposts.html) for more information.
* `backup_window` - (Optional) The daily time range (in UTC) during which automated backups are created if they are enabled.
  Example: "09:46-10:16". Must not overlap with `maintenance_window`.
* `blue_green_update` - (Optional) Enables low-downtime updates using [RDS Blue/Green deployments][blue-green].
  See [`blue_green_update`](#blue_green_update) below.
* `ca_cert_identifier` - (Optional) The identifier of the CA certificate for the DB instance.
* `character_set_name` - (Optional) The character set name to use for DB encoding in Oracle and Microsoft SQL instances (collation).
  This can't be changed.
  See [Oracle Character Sets Supported in Amazon RDS](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Appendix.OracleCharacterSets.html) or
  [Server-Level Collation for Microsoft SQL Server](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Appendix.SQLServer.CommonDBATasks.Collation.html) for more information.
  Cannot be set  with `replicate_source_db`, `restore_to_point_in_time`, `s3_import`, or `snapshot_identifier`.
* `copy_tags_to_snapshot` – (Optional, boolean) Copy all Instance `tags` to snapshots. Default is `false`.
* `custom_iam_instance_profile` - (Optional) The instance profile associated with the underlying Amazon EC2 instance of an RDS Custom DB instance.
* `db_name` - (Optional) The name of the database to create when the DB instance is created. If this parameter is not specified, no database is created in the DB instance. Note that this does not apply for Oracle or SQL Server engines. See the [AWS documentation](https://awscli.amazonaws.com/v2/documentation/api/latest/reference/rds/create-db-instance.html) for more details on what applies for those engines. If you are providing an Oracle db name, it needs to be in all upper case. Cannot be specified for a replica.
* `db_subnet_group_name` - (Optional) Name of [DB subnet group](/docs/providers/aws/r/db_subnet_group.html). DB instance will
be created in the VPC associated with the DB subnet group. If unspecified, will
be created in the `default` VPC, or in EC2 Classic, if available. When working
with read replicas, it should be specified only if the source database
specifies an instance in another AWS Region. See [DBSubnetGroupName in API
action CreateDBInstanceReadReplica](https://docs.aws.amazon.com/AmazonRDS/latest/APIReference/API_CreateDBInstanceReadReplica.html)
for additional read replica constraints.
* `dedicated_log_volume` - (Optional, boolean) Use a dedicated log volume (DLV) for the DB instance. Requires Provisioned IOPS. See the [AWS documentation](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_PIOPS.StorageTypes.html#USER_PIOPS.dlv) for more details.
* `delete_automated_backups` - (Optional) Specifies whether to remove automated backups immediately after the DB instance is deleted. Default is `true`.
* `deletion_protection` - (Optional) If the DB instance should have deletion protection enabled. The database can't be deleted when this value is set to `true`. The default is `false`.
* `domain` - (Optional) The ID of the Directory Service Active Directory domain to create the instance in. Conflicts with `domain_fqdn`, `domain_ou`, `domain_auth_secret_arn` and a `domain_dns_ips`.
* `domain_auth_secret_arn` - (Optional, but required if domain_fqdn is provided) The ARN for the Secrets Manager secret with the self managed Active Directory credentials for the user joining the domain. Conflicts with `domain` and `domain_iam_role_name`.
* `domain_dns_ips` - (Optional, but required if domain_fqdn is provided)  The IPv4 DNS IP addresses of your primary and secondary self managed Active Directory domain controllers. Two IP addresses must be provided. If there isn't a secondary domain controller, use the IP address of the primary domain controller for both entries in the list. Conflicts with `domain` and `domain_iam_role_name`.
* `domain_fqdn` - (Optional) The fully qualified domain name (FQDN) of the self managed Active Directory domain. Conflicts with `domain` and `domain_iam_role_name`.
* `domain_iam_role_name` - (Optional, but required if domain is provided) The name of the IAM role to be used when making API calls to the Directory Service. Conflicts with `domain_fqdn`, `domain_ou`, `domain_auth_secret_arn` and a `domain_dns_ips`.
* `domain_ou` - (Optional, but required if domain_fqdn is provided) The self managed Active Directory organizational unit for your DB instance to join. Conflicts with `domain` and `domain_iam_role_name`.
* `enabled_cloudwatch_logs_exports` - (Optional) Set of log types to enable for exporting to CloudWatch logs. If omitted, no logs will be exported. For supported values, see the EnableCloudwatchLogsExports.member.N parameter in [API action CreateDBInstance](https://docs.aws.amazon.com/AmazonRDS/latest/APIReference/API_CreateDBInstance.html).
* `engine` - (Required unless a `snapshot_identifier` or `replicate_source_db` is provided) The database engine to use. For supported values, see the Engine parameter in [API action CreateDBInstance](https://docs.aws.amazon.com/AmazonRDS/latest/APIReference/API_CreateDBInstance.html). Note that for Amazon Aurora instances the engine must match the [DB cluster](/docs/providers/aws/r/rds_cluster.html)'s engine'. For information on the difference between the available Aurora MySQL engines see [Comparison between Aurora MySQL 1 and Aurora MySQL 2](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/AuroraMySQL.Updates.20180206.html) in the Amazon RDS User Guide.
* `engine_version` - (Optional) The engine version to use. If `auto_minor_version_upgrade` is enabled, you can provide a prefix of the version such as `8.0` (for `8.0.36`). The actual engine version used is returned in the attribute `engine_version_actual`, see [Attribute Reference](#attribute-reference) below. For supported values, see the EngineVersion parameter in [API action CreateDBInstance](https://docs.aws.amazon.com/AmazonRDS/latest/APIReference/API_CreateDBInstance.html). Note that for Amazon Aurora instances the engine version must match the [DB cluster](/docs/providers/aws/r/rds_cluster.html)'s engine version'.
* `engine_lifecycle_support` - (Optional) The life cycle type for this DB instance. This setting applies only to RDS for MySQL and RDS for PostgreSQL. Valid values are `open-source-rds-extended-support`, `open-source-rds-extended-support-disabled`. Default value is `open-source-rds-extended-support`. [Using Amazon RDS Extended Support]: https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/extended-support.html
* `final_snapshot_identifier` - (Optional) The name of your final DB snapshot
when this DB instance is deleted. Must be provided if `skip_final_snapshot` is
set to `false`. The value must begin with a letter, only contain alphanumeric characters and hyphens, and not end with a hyphen or contain two consecutive hyphens. Must not be provided when deleting a read replica.
* `iam_database_authentication_enabled` - (Optional) Specifies whether mappings of AWS Identity and Access Management (IAM) accounts to database
accounts is enabled.
* `identifier` - (Optional) The name of the RDS instance, if omitted, Terraform will assign a random, unique identifier. Required if `restore_to_point_in_time` is specified.
* `identifier_prefix` - (Optional) Creates a unique identifier beginning with the specified prefix. Conflicts with `identifier`.
* `instance_class` - (Required) The instance type of the RDS instance.
* `iops` - (Optional) The amount of provisioned IOPS. Setting this implies a
storage_type of "io1" or "io2". Can only be set when `storage_type` is `"io1"`, `"io2` or `"gp3"`.
Cannot be specified for gp3 storage if the `allocated_storage` value is below a per-`engine` threshold.
See the [RDS User Guide](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/CHAP_Storage.html#gp3-storage) for details.
* `kms_key_id` - (Optional) The ARN for the KMS encryption key. If creating an
encrypted replica, set this to the destination KMS ARN.
* `license_model` - (Optional, but required for some DB engines, i.e., Oracle SE1) License model information for this DB instance. Valid values for this field are as follows:
    * RDS for MariaDB: `general-public-license`
    * RDS for Microsoft SQL Server: `license-included`
    * RDS for MySQL: `general-public-license`
    * RDS for Oracle: `bring-your-own-license | license-included`
    * RDS for PostgreSQL: `postgresql-license`
* `maintenance_window` - (Optional) The window to perform maintenance in.
Syntax: "ddd:hh24:mi-ddd:hh24:mi". Eg: "Mon:00:00-Mon:03:00". See [RDS
Maintenance Window
docs](http://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_UpgradeDBInstance.Maintenance.html#AdjustingTheMaintenanceWindow)
for more information.
* `manage_master_user_password` - (Optional) Set to true to allow RDS to manage the master user password in Secrets Manager. Cannot be set if `password` is provided.
* `master_user_secret_kms_key_id` - (Optional) The Amazon Web Services KMS key identifier is the key ARN, key ID, alias ARN, or alias name for the KMS key. To use a KMS key in a different Amazon Web Services account, specify the key ARN or alias ARN. If not specified, the default KMS key for your Amazon Web Services account is used.
* `max_allocated_storage` - (Optional) When configured, the upper limit to which Amazon RDS can automatically scale the storage of the DB instance. Configuring this will automatically ignore differences to `allocated_storage`. Must be greater than or equal to `allocated_storage` or `0` to disable Storage Autoscaling.
* `monitoring_interval` - (Optional) The interval, in seconds, between points
when Enhanced Monitoring metrics are collected for the DB instance. To disable
collecting Enhanced Monitoring metrics, specify 0. The default is 0. Valid
Values: 0, 1, 5, 10, 15, 30, 60.
* `monitoring_role_arn` - (Optional) The ARN for the IAM role that permits RDS
to send enhanced monitoring metrics to CloudWatch Logs. You can find more
information on the [AWS
Documentation](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_Monitoring.html)
what IAM permissions are needed to allow Enhanced Monitoring for RDS Instances.
* `multi_az` - (Optional) Specifies if the RDS instance is multi-AZ
* `nchar_character_set_name` - (Optional, Forces new resource) The national character set is used in the NCHAR, NVARCHAR2, and NCLOB data types for Oracle instances. This can't be changed. See [Oracle Character Sets
Supported in Amazon RDS](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Appendix.OracleCharacterSets.html).
* `network_type` - (Optional) The network type of the DB instance. Valid values: `IPV4`, `DUAL`.
* `option_group_name` - (Optional) Name of the DB option group to associate.
* `parameter_group_name` - (Optional) Name of the DB parameter group to associate.
* `password` - (Required unless `manage_master_user_password` is set to true or unless a `snapshot_identifier` or `replicate_source_db`
is provided or `manage_master_user_password` is set.) Password for the master DB user. Note that this may show up in
logs, and it will be stored in the state file. Cannot be set if `manage_master_user_password` is set to `true`.
* `performance_insights_enabled` - (Optional) Specifies whether Performance Insights are enabled. Defaults to false.
* `performance_insights_kms_key_id` - (Optional) The ARN for the KMS key to encrypt Performance Insights data. When specifying `performance_insights_kms_key_id`, `performance_insights_enabled` needs to be set to true. Once KMS key is set, it can never be changed.
* `performance_insights_retention_period` - (Optional) Amount of time in days to retain Performance Insights data. Valid values are `7`, `731` (2 years) or a multiple of `31`. When specifying `performance_insights_retention_period`, `performance_insights_enabled` needs to be set to true. Defaults to '7'.
* `port` - (Optional) The port on which the DB accepts connections.
* `publicly_accessible` - (Optional) Bool to control if instance is publicly
accessible. Default is `false`.
* `replica_mode` - (Optional) Specifies whether the replica is in either `mounted` or `open-read-only` mode. This attribute
is only supported by Oracle instances. Oracle replicas operate in `open-read-only` mode unless otherwise specified. See [Working with Oracle Read Replicas](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/oracle-read-replicas.html) for more information.
* `replicate_source_db` - (Optional) Specifies that this resource is a Replicate
database, and to use this value as the source database. This correlates to the
`identifier` of another Amazon RDS Database to replicate (if replicating within
a single region) or ARN of the Amazon RDS Database to replicate (if replicating
cross-region). Note that if you are
creating a cross-region replica of an encrypted database you will also need to
specify a `kms_key_id`. See [DB Instance Replication][instance-replication] and [Working with
PostgreSQL and MySQL Read Replicas](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_ReadRepl.html)
for more information on using Replication.
* `upgrade_storage_config` - (Optional) Whether to upgrade the storage file system configuration on the read replica. Can only be set with `replicate_source_db`.
* `restore_to_point_in_time` - (Optional, Forces new resource) A configuration block for restoring a DB instance to an arbitrary point in time. Requires the `identifier` argument to be set with the name of the new DB instance to be created. See [Restore To Point In Time](#restore-to-point-in-time) below for details.
* `s3_import` - (Optional) Restore from a Percona Xtrabackup in S3.  See [Importing Data into an Amazon RDS MySQL DB Instance](http://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/MySQL.Procedural.Importing.html)
* `skip_final_snapshot` - (Optional) Determines whether a final DB snapshot is
created before the DB instance is deleted. If true is specified, no DBSnapshot
is created. If false is specified, a DB snapshot is created before the DB
instance is deleted, using the value from `final_snapshot_identifier`. Default
is `false`.
* `snapshot_identifier` - (Optional) Specifies whether or not to create this
database from a snapshot. This correlates to the snapshot ID you'd find in the
RDS console, e.g: rds:production-2015-06-26-06-05.
* `storage_encrypted` - (Optional) Specifies whether the DB instance is
encrypted. Note that if you are creating a cross-region read replica this field
is ignored and you should instead declare `kms_key_id` with a valid ARN. The
default is `false` if not specified.
* `storage_type` - (Optional) One of "standard" (magnetic), "gp2" (general
purpose SSD), "gp3" (general purpose SSD that needs `iops` independently)
"io1" (provisioned IOPS SSD) or "io2" (block express storage provisioned IOPS
SSD). The default is "io1" if `iops` is specified, "gp2" if not.
* `storage_throughput` - (Optional) The storage throughput value for the DB instance. Can only be set when `storage_type` is `"gp3"`. Cannot be specified if the `allocated_storage` value is below a per-`engine` threshold. See the [RDS User Guide](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/CHAP_Storage.html#gp3-storage) for details.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `timezone` - (Optional) Time zone of the DB instance. `timezone` is currently
only supported by Microsoft SQL Server. The `timezone` can only be set on
creation. See [MSSQL User
Guide](http://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/CHAP_SQLServer.html#SQLServer.Concepts.General.TimeZone)
for more information.
* `username` - (Required unless a `snapshot_identifier` or `replicate_source_db`
is provided) Username for the master DB user. Cannot be specified for a replica.
* `vpc_security_group_ids` - (Optional) List of VPC security groups to
associate.
* `customer_owned_ip_enabled` - (Optional) Indicates whether to enable a customer-owned IP address (CoIP) for an RDS on Outposts DB instance. See [CoIP for RDS on Outposts](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/rds-on-outposts.html#rds-on-outposts.coip) for more information.

~> **NOTE:** Removing the `replicate_source_db` attribute from an existing RDS
Replicate database managed by Terraform will promote the database to a fully
standalone database.

### Restore To Point In Time

-> **Note:** You can restore to any point in time before the source DB instance's `latest_restorable_time` or a point up to the number of days specified in the source DB instance's `backup_retention_period`.
For more information, please refer to the [Developer Guide](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_PIT.html).
This setting does not apply to `aurora-mysql` or `aurora-postgresql` DB engines. For Aurora, refer to the [`aws_rds_cluster` resource documentation](/docs/providers/aws/r/rds_cluster.html#restore_in_time).

The `restore_to_point_in_time` block supports the following arguments:

* `restore_time` - (Optional) The date and time to restore from. Value must be a time in Universal Coordinated Time (UTC) format and must be before the latest restorable time for the DB instance. Cannot be specified with `use_latest_restorable_time`.
* `source_db_instance_identifier` - (Optional) The identifier of the source DB instance from which to restore. Must match the identifier of an existing DB instance. Required if `source_db_instance_automated_backups_arn` or `source_dbi_resource_id` is not specified.
* `source_db_instance_automated_backups_arn` - (Optional) The ARN of the automated backup from which to restore. Required if `source_db_instance_identifier` or `source_dbi_resource_id` is not specified.
* `source_dbi_resource_id` - (Optional) The resource ID of the source DB instance from which to restore. Required if `source_db_instance_identifier` or `source_db_instance_automated_backups_arn` is not specified.
* `use_latest_restorable_time` - (Optional) A boolean value that indicates whether the DB instance is restored from the latest backup time. Defaults to `false`. Cannot be specified with `restore_time`.

### S3 Import Options

Full details on the core parameters and impacts are in the API Docs: [RestoreDBInstanceFromS3](http://docs.aws.amazon.com/AmazonRDS/latest/APIReference/API_RestoreDBInstanceFromS3.html).  Sample

```terraform
resource "aws_db_instance" "db" {
  s3_import {
    source_engine         = "mysql"
    source_engine_version = "5.6"
    bucket_name           = "mybucket"
    bucket_prefix         = "backups"
    ingestion_role        = "arn:aws:iam::1234567890:role/role-xtrabackup-rds-restore"
  }
}
```

* `bucket_name` - (Required) The bucket name where your backup is stored
* `bucket_prefix` - (Optional) Can be blank, but is the path to your backup
* `ingestion_role` - (Required) Role applied to load the data.
* `source_engine` - (Required, as of Feb 2018 only 'mysql' supported) Source engine for the backup
* `source_engine_version` - (Required, as of Feb 2018 only '5.6' supported) Version of the source engine used to make the backup

This will not recreate the resource if the S3 object changes in some way.  It's only used to initialize the database.

### `blue_green_update`

* `enabled` - (Optional) Enables [low-downtime updates](#low-downtime-updates) when `true`.
  Default is `false`.

[instance-replication]:
https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Overview.Replication.html
[instance-maintenance]:
https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_UpgradeDBInstance.Maintenance.html
[blue-green]:
https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/blue-green-deployments.html

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `address` - The hostname of the RDS instance. See also `endpoint` and `port`.
* `arn` - The ARN of the RDS instance.
* `allocated_storage` - The amount of allocated storage.
* `availability_zone` - The availability zone of the instance.
* `backup_retention_period` - The backup retention period.
* `backup_window` - The backup window.
* `ca_cert_identifier` - Identifier of the CA certificate for the
DB instance.
* `db_name` - The database name.
* `domain` - The ID of the Directory Service Active Directory domain the instance is joined to
* `domain_auth_secret_arn` - The ARN for the Secrets Manager secret with the self managed Active Directory credentials for the user joining the domain.
* `domain_dns_ips` - The IPv4 DNS IP addresses of your primary and secondary self managed Active Directory domain controllers.
* `domain_fqdn` - The fully qualified domain name (FQDN) of an self managed Active Directory domain.
* `domain_iam_role_name` - The name of the IAM role to be used when making API calls to the Directory Service.
* `domain_ou` - The self managed Active Directory organizational unit for your DB instance to join.
* `endpoint` - The connection endpoint in `address:port` format.
* `engine` - The database engine.
* `engine_version_actual` - The running version of the database.
* `hosted_zone_id` - The canonical hosted zone ID of the DB instance (to be used
in a Route 53 Alias record).
* `id` - RDS DBI resource ID.
* `instance_class`- The RDS instance class.
* `latest_restorable_time` - The latest time, in UTC [RFC3339 format](https://tools.ietf.org/html/rfc3339#section-5.8), to which a database can be restored with point-in-time restore.
* `listener_endpoint` - Specifies the listener connection endpoint for SQL Server Always On. See [endpoint](#endpoint) below.
* `maintenance_window` - The instance maintenance window.
* `master_user_secret` - A block that specifies the master user secret. Only available when `manage_master_user_password` is set to true. [Documented below](#master_user_secret).
* `multi_az` - If the RDS instance is multi AZ enabled.
* `port` - The database port.
* `resource_id` - The RDS Resource ID of this instance.
* `status` - The RDS instance status.
* `storage_encrypted` - Whether the DB instance is encrypted.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `username` - The master username for the database.

On Oracle and Microsoft SQL instances the following is exported additionally:

* `character_set_name` - The character set (collation) used on Oracle and Microsoft SQL instances.

### Endpoint

* `address` - Specifies the DNS address of the DB instance.
* `hosted_zone_id` - Specifies the ID that Amazon Route 53 assigns when you create a hosted zone.
* `port` - Specifies the port that the database engine is listening on.

### master_user_secret

The `master_user_secret` configuration block supports the following attributes:

* `kms_key_id` - The Amazon Web Services KMS key identifier that is used to encrypt the secret.
* `secret_arn` - The Amazon Resource Name (ARN) of the secret.
* `secret_status` - The status of the secret. Valid Values: `creating` | `active` | `rotating` | `impaired`.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `40m`)
- `update` - (Default `80m`)
- `delete` - (Default `60m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import DB Instances using the `identifier`. For example:

```terraform
import {
  to = aws_db_instance.default
  id = "mydb-rds-instance"
}
```

Using `terraform import`, import DB Instances using the `identifier`. For example:

```console
% terraform import aws_db_instance.default mydb-rds-instance
```
