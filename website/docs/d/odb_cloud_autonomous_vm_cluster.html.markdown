---
subcategory: "Oracle Database@AWS"
layout: "AWS: aws_odb_cloud_autonomous_vm_cluster"
page_title: "AWS: aws_odb_cloud_autonomous_vm_cluster"
description: |-
  Terraform data source for managing cloud autonomous vm cluster resource in AWS for Oracle Database@AWS.
---

# Data Source: aws_odb_cloud_autonomous_vm_cluster

Terraform data source for managing cloud autonomous vm cluster resource in AWS for Oracle Database@AWS.

You can find out more about Oracle Database@AWS from [User Guide](https://docs.aws.amazon.com/odb/latest/UserGuide/what-is-odb.html).

## Example Usage

### Basic Usage

```terraform
data "aws_odb_cloud_autonomous_vm_cluster" "example" {
  id = "example"
}
```

## Argument Reference

The following arguments are optional:

* `id` - (Required) Unique identifier of the cloud autonomous vm cluster.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - Amazon Resource Name (ARN) for the Exadata infrastructure.
* `autonomous_data_storage_percentage` - Percentage of data storage currently in use for Autonomous Databases in the Autonomous VM cluster.
* `autonomous_data_storage_size_in_tbs` - Data storage size allocated for Autonomous Databases in the Autonomous VM cluster, in TB.
* `available_autonomous_data_storage_size_in_tbs` - Available data storage space for Autonomous Databases in the Autonomous VM cluster, in TB.
* `available_container_databases` - Number of Autonomous CDBs that you can create with the currently available storage.
* `available_cpus` - Number of CPU cores available for allocation to Autonomous Databases.
* `cloud_exadata_infrastructure_arn` - Cloud exadata infrastructure ARN associated with this cloud autonomous VM cluster.
* `cloud_exadata_infrastructure_id` - Cloud exadata infrastructure id associated with this cloud autonomous VM cluster.
* `compute_model` - Compute model of the Autonomous VM cluster: ECPU or OCPU.
* `cpu_core_count` - Total number of CPU cores in the Autonomous VM cluster.
* `cpu_core_count_per_node` - Number of CPU cores enabled per node in the Autonomous VM cluster.
* `cpu_percentage` - Percentage of total CPU cores currently in use in the Autonomous VM cluster.
* `created_at` - Date and time when the Autonomous VM cluster was created.
* `data_storage_size_in_gbs` - Total data storage allocated to the Autonomous VM cluster, in GB.
* `data_storage_size_in_tbs` - Total data storage allocated to the Autonomous VM cluster, in TB.
* `db_servers` - List of database servers associated with the Autonomous VM cluster.
* `description` - User-provided description of the Autonomous VM cluster.
* `display_name` - Display name of the Autonomous VM cluster.
* `domain` - Domain name of the Autonomous VM cluster.
* `exadata_storage_in_tbs_lowest_scaled_value` - Minimum value to which you can scale down the Exadata storage, in TB.
* `hostname` - Hostname of the Autonomous VM cluster.
* `is_mtls_enabled_vm_cluster` - Whether mutual TLS (mTLS) authentication is enabled for the Autonomous VM cluster.
* `license_model` - Oracle license model that applies to the Autonomous VM cluster. Valid values are LICENSE_INCLUDED or BRING_YOUR_OWN_LICENSE.
* `maintenance_window` - Maintenance window for the Autonomous VM cluster.
* `max_acds_lowest_scaled_value` - Minimum value to which you can scale down the maximum number of Autonomous CDBs.
* `memory_per_oracle_compute_unit_in_gbs` - Amount of memory allocated per Oracle Compute Unit, in GB.
* `memory_size_in_gbs` - Total amount of memory allocated to the Autonomous VM cluster, in gigabytes (GB).
* `node_count` - Number of database server nodes in the Autonomous VM cluster.
* `non_provisionable_autonomous_container_databases` - Number of Autonomous CDBs that can't be provisioned because of resource constraints.
* `oci_resource_anchor_name` - Name of the OCI resource anchor associated with this Autonomous VM cluster.
* `oci_url` - URL for accessing the OCI console page for this Autonomous VM cluster.
* `ocid` - Oracle Cloud Identifier (OCID) of the Autonomous VM cluster.
* `odb_network_arn` - ARN of the ODB network associated with this Autonomous VM cluster.
* `odb_network_id` - Unique identifier of the ODB network associated with this Autonomous VM cluster.
* `odb_node_storage_size_in_gbs` - Local node storage allocated to the Autonomous VM cluster, in gigabytes (GB).
* `percent_progress` - Progress of the current operation on the Autonomous VM cluster, as a percentage.
* `provisionable_autonomous_container_databases` - Number of Autonomous CDBs that can be provisioned in the Autonomous VM cluster.
* `provisioned_autonomous_container_databases` - Number of Autonomous CDBs currently provisioned in the Autonomous VM cluster.
* `provisioned_cpus` - Number of CPU cores currently provisioned in the Autonomous VM cluster.
* `reclaimable_cpus` - Number of CPU cores that can be reclaimed from terminated or scaled-down Autonomous Databases.
* `reserved_cpus` - Number of CPU cores reserved for system operations and redundancy.
* `scan_listener_port_non_tls` - SCAN listener port for non-TLS (TCP) protocol. The default is 1521.
* `scan_listener_port_tls` - SCAN listener port for TLS (TCP) protocol. The default is 2484.
* `shape` - Shape of the Exadata infrastructure for the Autonomous VM cluster.
* `status` - Status of the Autonomous VM cluster.
* `status_reason` - Additional information about the current status of the Autonomous VM cluster.
* `tags` - Map of tags to assign to the exadata infrastructure. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `time_database_ssl_certificate_expires` - Expiration date and time of the database SSL certificate.
* `time_ords_certificate_expires` - Expiration date and time of the Oracle REST Data Services (ORDS) certificate.
* `time_zone` - Time zone of the Autonomous VM cluster.
* `total_container_databases` - Total number of Autonomous Container Databases that can be created with the allocated local storage.
