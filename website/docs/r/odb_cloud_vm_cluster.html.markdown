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

## Argument Reference

The following arguments are required:

* `cpu_core_count` - (Required) The number of CPU cores to enable on the VM cluster. Changing this will create a new resource.
* `db_servers` - (Required) The list of database servers for the VM cluster. Changing this will create a new resource.
* `display_name` - (Required) A user-friendly name for the VM cluster. Changing this will create a new resource.
* `gi_version` - (Required) A valid software version of Oracle Grid Infrastructure (GI). To get the list of valid values, use the ListGiVersions operation and specify the shape of the Exadata infrastructure. Example: 19.0.0.0 Changing this will create a new resource.
* `hostname_prefix` - (Required) The host name prefix for the VM cluster. Constraints: - Can't be "localhost" or "hostname". - Can't contain "-version". - The maximum length of the combined hostname and domain is 63 characters. - The hostname must be unique within the subnet. Changing this will create a new resource.
* `ssh_public_keys` - (Required) The public key portion of one or more key pairs used for SSH access to the VM cluster. Changing this will create a new resource.
* `data_collection_options` - (Required) The set of preferences for the various diagnostic collection options for the VM cluster.
* `data_storage_size_in_tbs` - (Required) The size of the data disk group, in terabytes (TBs), to allocate for the VM cluster. Changing this will create a new resource.

The following arguments are optional:

* `odb_network_id` - (Optional) The unique identifier of the ODB network for the VM cluster. Changing this will create a new resource. Either the combination of cloud_exadata_infrastructure_id and odb_network_id or cloud_exadata_infrastructure_arn and odb_network_arn must be used.
* `cloud_exadata_infrastructure_id` - (Optional) The unique identifier of the Exadata infrastructure for this VM cluster. Changing this will create a new resource. Either the combination of cloud_exadata_infrastructure_id and odb_network_id or cloud_exadata_infrastructure_arn and odb_network_arn must be used.
* `odb_network_arn` - (Optional) The ARN of the ODB network for the VM cluster. Changing this will create a new resource. Either the combination of cloud_exadata_infrastructure_id and odb_network_id or cloud_exadata_infrastructure_arn and odb_network_arn must be used.
* `cloud_exadata_infrastructure_arn` - (Optional) The ARN of the Exadata infrastructure for this VM cluster. Changing this will create a new resource. Either the combination of cloud_exadata_infrastructure_id and odb_network_id or cloud_exadata_infrastructure_arn and odb_network_arn must be used.
* `cluster_name` - (Optional) The name of the Grid Infrastructure (GI) cluster. Changing this will create a new resource.
* `db_node_storage_size_in_gbs` - (Optional) The amount of local node storage, in gigabytes (GBs), to allocate for the VM cluster. Changing this will create a new resource.
* `is_local_backup_enabled` - (Optional) Specifies whether to enable database backups to local Exadata storage for the VM cluster. Changing this will create a new resource.
* `is_sparse_diskgroup_enabled` - (Optional) Specifies whether to create a sparse disk group for the VM cluster. Changing this will create a new resource.
* `license_model` - (Optional) The Oracle license model to apply to the VM cluster. Default: LICENSE_INCLUDED. Changing this will create a new resource.
* `memory_size_in_gbs` - (Optional) The amount of memory, in gigabytes (GBs), to allocate for the VM cluster. Changing this will create a new resource.
* `scan_listener_port_tcp` - (Optional) The port number for TCP connections to the single client access name (SCAN) listener. Valid values: 1024–8999, except 2484, 6100, 6200, 7060, 7070, 7085, and 7879. Default: 1521. Changing this will create a new resource.
* `timezone` - (Optional) The configured time zone of the VM cluster. Changing this will create a new resource.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) A map of tags to assign to the exadata infrastructure. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - Unique identifier of vm cluster.
* `arn` - The Amazon Resource Name (ARN) for the cloud vm cluster.
* `disk_redundancy` - The type of redundancy for the VM cluster: NORMAL (2-way) or HIGH (3-way).
* `AttrDomain` - The domain name associated with the VM cluster.
* `hostname_prefix_computed` - The host name for the VM cluster. Constraints: - Can't be "localhost" or "hostname". - Can't contain "-version". - The maximum length of the combined hostname and domain is 63 characters. - The hostname must be unique within the subnet. This member is required. Changing this will create a new resource.
* `iorm_config_cache` - The Exadata IORM (I/O Resource Manager) configuration cache details for the VM cluster.
* `last_update_history_entry_id` - The OCID of the most recent maintenance update history entry.
* `listener_port` - The listener port number configured on the VM cluster.
* `node_count` - The total number of nodes in the VM cluster.
* `ocid` - The OCID (Oracle Cloud Identifier) of the VM cluster.
* `oci_resource_anchor_name` - The name of the OCI resource anchor associated with the VM cluster.
* `oci_url` - The HTTPS link to the VM cluster resource in OCI.
* `percent_progress` - The percentage of progress made on the current operation for the VM cluster.
* `scan_dns_name` - The fully qualified domain name (FQDN) for the SCAN IP addresses associated with the VM cluster.
* `scan_dns_record_id` - The OCID of the DNS record for the SCAN IPs linked to the VM cluster.
* `scan_ip_ids` - The list of OCIDs for SCAN IP addresses associated with the VM cluster.
* `shape` - The hardware model name of the Exadata infrastructure running the VM cluster.
* `status` - The current lifecycle status of the VM cluster.
* `status_reason` - Additional information regarding the current status of the VM cluster.
* `storage_size_in_gbs` - The local node storage allocated to the VM cluster, in gigabytes (GB).
* `system_version` - The operating system version of the image chosen for the VM cluster.
* `vip_ids` - The virtual IP (VIP) addresses assigned to the VM cluster. CRS assigns one VIP per node for failover support.
* `created_at` - The timestamp when the VM cluster was created.
* `gi_version_computed` - A complete software version of Oracle Grid Infrastructure (GI).
* `compute_model` - The compute model used when the instance is created or cloned — either ECPU or OCPU. ECPU is a virtualized compute unit; OCPU is a physical processor core with hyper-threading.
* `tags_all` - The combined set of user-defined and provider-defined tags.

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
