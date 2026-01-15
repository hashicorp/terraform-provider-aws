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

* `id` - (Required) The unique identifier of the cloud autonomous vm cluster.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - The Amazon Resource Name (ARN) for the Exadata infrastructure.
* `cloud_exadata_infrastructure_id` - Cloud exadata infrastructure id associated with this cloud autonomous VM cluster.
* `cloud_exadata_infrastructure_arn` - Cloud exadata infrastructure ARN associated with this cloud autonomous VM cluster.
* `autonomous_data_storage_percentage` - The percentage of data storage currently in use for Autonomous Databases in the Autonomous VM cluster.
* `autonomous_data_storage_size_in_tbs` - The data storage size allocated for Autonomous Databases in the Autonomous VM cluster, in TB.
* `available_autonomous_data_storage_size_in_tbs` - The available data storage space for Autonomous Databases in the Autonomous VM cluster, in TB.
* `available_container_databases` - The number of Autonomous CDBs that you can create with the currently available storage.
* `available_cpus` - The number of CPU cores available for allocation to Autonomous Databases.
* `compute_model` - The compute model of the Autonomous VM cluster: ECPU or OCPU.
* `cpu_core_count` - The total number of CPU cores in the Autonomous VM cluster.
* `cpu_core_count_per_node` - The number of CPU cores enabled per node in the Autonomous VM cluster.
* `cpu_percentage` - he percentage of total CPU cores currently in use in the Autonomous VM cluster.
* `created_at` - The date and time when the Autonomous VM cluster was created.
* `data_storage_size_in_gbs` - The total data storage allocated to the Autonomous VM cluster, in GB.
* `data_storage_size_in_tbs` - The total data storage allocated to the Autonomous VM cluster, in TB.
* `odb_node_storage_size_in_gbs` - The local node storage allocated to the Autonomous VM cluster, in gigabytes (GB).
* `db_servers` - The list of database servers associated with the Autonomous VM cluster.
* `description` - The user-provided description of the Autonomous VM cluster.
* `display_name` - The display name of the Autonomous VM cluster.
* `domain` - The domain name of the Autonomous VM cluster.
* `exadata_storage_in_tbs_lowest_scaled_value` - The minimum value to which you can scale down the Exadata storage, in TB.
* `hostname` - The hostname of the Autonomous VM cluster.
* `is_mtls_enabled_vm_cluster` - Indicates whether mutual TLS (mTLS) authentication is enabled for the Autonomous VM cluster.
* `license_model` - The Oracle license model that applies to the Autonomous VM cluster. Valid values are LICENSE_INCLUDED or BRING_YOUR_OWN_LICENSE.
* `max_acds_lowest_scaled_value` - The minimum value to which you can scale down the maximum number of Autonomous CDBs.
* `memory_per_oracle_compute_unit_in_gbs` - The amount of memory allocated per Oracle Compute Unit, in GB.
* `memory_size_in_gbs` - The total amount of memory allocated to the Autonomous VM cluster, in gigabytes (GB).
* `node_count` - The number of database server nodes in the Autonomous VM cluster.
* `non_provisionable_autonomous_container_databases` - The number of Autonomous CDBs that can't be provisioned because of resource  constraints.
* `oci_resource_anchor_name` - The name of the OCI resource anchor associated with this Autonomous VM cluster.
* `oci_url` - The URL for accessing the OCI console page for this Autonomous VM cluster.
* `ocid` - The Oracle Cloud Identifier (OCID) of the Autonomous VM cluster.
* `odb_network_id` - The unique identifier of the ODB network associated with this Autonomous VM cluster.
* `odb_network_arn` - The arn of the ODB network associated with this Autonomous VM cluster.
* `percent_progress` - The progress of the current operation on the Autonomous VM cluster, as a percentage.
* `provisionable_autonomous_container_databases` - The number of Autonomous CDBs that can be provisioned in the Autonomous VM cluster.
* `provisioned_autonomous_container_databases` - The number of Autonomous CDBs currently provisioned in the Autonomous VM cluster.
* `provisioned_cpus` - The number of CPU cores currently provisioned in the Autonomous VM cluster.
* `reclaimable_cpus` - The number of CPU cores that can be reclaimed from terminated or scaled-down Autonomous Databases.
* `reserved_cpus` - The number of CPU cores reserved for system operations and redundancy.
* `scan_listener_port_non_tls` - The SCAN listener port for non-TLS (TCP) protocol. The default is 1521.
* `scan_listener_port_tls` - The SCAN listener port for TLS (TCP) protocol. The default is 2484.
* `shape` - The shape of the Exadata infrastructure for the Autonomous VM cluster.
* `status` - The status of the Autonomous VM cluster.
* `status_reason` - Additional information about the current status of the Autonomous VM cluster.
* `time_database_ssl_certificate_expires` - The expiration date and time of the database SSL certificate.
* `time_ords_certificate_expires` - The expiration date and time of the Oracle REST Data Services (ORDS)certificate.
* `time_zone` - The time zone of the Autonomous VM cluster.
* `total_container_databases` - The total number of Autonomous Container Databases that can be created with the allocated local storage.
* `tags` - A map of tags to assign to the exadata infrastructure. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `maintenance_window` - The maintenance window for the Autonomous VM cluster.
