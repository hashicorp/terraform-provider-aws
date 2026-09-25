---
subcategory: "FSx"
layout: "aws"
page_title: "AWS: aws_fsx_lustre_file_system"
description: |-
  Provides details about an FSx for Lustre File System.
---

# Data Source: aws_fsx_lustre_file_system

Provides details about an FSx for Lustre File System.

## Example Usage

### Basic Usage

```terraform
data "aws_fsx_lustre_file_system" "example" {
  id = "fs-12345678"
}
```

## Argument Reference

This data source supports the following arguments:

* `id` - (Required) Identifier of the file system (e.g. `fs-12345678`).
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - Amazon Resource Name of the file system.
* `auto_import_policy` - How Amazon FSx keeps your file and directory listings up to date as you add or modify objects in your linked S3 bucket.
* `automatic_backup_retention_days` - Number of days to retain automatic backups.
* `copy_tags_to_backups` - Whether tags on the file system are copied to backups.
* `daily_automatic_backup_start_time` - Preferred time (in `HH:MM` format) to take daily automatic backups, in the UTC time zone.
* `data_compression_type` - Data compression configuration for the file system.
* `data_read_cache_configuration` - Data read cache configuration for the file system. See [Data Read Cache Configuration](#data-read-cache-configuration) below.
* `deployment_type` - File system deployment type.
* `dns_name` - DNS name for the file system.
* `drive_cache_type` - Type of drive cache used by `PERSISTENT_1` file systems that are provisioned with HDD storage devices.
* `efa_enabled` - Whether Elastic Fabric Adapter (EFA) is enabled for the file system.
* `export_path` - S3 URI (with optional prefix) where the root of the file system is exported.
* `file_system_type_version` - Lustre version for the file system.
* `imported_file_chunk_size` - Chunk size (in MiB) for data imported from the linked S3 bucket.
* `kms_key_id` - ARN for the KMS Key to encrypt the file system at rest.
* `log_configuration` - Lustre logging configuration. See [Log Configuration](#log-configuration) below.
* `metadata_configuration` - Lustre metadata configuration. See [Metadata Configuration](#metadata-configuration) below.
* `mount_name` - Value to use when mounting the file system.
* `network_interface_ids` - IDs of the elastic network interfaces from which a specific file system is accessible.
* `owner_id` - AWS account identifier that created the file system.
* `per_unit_storage_throughput` - Amount of read and write throughput for each 1 tebibyte of storage, in MB/s/TiB.
* `root_squash_configuration` - Lustre root squash configuration. See [Root Squash Configuration](#root-squash-configuration) below.
* `storage_capacity` - Storage capacity of the file system in gibibytes (GiB).
* `storage_type` - Type of storage the file system is using. If set to `SSD`, the file system uses solid state drive storage. If set to `HDD`, the file system uses hard disk drive storage.
* `subnet_ids` - IDs of the subnets that the file system is accessible from.
* `tags` - Tags associated with the file system.
* `throughput_capacity` - Sustained throughput of the file system in Megabytes per second (MBps).
* `vpc_id` - ID of the primary virtual private cloud (VPC) for the file system.
* `weekly_maintenance_start_time` - Preferred start time (in `D:HH:MM` format) to perform weekly maintenance, in the UTC time zone.

### Data Read Cache Configuration

* `size` - Size of the data read cache in gibibytes (GiB).
* `sizing_mode` - How the cache size is determined. Valid values are `NO_CACHE`, `USER_PROVISIONED`, and `PROPORTIONAL_TO_THROUGHPUT_CAPACITY`.

### Log Configuration

* `destination` - Amazon Resource Name (ARN) that specifies the destination of the logs.
* `level` - Data repository association audit logging level. Valid values are `DISABLED`, `WARN_ONLY`, `ERROR_ONLY`, and `WARN_ERROR`.

### Metadata Configuration

* `iops` - Number of metadata IOPS provisioned for the file system.
* `mode` - Metadata configuration mode. Valid values are `AUTOMATIC` and `USER_PROVISIONED`.

### Root Squash Configuration

* `no_squash_nids` - Set of NIDs that are not subject to root squash.
* `root_squash` - User ID and group ID values (in the format `UID:GID`) for mapping the root user to when root squash is enabled.
