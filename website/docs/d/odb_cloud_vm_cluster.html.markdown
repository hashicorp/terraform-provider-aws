---
subcategory: "Oracle Database@AWS"
layout: "AWS: aws_odb_cloud_vm_cluster"
page_title: "AWS: aws_odb_cloud_vm_cluster"
description: |-
  Terraform data source for managing cloud vm cluster resource in AWS for Oracle Database@AWS.
---

# Data Source: aws_odb_cloud_vm_cluster

Terraform data source for cloud vm cluster in AWS for Oracle Database@AWS.

You can find out more about Oracle Database@AWS from [User Guide](https://docs.aws.amazon.com/odb/latest/UserGuide/what-is-odb.html).

## Example Usage

### Basic Usage

```terraform
data "aws_odb_cloud_vm_cluster" "example" {
  id = "example-id"
}
```

## Argument Reference

The following arguments are required:

* `id` - (Required) Unique identifier of the cloud vm cluster.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - Amazon Resource Name (ARN) for the cloud vm cluster.
* `cloud_exadata_infrastructure_arn` - ARN of the Cloud Exadata Infrastructure.
* `cloud_exadata_infrastructure_id` - ID of the Cloud Exadata Infrastructure.
* `cluster_name` - Name of the Grid Infrastructure (GI) cluster.
* `compute_model` - OCI model compute model used when you create or clone an instance: ECPU or OCPU. An ECPU is an abstracted measure of compute resources. ECPUs are based on the number of cores elastically allocated from a pool of compute and storage servers. An OCPU is a legacy physical measure of compute resources. OCPUs are based on the physical core of a processor with hyper-threading enabled.
* `cpu_core_count` - Number of CPU cores enabled on the VM cluster.
* `created_at` - Time when the VM cluster was created.
* `data_collection_options` - Set of diagnostic collection options enabled for the VM cluster.
* `data_storage_size_in_tbs` - Size of the data disk group, in terabytes (TB), that's allocated for the VM cluster.
* `db_node_storage_size_in_gbs` - Amount of local node storage, in gigabytes (GB), that's allocated for the VM cluster.
* `db_servers` - List of database servers for the VM cluster.
* `disk_redundancy` - Type of redundancy configured for the VM cluster. NORMAL is 2-way redundancy. HIGH is 3-way redundancy.
* `display_name` - Display name of the VM cluster.
* `domain` - Domain name of the VM cluster.
* `gi_version` - Software version of the Oracle Grid Infrastructure (GI) for the VM cluster.
* `hostname_prefix_computed` - Computed hostname prefix for the VM cluster.
* `iorm_config_cache` - ExadataIormConfig cache details for the VM cluster.
* `is_local_backup_enabled` - Whether database backups to local Exadata storage is enabled for the VM cluster.
* `is_sparse_disk_group_enabled` - Whether the VM cluster is configured with a sparse disk group.
* `last_update_history_entry_id` - Oracle Cloud ID (OCID) of the last maintenance update history entry.
* `license_model` - Oracle license model applied to the VM cluster.
* `listener_port` - Port number configured for the listener on the VM cluster.
* `memory_size_in_gbs` - Amount of memory, in gigabytes (GB), that's allocated for the VM cluster.
* `node_count` - Number of nodes in the VM cluster.
* `oci_resource_anchor_name` - Name of the OCI Resource Anchor.
* `oci_url` - HTTPS link to the VM cluster in OCI.
* `ocid` - OCID of the VM cluster.
* `odb_network_arn` - ARN of the ODB network.
* `odb_network_id` - ID of the ODB network.
* `percent_progress` - Amount of progress made on the current operation on the VM cluster, expressed as a percentage.
* `scan_dns_name` - FQDN of the DNS record for the Single Client Access Name (SCAN) IP addresses that are associated with the VM cluster.
* `scan_dns_record_id` - OCID of the DNS record for the SCAN IP addresses that are associated with the VM cluster.
* `scan_ip_ids` - OCID of the SCAN IP addresses that are associated with the VM cluster.
* `shape` - Hardware model name of the Exadata infrastructure that's running the VM cluster.
* `ssh_public_keys` - Public key portion of one or more key pairs used for SSH access to the VM cluster.
* `status` - Status of the VM cluster.
* `status_reason` - Additional information about the status of the VM cluster.
* `storage_size_in_gbs` - Amount of local node storage, in gigabytes (GB), that's allocated to the VM cluster.
* `system_version` - Operating system version of the image chosen for the VM cluster.
* `tags` - Map of tags assigned to the resource.
* `timezone` - Time zone of the VM cluster.
* `vip_ids` - Virtual IP (VIP) addresses that are associated with the VM cluster. Oracle's Cluster Ready Services (CRS) creates and maintains one VIP address for each node in the VM cluster to enable failover. If one node fails, the VIP is reassigned to another active node in the cluster.
