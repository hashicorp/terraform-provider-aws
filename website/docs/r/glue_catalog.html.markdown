---
subcategory: "Glue"
layout: "aws"
page_title: "AWS: aws_glue_catalog"
description: |-
  Manages an AWS Glue Catalog.
---

# Resource: aws_glue_catalog

Manages an AWS Glue Catalog.

## Example Usage

### Basic Usage

```terraform
resource "aws_glue_catalog" "example" {
  name        = "example"
  description = "Example Glue Catalog"
}
```

### With Parameters

```terraform
resource "aws_glue_catalog" "example" {
  name        = "example"
  description = "Example Glue Catalog"

  parameters = {
    "key1" = "value1"
    "key2" = "value2"
  }
}
```

### With Catalog Properties

```terraform
resource "aws_glue_catalog" "example" {
  name        = "example"
  description = "Example Glue Catalog with data lake access"

  catalog_properties {
    custom_properties = {
      "property1" = "value1"
    }

    data_lake_access_properties {
      data_lake_access = true
      catalog_type     = "aws:glue:datacatalog"
    }
  }
}
```

### With Federated Catalog

```terraform
resource "aws_glue_catalog" "example" {
  name = "example"

  federated_catalog {
    connection_name = aws_glue_connection.example.name
    identifier      = "arn:aws:glue:us-east-1:123456789012:catalog"
  }
}
```

### With Default Permissions

```terraform
resource "aws_glue_catalog" "example" {
  name        = "example"
  description = "Example Glue Catalog"

  create_database_default_permissions {
    permissions = ["ALL"]

    principal {
      data_lake_principal_identifier = "IAM_ALLOWED_PRINCIPALS"
    }
  }

  create_table_default_permissions {
    permissions = ["ALL"]

    principal {
      data_lake_principal_identifier = "IAM_ALLOWED_PRINCIPALS"
    }
  }
}
```

## Argument Reference

The following arguments are optional:

* `allow_full_table_external_data_access` - (Optional) Whether third-party engines can access data in Amazon S3 locations that are registered with Lake Formation. Valid values are `True` and `False`.
* `catalog_properties` - (Optional) Configuration block of properties for the catalog. See [`catalog_properties`](#catalog_properties) below.
* `create_database_default_permissions` - (Optional) List of default permissions on databases for principals. See [`create_database_default_permissions`](#create_database_default_permissions) below.
* `create_table_default_permissions` - (Optional) List of default permissions on tables for principals. See [`create_table_default_permissions`](#create_table_default_permissions) below.
* `description` - (Optional) Description of the catalog.
* `federated_catalog` - (Optional) Configuration block for a federated catalog. See [`federated_catalog`](#federated_catalog) below.
* `name` - (Required, Forces new resource) Name of the catalog.
* `overwrite_child_resource_permissions_with_default` - (Optional) Whether to overwrite existing Lake Formation permissions on child resources with the default permissions. Valid values are `Accept` and `Deny`.
* `parameters` - (Optional) Map of key-value pairs that define parameters and properties of the catalog.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `target_redshift_catalog` - (Optional) Configuration block for a target Redshift catalog. See [`target_redshift_catalog`](#target_redshift_catalog) below.

### catalog_properties

* `custom_properties` - (Optional) Map of custom key-value pairs for the catalog properties.
* `data_lake_access_properties` - (Optional) Configuration block for data lake access properties. See [`data_lake_access_properties`](#data_lake_access_properties) below.
* `iceberg_optimization_properties` - (Optional) Configuration block for Iceberg optimization properties. See [`iceberg_optimization_properties`](#iceberg_optimization_properties) below.

#### data_lake_access_properties

* `catalog_type` - (Optional) Type of the catalog.
* `data_lake_access` - (Optional) Whether data lake access is enabled.
* `data_transfer_role` - (Optional) ARN of the IAM role used for data transfer.
* `kms_key` - (Optional) ARN of the KMS key used for encryption.

#### iceberg_optimization_properties

* `compaction` - (Optional) Map of key-value pairs for compaction settings.
* `orphan_file_deletion` - (Optional) Map of key-value pairs for orphan file deletion settings.
* `retention` - (Optional) Map of key-value pairs for retention settings.
* `role_arn` - (Optional) ARN of the IAM role for Iceberg optimization.

### federated_catalog

* `connection_name` - (Optional) Name of the connection to the external metastore.
* `connection_type` - (Optional) Type of connection used to access the federated catalog.
* `identifier` - (Optional) Unique identifier for the federated catalog.

### target_redshift_catalog

* `catalog_arn` - (Required) ARN of the target Redshift catalog.

### create_database_default_permissions

* `permissions` - (Optional) Permissions that are granted to the principal. Valid values include `ALL`, `SELECT`, `ALTER`, `DROP`, `DELETE`, `INSERT`, `CREATE_DATABASE`, `CREATE_TABLE`, `DATA_LOCATION_ACCESS`.
* `principal` - (Optional) Principal who is granted permissions. See [`principal`](#principal) below.

### create_table_default_permissions

* `permissions` - (Optional) Permissions that are granted to the principal. Valid values include `ALL`, `SELECT`, `ALTER`, `DROP`, `DELETE`, `INSERT`, `CREATE_DATABASE`, `CREATE_TABLE`, `DATA_LOCATION_ACCESS`.
* `principal` - (Optional) Principal who is granted permissions. See [`principal`](#principal) below.

#### principal

* `data_lake_principal_identifier` - (Optional) Identifier for the Lake Formation principal.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Glue Catalog.
* `catalog_id` - ID of the parent catalog.
* `create_time` - Time at which the catalog was created.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `update_time` - Time at which the catalog was last updated.

The `catalog_properties[0].data_lake_access_properties[0]` block also exports:

* `managed_workgroup_name` - Managed workgroup name.
* `managed_workgroup_status` - Managed workgroup status.
* `redshift_database_name` - Redshift database name.
* `status_message` - Status message.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_glue_catalog.example
  identity = {
    name = "example"
  }
}

resource "aws_glue_catalog" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `name` (String) Name of the Glue Catalog.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Glue Catalog using the catalog name. For example:

```terraform
import {
  to = aws_glue_catalog.example
  id = "example"
}
```

Using `terraform import`, import Glue Catalog using the catalog name. For example:

```console
% terraform import aws_glue_catalog.example example
```
