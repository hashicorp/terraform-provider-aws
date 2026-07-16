---
subcategory: "Oracle Database@AWS"
layout: "AWS: aws_odb_autonomous_database"
page_title: "AWS: aws_odb_autonomous_database"
description: |-
  Terraform resource for managing an Oracle Database@AWS Autonomous Database Serverless instance.
---

# Resource: aws_odb_autonomous_database

Manages an Oracle Database@AWS Autonomous Database Serverless (ADB-S) instance. ADB-S uses shared Exadata infrastructure managed by Oracle and requires an existing ODB network; it does not require a customer-managed Exadata infrastructure or VM cluster.

Provisioning, updating, and deleting an Autonomous Database are asynchronous and can take several hours. The AWS account and Region must be onboarded for Oracle Database@AWS and have sufficient service quotas.

## Example Usage

### Basic Usage

```terraform
variable "autonomous_database_admin_password" {
  type      = string
  sensitive = true
}

resource "aws_odb_autonomous_database" "example" {
  admin_password_wo         = var.autonomous_database_admin_password
  admin_password_wo_version = 1
  compute_count             = 2
  data_storage_size_in_tbs  = 1
  db_name                   = "TFADBEXAMPLE"
  db_workload               = "OLTP"
  display_name              = "terraform-adbs-example"
  license_model             = "LICENSE_INCLUDED"
  odb_network_id            = aws_odb_network.example.id
  source                    = "NONE"

  tags = {
    Environment = "example"
  }
}
```

### Reading the Created Database

```terraform
data "aws_odb_autonomous_database" "example" {
  id = aws_odb_autonomous_database.example.id
}
```

## Argument Reference

The AWS API defines all create parameters as optional because the valid combination depends on `source`. For a new database (`source = "NONE"`), configure the ODB network, database identity, workload, compute, storage, license, and ADMIN password values required by your Oracle Database@AWS tenancy.

The following arguments are optional:

* `admin_password_wo` - (Optional, Sensitive, Write-only) Password for the `ADMIN` user. Must be between 12 and 30 characters. The value is sent to AWS but is never stored in Terraform plan or state. Set `admin_password_wo_version` with this argument.
* `admin_password_wo_version` - (Optional) Arbitrary integer stored in state. Change this value together with `admin_password_wo` to rotate the ADMIN password.
* `allowlisted_ips` - (Optional) List of between 1 and 1024 IP addresses allowed to access the database.
* `auto_refresh_frequency_in_seconds` - (Optional) Automatic refresh frequency, in seconds, for a refreshable clone.
* `auto_refresh_point_lag_in_seconds` - (Optional) Refresh lag, in seconds, between a refreshable clone and its source.
* `autonomous_maintenance_schedule_type` - (Optional) Maintenance schedule type. Valid values are `EARLY` and `REGULAR`.
* `backup_retention_period_in_days` - (Optional) Automatic backup retention period, in days.
* `byol_compute_count_limit` - (Optional) Maximum compute capacity under the bring-your-own-license model. Minimum value is `2`.
* `character_set` - (Optional) Database character set. Changing this value creates a new resource.
* `compute_count` - (Optional) Compute capacity in ECPUs or OCPUs. Valid values are from `0.1` through `512`.
* `cpu_core_count` - (Optional) Allocated CPU core count. Valid values are from `1` through `128`.
* `data_storage_size_in_gbs` - (Optional) Data volume size in GB. Valid values are from `20` through `393216`.
* `data_storage_size_in_tbs` - (Optional) Data volume size in whole TB. Valid values are from `1` through `384`.
* `database_edition` - (Optional) Oracle Database edition. Valid values are `STANDARD_EDITION` and `ENTERPRISE_EDITION`.
* `db_name` - (Optional) Database name. Must begin with a letter, contain only alphanumeric characters, and contain at most 30 characters.
* `db_version` - (Optional) Oracle Database software version.
* `db_workload` - (Optional) Database workload. Valid values are `OLTP`, `AJD`, `APEX`, and `LH`.
* `display_name` - (Optional) User-friendly database name.
* `encryption_key_provider` - (Optional) Encryption key provider. Valid configurable values are `ORACLE_MANAGED` and `AWS_KMS`.
* `is_auto_scaling_enabled` - (Optional) Whether automatic compute scaling is enabled.
* `is_auto_scaling_for_storage_enabled` - (Optional) Whether automatic storage scaling is enabled.
* `is_backup_retention_locked` - (Optional) Whether the backup retention period is locked against shortening.
* `is_local_data_guard_enabled` - (Optional) Whether local Oracle Data Guard is enabled.
* `is_mtls_connection_required` - (Optional) Whether mutual TLS is required for database connections.
* `is_refreshable_clone` - (Optional) Whether the database is a refreshable clone.
* `kms_key_id` - (Optional) ARN of the AWS KMS key used when `encryption_key_provider` is `AWS_KMS`.
* `license_model` - (Optional) Oracle license model. Valid values are `BRING_YOUR_OWN_LICENSE` and `LICENSE_INCLUDED`.
* `local_adg_auto_failover_max_data_loss_limit` - (Optional) Maximum data-loss limit, in seconds, for automatic local Data Guard failover.
* `ncharacter_set` - (Optional) National character set. Changing this value creates a new resource.
* `odb_network_id` - (Optional) ID or ARN of the ODB network used by the database. Changing this value creates a new resource.
* `open_mode` - (Optional) Database open mode.
* `permission_level` - (Optional) Database permission level.
* `private_endpoint_ip` - (Optional) Private endpoint IP address.
* `private_endpoint_label` - (Optional) Private endpoint label.
* `refreshable_mode` - (Optional) Refresh mode for a refreshable clone.
* `resource_pool_leader_id` - (Optional) ID or ARN of the resource-pool leader Autonomous Database.
* `source` - (Optional) Source from which to create the database. Valid values are `NONE`, `DATABASE`, `BACKUP_FROM_ID`, `BACKUP_FROM_TIMESTAMP`, `CROSS_REGION_DATAGUARD`, `CROSS_REGION_DISASTER_RECOVERY`, and `CLONE_TO_REFRESHABLE`. Changing this value creates a new resource.
* `standby_allowlisted_ips` - (Optional) List of between 1 and 1024 IP addresses allowed to access the standby database.
* `standby_allowlisted_ips_source` - (Optional) Source of the standby allowlist. Valid values are `PRIMARY`, `SEPARATE`, and `NOT_APPLICABLE`.
* `time_of_auto_refresh_start` - (Optional) RFC3339 timestamp at which automatic refresh begins.
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block), tags with matching keys overwrite those defined at the provider level.

The following nested blocks are supported:

### `customer_contacts_to_send_to_oci`

* `email` - (Required) Email address that receives operational notifications from OCI.

### `db_tools_details`

* `compute_count` - (Optional) Compute capacity allocated to the database tool.
* `is_enabled` - (Optional) Whether the database tool is enabled.
* `max_idle_time_in_minutes` - (Optional) Maximum idle time before the tool is shut down.
* `name` - (Optional) Database tool name.

### `long_term_backup_schedule`

At most one block can be configured.

* `is_disabled` - (Optional) Whether the long-term backup schedule is disabled.
* `repeat_cadence` - (Optional) Backup cadence. Valid values are `ONE_TIME`, `WEEKLY`, `MONTHLY`, and `YEARLY`.
* `retention_period_in_days` - (Optional) Backup retention period. Valid values are from `90` through `3650`.
* `time_of_backup` - (Optional) RFC3339 timestamp at which the backup is taken.

### `resource_pool_summary`

At most one block can be configured.

* `is_disabled` - (Optional) Whether the resource pool is disabled.
* `pool_size` - (Optional) Number of Autonomous Databases the pool can contain.
* `pool_storage_size_in_tbs` - (Optional) Pool storage size in TB.

The `available_compute_capacity`, `available_storage_capacity_in_tbs`, and `total_compute_capacity` fields are computed.

### `scheduled_operations`

* `day_of_week` - (Required) Day of the week.
* `scheduled_start_time` - (Optional) Scheduled start time in UTC.
* `scheduled_stop_time` - (Optional) Scheduled stop time in UTC.

### `source_configuration`

At most one `source_configuration` block can be configured. The block must contain exactly one of the source-specific blocks below. Changing any source configuration creates a new resource.

#### `clone_to_refreshable`

* `source_autonomous_database_id` - (Required) ID of the source Autonomous Database.
* `auto_refresh_frequency_in_seconds` - (Optional) Automatic refresh frequency in seconds.
* `auto_refresh_point_lag_in_seconds` - (Optional) Refresh lag in seconds.
* `clone_type` - (Optional) Clone type.
* `open_mode` - (Optional) Clone open mode.
* `refreshable_mode` - (Optional) Refresh mode.
* `time_of_auto_refresh_start` - (Optional) RFC3339 automatic refresh start timestamp.

#### `cross_region_data_guard`

* `source_autonomous_database_arn` - (Required) ARN of the source Autonomous Database.

#### `cross_region_disaster_recovery`

* `remote_disaster_recovery_type` - (Required) Remote disaster recovery type.
* `source_autonomous_database_arn` - (Required) ARN of the source Autonomous Database.
* `is_replicate_automatic_backups` - (Optional) Whether automatic backups are replicated.

#### `database_clone`

* `clone_type` - (Required) Clone type.
* `source_autonomous_database_id` - (Required) ID of the source Autonomous Database.

#### `point_in_time_restore`

* `clone_type` - (Required) Clone type.
* `source_autonomous_database_id` - (Required) ID of the source Autonomous Database.
* `clone_table_space_list` - (Optional) List of tablespace IDs to clone.
* `timestamp` - (Optional) RFC3339 timestamp to which the database is restored.
* `use_latest_available_backup_timestamp` - (Optional) Whether to use the latest available backup timestamp.

#### `restore_from_backup`

* `autonomous_database_backup_id` - (Required) ID of the Autonomous Database backup.
* `clone_type` - (Required) Clone type.
* `clone_table_space_list` - (Optional) List of tablespace IDs to clone.

### `transportable_tablespace`

At most one block can be configured. Changing this block creates a new resource.

* `tts_bundle_url` - (Optional) URL of the transportable tablespace bundle.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Unique Autonomous Database identifier. This is the Terraform resource ID.
* `arn` - Amazon Resource Name (ARN) of the Autonomous Database.
* `actual_used_data_storage_size_in_tbs` - Actual data storage currently in use, in TB.
* `allocated_storage_size_in_tbs` - Storage currently allocated, in TB.
* `availability_zone` - Availability Zone of the database.
* `availability_zone_id` - Availability Zone ID of the database.
* `available_upgrade_versions` - Oracle Database versions available for upgrade.
* `compute_model` - Compute model, either `ECPU` or `OCPU`.
* `created_at` - Creation date and time.
* `database_type` - Autonomous Database type.
* `oci_resource_anchor_name` - OCI resource anchor name.
* `oci_url` - OCI console URL.
* `ocid` - Oracle Cloud Identifier.
* `odb_network_arn` - ARN of the associated ODB network.
* `percent_progress` - Progress of the current operation.
* `private_endpoint` - Private endpoint hostname.
* `service_console_url` - Oracle service console URL.
* `source_id` - ID of the source database or backup.
* `sql_web_developer_url` - Oracle SQL Developer Web URL.
* `status` - Current database lifecycle status.
* `status_reason` - Additional lifecycle status information.
* `tags_all` - Map of tags assigned to the resource, including tags inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `24h`)
* `update` - (Default `24h`)
* `delete` - (Default `24h`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) with the Autonomous Database ID. For example:

```terraform
import {
  to = aws_odb_autonomous_database.example
  id = "adb-example123"
}
```

Using `terraform import`, import an Autonomous Database using its ID. For example:

```console
% terraform import aws_odb_autonomous_database.example adb-example123
```

The ADMIN password and creation-only source configuration are not returned by AWS. After import, configure `admin_password_wo` only when rotating the password.
