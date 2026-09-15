---
subcategory: "DynamoDB"
layout: "aws"
page_title: "AWS: aws_dynamodb_table_item"
description: |-
  Lists DynamoDB Table Item resources.
---

# List Resource: aws_dynamodb_table_item

Lists DynamoDB Table Item resources.

## Example Usage

```terraform
list "aws_dynamodb_table_item" "example" {
  provider = aws

  config {
    table_name = "example"
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
* `table_name` - (Required) Name of the DynamoDB table to list items from.
