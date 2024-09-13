---
subcategory: "Lake Formation"
layout: "aws"
page_title: "AWS: aws_lakeformation_data_cells_filter"
description: |-
  Terraform resource for managing an AWS Lake Formation Data Cells Filter.
---
# Resource: aws_lakeformation_data_cells_filter

Terraform resource for managing an AWS Lake Formation Data Cells Filter.

## Example Usage

### Basic Usage

```terraform
resource "aws_lakeformation_data_cells_filter" "example" {
  table_data {
    database_name    = aws_glue_catalog_database.test.name
    name             = "example"
    table_catalog_id = data.aws_caller_identity.current.account_id
    table_name       = aws_glue_catalog_table.test.name

    column_names = ["my_column"]

    row_filter {
      filter_expression = "my_column='example'"
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `table_data` - (Required) Information about the data cells filter. See [Table Data](#table-data) below for details.

### Table Data

* `database_name` - (Required) The name of the database.
* `name` - (Required) The name of the data cells filter.
* `table_catalog_id` - (Required) The ID of the Data Catalog.
* `table_name` - (Required) The name of the table.
* `column_names` - (Optional) A list of column names and/or nested column attributes.
* `column_wildcard` - (Optional) A wildcard with exclusions. See [Column Wildcard](#column-wildcard) below for details.
* `row_filter` - (Optional) A PartiQL predicate. See [Row Filter](#row-filter) below for details.
* `version_id` - (Optional) ID of the data cells filter version.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Provider composed identifier: `database_name,name,table_catalog_id,table_name`.

#### Column Wildcard

* `excluded_column_names` - (Optional) Excludes column names. Any column with this name will be excluded.

#### Row Filter

* `all_rows_wildcard` - (Optional) A wildcard that matches all rows.
* `filter_expression` - (Optional) A filter expression.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `2m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Lake Formation Data Cells Filter using the `example_id_arg`. For example:

```terraform
import {
  to = aws_lakeformation_data_cells_filter.example
  id = "database_name,name,table_catalog_id,table_name"
}
```

Using `terraform import`, import Lake Formation Data Cells Filter using the `id`. For example:

```console
% terraform import aws_lakeformation_data_cells_filter.example database_name,name,table_catalog_id,table_name
```
