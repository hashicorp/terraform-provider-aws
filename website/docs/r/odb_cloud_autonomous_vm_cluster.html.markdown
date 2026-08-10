---
subcategory: "Oracle Database@AWS"
layout: "AWS: aws_odb_cloud_autonomous_vm_cluster"
page_title: "AWS: aws_odb_cloud_autonomous_vm_cluster"
description: |-
  Terraform resource managing cloud autonomous vm cluster in AWS for Oracle Database@AWS.
---

# Resource: aws_odb_cloud_autonomous_vm_cluster

Terraform resource managing cloud autonomous vm cluster in AWS for Oracle Database@AWS.

You can find out more about Oracle Database@AWS from [User Guide](https://docs.aws.amazon.com/odb/latest/UserGuide/what-is-odb.html).

## Example Usage

### Basic Usage

```terraform
resource "aws_odb_cloud_autonomous_vm_cluster" "avmc_with_minimum_parameters" {
  cloud_exadata_infrastructure_id       = "<aws_odb_cloud_exadata_infrastructure_id>"
  odb_network_id                        = "<aws_odb_network_id>"
  display_name                          = "my_autonomous_vm_cluster"
  autonomous_data_storage_size_in_tbs   = 5
  memory_per_oracle_compute_unit_in_gbs = 2
  total_container_databases             = 1
  cpu_core_count_per_node               = 40
  license_model                         = "LICENSE_INCLUDED"
  # ids of db server. refer your exa infra. This is a manadatory fileld. Refer your cloud exadata infrastructure for db server id
  db_servers                 = ["<my_db_server_id>"]
  scan_listener_port_tls     = 8561
  scan_listener_port_non_tls = 1024
  maintenance_window {
    preference = "NO_PREFERENCE"
  }

}

resource "aws_odb_cloud_autonomous_vm_cluster" "avmc_with_all_params" {
  description                           = "my first avmc"
  time_zone                             = "UTC"
  cloud_exadata_infrastructure_id       = "<aws_odb_cloud_exadata_infrastructure_id>"
  odb_network_id                        = "<aws_odb_network_id>"
  display_name                          = "my_autonomous_vm_cluster"
  autonomous_data_storage_size_in_tbs   = 5
  memory_per_oracle_compute_unit_in_gbs = 2
  total_container_databases             = 1
  cpu_core_count_per_node               = 40
  license_model                         = "LICENSE_INCLUDED"
  db_servers                            = ["<my_db_server_1>", "<my_db_server_2>"]
  scan_listener_port_tls                = 8561
  scan_listener_port_non_tls            = 1024
  maintenance_window {
    days_of_week       = [{ name = "MONDAY" }, { name = "TUESDAY" }]
    hours_of_day       = [4, 16]
    lead_time_in_weeks = 3
    months             = [{ name = "FEBRUARY" }, { name = "MAY" }, { name = "AUGUST" }, { name = "NOVEMBER" }]
    preference         = "CUSTOM_PREFERENCE"
    weeks_of_month     = [2, 4]
  }
  tags = {
    "env" = "dev"
  }

}

```

## Argument Reference

The following arguments are required:

* `autonomous_data_storage_size_in_tbs` - (Required) Data storage size allocated for Autonomous Databases in the Autonomous VM cluster, in TB. Changing this will force terraform to create new resource.
* `cpu_core_count_per_node` - (Required) Number of CPU cores enabled per node in the Autonomous VM cluster. Changing this will force terraform to create new resource.
* `db_servers` - (Required) Database servers in the Autonomous VM cluster. Changing this will force terraform to create new resource.
* `display_name` - (Required) Display name of the Autonomous VM cluster. Changing this will force terraform to create new resource.
* `maintenance_window` - (Required) Maintenance window of the Autonomous VM cluster. Changing this will force terraform to create new resource.
* `memory_per_oracle_compute_unit_in_gbs` - (Required) Amount of memory allocated per Oracle Compute Unit, in GB. Changing this will force terraform to create new resource.
* `scan_listener_port_non_tls` - (Required) SCAN listener port for non-TLS (TCP) protocol. The default is 1521. Changing this will force terraform to create new resource.
* `scan_listener_port_tls` - (Required) SCAN listener port for TLS (TCP) protocol. The default is 2484. Changing this will force terraform to create new resource.
* `total_container_databases` - (Required) Total number of Autonomous Container Databases that can be created with the allocated local storage. Changing this will force terraform to create new resource.

The following arguments are optional:

* `cloud_exadata_infrastructure_arn` - (Optional) Exadata infrastructure ARN. Changing this will force Terraform to create a new resource. Either the combination of `cloud_exadata_infrastructure_id` and `odb_network_id` or `cloud_exadata_infrastructure_arn` and `odb_network_arn` must be used.
* `cloud_exadata_infrastructure_id` - (Optional) Exadata infrastructure id. Changing this will force Terraform to create a new resource. Either the combination of `cloud_exadata_infrastructure_id` and `odb_network_id` or `cloud_exadata_infrastructure_arn` and `odb_network_arn` must be used.
* `description` - (Optional) Description of the Autonomous VM cluster.
* `is_mtls_enabled_vm_cluster` - (Optional) Whether mutual TLS (mTLS) authentication is enabled for the Autonomous VM cluster. Changing this will force terraform to create new resource.
* `license_model` - (Optional) License model for the Autonomous VM cluster. Valid values are LICENSE_INCLUDED or BRING_YOUR_OWN_LICENSE. Changing this will force terraform to create new resource.
* `odb_network_arn` - (Optional) ARN of the ODB network associated with this Autonomous VM Cluster. Changing this will force Terraform to create a new resource. Either the combination of `cloud_exadata_infrastructure_id` and `odb_network_id` or `cloud_exadata_infrastructure_arn` and `odb_network_arn` must be used.
* `odb_network_id` - (Optional) Unique identifier of the ODB network associated with this Autonomous VM Cluster. Changing this will force Terraform to create a new resource. Changing this will create a new resource. Either the combination of `cloud_exadata_infrastructure_id` and `odb_network_id` or `cloud_exadata_infrastructure_arn` and `odb_network_arn` must be used.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Map of tags to assign to the exadata infrastructure. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `time_zone` - (Optional) Time zone of the Autonomous VM cluster. Changing this will force terraform to create new resource.

### `maintenance_window` Block

* `days_of_week` - (Optional) Days of the week when maintenance can be performed. Changing this will force terraform to create new resource.
* `hours_of_day` - (Optional) Hours of the day when maintenance can be performed. Changing this will force terraform to create new resource.
* `lead_time_in_weeks` - (Optional) Lead time in weeks before the maintenance window. Changing this will force terraform to create new resource.
* `months` - (Optional) Months when maintenance can be performed. Changing this will force terraform to create new resource.
* `preference` - (Required) Preference for the maintenance window scheduling. Changing this will force terraform to create new resource.
* `weeks_of_month` - (Optional) Whether to skip release updates during maintenance. Changing this will force terraform to create new resource.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Amazon Resource Name (ARN) for the Exadata infrastructure.
* `autonomous_data_storage_percentage` - Progress of the current operation on the Autonomous VM cluster, as a percentage.
* `available_autonomous_data_storage_size_in_tbs` - Available data storage space for Autonomous Databases in the Autonomous VM cluster, in TB.
* `available_container_databases` - Number of Autonomous CDBs that you can create with the currently available storage.
* `available_cpus` - Number of CPU cores available for allocation to Autonomous Databases.
* `compute_model` - Compute model of the Autonomous VM cluster: ECPU or OCPU.
* `cpu_core_count` - Total number of CPU cores in the Autonomous VM cluster.
* `cpu_percentage` - Percentage of total CPU cores currently in use in the Autonomous VM cluster.
* `created_at` - Date and time when the Autonomous VM cluster was created.
* `data_storage_size_in_gbs` - Total data storage allocated to the Autonomous VM cluster, in GB.
* `data_storage_size_in_tbs` - Total data storage allocated to the Autonomous VM cluster, in TB.
* `domain` - Domain name of the Autonomous VM cluster.
* `exadata_storage_in_tbs_lowest_scaled_value` - Minimum value to which you can scale down the Exadata storage, in TB.
* `hostname` - Hostname of the Autonomous VM cluster.
* `id` - Unique identifier of autonomous vm cluster.
* `license_model` - License model for the Autonomous VM cluster. Valid values are LICENSE_INCLUDED or BRING_YOUR_OWN_LICENSE.
* `max_acds_lowest_scaled_value` - Minimum value to which you can scale down the maximum number of Autonomous CDBs.
* `memory_size_in_gbs` - Total amount of memory allocated to the Autonomous VM cluster, in gigabytes(GB).
* `node_count` - Number of database server nodes in the Autonomous VM cluster.
* `non_provisionable_autonomous_container_databases` - Number of Autonomous CDBs that can't be provisioned because of resource constraints.
* `oci_resource_anchor_name` - Name of the OCI resource anchor associated with this Autonomous VM cluster.
* `oci_url` - URL for accessing the OCI console page for this Autonomous VM cluster.
* `ocid` - Oracle Cloud Identifier (OCID) of the Autonomous VM cluster.
* `odb_node_storage_size_in_gbs` - Local node storage allocated to the Autonomous VM cluster, in gigabytes (GB).
* `percent_progress` - Progress of the current operation on the Autonomous VM cluster, as a percentage.
* `provisionable_autonomous_container_databases` - Number of Autonomous CDBs that can be provisioned in the Autonomous VM cluster.
* `provisioned_autonomous_container_databases` - Number of Autonomous CDBs currently provisioned in the Autonomous VM cluster.
* `provisioned_cpus` - Number of CPUs provisioned in the Autonomous VM cluster.
* `reclaimable_cpus` - Number of CPU cores that can be reclaimed from terminated or scaled-down Autonomous Databases.
* `reserved_cpus` - Number of CPU cores reserved for system operations and redundancy.
* `shape` - Shape of the Exadata infrastructure for the Autonomous VM cluster.
* `status` - Status of the Autonomous VM cluster. Possible values include CREATING, AVAILABLE, UPDATING, DELETING, DELETED, FAILED.
* `status_reason` - Additional information about the current status of the Autonomous VM cluster.
* `tags_all` - Combined set of user-defined and provider-defined tags.
* `time_database_ssl_certificate_expires` - Expiration date and time of the database SSL certificate.
* `time_ords_certificate_expires` - Expiration date and time of the ORDS certificate.
* `time_zone` - Time zone of the Autonomous VM cluster.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `24h`)
* `update` - (Default `24h`)
* `delete` - (Default `24h`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import OpenSearch Ingestion Pipeline using the `id`. For example:

```terraform
import {
  to = aws_odb_cloud_autonomous_vm_cluster.example
  id = "example"
}
```

Using `terraform import`, import cloud autonomous vm cluster `id`. For example:

```console
% terraform import aws_odb_cloud_autonomous_vm_cluster.example example
```
