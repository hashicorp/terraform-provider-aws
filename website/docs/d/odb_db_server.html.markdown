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

* `cloud_exadata_infrastructure_id` - (Required) The unique identifier of the cloud vm cluster.
* `id` - (Required) The unique identifier of db node associated with vm cluster.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `autonomous_virtual_machine_ids` - The list of unique identifiers for the Autonomous VMs associated with this database server.
* `autonomous_vm_cluster_ids` - The OCID of the autonomous VM clusters that are associated with the database server.
* `compute_model` - The compute model of the database server.
* `status` - The status of the database server.
* `status_reason` - Additional information about the current status of the database server.
* `cpu_core_count` - The number of CPU cores enabled on the database server.
* `db_node_storage_size_in_gbs` - The allocated local node storage in GBs on the database server.
* `db_server_patching_details` - The scheduling details for the quarterly maintenance window. Patching and system updates take place during the maintenance window.
* `display_name` - The display name of the database server.
* `exadata_infrastructure_id` - The exadata infrastructure ID of the database server.
* `ocid` - The OCID of the database server to retrieve information about.
* `oci_resource_anchor_name` - The name of the OCI resource anchor.
* `max_cpu_count` - The total number of CPU cores available.
* `max_db_node_storage_in_gbs` - The total local node storage available in GBs.
* `max_memory_in_gbs` - The total memory available in GBs.
* `memory_size_in_gbs` - The allocated memory in GBs on the database server.
* `shape` - The shape of the database server. The shape determines the amount of CPU, storage, and memory resources available.
* `created_at` - The date and time when the database server was created.
* `vm_cluster_ids` - The OCID of the VM clusters that are associated with the database server.
