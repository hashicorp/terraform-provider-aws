---
subcategory: "DynamoDB"
layout: "aws"
page_title: "AWS: aws_dynamodb_table"
description: |-
  Lists DynamoDB Table resources.
---

# List Resource: aws_dynamodb_table

Lists DynamoDB Table resources.

## Example Usage

```terraform
list "aws_dynamodb_table" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
