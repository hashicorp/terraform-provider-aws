---
subcategory: "Oracle Database@AWS"
layout: "AWS: aws_odb_cloud_exadata_infrastructure"
page_title: "AWS: aws_odb_cloud_exadata_infrastructure"
description: |-
  Terraform data source for managing exadata infrastructure resource in AWS for Oracle Database@AWS.
---

# Data Source: aws_odb_cloud_exadata_infrastructure

Terraform data source for exadata infrastructure resource in AWS for Oracle Database@AWS.

You can find out more about Oracle Database@AWS from [User Guide](https://docs.aws.amazon.com/odb/latest/UserGuide/what-is-odb.html).

## Example Usage

### Basic Usage

```terraform
data "aws_odb_cloud_exadata_infrastructure" "example" {
  id = "example"
}
```

## Argument Reference

The following arguments are required:

* `id` - (Required) Unique identifier of the Exadata infrastructure.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `activated_storage_count` - Number of storage servers requested for the Exadata infrastructure.
* `additional_storage_count` - Number of storage servers requested for the Exadata infrastructure.
* `arn` - Amazon Resource Name (ARN) for the Exadata infrastructure.
* `availability_zone` - Name of the Availability Zone (AZ) where the Exadata infrastructure is located.
* `availability_zone_id` - AZ ID of the AZ where the Exadata infrastructure is located.
* `available_storage_size_in_gbs` - Amount of available storage, in gigabytes (GB), for the Exadata infrastructure.
* `compute_count` - Number of database servers for the Exadata infrastructure.
* `compute_model` - OCI compute model used when you create or clone an instance: ECPU or OCPU. An ECPU is an abstracted measure of compute resources. ECPUs are based on the number of cores elastically allocated from a pool of compute and storage servers. An OCPU is a legacy physical measure of compute resources. OCPUs are based on the physical core of a processor with hyper-threading enabled.
* `cpu_count` - Total number of CPU cores that are allocated to the Exadata infrastructure.
* `created_at` - Time when the Exadata infrastructure was created.
* `customer_contacts_to_send_to_oci` - Email addresses of contacts to receive notification from Oracle about maintenance updates for the Exadata infrastructure.
* `data_storage_size_in_tbs` - Size of the Exadata infrastructure's data disk group, in terabytes (TB).
* `database_server_type` - Database server model type of the Exadata infrastructure. For the list of valid model names, use the ListDbSystemShapes operation.
* `db_node_storage_size_in_gbs` - Size of the storage available on each database node, in gigabytes (GB).
* `db_server_version` - Version of the Exadata infrastructure.
* `display_name` - Display name of the Exadata infrastructure.
* `id` - Unique identifier of the Exadata infrastructure.
* `last_maintenance_run_id` - Oracle Cloud Identifier (OCID) of the last maintenance run for the Exadata infrastructure.
* `maintenance_window` - Scheduling details of the maintenance window. Patching and system updates take place during the maintenance window.
* `max_cpu_count` - Total number of CPU cores available on the Exadata infrastructure.
* `max_data_storage_in_tbs` - Total amount of data disk group storage, in terabytes (TB), that's available on the Exadata infrastructure.
* `max_db_node_storage_size_in_gbs` - Total amount of local node storage, in gigabytes (GB), that's available on the Exadata infrastructure.
* `max_memory_in_gbs` - Total amount of memory, in gigabytes (GB), that's available on the Exadata infrastructure.
* `memory_size_in_gbs` - Amount of memory, in gigabytes (GB), that's allocated on the Exadata infrastructure.
* `monthly_db_server_version` - Monthly software version of the database servers installed on the Exadata infrastructure.
* `monthly_storage_server_version` - Monthly software version of the storage servers installed on the Exadata infrastructure.
* `next_maintenance_run_id` - OCID of the next maintenance run for the Exadata infrastructure.
* `oci_resource_anchor_name` - Name of the OCI resource anchor for the Exadata infrastructure.
* `oci_url` - HTTPS link to the Exadata infrastructure in OCI.
* `ocid` - OCID of the Exadata infrastructure in OCI.
* `percent_progress` - Amount of progress made on the current operation on the Exadata infrastructure expressed as a percentage.
* `shape` - Model name of the Exadata infrastructure.
* `status` - Status of the Exadata infrastructure.
* `status_reason` - Additional information about the status of the Exadata infrastructure.
* `storage_count` - Number of storage servers that are activated for the Exadata infrastructure.
* `storage_server_type` - Storage server model type of the Exadata infrastructure. For the list of valid model names, use the ListDbSystemShapes operation.
* `storage_server_version` - Software version of the storage servers on the Exadata infrastructure.
* `tags` - Map of tags assigned to the Exadata infrastructure.
* `total_storage_size_in_gbs` - Total amount of storage, in gigabytes (GB), on the Exadata infrastructure.
