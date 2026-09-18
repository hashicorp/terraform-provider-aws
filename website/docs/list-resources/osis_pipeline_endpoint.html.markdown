---
subcategory: "OpenSearch Ingestion (OSIS)"
layout: "aws"
page_title: "AWS: aws_osis_pipeline_endpoint"
description: |-
  Lists OpenSearch Ingestion (OSIS) Pipeline Endpoint resources.
---

# List Resource: aws_osis_pipeline_endpoint

Lists OpenSearch Ingestion (OSIS) Pipeline Endpoint resources.

## Example Usage

```terraform
list "aws_osis_pipeline_endpoint" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
