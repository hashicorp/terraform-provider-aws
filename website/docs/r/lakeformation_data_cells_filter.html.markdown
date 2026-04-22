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
    database_name    = aws_glue_catalog_database.example.name
    name             = "example"
    table_catalog_id = data.aws_caller_identity.current.account_id
    table_name       = aws_glue_catalog_table.example.name

    column_names = ["my_column"]

    row_filter {
      filter_expression = "my_column='example'"
    }
  }
}
```

### Filter with Excluded Columns Only (No Row Filter)

When excluding columns without a row filter, you must include `all_rows_wildcard {}`:

```terraform
resource "aws_lakeformation_data_cells_filter" "excluded_columns" {
  table_data {
    database_name    = aws_glue_catalog_database.example.name
    name             = "exclude-pii"
    table_catalog_id = data.aws_caller_identity.current.account_id
    table_name       = aws_glue_catalog_table.example.name

    column_wildcard {
      excluded_column_names = ["ssn", "credit_card"]
    }

    row_filter {
      all_rows_wildcard {}
    }
  }
}
```

### Filter with Row Filter and Excluded Columns

```terraform
resource "aws_lakeformation_data_cells_filter" "row_and_column" {
  table_data {
    database_name    = aws_glue_catalog_database.example.name
    name             = "marketing-filtered"
    table_catalog_id = data.aws_caller_identity.current.account_id
    table_name       = aws_glue_catalog_table.example.name

    column_wildcard {
      excluded_column_names = ["salary", "bonus"]
    }

    row_filter {
      filter_expression = "department = 'Marketing'"
    }
  }
}
```

### Filter with Row Filter Only (All Columns Included)

To include all columns with a row filter, set `excluded_column_names` to an empty list:

```terraform
resource "aws_lakeformation_data_cells_filter" "row_only" {
  table_data {
    database_name    = aws_glue_catalog_database.example.name
    name             = "regional-filter"
    table_catalog_id = data.aws_caller_identity.current.account_id
    table_name       = aws_glue_catalog_table.example.name

    column_wildcard {
      excluded_column_names = []
    }

    row_filter {
      filter_expression = "region = 'US-WEST'"
    }
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
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

**Note:** Exactly one of `filter_expression` or `all_rows_wildcard` must be specified.

* `all_rows_wildcard` - (Optional) A wildcard that matches all rows. Required when applying column-level filtering without row-level filtering. Use an empty block: `all_rows_wildcard {}`.
* `filter_expression` - (Optional) A PartiQL predicate expression for row-level filtering.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `2m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Lake Formation Data Cells Filter using the `database_name`, `name`, `table_catalog_id`, and `table_name` separated by `,`. For example:

```terraform
import {
  to = aws_lakeformation_data_cells_filter.example
  id = "database_name,name,table_catalog_id,table_name"
}
```

Using `terraform import`, import Lake Formation Data Cells Filter using the `database_name`, `name`, `table_catalog_id`, and `table_name` separated by `,`. For example:

```console
% terraform import aws_lakeformation_data_cells_filter.example database_name,name,table_catalog_id,table_name
```
