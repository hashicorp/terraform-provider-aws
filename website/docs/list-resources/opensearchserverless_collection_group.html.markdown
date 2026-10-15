---
subcategory: "OpenSearch Serverless"
layout: "aws"
page_title: "AWS: aws_opensearchserverless_collection_group"
description: |-
  Lists OpenSearch Serverless Collection Group resources.
---

# List Resource: aws_opensearchserverless_collection_group

Lists OpenSearch Serverless Collection Group resources.

## Example Usage

```terraform
list "aws_opensearchserverless_collection_group" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
