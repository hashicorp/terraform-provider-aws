---
subcategory: "Oracle Database@AWS"
layout: "AWS: aws_odb_autonomous_database"
page_title: "AWS: aws_odb_autonomous_database"
description: |-
  Terraform data source for reading an Oracle Database@AWS Autonomous Database Serverless instance.
---

# Data Source: aws_odb_autonomous_database

Reads an Oracle Database@AWS Autonomous Database Serverless (ADB-S) instance by its unique identifier.

## Example Usage

```terraform
data "aws_odb_autonomous_database" "example" {
  id = aws_odb_autonomous_database.example.id
}
```

## Argument Reference

The following argument is required:

* `id` - (Required) Unique Autonomous Database identifier.

The following argument is optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - Amazon Resource Name (ARN) of the Autonomous Database.
* `actual_used_data_storage_size_in_tbs` - Actual data storage currently in use, in TB.
* `allocated_storage_size_in_tbs` - Storage currently allocated, in TB.
* `allowlisted_ips` - IP addresses allowed to access the database.
* `auto_refresh_frequency_in_seconds` - Automatic refresh frequency in seconds.
* `auto_refresh_point_lag_in_seconds` - Refresh lag from the source in seconds.
* `autonomous_maintenance_schedule_type` - Maintenance schedule type.
* `availability_zone` - Availability Zone of the database.
* `availability_zone_id` - Availability Zone ID of the database.
* `available_upgrade_versions` - Oracle Database versions available for upgrade.
* `backup_retention_period_in_days` - Automatic backup retention period in days.
* `byol_compute_count_limit` - Maximum BYOL compute capacity.
* `character_set` - Database character set.
* `compute_count` - Compute capacity in ECPUs or OCPUs.
* `compute_model` - Compute model, either `ECPU` or `OCPU`.
* `cpu_core_count` - Allocated CPU core count.
* `created_at` - Creation date and time.
* `customer_contacts_to_send_to_oci` - Customer contacts that receive operational notifications from OCI.
* `data_storage_size_in_gbs` - Data volume size in GB.
* `data_storage_size_in_tbs` - Data volume size in TB.
* `database_edition` - Oracle Database edition.
* `database_type` - Autonomous Database type.
* `db_name` - Database name.
* `db_tools_details` - Database management tools enabled for the database.
* `db_version` - Oracle Database software version.
* `db_workload` - Database workload.
* `display_name` - User-friendly database name.
* `encryption_key_provider` - Encryption key provider.
* `is_auto_scaling_enabled` - Whether automatic compute scaling is enabled.
* `is_auto_scaling_for_storage_enabled` - Whether automatic storage scaling is enabled.
* `is_backup_retention_locked` - Whether the backup retention period is locked.
* `is_local_data_guard_enabled` - Whether local Oracle Data Guard is enabled.
* `is_mtls_connection_required` - Whether mutual TLS is required.
* `is_refreshable_clone` - Whether the database is a refreshable clone.
* `kms_key_id` - ARN of the AWS KMS encryption key, when configured.
* `license_model` - Oracle license model.
* `local_adg_auto_failover_max_data_loss_limit` - Maximum automatic-failover data-loss limit in seconds.
* `long_term_backup_schedule` - Long-term backup schedule.
* `ncharacter_set` - National character set.
* `oci_resource_anchor_name` - OCI resource anchor name.
* `oci_url` - OCI console URL.
* `ocid` - Oracle Cloud Identifier.
* `odb_network_arn` - ARN of the associated ODB network.
* `odb_network_id` - ID of the associated ODB network.
* `open_mode` - Database open mode.
* `percent_progress` - Progress of the current operation.
* `permission_level` - Database permission level.
* `private_endpoint` - Private endpoint hostname.
* `private_endpoint_ip` - Private endpoint IP address.
* `private_endpoint_label` - Private endpoint label.
* `refreshable_mode` - Refresh mode of a refreshable clone.
* `resource_pool_leader_id` - Resource-pool leader database ID.
* `resource_pool_summary` - Resource pool configuration.
* `scheduled_operations` - Scheduled database start and stop times.
* `service_console_url` - Oracle service console URL.
* `source_id` - ID of the source database or backup.
* `sql_web_developer_url` - Oracle SQL Developer Web URL.
* `standby_allowlisted_ips` - IP addresses allowed to access the standby database.
* `standby_allowlisted_ips_source` - Source of standby allowlisted IPs.
* `status` - Current database lifecycle status.
* `status_reason` - Additional lifecycle status information.
* `tags` - Map of tags assigned to the Autonomous Database.
* `time_of_auto_refresh_start` - Automatic refresh start date and time.

### `customer_contacts_to_send_to_oci`

* `email` - Customer contact email address.

### `db_tools_details`

* `compute_count` - Compute capacity allocated to the database tool.
* `is_enabled` - Whether the database tool is enabled.
* `max_idle_time_in_minutes` - Maximum idle time before the tool is shut down.
* `name` - Database tool name.

### `long_term_backup_schedule`

* `is_disabled` - Whether the schedule is disabled.
* `repeat_cadence` - Backup cadence.
* `retention_period_in_days` - Backup retention period in days.
* `time_of_backup` - Backup date and time.

### `resource_pool_summary`

* `available_compute_capacity` - Available compute capacity.
* `available_storage_capacity_in_tbs` - Available storage capacity in TB.
* `is_disabled` - Whether the resource pool is disabled.
* `pool_size` - Number of databases the pool can contain.
* `pool_storage_size_in_tbs` - Pool storage size in TB.
* `total_compute_capacity` - Total compute capacity.

### `scheduled_operations`

* `day_of_week` - Day of the week.
* `scheduled_start_time` - Scheduled start time in UTC.
* `scheduled_stop_time` - Scheduled stop time in UTC.
