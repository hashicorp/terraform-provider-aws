---
subcategory: "Glue"
layout: "aws"
page_title: "AWS: aws_glue_catalog"
description: |-
  Lists Glue Catalog resources.
---

# List Resource: aws_glue_catalog

Lists Glue Catalog resources.

## Example Usage

```terraform
list "aws_glue_catalog" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
