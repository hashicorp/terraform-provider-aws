---
subcategory: "Lake Formation"
layout: "aws"
page_title: "AWS: aws_lakeformation_opt_in"
description: |-
  Terraform resource for managing an AWS Lake Formation Opt In.
---

# Resource: aws_lakeformation_opt_in

Terraform resource for managing an AWS Lake Formation Opt In.

## Example Usage

### Basic Usage

```terraform
resource "aws_lakeformation_opt_in" "example" {
  principal {
    data_lake_principal_identifier = aws_iam_role.example.arn
  }

  resource_data {
    database {
      name       = aws_glue_catalog_database.example.name
      catalog_id = data.aws_caller_identity.current.account_id
    }
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `principal` - (Required) Lake Formation principal. Supported principals are IAM users or IAM roles. See [`principal` Block](#principal-block) for more details.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `resource_data` - (Required) Structure for the resource. See [`resource_data` Block](#resource_data-block) for more details.

### `principal` Block

* `data_lake_principal_identifier` - (Required) Identifier for the Lake Formation principal.

### `resource_data` Block

* `catalog` - (Optional) Identifier for the Data Catalog. By default, the account ID. The Data Catalog is the persistent metadata store. It contains database definitions, table definitions, and other control information to manage your Lake Formation environment. See [`catalog` Block](#catalog-block) for more details.
* `data_cells_filter` - (Optional) Data cell filter. See [`data_cells_filter` Block](#data_cells_filter-block) for more details.
* `data_location` - (Optional) Location of an Amazon S3 path where permissions are granted or revoked. See [`data_location` Block](#data_location-block) for more details.
* `database` - (Optional) Database for the resource. Unique to the Data Catalog. A database is a set of associated table definitions organized into a logical group. You can Grant and Revoke database permissions to a principal. See [`database` Block](#database-block) for more details.
* `lf_tag` - (Optional) LF-tag key and values attached to a resource.
* `lf_tag_expression` - (Optional) Logical expression composed of one or more LF-Tag key:value pairs. See [`lf_tag_expression` Block](#lf_tag_expression-block) for more details.
* `lf_tag_policy` - (Optional) List of LF-Tag conditions or saved LF-Tag expressions that define a resource's LF-Tag policy. See [`lf_tag_policy` Block](#lf_tag_policy-block) for more details.
* `table` - (Optional) Table for the resource. A table is a metadata definition that represents your data. You can Grant and Revoke table privileges to a principal. See [`table` Block](#table-block) for more details.
* `table_with_columns` - (Optional) Table with columns for the resource. A principal with permissions to this resource can select metadata from the columns of a table in the Data Catalog and the underlying data in Amazon S3. See [`table_with_columns` Block](#table_with_columns-block) for more details.

### `catalog` Block

* `id` - (Optional) Identifier for the catalog resource.

### `data_cells_filter` Block

* `database_name` - (Optional) Database in the Glue Data Catalog.
* `name` - (Optional) Name of the data cells filter.
* `table_catalog_id` - (Optional) ID of the catalog to which the table belongs.
* `table_name` - (Optional) Name of the table.

### `data_location` Block

* `catalog_id` - (Optional) Identifier for the Data Catalog where the location is registered with Lake Formation. By default, it is the account ID of the caller.
* `resource_arn` - (Required) ARN that uniquely identifies the data location resource.

### `database` Block

* `catalog_id` - (Optional) Identifier for the Data Catalog. By default, it is the account ID of the caller.
* `name` - (Required) Name of the database resource. Unique to the Data Catalog.

### `lf_tag` Block

* `catalog_id` - (Optional) Identifier for the Data Catalog. By default, it is the account ID of the caller.
* `key` - (Required) Key name for the LF-Tag.
* `values` - (Required) Set of tag values for the LF-Tag key. At least one value is required. Each value can be 1-255 characters.

### `lf_tag_expression` Block

* `catalog_id` - (Optional) Identifier for the Data Catalog. By default, it is the account ID of the caller.
* `name` - (Required) Name of the LF-Tag expression to grant permissions on.

### `lf_tag_policy` Block

* `catalog_id` - (Optional) Identifier for the Data Catalog. By default, it is the account ID of the caller. The Data Catalog is the persistent metadata store. It contains database definitions, table definitions, and other control information to manage your Lake Formation environment.
* `expression` - (Optional) List of LF-tag conditions or a saved expression that apply to the resource's LF-Tag policy.
* `expression_name` - (Optional) Name of the saved expression to match. If provided, permissions are granted to the Data Catalog resources whose assigned LF-Tags match the expression body of the saved expression under the provided expression name.
* `resource_type` - (Required) Resource type for which the LF-tag policy applies.

### `table` Block

* `catalog_id` - (Optional) Identifier for the Data Catalog. By default, it is the account ID of the caller.
* `database_name` - (Required) Name of the database for the table. Unique to a Data Catalog. A database is a set of associated table definitions organized into a logical group. You can Grant and Revoke database privileges to a principal.
* `name` - (Optional) Name of the table.
* `wildcard` - (Optional) Boolean value that indicates whether to use a wildcard representing every table under the specified database. When set to true, this represents all tables within the specified database. At least one of TableResource$Name or TableResource$Wildcard is required.

### `table_with_columns` Block

* `catalog_id` - (Optional) Identifier for the Data Catalog. By default, it is the account ID of the caller.
* `column_names` - (Optional) List of column names for the table. At least one of ColumnNames or ColumnWildcard is required.
* `column_wildcard` - (Optional) Wildcard specified by a ColumnWildcard object. At least one of ColumnNames or ColumnWildcard is required. See [`column_wildcard` Block](#column_wildcard-block) for more details.
* `database_name` - (Required) Name of the database for the table. Unique to a Data Catalog. A database is a set of associated table definitions organized into a logical group. You can Grant and Revoke database privileges to a principal.
* `name` - (Required) Name of the table.

### `column_wildcard` Block

* `excluded_column_names` - (Optional) Excludes column names. Any column with this name will be excluded.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `condition` - Lake Formation condition, which applies to permissions and opt-ins that contain an expression. See [`condition` Block](#condition-block) for more details.
* `last_modified` - Last modified date and time of the record.
* `last_updated_by` - User who updated the record.

### `condition` Block

* `expression` - Expression written based on the Cedar Policy Language used to match the principal attributes.
