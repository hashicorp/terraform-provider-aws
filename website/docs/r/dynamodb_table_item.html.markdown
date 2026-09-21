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

This resource supports the following arguments:

* `hash_key` - (Required) Hash key to use for lookups and identification of the item
* `item` - (Required) JSON representation of a map of attribute name/value pairs, one for each attribute. Only the primary key attributes are required; you can optionally provide other attribute name-value pairs for the item.
* `range_key` - (Optional) Range key to use for lookups and identification of the item. Required if there is range key defined in the table.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `table_name` - (Required) Name or ARN of the table to contain the item.

~> **Note:** Names included in `item` are represented internally with everything but letters removed. There is the possibility of collisions if two names, once filtered, are the same. For example, the names `your-name-here` and `yournamehere` will overlap and cause an error.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `hash_key_value` - Canonical string representation of the hash key value. Binary values are base64-encoded; numbers and strings are taken verbatim.
* `range_key_value` - Canonical string representation of the range key value, when the table has a range key. Same encoding as `hash_key_value`.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute, which avoids any concerns about separator characters appearing inside key values:

```terraform
import {
  to = aws_dynamodb_table_item.example
  identity = {
    table_name      = "example-name"
    hash_key_value  = "something"
    range_key_value = "something-else" # omit for tables with no range key
  }
}

resource "aws_dynamodb_table_item" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `table_name` (String) Name of the DynamoDB table.
* `hash_key_value` (String) Canonical value of the hash key (base64 for `B`, verbatim for `N` and `S`).

#### Optional

* `range_key_value` (String) Canonical value of the range key, required for tables that define a range key.
* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import DynamoDB Table Items using the table name and key values, separated by commas (`,`). For example:

```terraform
import {
  to = aws_dynamodb_table_item.example
  id = "example-name,something"
}
```

For tables with a range key, append the range key value:

```terraform
import {
  to = aws_dynamodb_table_item.example
  id = "example-name,something,something-else"
}
```

Use [`terraform import`](https://developer.hashicorp.com/terraform/cli/commands/import) for the same effect on the command line:

```console
% terraform import aws_dynamodb_table_item.example example-name,something
```

~> **Note:** Importing requires `dynamodb:DescribeTable` in addition to `dynamodb:GetItem`. The DescribeTable call is used to recover the key attribute names and types from the table's schema.

~> **Note:** If a hash key or range key value contains the separator character (`,`), use the `import` block with the `identity` attribute. The legacy `terraform import` command and `id`-based `import` block cannot disambiguate separators from value content.

For Binary (`B`) key attributes, the value in the import ID and the identity attribute must be standard base64.
