---
subcategory: "Lake Formation"
layout: "aws"
page_title: "AWS: aws_lakeformation_permissions"
description: |-
  Manages the permissions that a principal has on an AWS Glue Data Catalog resource (such as AWS Glue database or AWS Glue tables)
---

# Resource: aws_lakeformation_resource

Manages the permissions that a principal has on an AWS Glue Data Catalog resource (such as AWS Glue database or AWS Glue tables).

## Example Usage

### Granting permissions on Lake Formation resource

```hcl
data "aws_iam_role" "example" {
  name = "existing_lakeformation_role"
}

data "aws_s3_bucket" "example" {
  bucket = "existing_bucket"
}

resource "aws_lakeformation_resource" "example" {
  resource_arn            = data.aws_s3_bucket.example.arn
  use_service_linked_role = true
}

resource "aws_lakeformation_permissions" "example" {
  permissions = ["DATA_LOCATION_ACCESS"]
  principal   = data.aws_iam_role.example.arn

  location = aws_lakeformation_resource.example.resource_arn
}
```

### Granting permissions on Lake Formation catalog

```hcl
data "aws_iam_role" "example" {
  name = "existing_lakeformation_role"
}

resource "aws_lakeformation_permissions" "example" {
  permissions = ["CREATE_DATABASE"]
  principal   = data.aws_iam_role.example.arn
}
```

### Granting permissions on Lake Formation database

```hcl
data "aws_iam_role" "example" {
  name = "existing_lakeformation_role"
}

resource "aws_glue_catalog_database" "example" {
  name = "example_database"
}

resource "aws_lakeformation_permissions" "example" {
  permissions = ["ALTER", "CREATE_TABLE", "DROP"]
  principal   = data.aws_iam_role.example.arn

  database = aws_glue_catalog_database.example.name
}
```

### Granting permissions on Lake Formation table

```hcl
data "aws_iam_role" "example" {
  name = "existing_lakeformation_role"
}

resource "aws_glue_catalog_database" "example" {
  name = "example_database"
}

resource "aws_glue_catalog_table" "example" {
  name          = "example_table"
  database_name = aws_glue_catalog_database.example.name
}

resource "aws_lakeformation_permissions" "example" {
  permissions                   = ["INSERT", "DELETE", "SELECT"]
  permissions_with_grant_option = ["SELECT"]
  principal                     = data.aws_iam_role.example.arn

  table {
    database = aws_glue_catalog_table.example.database_name
    name     = aws_glue_catalog_table.example.name
  }
}
```

### Granting permissions on Lake Formation columns

```hcl
data "aws_iam_role" "example" {
  name = "existing_lakeformation_role"
}

resource "aws_glue_catalog_database" "example" {
  name = "example_database"
}

resource "aws_glue_catalog_table" "example" {
  name          = "example_table"
  database_name = aws_glue_catalog_database.example.name

  storage_descriptor {
    columns {
      name = "event"
      type = "string"
    }
    columns {
      name = "timestamp"
      type = "date"
    }
    columns {
      name = "value"
      type = "double"
    }
  }
}

resource "aws_lakeformation_permissions" "example" {
  permissions = [""SELECT"]
  principal   = data.aws_iam_role.example.arn

  table {
    database     = aws_glue_catalog_table.example.database_name
    name         = aws_glue_catalog_table.example.name
    column_names = ["event", "timestamp"]
  }
}
```

## Argument Reference

The following arguments are required:

* `permissions` – (Required) The permissions granted.

* `principal` – (Required) The AWS Lake Formation principal.

The following arguments are optional:

* `catalog_id` – (Optional) The identifier for the Data Catalog. By default, the account ID.

* `permissions_with_grant_option` – (Optional) Indicates whether to grant the ability to grant permissions (as a subset of permissions granted)s.

* `database` – (Optional) The name of the database resource. Unique to the Data Catalog. A database is a set of associated table definitions organized into a logical group.

* `location` – (Optional) The Amazon Resource Name (ARN) of the resource (data location).

* `table` – (Optional) A structure for the table object. A table is a metadata definition that represents your data.

Only one of `database`, `location`, `table` can be specified at a time. If none of them is specified, permissions will be set at catalog level. See bellow for available permissions for each resource.

The `table` object supports the following:

* `database` – (Required) The name of the database for the table.

* `table` – (Required) The name of the table.

* `column_names` - (Optional) The list of column names for the table.

* `excluded_column_names` - (Optional) Excludes column names. Any column with this name will be excluded.

The following summarizes the available Lake Formation permissions on Data Catalog resources:

* `DATA_LOCATION_ACCESS` on registered location resources,

* `CREATE_DATABASE` on catalog,

* `CREATE_TABLE`, `ALTER`, `DROP` on databases,

* `ALTER`, `INSERT`, `DELETE`, `DROP`, `SELECT` on tables,

* `SELECT` on columns.

`INSERT`, `DELETE`, `SELECT` permissions apply to the underlying data, the others to the metadata.

There is also a special permission `ALL`, that enables a principal to perform every supported Lake Formation operation on the database or table on which it is granted.

For details on each permission, see [Lake Formation Permissions Reference](https://docs.aws.amazon.com/lake-formation/latest/dg/lf-permissions-reference.html).

~> **NOTE:** Data lake administrators and database creators have implicit Lake Formation permissions. See [Implicit Lake Formation Permissions](https://docs.aws.amazon.com/lake-formation/latest/dg/implicit-permissions.html) for more information.

