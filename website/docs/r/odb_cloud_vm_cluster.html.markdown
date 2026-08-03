---
subcategory: "Oracle Database@AWS"
layout: "AWS: aws_odb_cloud_vm_cluster"
page_title: "AWS: aws_odb_cloud_vm_cluster"
description: |-
  Terraform resource for managing cloud vm cluster resource in AWS for Oracle Database@AWS.
---

# Resource: aws_odb_cloud_vm_cluster

Terraform to manage cloud vm cluster resource in AWS for Oracle Database@AWS. If underlying odb network and cloud exadata infrastructure is shared, ARN must be used while creating VM cluster.

You can find out more about Oracle Database@AWS from [User Guide](https://docs.aws.amazon.com/odb/latest/UserGuide/what-is-odb.html).

## Example Usage

### Basic Usage

```terraform
resource "aws_odb_cloud_vm_cluster" "with_minimum_parameter" {
  display_name                    = "my_vm_cluster"
  cloud_exadata_infrastructure_id = "<aws_odb_cloud_exadata_infrastructure_id>"
  cpu_core_count                  = 6
  gi_version                      = "23.0.0.0"
  hostname_prefix                 = "apollo12"
  ssh_public_keys                 = ["public-ssh-key"]
  odb_network_id                  = "<aws_odb_network_id>"
  is_local_backup_enabled         = true
  is_sparse_diskgroup_enabled     = true
  license_model                   = "LICENSE_INCLUDED"
  data_storage_size_in_tbs        = 20.0
  db_servers                      = ["db-server-1", "db-server-2"]
  db_node_storage_size_in_gbs     = 120.0
  memory_size_in_gbs              = 60
  data_collection_options {
    is_diagnostics_events_enabled = false
    is_health_monitoring_enabled  = false
    is_incident_logs_enabled      = false
  }
}
```

### With Optional Arguments

```terraform
resource "aws_odb_cloud_vm_cluster" "with_all_parameters" {
  display_name                    = "my_vm_cluster"
  cloud_exadata_infrastructure_id = "<aws_odb_cloud_exadata_infrastructure_id>"
  cpu_core_count                  = 6
  gi_version                      = "23.0.0.0"
  hostname_prefix                 = "apollo12"
  ssh_public_keys                 = ["my-ssh-key"]
  odb_network_id                  = "<aws_odb_network_id>"
  is_local_backup_enabled         = true
  is_sparse_diskgroup_enabled     = true
  license_model                   = "LICENSE_INCLUDED"
  data_storage_size_in_tbs        = 20.0
  db_servers                      = ["my-dbserver-1", "my-db-server-2"]
  db_node_storage_size_in_gbs     = 120.0
  memory_size_in_gbs              = 60
  cluster_name                    = "julia-13"
  timezone                        = "UTC"
  scan_listener_port_tcp          = 1521
  tags = {
    "env" = "dev"
  }
  data_collection_options {
    is_diagnostics_events_enabled = true
    is_health_monitoring_enabled  = true
    is_incident_logs_enabled      = true
  }
}
```

### With GI Version Tag

```terraform
resource "aws_odb_cloud_vm_cluster" "gi_version_tag_example" {
  display_name                    = "my_vm_cluster"
  cloud_exadata_infrastructure_id = "<aws_odb_cloud_exadata_infrastructure_id>"
  cpu_core_count                  = 6
  gi_version                      = "23.0.0.0"
  hostname_prefix                 = "apollo12"
  ssh_public_keys                 = ["my-ssh-key"]
  odb_network_id                  = "<aws_odb_network_id>"
  is_local_backup_enabled         = true
  is_sparse_diskgroup_enabled     = true
  license_model                   = "LICENSE_INCLUDED"
  data_storage_size_in_tbs        = 20.0
  db_servers                      = ["my-dbserver-1", "my-db-server-2"]
  db_node_storage_size_in_gbs     = 120.0
  memory_size_in_gbs              = 60
  cluster_name                    = "julia-13"
  timezone                        = "UTC"
  scan_listener_port_tcp          = 1521
  tags = {
    "odb:input_gi_version" = "23.0.0.0"
  }
  data_collection_options {
    is_diagnostics_events_enabled = true
    is_health_monitoring_enabled  = true
    is_incident_logs_enabled      = true
  }
}
```

## Argument Reference

The following arguments are required:

* `cpu_core_count` - (Required) Number of CPU cores to enable on the VM cluster. Changing this will create a new resource.
* `data_collection_options` - (Required) Set of preferences for the various diagnostic collection options for the VM cluster. See [`data_collection_options` Block](#data_collection_options-block) below. Changing this will create a new resource.
* `data_storage_size_in_tbs` - (Required) Size of the data disk group, in terabytes (TBs), to allocate for the VM cluster. Changing this will create a new resource.
* `db_servers` - (Required) List of database servers for the VM cluster. Changing this will create a new resource.
* `display_name` - (Required) User-friendly name for the VM cluster. Changing this will create a new resource.
* `gi_version` - (Required) Valid Oracle Grid Infrastructure (GI) software version. To get valid values, use the ListGiVersions operation for the Exadata infrastructure shape. Example: `19.0.0.0`. Changing this creates a new resource. Prefer to provide `odb:input_gi_version` tag. If `odb:input_gi_version` tag is provided, its value must exactly match `gi_version`, otherwise Terraform returns an error. See the [`With GI Version Tag`](#with-gi-version-tag) example above.
* `hostname_prefix` - (Required) Host name prefix for the VM cluster. Constraints: - Can't be "localhost" or "hostname". - Can't contain "-version". - Maximum length of the combined hostname and domain is 63 characters. - Hostname must be unique within the subnet. Changing this will create a new resource.
* `ssh_public_keys` - (Required) Public key portion of one or more key pairs used for SSH access to the VM cluster. Changing this will create a new resource.

The following arguments are optional:

* `cloud_exadata_infrastructure_arn` - (Optional) ARN of the Exadata infrastructure for this VM cluster. Changing this will create a new resource. Either the combination of cloud_exadata_infrastructure_id and odb_network_id or cloud_exadata_infrastructure_arn and odb_network_arn must be used.
* `cloud_exadata_infrastructure_id` - (Optional) Unique identifier of the Exadata infrastructure for this VM cluster. Changing this will create a new resource. Either the combination of cloud_exadata_infrastructure_id and odb_network_id or cloud_exadata_infrastructure_arn and odb_network_arn must be used.
* `cluster_name` - (Optional) Name of the Grid Infrastructure (GI) cluster. Changing this will create a new resource.
* `db_node_storage_size_in_gbs` - (Optional) Amount of local node storage, in gigabytes (GBs), to allocate for the VM cluster. Changing this will create a new resource.
* `is_local_backup_enabled` - (Optional) Whether to enable database backups to local Exadata storage for the VM cluster. Changing this will create a new resource.
* `is_sparse_diskgroup_enabled` - (Optional) Whether to create a sparse disk group for the VM cluster. Changing this will create a new resource.
* `license_model` - (Optional) Oracle license model to apply to the VM cluster. Default: LICENSE_INCLUDED. Changing this will create a new resource.
* `memory_size_in_gbs` - (Optional) Amount of memory, in gigabytes (GBs), to allocate for the VM cluster. Changing this will create a new resource.
* `odb_network_arn` - (Optional) ARN of the ODB network for the VM cluster. Changing this will create a new resource. Either the combination of cloud_exadata_infrastructure_id and odb_network_id or cloud_exadata_infrastructure_arn and odb_network_arn must be used.
* `odb_network_id` - (Optional) Unique identifier of the ODB network for the VM cluster. Changing this will create a new resource. Either the combination of cloud_exadata_infrastructure_id and odb_network_id or cloud_exadata_infrastructure_arn and odb_network_arn must be used.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `scan_listener_port_tcp` - (Optional) Port number for TCP connections to the single client access name (SCAN) listener. Valid values: 1024–8999, except 2484, 6100, 6200, 7060, 7070, 7085, and 7879. Default: 1521. Changing this will create a new resource.
* `tags` - (Optional) Map of tags to assign to the exadata infrastructure. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `timezone` - (Optional) Configured time zone of the VM cluster. Changing this will create a new resource.

### `data_collection_options` Block

The `data_collection_options` block supports the following:

* `is_diagnostics_events_enabled` - (Required) Whether to enable diagnostic events for the VM cluster. Changing this will create a new resource.
* `is_health_monitoring_enabled` - (Required) Whether to enable health monitoring for the VM cluster. Changing this will create a new resource.
* `is_incident_logs_enabled` - (Required) Whether to enable incident logs for the VM cluster. Changing this will create a new resource.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Amazon Resource Name (ARN) for the cloud vm cluster.
* `compute_model` - Compute model used when the instance is created or cloned — either ECPU or OCPU. ECPU is a virtualized compute unit; OCPU is a physical processor core with hyper-threading.
* `created_at` - Timestamp when the VM cluster was created.
* `disk_redundancy` - Type of redundancy for the VM cluster: NORMAL (2-way) or HIGH (3-way).
* `domain` - Domain name associated with the VM cluster.
* `gi_version_computed` - Complete software version of Oracle Grid Infrastructure (GI).
* `hostname_prefix_computed` - Host name for the VM cluster. Constraints: - Can't be "localhost" or "hostname". - Can't contain "-version". - Maximum length of the combined hostname and domain is 63 characters. - Hostname must be unique within the subnet.
* `id` - Unique identifier of vm cluster.
* `iorm_config_cache` - Exadata IORM (I/O Resource Manager) configuration cache details for the VM cluster.
* `last_update_history_entry_id` - OCID of the most recent maintenance update history entry.
* `listener_port` - Listener port number configured on the VM cluster.
* `node_count` - Total number of nodes in the VM cluster.
* `oci_resource_anchor_name` - Name of the OCI resource anchor associated with the VM cluster.
* `oci_url` - HTTPS link to the VM cluster resource in OCI.
* `ocid` - OCID (Oracle Cloud Identifier) of the VM cluster.
* `percent_progress` - Percentage of progress made on the current operation for the VM cluster.
* `scan_dns_name` - Fully qualified domain name (FQDN) for the SCAN IP addresses associated with the VM cluster.
* `scan_dns_record_id` - OCID of the DNS record for the SCAN IPs linked to the VM cluster.
* `scan_ip_ids` - List of OCIDs for SCAN IP addresses associated with the VM cluster.
* `shape` - Hardware model name of the Exadata infrastructure running the VM cluster.
* `status` - Current lifecycle status of the VM cluster.
* `status_reason` - Additional information regarding the current status of the VM cluster.
* `storage_size_in_gbs` - Local node storage allocated to the VM cluster, in gigabytes (GB).
* `system_version` - Operating system version of the image chosen for the VM cluster.
* `tags_all` - Combined set of user-defined and provider-defined tags.
* `vip_ids` - Virtual IP (VIP) addresses assigned to the VM cluster. CRS assigns one VIP per node for failover support.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `24h`)
* `update` - (Default `24h`)
* `delete` - (Default `24h`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import OpenSearch Ingestion Pipeline using the `id`. For example:

```terraform
import {
  to = aws_odb_cloud_vm_cluster.example
  id = "example"
}
```

Using `terraform import`, import cloud vm cluster using the `id`. For example:

```console
% terraform import aws_odb_cloud_vm_cluster.example example
```
