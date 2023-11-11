---
subcategory: "DynamoDB"
layout: "aws"
page_title: "AWS: aws_dynamodb_table_item"
description: |-
  Terraform data source for retrieving a value from an AWS DynamoDB table.
---

# Data Source: aws_dynamodb_table_item

Terraform data source for retrieving a value from an AWS DynamoDB table.

## Example Usage

### Basic Usage

```terraform
data "aws_dynamodb_table_item" "test" {
  table_name = aws_dynamodb_table.example.name
  expression_attribute_names = {
    "#P" = "Percentile"
  }
  projection_expression = "#P"
  key                   = <<KEY
{
	"hashKey": {"S": "example"}
}
KEY
  depends_on            = [aws_dynamodb_table_item.example]
}
```

## Argument Reference

The following arguments are required:

* `table_name` - (Required) The name of the table containing the requested item.
* `key` - (Required) A map of attribute names to AttributeValue objects, representing the primary key of the item to retrieve.
  For the primary key, you must provide all of the attributes. For example, with a simple primary key, you only need to provide a value for the partition key. For a composite primary key, you must provide values for both the partition key and the sort key.

The following arguments are optional:

* `expression_attribute_name` - (Optional) - One or more substitution tokens for attribute names in an expression. Use the `#` character in an expression to dereference an attribute name.
* `projection_expression` - (Optional) A string that identifies one or more attributes to retrieve from the table. These attributes can include scalars, sets, or elements of a JSON document. The attributes in the expression must be separated by commas.
If no attribute names are specified, then all attributes are returned. If any of the requested attributes are not found, they do not appear in the result.
* `fail_on_missing` - (Optional) A boolean to fail item lookup if it doesn't exists.  Default is false

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `item` - JSON representation of a map of attribute names to [AttributeValue](https://docs.aws.amazon.com/amazondynamodb/latest/APIReference/API_AttributeValue.html) objects, as specified by ProjectionExpression.
