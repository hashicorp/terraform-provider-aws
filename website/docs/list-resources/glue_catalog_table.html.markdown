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
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
