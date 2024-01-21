---
subcategory: "Glue"
layout: "aws"
page_title: "AWS: aws_glue_catalog_database"
description: |-
  Get information on AWS Glue Data Catalog Database
---

# Data Source: aws_glue_catalog_database

This data source can be used to fetch information about an AWS Glue Data Catalog Table.

## Example Usage

```terraform
data "aws_glue_catalog_database" "example" {
  name          = "MyCatalogDatabase"
}
```

## Argument Reference

This data source supports the following arguments:

* `name` - (Required) Name of the table.
* `catalog_id` - (Optional) ID of the Glue Catalog and database where the table metadata resides. If omitted, this defaults to the current AWS Account ID.
