---
subcategory: "DynamoDB"
layout: "aws"
page_title: "AWS: aws_dynamodb_global_secondary_index"
description: |-
  Lists DynamoDB Global Secondary Index resources.
---

# List Resource: aws_dynamodb_global_secondary_index

Lists DynamoDB Global Secondary Index resources.

## Example Usage

```terraform
list "aws_dynamodb_global_secondary_index" "example" {
  provider = aws

  config {
    table_name = "my-table"
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
* `table_name` - (Required) Name of the DynamoDB table.
