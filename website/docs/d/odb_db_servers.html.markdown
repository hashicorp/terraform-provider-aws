---
subcategory: "Oracle Database@AWS"
layout: "AWS: aws_odb_db_servers"
page_title: "AWS: aws_odb_db_servers"
description: |-
  Terraform data source for managing db servers linked to exadata infrastructure of Oracle Database@AWS.
---

# Data Source: aws_odb_db_servers

Terraform data source for manging db servers linked to exadata infrastructure of Oracle Database@AWS.

You can find out more about Oracle Database@AWS from [User Guide](https://docs.aws.amazon.com/odb/latest/UserGuide/what-is-odb.html).

## Example Usage

### Basic Usage

```terraform
data "aws_odb_db_servers" "example" {
  cloud_exadata_infrastructure_id = "exadata_infra_id"
}
```

## Argument Reference

The following arguments are required:

* `cloud_exadata_infrastructure_id` - (Required) The unique identifier of the cloud vm cluster.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `db_servers` - the list of DB servers along with their properties.

### db_servers

* `autonomous_virtual_machine_ids` - A list of unique identifiers for the Autonomous VMs.
* `autonomous_vm_cluster_ids` - A list of identifiers for the Autonomous VM clusters.
* `compute_model` - The OCI compute model used when you create or clone an instance: **ECPU** or **OCPU**. ECPUs are based on the number of cores elastically allocated from a pool of compute and storage servers, while OCPUs are based on the physical core of a processor with hyper-threading enabled.
* `cpu_core_count` - The number of CPU cores enabled on the database server.
* `created_at` - The date and time when the database server was created.
* `db_node_storage_size_in_gbs` - The amount of local node storage, in gigabytes (GB), that's allocated on the database server.
* `id` - The unique identifier of the database server.
* `db_server_patching_details` - The scheduling details for the quarterly maintenance window. Patching and system updates take place during the maintenance window.
* `display_name` - The user-friendly name of the database server. The name doesn't need to be unique.
* `exadata_infrastructure_id` - The ID of the Exadata infrastructure that hosts the database server.
* `max_cpu_count` - The total number of CPU cores available on the database server.
* `max_db_node_storage_in_gbs` - The total amount of local node storage, in gigabytes (GB), that's available on the database server.
* `max_memory_in_gbs` - The total amount of memory, in gigabytes (GB), that's available on the database server.
* `memory_size_in_gbs` - The amount of memory, in gigabytes (GB), that's allocated on the database server.
* `oci_resource_anchor_name` - The name of the OCI resource anchor for the database server.
* `ocid` - The OCID of the database server.
* `shape` - The hardware system model of the Exadata infrastructure that the database server is hosted on. The shape determines the amount of CPU, storage, and memory resources available.
* `status` - The current status of the database server.
* `status_reason` - Additional information about the status of the database server.
* `vm_cluster_ids` - The IDs of the VM clusters that are associated with the database server.
