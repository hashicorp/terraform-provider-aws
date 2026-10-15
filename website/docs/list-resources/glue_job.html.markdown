---
subcategory: "Glue"
layout: "aws"
page_title: "AWS: aws_glue_job"
description: |-
  Lists Glue Job resources.
---

# List Resource: aws_glue_job

Lists Glue Job resources.

## Example Usage

```terraform
list "aws_glue_job" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
