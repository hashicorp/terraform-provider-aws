---
subcategory: "FSx"
layout: "aws"
page_title: "AWS: aws_fsx_ontap_file_system"
description: |-
  Retrieve information on FSx ONTAP File System.
---

# Data Source: aws_fsx_ontap_file_system

Retrieve information on FSx ONTAP File System.

## Example Usage

### Basic Usage

```terraform
data "aws_fsx_ontap_file_system" "example" {
  id = "fs-12345678"
}
```

## Argument Reference

The following arguments are required:

* `id` - (Required) Identifier of the file system (e.g. `fs-12345678`).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name of the file system.
* `automatic_backup_retention_days` - The number of days to retain automatic backups.
* `daily_automatic_backup_start_time` - The preferred time (in `HH:MM` format) to take daily automatic backups, in the UTC time zone.
* `deployment_type` - The file system deployment type.
* `disk_iops_configuration` - The SSD IOPS configuration for the Amazon FSx for NetApp ONTAP file system, specifying the number of provisioned IOPS and the provision mode. See [Disk IOPS](#disk-iops) Below.
* `dns_name` - DNS name for the file system.

  **Note:** This attribute does not apply to FSx for ONTAP file systems and is consequently not set. You can access your FSx for ONTAP file system and volumes via a [Storage Virtual Machine (SVM)](fsx_ontap_storage_virtual_machine.html) using its DNS name or IP address.
* `endpoint_ip_address_range` - (Multi-AZ only) Specifies the IP address range in which the endpoints to access your file system exist.
* `endpoints` - The Management and Intercluster FileSystemEndpoints that are used to access data or to manage the file system using the NetApp ONTAP CLI, REST API, or NetApp SnapMirror. See [FileSystemEndpoints](#file-system-endpoints) below.
* `ha_pairs` - The number of HA pairs for the file system.
* `id` - Identifier of the file system (e.g. `fs-12345678`).
* `kms_key_id` - ARN for the KMS Key to encrypt the file system at rest.
* `network_interface_ids` - The IDs of the elastic network interfaces from which a specific file system is accessible.
* `owner_id` - AWS account identifier that created the file system.
* `preferred_subnet_id` - Specifies the subnet in which you want the preferred file server to be located.
* `route_table_ids` - (Multi-AZ only) The VPC route tables in which your file system's endpoints exist.
* `storage_capacity` - The storage capacity of the file system in gibibytes (GiB).
* `storage_type` - The type of storage the file system is using. If set to `SSD`, the file system uses solid state drive storage. If set to `HDD`, the file system uses hard disk drive storage.
* `subnet_ids` - Specifies the IDs of the subnets that the file system is accessible from. For the MULTI_AZ_1 file system deployment type, there are two subnet IDs, one for the preferred file server and one for the standby file server. The preferred file server subnet identified in the `preferred_subnet_id` property.
* `tags` - The tags associated with the file system.
* `throughput_capacity` - The sustained throughput of an Amazon FSx file system in Megabytes per second (MBps). If the file system uses multiple HA pairs this will equal throuthput_capacity_per_ha_pair x ha_pairs
* `throughput_capacity_per_ha_pair` - The sustained throughput of each HA pair for an Amazon FSx file system in Megabytes per second (MBps).
* `vpc_id` - The ID of the primary virtual private cloud (VPC) for the file system.
* `weekly_maintenance_start_time` - The preferred start time (in `D:HH:MM` format) to perform weekly maintenance, in the UTC time zone.

### Disk IOPS

* `iops` - The total number of SSD IOPS provisioned for the file system.
* `mode` - Specifies whether the file system is using the `AUTOMATIC` setting of SSD IOPS of 3 IOPS per GB of storage capacity, or if it using a `USER_PROVISIONED` value.

### File System Endpoints

* `intercluster` - A FileSystemEndpoint for managing your file system by setting up NetApp SnapMirror with other ONTAP systems. See [FileSystemEndpoint](#file-system-endpoint) below.
* `management` - A FileSystemEndpoint for managing your file system using the NetApp ONTAP CLI and NetApp ONTAP API. See [FileSystemEndpoint](#file-system-endpoint) below.

### File System Endpoint

* `DNSName` - The file system's DNS name. You can mount your file system using its DNS name.
* `IpAddresses` - IP addresses of the file system endpoint.
