---
subcategory: "Lake Formation"
layout: "aws"
page_title: "AWS: aws_lakeformation_data_cells_filter"
description: |-
  Terraform resource for managing an AWS Lake Formation Data Cells Filter.
---

# Resource: aws_lakeformation_data_cells_filter

Provides a Lake Formation Data Cells Filter resource. (In the AWS Console, this resource is sometimes referred to as "Data Filters"). A Data Cells Filter is applied to a single Glue Database Table. 

"Column-level security" is provided when a Data Cells Filter is configured to include or exclude columns in a Table.  "Row-level security" is provided when a Data Cells Filter is configured to filter (remove) rows in a Table with a Row Filter Expression.  "Cell-level security" combines row filtering and column filtering.

~> **NOTE on Lake Formation Data Cell Filters ID:** AWS does not offer an API to get a single Data Cells Filter object.  Instead, a ListDataCellsFilterPages API is provided.  This API lists all Data Cells Filters for a single Table. The Data Cells Filter resource must iterate through the results of the List API to reference the Data Cells Filter's attributes.  If there are many Data Cells Filters configured for a Table, this may result in slowness.

~> **NOTE on Lake Formation Data Cell Filters Maintenance:** AWS does not offer an API to update a single Data Cells Filter.  Instead, any changes to a Data Cells Filter will delete the existing Data Cells Filter and create a new one.

## Example Usage

### Column-level security - Filter (Exclude) some Columns and allow all Row values

```terraform
resource "aws_lakeformation_data_cells_filter" "animal_dog_exclude_birthdate" {
  table_catalog_id = "111111111111"
  database_name    = "animal"
  table_name       = "dog"
  name             = "animal_dog_exclude_birthdate"

  row_filter {
    all_rows_wildcard = "true"
  }

  column_wildcard {
    excluded_column_names = ["birthdate"]
  }
}
```

### Column-level security - Filter (Include) some Columns and allow all Row values

```terraform
resource "aws_lakeformation_data_cells_filter" "animal_dog_include_fourcolumns" {
  table_catalog_id = "111111111111"
  database_name    = "animal"
  table_name       = "dog"
  name             = "animal_dog_include_fourcolumns"

  row_filter {
    all_rows_wildcard = "true"
  }

  column_names = ["id","name","hair","lastupdate"]
  
}
```

### Row-level security - Include all Columns and filter specific Row values

```terraform
resource "aws_lakeformation_data_cells_filter" "animal_dog_hair_short" {
  table_catalog_id = "111111111111"
  database_name    = "animal"
  table_name       = "dog"
  name             = "animal_dog_hair_short"

  row_filter {
    filter_expression = "hair = 'short'"
  }

  column_wildcard {}

}
```

### Cell-level security - Filter (Include) some Columns and filter specific Row values

```terraform
resource "aws_lakeformation_data_cells_filter" "animal_dog_hair_short_include_fourcolumns" {
  table_catalog_id = "111111111111"
  database_name    = "animal"
  table_name       = "dog"
  name             = "animal_dog_hair_short_include_fourcolumns"

  row_filter {
    filter_expression = "hair = 'short'"
  }

  column_names = ["id","name","hair","lastupdate"]

}
```

## Argument Reference

The following arguments are required:

* `database_name` - (Required) Name of the database resource. Unique to the Data Catalog.

* `table_name` - (Required) Name of the table resource. Unique to the database.

* `name` - (Required) Name of the data cells filter resource. Unique to the Data Catalog.

* `row_filter`- (Required) Configuration block for table row filtering.  See below.

Exactly one of the following is required:

* `column_names` - (Optional) Set of column names in the table to INCLUDE.

* `column_wildcard` - (Optional) Configuration block for column names in the table to EXCLUDE. See below.

~> **NOTE include EMPTY column_wildcard if there is no Column-level security needed:** set `column_wildcard` to `{}` 

The following arguments are optional:

* `table_catalog_id` - (Optional) Identifier for the Data Catalog. By default, it is the account ID of the caller.

### row_filter

One of the following arguments are required:

* `filter_expression` - (Required, at least one of `all_rows_wildcard` or `filter_expression`). "WHERE clause" to filter table rows

* `all_rows_wildcard` - (Required, at least one of `all_rows_wildcard` or `filter_expression`). Boolean value. `true` = returns all rows (used in combination with `column_names` or `column_wildcard` to provide Column-level security only).  The default value = "false"

### column_wildcard

The following arguments are optional:

* `excluded_column_names` - (Optional) Set of column names in the table to exclude.

## Attributes Reference

No additional attributes are exported.