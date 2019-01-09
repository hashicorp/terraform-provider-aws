---
layout: "aws"
page_title: "AWS: aws_fsx_file_system"
sidebar_current: "docs-aws-resource-fsx-file-system"
description: |-
  Provides an FSx filesystem resource.
---

# aws_fsx_file_system

Provides an FSx filesystem resource.

## Example Usage

```hcl
resource "aws_fsx_file_system" "lustre" {
  type 				= "LUSTRE"
  capacity 		= 3600
	kms_key_id 	= "${aws_kms_key.example.key_id}"
	subnet_ids 	= ["${aws_subnet.example.id}"]

	lustre_configuration {
		import_path = "s3://example-bucket"
		chunk_size 	= 2048
	}
}
```

```hcl
resource "aws_fsx_file_system" "windows" {
  type 				= "WINDOWS"
  capacity 		= 300
	kms_key_id 	= "${aws_kms_key.example.arn}"
	subnet_ids 	= ["${aws_subnet.example.id}"]

	windows_configuration {
		active_directory_id		= "${aws_directory_service_directory.example.id}"
		backup_retention 			= 7
		copy_tags_to_backups 	= true
		throughput_capacity 	= 1024
	}
}
```

## Argument Reference

The following arguments are supported:

* `type` - (Required) The type of file system.  `LUSTRE`
* `capacity` - (Required) The storage capacity of the file system. For Windows file systems, the storage capacity has a minimum of 300 GiB, and a maximum of 65,536 GiB.  For Lustre file systems, the storage capacity has a minimum of 3,600 GiB. Storage capacity is provisioned in increments of 3,600 GiB.
* `kms_key_id` - (Required) The ARN for the KMS encryption key.
* `subnet_ids` - (Required) A list of IDs for the subnets that the file system will be accessible from. File systems support only one subnet. The file server is also launched in that subnet's Availability Zone.
* `security_group_ids` - (Optional) A list of IDs for the security groups that apply to the specified network interfaces created for file system access. These security groups will apply to all network interfaces. 
* `lustre_configuration` - (Optional) The configuration for this Lustre file system.
* `windows_configuration` - (Optional) The configuration for this Microsoft Windows file system.
* `tags` - (Optional) A mapping of tags to assign to the file system.

## lustre_configuration

Attributes for Lustre configuration

* `import_path` - (Required) The path to the Amazon S3 bucket (and optional prefix) that you're using as the data repository for your FSx for Lustre file system.
* `chunk_size` - (Optional) For files imported from a data repository, this value determines the stripe count and maximum amount of data per file (in MiB) stored on a single physical disk. 
* `weekly_maintenance_start_time` - (Optional) The preferred time to perform weekly maintenance, in the UTC time zone.

## windows_configuration

Attributes for Windows configuration

* `active_directory_id` - (Required) The ID for an existing Microsoft Active Directory instance that the file system should join when it's created.
* `backup_retention` - (Optional) The number of days to retain automatic backups. 
* `copy_tags_to_backups` - (Optional) A boolean flag indicating whether tags on the file system should be copied to backups.
* `daily_backup_start_time` - (Optional) The preferred time to take daily automatic backups, in the UTC time zone.
* `throughput_capacity` - (Required) The throughput of an Amazon FSx file system, measured in megabytes per second.
* `weekly_maintenance_start_time` - (Optional) The preferred start time to perform weekly maintenance, in the UTC time zone.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID that identifies the file system.
* `arn` - Amazon Resource Name of the file system.
* `dns_name` - The DNS name for the filesystem per [documented convention](https://docs.aws.amazon.com/fsx/index.html#lang/en_us).

## Import

The FSx file systems can be imported using the `id`, e.g.

```
$ terraform import aws_fsx_file_system.example fs-543ab12b1ca672f33
```
