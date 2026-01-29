---
subcategory: "OpenSearch Serverless"
layout: "aws"
page_title: "AWS: aws_opensearchserverless_collection"
description: |-
  Lists OpenSearch Serverless Collection resources.
---

# List Resource: aws_opensearchserverless_collection

Lists OpenSearch Serverless Collection resources.

## Example Usage

```terraform
list "aws_opensearchserverless_collection" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
