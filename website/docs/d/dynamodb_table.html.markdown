---
subcategory: "DynamoDB"
layout: "aws"
page_title: "AWS: aws_dynamodb_table"
description: |-
  Provides a DynamoDB table data source.
---

# Data Source: aws_dynamodb_table

Provides information about a DynamoDB table.

## Example Usage

```terraform
data "aws_dynamodb_table" "tableName" {
  name = "tableName"
}
```

## Argument Reference

This data source supports the following arguments:

* `name` - (Required) Name of the DynamoDB table.

## Attribute Reference

See the [DynamoDB Table Resource](/docs/providers/aws/r/dynamodb_table.html) for details on the
returned attributes - they are identical.
