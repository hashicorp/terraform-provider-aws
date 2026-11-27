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

* `account` - (Optional) AWS account ID to query. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `region` - (Optional) Region to query. Defaults to provider region.
