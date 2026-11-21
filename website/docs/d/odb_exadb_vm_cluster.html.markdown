---
subcategory: "Oracle Database@AWS"
layout: "aws"
page_title: "AWS: aws_odb_exadb_vm_cluster"
description: |-
  Provides details about an Oracle Database@AWS ExaDB VM Cluster.
---

# Data Source: aws_odb_exadb_vm_cluster

Provides details about an [Oracle Database@AWS ExaDB VM Cluster](https://docs.aws.amazon.com/odb/latest/APIReference/API_GetExadbVmCluster.html).

## Example Usage

### Basic Usage

```terraform
data "aws_odb_exadb_vm_cluster" "example" {
  id = "exadbvmcluster_0123456789"
}
```

## Argument Reference

The following arguments are required:

* `id` - (Required) Unique identifier of the ExaDB VM Cluster. Length must be between `6` and `2048` characters.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the ExaDB VM Cluster.
* `cluster_name` - Name of the Grid Infrastructure cluster.
* `created_at` - Date and time when the ExaDB VM Cluster was created.
* `data_collection_options` - Diagnostic collection preferences for the ExaDB VM Cluster. See [`data_collection_options` Block](#data_collection_options-block) below.
* `display_name` - User-friendly name for the ExaDB VM Cluster.
* `domain` - Domain of the ExaDB VM Cluster.
* `enabled_ecpu_count` - Number of ECPUs enabled for the ExaDB VM Cluster.
* `exascale_db_storage_vault_arn` - ARN of the Exascale DB Storage Vault associated with the ExaDB VM Cluster.
* `exascale_db_storage_vault_id` - ID of the Exascale DB Storage Vault associated with the ExaDB VM Cluster.
* `gi_version` - Oracle Grid Infrastructure software version for the ExaDB VM Cluster.
* `grid_image_id` - Grid Infrastructure software image ID for the ExaDB VM Cluster.
* `grid_image_type` - Type of Grid Infrastructure image used by the ExaDB VM Cluster.
* `hostname` - Host name for the ExaDB VM Cluster.
* `iam_roles` - IAM service roles associated with the ExaDB VM Cluster. See [`iam_roles` Block](#iam_roles-block) below.
* `iorm_config_cache` - IORM configuration cache details for the ExaDB VM Cluster. See [`iorm_config_cache` Block](#iorm_config_cache-block) below.
* `last_update_history_entry_id` - OCID of the last maintenance update history entry.
* `license_model` - Oracle license model applied to the ExaDB VM Cluster.
* `listener_port` - Listener port configured for the ExaDB VM Cluster.
* `memory_size_in_gbs` - Amount of memory allocated to the ExaDB VM Cluster, in GB.
* `node_count` - Number of nodes in the ExaDB VM Cluster.
* `oci_resource_anchor_name` - Name of the OCI resource anchor for the ExaDB VM Cluster.
* `oci_url` - HTTPS URL of the ExaDB VM Cluster in OCI.
* `ocid` - OCID of the ExaDB VM Cluster.
* `odb_network_arn` - ARN of the ODB network associated with the ExaDB VM Cluster.
* `odb_network_id` - ID of the ODB network associated with the ExaDB VM Cluster.
* `percent_progress` - Progress of the current operation, expressed as a percentage.
* `scan_dns_name` - FQDN of the SCAN IP addresses associated with the ExaDB VM Cluster.
* `scan_dns_record_id` - OCID of the DNS record for the SCAN IP addresses.
* `scan_ip_ids` - OCIDs of the SCAN IP addresses associated with the ExaDB VM Cluster.
* `scan_listener_port_tcp` - Port for TCP connections to the SCAN listener.
* `scan_listener_port_tcp_ssl` - Port for SSL/TCP connections to the SCAN listener.
* `shape` - Shape of the ExaDB VM Cluster.
* `shape_attribute` - Shape attribute for the ExaDB VM Cluster.
* `snapshot_file_system_storage` - Snapshot file system storage details for the ExaDB VM Cluster. See [`snapshot_file_system_storage` Block](#snapshot_file_system_storage-block) below.
* `ssh_public_keys` - Public keys used for SSH access to the ExaDB VM Cluster.
* `status` - Current status of the ExaDB VM Cluster.
* `status_reason` - Additional information about the current ExaDB VM Cluster status.
* `system_version` - Operating system version of the image for the ExaDB VM Cluster.
* `tags` - Map of tags assigned to the resource.
* `time_zone` - Time zone for the ExaDB VM Cluster.
* `total_ecpu_count` - Total number of ECPUs for the ExaDB VM Cluster.
* `total_file_system_storage` - Total file system storage details for the ExaDB VM Cluster. See [`total_file_system_storage` Block](#total_file_system_storage-block) below.
* `vip_ids` - OCIDs of the virtual IP addresses associated with the ExaDB VM Cluster.
* `vm_file_system_storage` - VM file system storage details for the ExaDB VM Cluster. See [`vm_file_system_storage` Block](#vm_file_system_storage-block) below.
* `vm_file_system_storage_total_size_in_gbs` - Total amount of VM file system storage for the ExaDB VM Cluster, in GB.

### `data_collection_options` Block

* `is_diagnostics_events_enabled` - Whether diagnostic event collection is enabled.
* `is_health_monitoring_enabled` - Whether health monitoring is enabled.
* `is_incident_logs_enabled` - Whether incident log collection is enabled.

### `iam_roles` Block

* `aws_integration` - AWS integration supported by the IAM role.
* `iam_role_arn` - ARN of the IAM role.
* `status` - Status of the IAM role association.
* `status_reason` - Additional information about the IAM role association status.

### `iorm_config_cache` Block

* `db_plans` - IORM database plans for the ExaDB VM Cluster. See [`db_plans` Block](#db_plans-block) below.
* `lifecycle_details` - Additional information about the current IORM configuration lifecycle state.
* `lifecycle_state` - Current lifecycle state of the IORM configuration.
* `objective` - Current IORM objective.

### `db_plans` Block

* `db_name` - Database name to which the IORM plan applies.
* `flash_cache_limit` - Flash cache limit for the database plan.
* `share` - Relative priority of the database in the IORM plan.

### `snapshot_file_system_storage` Block

* `total_size_in_gbs` - Total storage size, in GB.

### `total_file_system_storage` Block

* `total_size_in_gbs` - Total storage size, in GB.

### `vm_file_system_storage` Block

* `total_size_in_gbs` - Total storage size, in GB.
