---
subcategory: "Oracle Database@AWS"
layout: "aws"
page_title: "AWS: aws_odb_exascale_db_storage_vault"
description: |-
  Provides details about an Oracle Database@AWS Exascale DB storage vault.
---

# Data Source: aws_odb_exascale_db_storage_vault

Provides details about an [Oracle Database@AWS Exascale DB storage vault](https://docs.aws.amazon.com/odb/latest/APIReference/API_GetExascaleDbStorageVault.html).

## Example Usage

### Basic Usage

```terraform
data "aws_odb_exascale_db_storage_vault" "example" {
  id = "xsvault_0123456789"
}
```

## Argument Reference

The following arguments are required:

* `id` - (Required) Unique identifier of the Exascale DB storage vault. Must be between `6` and `2048` characters.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `additional_flash_cache_in_percent` - Additional flash cache percentage for the Exascale DB storage vault.
* `arn` - ARN of the Exascale DB storage vault.
* `attached_shape_attributes` - Shape attributes attached to the Exascale DB storage vault.
* `autoscale_limit_in_gbs` - Autoscale limit, in GB, for the Exascale DB storage vault.
* `availability_zone` - Availability Zone for the Exascale DB storage vault.
* `availability_zone_id` - Availability Zone ID for the Exascale DB storage vault.
* `created_at` - Date and time when the Exascale DB storage vault was created.
* `description` - Description of the Exascale DB storage vault.
* `display_name` - User-friendly name for the Exascale DB storage vault.
* `high_capacity_database_storage` - High-capacity database storage details for the Exascale DB storage vault. See [`high_capacity_database_storage`](#high_capacity_database_storage) below.
* `is_autoscale_enabled` - Whether autoscaling is enabled for the Exascale DB storage vault.
* `oci_resource_anchor_name` - Name of the OCI resource anchor for the Exascale DB storage vault.
* `oci_url` - HTTPS URL for the Exascale DB storage vault in OCI.
* `ocid` - OCID of the Exascale DB storage vault.
* `percent_progress` - Progress of the current operation on the Exascale DB storage vault, expressed as a percentage.
* `status` - Current status of the Exascale DB storage vault.
* `status_reason` - Additional information about the status of the Exascale DB storage vault.
* `tags` - Map of tags assigned to the resource.
* `time_zone` - Time zone of the Exascale DB storage vault.
* `vm_cluster_arns` - ARNs of the VM clusters associated with the Exascale DB storage vault.
* `vm_cluster_count` - Number of VM clusters associated with the Exascale DB storage vault.
* `vm_cluster_ids` - Unique identifiers of the VM clusters associated with the Exascale DB storage vault.

### `high_capacity_database_storage` Block

* `available_size_in_gbs` - Available storage size, in GB.
* `total_size_in_gbs` - Total storage size, in GB.
