---
subcategory: "Oracle Database@AWS"
layout: "AWS: aws_odb_db_nodes"
page_title: "AWS: aws_odb_db_nodes"
description: |-
  Terraform data source for managing db nodes linked to cloud vm cluster of Oracle Database@AWS.
---

# Data Source: aws_odb_db_nodes

Terraform data source for manging db nodes linked to cloud vm cluster of Oracle Database@AWS.

You can find out more about Oracle Database@AWS from [User Guide](https://docs.aws.amazon.com/odb/latest/UserGuide/what-is-odb.html).

## Example Usage

### Basic Usage

```terraform
data "aws_odb_db_nodes" "example" {
  cloud_vm_cluster_id = "example"
}
```

## Argument Reference

The following arguments are required:

* `cloud_vm_cluster_id` - (Required) Unique identifier of the cloud vm cluster.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `db_nodes` - List of DB nodes along with their properties.

### db_nodes

* `additional_details` - Additional information about the planned maintenance.
* `backup_ip_id` - Oracle Cloud ID (OCID) of the backup IP address that's associated with the DB node.
* `backup_vnic_2_id` - OCID of the second backup virtual network interface card (VNIC) for the DB node.
* `backup_vnic_id` - OCID of the backup VNIC for the DB node.
* `cpu_core_count` - Number of CPU cores enabled on the DB node.
* `created_at` - Date and time when the DB node was created.
* `db_node_arn` - Amazon Resource Name (ARN) of the DB node.
* `db_node_id` - Unique identifier of the DB node.
* `db_node_storage_size_in_gbs` - Amount of local node storage, in gigabytes (GB), that's allocated on the DB node.
* `db_server_id` - Unique identifier of the database server that's associated with the DB node.
* `db_system_id` - OCID of the DB system.
* `fault_domain` - Name of the fault domain where the DB node is located.
* `host_ip_id` - OCID of the host IP address that's associated with the DB node.
* `hostname` - Host name for the DB node.
* `maintenance_type` - Type of maintenance the DB node is undergoing.
* `memory_size_in_gbs` - Amount of memory, in gigabytes (GB), that's allocated on the DB node.
* `oci_resource_anchor_name` - Name of the OCI resource anchor for the DB node.
* `ocid` - OCID of the DB node.
* `software_storage_size_in_gb` - Size of the block storage volume, in gigabytes (GB), that's allocated for the DB system. This attribute applies only for virtual machine DB systems.
* `status` - Current status of the DB node.
* `status_reason` - Additional information about the status of the DB node.
* `time_maintenance_window_end` - End date and time of the maintenance window.
* `time_maintenance_window_start` - Start date and time of the maintenance window.
* `total_cpu_core_count` - Total number of CPU cores reserved on the DB node.
* `vnic_2_id` - OCID of the second VNIC.
* `vnic_id` - OCID of the VNIC.
