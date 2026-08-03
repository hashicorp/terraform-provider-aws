---
subcategory: "Oracle Database@AWS"
layout: "AWS: aws_odb_db_server"
page_title: "AWS: aws_odb_db_server"
description: |-
  Terraform data source for managing db server linked to exadata infrastructure of Oracle Database@AWS.
---

# Data Source: aws_odb_db_server

Terraform data source for manging db server linked to exadata infrastructure of Oracle Database@AWS.

You can find out more about Oracle Database@AWS from [User Guide](https://docs.aws.amazon.com/odb/latest/UserGuide/what-is-odb.html).

## Example Usage

### Basic Usage

```terraform
data "aws_odb_db_server" "example" {
  cloud_exadata_infrastructure_id = "exadata_infra_id"
  id                              = "db_server_id"
}
```

## Argument Reference

The following arguments are required:

* `cloud_exadata_infrastructure_id` - (Required) Unique identifier of the cloud vm cluster.
* `id` - (Required) Unique identifier of db node associated with vm cluster.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `autonomous_virtual_machine_ids` - List of unique identifiers for the Autonomous VMs associated with this database server.
* `autonomous_vm_cluster_ids` - OCID of the autonomous VM clusters that are associated with the database server.
* `compute_model` - Compute model of the database server.
* `cpu_core_count` - Number of CPU cores enabled on the database server.
* `created_at` - Date and time when the database server was created.
* `db_node_storage_size_in_gbs` - Allocated local node storage in GBs on the database server.
* `db_server_patching_details` - Scheduling details for the quarterly maintenance window. Patching and system updates take place during the maintenance window.
* `display_name` - Display name of the database server.
* `exadata_infrastructure_id` - Exadata infrastructure ID of the database server.
* `max_cpu_count` - Total number of CPU cores available.
* `max_db_node_storage_in_gbs` - Total local node storage available in GBs.
* `max_memory_in_gbs` - Total memory available in GBs.
* `memory_size_in_gbs` - Allocated memory in GBs on the database server.
* `oci_resource_anchor_name` - Name of the OCI resource anchor.
* `ocid` - OCID of the database server to retrieve information about.
* `shape` - Shape of the database server. The shape determines the amount of CPU, storage, and memory resources available.
* `status` - Status of the database server.
* `status_reason` - Additional information about the current status of the database server.
* `vm_cluster_ids` - OCID of the VM clusters that are associated with the database server.
