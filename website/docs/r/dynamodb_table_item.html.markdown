---
subcategory: "DynamoDB"
layout: "aws"
page_title: "AWS: aws_dynamodb_table_item"
description: |-
  Provides a DynamoDB table item resource
---

# Resource: aws_dynamodb_table_item

Provides a DynamoDB table item resource

-> **Note:** This resource is not meant to be used for managing large amounts of data in your table, it is not designed to scale.
  You should perform **regular backups** of all data in the table, see [AWS docs for more](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/BackupRestore.html).

## Example Usage

```terraform
resource "aws_dynamodb_table_item" "example" {
  table_name = aws_dynamodb_table.example.name
  hash_key   = aws_dynamodb_table.example.hash_key

  item = <<ITEM
{
  "exampleHashKey": {"S": "something"},
  "one": {"N": "11111"},
  "two": {"N": "22222"},
  "three": {"N": "33333"},
  "four": {"N": "44444"}
}
ITEM
}

resource "aws_dynamodb_table" "example" {
  name           = "example-name"
  read_capacity  = 10
  write_capacity = 10
  hash_key       = "exampleHashKey"

  attribute {
    name = "exampleHashKey"
    type = "S"
  }
}
```

## Argument Reference

~> **Note:** Names included in `item` are represented internally with everything but letters removed. There is the possibility of collisions if two names, once filtered, are the same. For example, the names `your-name-here` and `yournamehere` will overlap and cause an error.

This resource supports the following arguments:

* `hash_key` - (Required) Hash key to use for lookups and identification of the item
* `item` - (Required) JSON representation of a map of attribute name/value pairs, one for each attribute. Only the primary key attributes are required; you can optionally provide other attribute name-value pairs for the item.
* `range_key` - (Optional) Range key to use for lookups and identification of the item. Required if there is range key defined in the table.
* `table_name` - (Required) Name of the table to contain the item.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

## Import

You cannot import DynamoDB table items.
