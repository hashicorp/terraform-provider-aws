---
subcategory: "OpenSearch Ingestion (OSIS)"
layout: "aws"
page_title: "AWS: aws_osis_resource_policy"
description: |-
  Lists OpenSearch Ingestion (OSIS) Resource Policy resources.
---

# List Resource: aws_osis_resource_policy

Lists OpenSearch Ingestion (OSIS) Resource Policy resources.

## Example Usage

```terraform
list "aws_osis_resource_policy" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
