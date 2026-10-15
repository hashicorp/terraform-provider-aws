---
subcategory: "OpenSearch Ingestion (OSIS)"
layout: "aws"
page_title: "AWS: aws_osis_pipeline"
description: |-
  Lists OpenSearch Ingestion (OSIS) Pipeline resources.
---

# List Resource: aws_osis_pipeline

Lists OpenSearch Ingestion (OSIS) Pipeline resources.

## Example Usage

```terraform
list "aws_osis_pipeline" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
