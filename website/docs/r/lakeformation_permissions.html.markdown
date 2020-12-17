---
subcategory: "Lake Formation"
layout: "aws"
page_title: "AWS: aws_lakeformation_permissions"
description: |-
    Grants permissions to the principal to access metadata in the Data Catalog and data organized in underlying data storage such as Amazon S3.
---

# Resource: aws_lakeformation_permissions

Grants permissions to the principal to access metadata in the Data Catalog and data organized in underlying data storage such as Amazon S3. Permissions are granted to a principal, in a Data Catalog, relative to a Lake Formation resource, which includes the Data Catalog, databases, and tables. For more information, see [Security and Access Control to Metadata and Data in Lake Formation](https://docs.aws.amazon.com/lake-formation/latest/dg/security-data-access.html).

~> **NOTE:** This resource deals with explicitly granted permissions. Lake Formation grants implicit permissions to data lake administrators, database creators, and table creators. For more information, see [Implicit Lake Formation Permissions](https://docs.aws.amazon.com/lake-formation/latest/dg/implicit-permissions.html).

## Example Usage

### Grant Permissions For A Lake Formation S3 Resource

```hcl
resource "aws_lakeformation_permissions" "test" {
  principal_arn = aws_iam_role.workflow_role.arn
  permissions   = ["ALL"]

  data_location {
    resource_arn = aws_lakeformation_resource.test.resource_arn
  }
}
```

### Grant Permissions For A Glue Catalog Database

```hcl
resource "aws_lakeformation_permissions" "test" {
  role        = aws_iam_role.workflow_role.arn
  permissions = ["CREATE_TABLE", "ALTER", "DROP"]

  database {
    name       = aws_glue_catalog_database.test.name
    catalog_id = "110376042874"
  }
}
```

## Argument Reference

The following arguments are required:

* `permissions` – (Required) List of permissions granted to the principal. Valid values include `ALL`, `ALTER`, `CREATE_DATABASE`, `CREATE_TABLE`, `DATA_LOCATION_ACCESS`, `DELETE`, `DESCRIBE`, `DROP`, `INSERT`, and `SELECT`. For details on each permission, see [Lake Formation Permissions Reference](https://docs.aws.amazon.com/lake-formation/latest/dg/lf-permissions-reference.html).
* `principal_arn` – (Required) Principal to be granted the permissions on the resource. Supported principals are IAM users or IAM roles.

The following arguments are optional:

* `catalog_id` – (Optional) Identifier for the Data Catalog. By default, the account ID. The Data Catalog is the persistent metadata store. It contains database definitions, table definitions, and other control information to manage your Lake Formation environment.

One of the following is required:

* `catalog_resource` - (Optional) Whether the permissions are to be granted for the Data Catalog. Defaults to `false`.
* `data_location` - (Optional) Configuration block for a data location resource. Detailed below.
* `database` - (Optional) Configuration block for a database resource. Detailed below.
* `table` - (Optional) Configuration block for a table resource. Detailed below.
* `table_with_columns` - (Optional) Configuration block for a table with columns resource. Detailed below.

### data_location

The following argument is required:

* `resource_arn` – (Required) Amazon Resource Name (ARN) that uniquely identifies the data location resource.

The following argument is optional:

* `catalog_id` - (Optional) Identifier for the Data Catalog where the location is registered with Lake Formation. By default, it is the account ID of the caller.

### database

The following argument is required:

* `name` – (Required) Name of the database resource. Unique to the Data Catalog.

The following argument is optional:

* `catalog_id` - (Optional) Identifier for the Data Catalog. By default, it is the account ID of the caller.

### table

The following argument is required:

* `database_name` – (Required) Name of the database for the table. Unique to a Data Catalog.

The following arguments are optional:

* `catalog_id` - (Optional) Identifier for the Data Catalog. By default, it is the account ID of the caller.
* `name` - (Optional) Name of the table. Not including the table name results in a wildcard representing every table under a database.

### table_with_columns

The following arguments are required:

* `database_name` – (Required) Name of the database for the table with columns resource. Unique to the Data Catalog.
* `name` – (Required) Name of the table resource.

The following arguments are optional:

* `catalog_id` - (Optional) Identifier for the Data Catalog. By default, it is the account ID of the caller.
* `column_names` - (Optional) List of column names for the table. At least one of `column_names` or `excluded_column_names` is required.
* `excluded_column_names` - (Optional) List of column names for the table to exclude. At least one of `column_names` or `excluded_column_names` is required.

## Attributes Reference

In addition to the above arguments, no attributes are exported.
