---
subcategory: "Lake Formation"
layout: "aws"
page_title: "AWS: aws_lakeformation_resource_lf_tags"
description: |-
    Manages an attachment between one or more LF-tags and an existing Lake Formation resource.
---

# Resource: aws_lakeformation_resource_lf_tags

Manages an attachment between one or more existing LF-tags and an existing Lake Formation resource.

## Example Usage

### Database Example

```terraform
resource "aws_lakeformation_lf_tag" "example" {
  key    = "copse"
  values = ["luffield"]
}

resource "aws_lakeformation_resource_lf_tags" "example" {
  catalog_id = data.aws_caller_identity.current.account_id

  database {
    name = aws_glue_catalog_database.example.name
  }

  lf_tag {
    key    = aws_lakeformation_lf_tag.example.key
    values = aws_lakeformation_lf_tag.example.values
  }
}
```

### Multiple Tags Example

```terraform
resource "aws_lakeformation_lf_tag" "example" {
  key    = "copse"
  values = ["luffield"]
}

resource "aws_lakeformation_lf_tag" "example2" {
  key    = "stowe"
  values = ["brooklands"]
}

resource "aws_lakeformation_resource_lf_tags" "example" {
  catalog_id = data.aws_caller_identity.current.account_id

  database {
    name = aws_glue_catalog_database.example.name
  }

  lf_tag {
    key    = aws_lakeformation_lf_tag.example.key
    values = aws_lakeformation_lf_tag.example.values
  }

  lf_tag {
    key    = aws_lakeformation_lf_tag.example2.key
    values = aws_lakeformation_lf_tag.example2.values
  }  
}
```

## Argument Reference

The following arguments are required:

* `lf_tag` – (Required) Set of LF-tags to attach to the resource.

One of the following is required:

* `database` - (Optional) Configuration block for a database resource. See below.
* `table` - (Optional) Configuration block for a table resource. See below.
* `table_with_columns` - (Optional) Configuration block for a table with columns resource. See below.

The following arguments are optional:

* `catalog_id` – (Optional) Identifier for the Data Catalog. By default, the account ID. The Data Catalog is the persistent metadata store. It contains database definitions, table definitions, and other control information to manage your Lake Formation environment.

### database

The following argument is required:

* `name` – (Required) Name of the database resource. Unique to the Data Catalog.

The following argument is optional:

* `catalog_id` - (Optional) Identifier for the Data Catalog. By default, it is the account ID of the caller.

### table

The following argument is required:

* `database_name` – (Required) Name of the database for the table. Unique to a Data Catalog.
* `name` - (Required, at least one of `name` or `wildcard`) Name of the table.
* `wildcard` - (Required, at least one of `name` or `wildcard`) Whether to use a wildcard representing every table under a database. Defaults to `false`.

The following arguments are optional:

* `catalog_id` - (Optional) Identifier for the Data Catalog. By default, it is the account ID of the caller.

### table_with_columns

The following arguments are required:

* `column_names` - (Required, at least one of `column_names` or `wildcard`) Set of column names for the table.
* `database_name` – (Required) Name of the database for the table with columns resource. Unique to the Data Catalog.
* `name` – (Required) Name of the table resource.
* `wildcard` - (Required, at least one of `column_names` or `wildcard`) Whether to use a column wildcard. If `excluded_column_names` is included, `wildcard` must be set to `true` to avoid Terraform reporting a difference.

The following arguments are optional:

* `catalog_id` - (Optional) Identifier for the Data Catalog. By default, it is the account ID of the caller.
* `excluded_column_names` - (Optional) Set of column names for the table to exclude. If `excluded_column_names` is included, `wildcard` must be set to `true` to avoid Terraform reporting a difference.

## Attributes Reference

No additional attributes are exported.
