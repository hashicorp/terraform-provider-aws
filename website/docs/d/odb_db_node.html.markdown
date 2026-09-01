---
subcategory: "Oracle Database@AWS"
layout: "AWS: aws_odb_db_node"
page_title: "AWS: aws_odb_db_node"
description: |-
  Terraform data source for managing db node linked to cloud vm cluster of Oracle Database@AWS.
---

# Data Source: aws_odb_db_node

Terraform data source for manging db nodes linked to cloud vm cluster of Oracle Database@AWS.

You can find out more about Oracle Database@AWS from [User Guide](https://docs.aws.amazon.com/odb/latest/UserGuide/what-is-odb.html).

## Example Usage

### Basic Usage

```terraform
data "aws_odb_db_node" "example" {
  cloud_vm_cluster_id = "cloud_vm_cluster_id"
  id                  = "db_node_id"
}
```

## Argument Reference

The following arguments are required:

* `cloud_vm_cluster_id` - (Required) Unique identifier of the cloud vm cluster.
* `id` - (Required) Unique identifier of db node associated with vm cluster.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `additional_details` - Additional information about the planned maintenance.
* `arn` - ARN of the DB node.
* `backup_ip_id` - Oracle Cloud ID (OCID) of the backup IP address that's associated with the DB node.
* `backup_vnic2_id` - OCID of the second backup VNIC.
* `backup_vnic_id` - OCID of the backup VNIC.
* `cloud_vm_cluster_id` - ID of the cloud VM cluster.
* `cpu_core_count` - Number of CPU cores enabled on the DB node.
* `created_at` - Date and time when the DB node was created.
* `db_server_id` - Unique identifier of the DB server that is associated with the DB node.
* `db_storage_size_in_gbs` - Amount of local node storage, in gigabytes (GB), allocated on the DB node.
* `db_system_id` - OCID of the DB system.
* `fault_domain` - Name of the fault domain the instance is contained in.
* `floating_ip_address` - Floating IP address assigned to the DB node.
* `host_ip_id` - OCID of the host IP address that's associated with the DB node.
* `hostname` - Host name for the DB node.
* `maintenance_type` - Type of database node maintenance. Either VMDB_REBOOT_MIGRATION or EXADBXS_REBOOT_MIGRATION.
* `memory_size_in_gbs` - Allocated memory in GBs on the DB node.
* `oci_resource_anchor_name` - Name of the OCI resource anchor for the DB node.
* `ocid` - OCID of the DB node.
* `private_ip_address` - Private IP address assigned to the DB node.
* `software_storage_size_in_gbs` - Size (in GB) of the block storage volume allocation for the DB system.
* `status` - Current status of the DB node.
* `status_reason` - Additional information about the status of the DB node.
* `time_maintenance_window_end` - End date and time of the maintenance window.
* `time_maintenance_window_start` - Start date and time of the maintenance window.
* `total_cpu_core_count` - Total number of CPU cores reserved on the DB node.
* `vnic2_id` - OCID of the second VNIC.
* `vnic_id` - OCID of the VNIC.
