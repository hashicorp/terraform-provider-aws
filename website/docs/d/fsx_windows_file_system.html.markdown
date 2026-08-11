---
subcategory: "FSx"
layout: "aws"
page_title: "AWS: aws_fsx_windows_file_system"
description: |-
  Retrieve information on FSx Windows File System.
---

# Data Source: aws_fsx_windows_file_system

Retrieve information on FSx Windows File System.

## Example Usage

### Root volume Example

```terraform
data "aws_fsx_windows_file_system" "example" {
  id = "fs-12345678"
}
```

## Argument Reference

This data source supports the following arguments:

* `id` - (Required) Identifier of the file system (e.g. `fs-12345678`).
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `active_directory_id` - ID for Microsoft Active Directory instance that the file system is joined to.
* `aliases` - Set of DNS alias names associated with the Amazon FSx file system.
* `arn` - Amazon Resource Name of the file system.
* `audit_log_configuration` - Configuration that Amazon FSx for Windows File Server uses to audit and log user accesses of files, folders, and file shares on the Amazon FSx for Windows File Server file system.
* `automatic_backup_retention_days` - Number of days to retain automatic backups.
* `backup_id` - Identifier of the source backup used to create the file system.
* `copy_tags_to_backups` - Whether tags on the file system are copied to backups.
* `daily_automatic_backup_start_time` - Preferred time (in `HH:MM` format) to take daily automatic backups, in the UTC time zone.
* `deployment_type` - File system deployment type.
* `disk_iops_configuration` - SSD IOPS configuration for the file system.
* `dns_name` - DNS name for the file system (e.g. `fs-12345678.corp.example.com`).
* `id` - Identifier of the file system (e.g. `fs-12345678`).
* `kms_key_id` - ARN for the KMS Key to encrypt the file system at rest.
* `network_interface_ids` - Set of network interface identifiers for the file system.
* `owner_id` - AWS account identifier that created the file system.
* `preferred_file_server_ip` - IP address of the primary, or preferred, file server.
* `preferred_subnet_id` - Subnet in which the preferred file server is located.
* `security_group_ids` - Set of security group identifiers associated with the file system.
* `skip_final_backup` - Whether a final backup is skipped when the file system is deleted.
* `storage_capacity` - Storage capacity of the file system in gibibytes (GiB).
* `storage_type` - Type of storage the file system is using. If set to `SSD`, the file system uses solid state drive storage. If set to `HDD`, the file system uses hard disk drive storage.
* `subnet_ids` - IDs of the subnets that the file system is accessible from.
* `tags` - Tags to associate with the file system.
* `throughput_capacity` - Throughput (megabytes per second) of the file system in power of 2 increments. Minimum of `8` and maximum of `2048`.
* `vpc_id` - ID of the primary virtual private cloud (VPC) for the file system.
* `weekly_maintenance_start_time` - Preferred start time (in `d:HH:MM` format) to perform weekly maintenance, in the UTC time zone.
