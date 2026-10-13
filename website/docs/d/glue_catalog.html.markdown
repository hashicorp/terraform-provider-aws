---
subcategory: "Glue"
layout: "aws"
page_title: "AWS: aws_glue_catalog"
description: |-
  Provides details about an AWS Glue Catalog.
---

# Data Source: aws_glue_catalog

Provides details about an AWS Glue Catalog.

## Example Usage

```terraform
data "aws_glue_catalog" "example" {
  name = "example"
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the catalog to look up.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `allow_full_table_external_data_access` - Whether third-party engines can access data in Amazon S3 locations that are registered with Lake Formation.
* `arn` - ARN of the Glue Catalog.
* `catalog_id` - ID of the parent catalog.
* `catalog_properties` - Catalog properties. See [`catalog_properties` Block](#catalog_properties-block) below.
* `create_database_default_permissions` - Default permissions on databases for principals. See [`create_database_default_permissions` Block](#create_database_default_permissions-block) below.
* `create_table_default_permissions` - Default permissions on tables for principals. See [`create_table_default_permissions` Block](#create_table_default_permissions-block) below.
* `create_time` - Time at which the catalog was created.
* `description` - Description of the catalog.
* `federated_catalog` - Federated catalog configuration. See [`federated_catalog` Block](#federated_catalog-block) below.
* `parameters` - Map of key-value pairs that define parameters and properties of the catalog.
* `tags` - Key-value map of resource tags.
* `target_redshift_catalog` - Target Redshift catalog configuration. See [`target_redshift_catalog` Block](#target_redshift_catalog-block) below.
* `update_time` - Time at which the catalog was last updated.

### `catalog_properties` Block

* `custom_properties` - Map of custom key-value pairs for the catalog properties.
* `data_lake_access_properties` - Data lake access properties. See [`data_lake_access_properties` Block](#data_lake_access_properties-block) below.
* `iceberg_optimization_properties` - Iceberg optimization properties. See [`iceberg_optimization_properties` Block](#iceberg_optimization_properties-block) below.

#### `data_lake_access_properties` Block

* `catalog_type` - Type of the catalog.
* `data_lake_access` - Whether data lake access is enabled.
* `data_transfer_role` - ARN of the IAM role used for data transfer.
* `kms_key` - ARN of the KMS key used for encryption.
* `managed_workgroup_name` - Managed workgroup name.
* `managed_workgroup_status` - Managed workgroup status.
* `redshift_database_name` - Redshift database name.
* `status_message` - Status message.

#### `iceberg_optimization_properties` Block

* `compaction` - Map of key-value pairs for compaction settings.
* `orphan_file_deletion` - Map of key-value pairs for orphan file deletion settings.
* `retention` - Map of key-value pairs for retention settings.
* `role_arn` - ARN of the IAM role for Iceberg optimization.

### `federated_catalog` Block

* `connection_name` - Name of the connection to the external metastore.
* `connection_type` - Type of connection used to access the federated catalog.
* `identifier` - Unique identifier for the federated catalog.

### `target_redshift_catalog` Block

* `catalog_arn` - ARN of the target Redshift catalog.

### `create_database_default_permissions` Block

* `permissions` - Permissions that are granted to the principal.
* `principal` - Principal who is granted permissions. See [`principal` Block](#principal-block) below.

### `create_table_default_permissions` Block

* `permissions` - Permissions that are granted to the principal.
* `principal` - Principal who is granted permissions. See [`principal` Block](#principal-block) below.

#### `principal` Block

* `data_lake_principal_identifier` - Identifier for the Lake Formation principal.
