---
subcategory: "Glue"
layout: "aws"
page_title: "AWS: aws_glue_catalog_table"
description: |-
  Get information on AWS Glue Data Catalog Table
---

# Data Source: aws_glue_catalog_tables

This data source can be used to fetch table names from an AWS Glue Data Catalog Database.

## Example Usage

```terraform
data "aws_glue_catalog_tables" "example" {
  database_name          = "MyCatalogTable"
}
```

## Argument Reference

This data source supports the following arguments:

* `database_name` - (Required) Name of the table.
* `catalog_id` - (Optional) ID of the Glue Catalog and database where the table metadata resides. If omitted, this defaults to the current AWS Account ID.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - Catalog ID and Database name of the tables separated by `:`.
* `ids` - List of table names in the database.
