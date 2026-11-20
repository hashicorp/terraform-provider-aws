---
subcategory: "Glue"
layout: "aws"
page_title: "AWS: aws_glue_catalog_table"
description: |-
  Lists Glue Catalog Table resources.
---

# List Resource: aws_glue_catalog_table

Lists Glue Catalog Table resources.

## Example Usage

```terraform
list "aws_glue_catalog_table" "example" {
  provider = aws

  config {
    database_name = "example"
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `catalog_id` - (Optional) ID of the Glue Catalog to list tables from. Defaults to the AWS account ID.
* `database_name` - (Required) Name of the Glue Catalog Database to list tables from.
* `region` - (Optional) Region to query. Defaults to provider region.
