---
subcategory: "DynamoDB"
layout: "aws"
page_title: "AWS: aws_dynamodb_tables"
description: |-
  Provides a list of all AWS DynamoDB Table Names in a Region
---

# Data Source: aws_dynamodb_tables

Returns a list of all AWS DynamoDB table names in a region.

## Example Usage

The following example retrieves a list of all DynamoDB table names in a region.

```hcl
data "aws_dynamodb_tables" "all" {}

output "table_names" {
  value = data.aws_dynamodb_tables.all.ids
}
```

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `ids` - A list of all the DynamoDB table names found.
