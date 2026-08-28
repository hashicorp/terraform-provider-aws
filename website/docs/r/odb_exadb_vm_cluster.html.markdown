---
subcategory: "Oracle Database@AWS"
layout: "aws"
page_title: "AWS: aws_odb_exadb_vm_cluster"
description: |-
  Manages an Oracle Database@AWS ExaDB VM Cluster.
---

# Resource: aws_odb_exadb_vm_cluster

Manages an Oracle Database@AWS ExaDB VM Cluster.

~> **NOTE:** The resource schema and lifecycle implementation are under development.

## Example Usage

```terraform
resource "aws_odb_exadb_vm_cluster" "example" {
  display_name                             = "example_exadb_vm_cluster"
  enabled_ecpu_count                       = 4
  exascale_db_storage_vault_id             = "<exascale-db-storage-vault-id>"
  grid_image_id                            = "<grid-image-id>"
  hostname                                 = "exadbvm1"
  node_count                               = 2
  odb_network_id                           = "<odb-network-id>"
  shape                                    = "<shape>"
  ssh_public_keys                          = ["ssh-rsa AAAA..."]
  total_ecpu_count                         = 4
  vm_file_system_storage_total_size_in_gbs = 100

  data_collection_options {
    is_diagnostics_events_enabled = true
    is_health_monitoring_enabled  = true
    is_incident_logs_enabled      = true
  }

  tags = {
    Name = "example"
  }
}
```

## Argument Reference

The following arguments are required:

* `display_name` - (Required) User-friendly name for the ExaDB VM Cluster. Length must be between `1` and `255` characters. Must start with a letter or underscore and contain only letters, numbers, underscores, and hyphens.
* `enabled_ecpu_count` - (Required) Number of ECPUs enabled for the ExaDB VM Cluster. Value must be at least `0`.
* `exascale_db_storage_vault_id` - (Required) ID of the Exascale DB Storage Vault for the ExaDB VM Cluster. Length must be between `6` and `2048` characters. Changing this value creates a new resource.
* `grid_image_id` - (Required) Grid Infrastructure software image ID for the ExaDB VM Cluster. Length must be between `1` and `255` characters.
* `hostname` - (Required) Host name for the ExaDB VM Cluster. Length must be between `1` and `12` characters. Must start with a letter, end with a letter or number, and contain only letters, numbers, and hyphens. Changing this value creates a new resource.
* `node_count` - (Required) Number of nodes in the ExaDB VM Cluster. Value must be at least `1`. Changing this value creates a new resource.
* `odb_network_id` - (Required) ID of the ODB network for the ExaDB VM Cluster. Length must be between `6` and `2048` characters. Changing this value creates a new resource.
* `shape` - (Required) Shape of the ExaDB VM Cluster. Length must be between `1` and `255` characters. Changing this value creates a new resource.
* `ssh_public_keys` - (Required) Public keys used for SSH access to the ExaDB VM Cluster. Must contain between `1` and `1024` elements.
* `total_ecpu_count` - (Required) Total number of ECPUs for the ExaDB VM Cluster. Value must be at least `2`.
* `vm_file_system_storage_total_size_in_gbs` - (Required) Total amount of VM file system storage for the ExaDB VM Cluster, in GB. Value must be at least `0`.

The following arguments are optional:

* `cluster_name` - (Optional) Name of the Grid Infrastructure cluster. Length must be between `1` and `11` characters. Must start with a letter and contain only letters, numbers, and hyphens. Changing this value creates a new resource.
* `data_collection_options` - (Optional) Diagnostic collection preferences for the ExaDB VM Cluster. Configure at most one block. See [`data_collection_options` Block](#data_collection_options-block) below.
* `license_model` - (Optional) Oracle license model applied to the ExaDB VM Cluster. Valid values are `BRING_YOUR_OWN_LICENSE` and `LICENSE_INCLUDED`.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `scan_listener_port_tcp` - (Optional) Port for TCP connections to the SCAN listener. Valid values are from `1024` through `8999`. Changing this value creates a new resource.
* `scan_listener_port_tcp_ssl` - (Optional) Port for SSL/TCP connections to the SCAN listener. Valid values are from `1024` through `8999`. Changing this value creates a new resource.
* `shape_attribute` - (Optional) Shape attribute for the ExaDB VM Cluster. Valid values are `SMART_STORAGE` and `BLOCK_STORAGE`. Changing this value creates a new resource.
* `system_version` - (Optional) Operating system version of the image for the ExaDB VM Cluster. Length must be between `1` and `255` characters.
* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider level.
* `time_zone` - (Optional) Time zone for the ExaDB VM Cluster. Length must be between `1` and `255` characters. Changing this value creates a new resource.

### `data_collection_options` Block

The `data_collection_options` block supports the following:

* `is_diagnostics_events_enabled` - (Optional) Whether diagnostic event collection is enabled.
* `is_health_monitoring_enabled` - (Optional) Whether health monitoring is enabled.
* `is_incident_logs_enabled` - (Optional) Whether incident log collection is enabled.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the ExaDB VM Cluster.
* `created_at` - Date and time when the ExaDB VM Cluster was created.
* `domain` - Domain of the ExaDB VM Cluster.
* `exascale_db_storage_vault_arn` - ARN of the Exascale DB Storage Vault associated with the ExaDB VM Cluster.
* `gi_version` - Oracle Grid Infrastructure software version for the ExaDB VM Cluster.
* `grid_image_type` - Type of Grid Infrastructure image used by the ExaDB VM Cluster.
* `iam_roles` - IAM service roles associated with the ExaDB VM Cluster. See [`iam_roles` Block](#iam_roles-block) below.
* `id` - ID of the ExaDB VM Cluster.
* `iorm_config_cache` - IORM configuration cache details for the ExaDB VM Cluster. See [`iorm_config_cache` Block](#iorm_config_cache-block) below.
* `last_update_history_entry_id` - OCID of the last maintenance update history entry.
* `listener_port` - Listener port configured for the ExaDB VM Cluster.
* `memory_size_in_gbs` - Amount of memory allocated to the ExaDB VM Cluster, in GB.
* `ocid` - OCID of the ExaDB VM Cluster.
* `oci_resource_anchor_name` - Name of the OCI resource anchor for the ExaDB VM Cluster.
* `oci_url` - HTTPS URL of the ExaDB VM Cluster in OCI.
* `odb_network_arn` - ARN of the ODB network associated with the ExaDB VM Cluster.
* `percent_progress` - Progress of the current operation, expressed as a percentage.
* `scan_dns_name` - FQDN of the SCAN IP addresses associated with the ExaDB VM Cluster.
* `scan_dns_record_id` - OCID of the DNS record for the SCAN IP addresses.
* `scan_ip_ids` - OCIDs of the SCAN IP addresses associated with the ExaDB VM Cluster.
* `snapshot_file_system_storage` - Snapshot file system storage details for the ExaDB VM Cluster. See [`storage_details` Block](#storage_details-block) below.
* `status` - Current status of the ExaDB VM Cluster.
* `status_reason` - Additional information about the current ExaDB VM Cluster status.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `total_file_system_storage` - Total file system storage details for the ExaDB VM Cluster. See [`storage_details` Block](#storage_details-block) below.
* `vip_ids` - OCIDs of the virtual IP addresses associated with the ExaDB VM Cluster.
* `vm_file_system_storage` - VM file system storage details for the ExaDB VM Cluster. See [`storage_details` Block](#storage_details-block) below.

### `iam_roles` Block

The `iam_roles` block exports the following attributes:

* `aws_integration` - AWS integration supported by the IAM role.
* `iam_role_arn` - ARN of the IAM role.
* `status` - Status of the IAM role association.
* `status_reason` - Additional information about the IAM role association status.

### `iorm_config_cache` Block

The `iorm_config_cache` block exports the following attributes:

* `db_plans` - IORM database plans for the ExaDB VM Cluster. See [`db_plans` Block](#db_plans-block) below.
* `lifecycle_details` - Additional information about the current IORM configuration lifecycle state.
* `lifecycle_state` - Current lifecycle state of the IORM configuration.
* `objective` - Current IORM objective.

### `db_plans` Block

The `db_plans` block exports the following attributes:

* `db_name` - Database name to which the IORM plan applies.
* `flash_cache_limit` - Flash cache limit for the database plan.
* `share` - Relative priority of the database in the IORM plan.

### `storage_details` Block

The `snapshot_file_system_storage`, `total_file_system_storage`, and `vm_file_system_storage` blocks export the following attribute:

* `total_size_in_gbs` - Total storage size, in GB.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `24h`)
* `update` - (Default `24h`)
* `delete` - (Default `24h`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_odb_exadb_vm_cluster.example
  identity = {
    id = "example"
  }
}
```

### Identity Schema

#### Required

* `id` (String) ID of the ExaDB VM Cluster.

#### Optional

* `account_id` (String) AWS account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import an ExaDB VM Cluster using its ID. For example:

```terraform
import {
  to = aws_odb_exadb_vm_cluster.example
  id = "example"
}
```

Using `terraform import`, import an ExaDB VM Cluster using its ID. For example:

```console
% terraform import aws_odb_exadb_vm_cluster.example example
```
