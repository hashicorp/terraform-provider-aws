---
subcategory: "DynamoDB"
layout: "aws"
page_title: "AWS: aws_dynamodb_table_query"
description: |-
  Terraform data source for querying values from an AWS DynamoDB table.
---

# Data Source: aws_dynamodb_table_query

Terraform data source for querying values from an AWS DynamoDB table.

## Example Usage

### Basic Usage

```terraform
data "aws_dynamodb_table_query" "test" {
  table_name                  = aws_dynamodb_table.example.name
	key_condition_expression    = "hashKey = :value"
	expression_attribute_values = {
		":value"= jsonencode({"S" = "something"})
	}
  depends_on                  = [aws_dynamodb_table_item.example]
}
```

## Argument Reference (following [Query API documentation](https://docs.aws.amazon.com/amazondynamodb/latest/APIReference/API_Query.html))

The following arguments are required:

* `table_name` - (Required) The name of the table containing the requested item.
* `key_condition_expression` - (Required) The condition that specifies the key values for items to be retrieved by the Query action.
* `expression_attribute_values` - (Required) - One or more values that can be substituted in an expression. Use the `:` character in an expression to dereference an attribute value. The values must be valid JSON, for example: {":value" = "{\"S\": \"0000A\"}"} or {":value"= jsonencode({"S" = %q})}

The following arguments are optional:

* `expression_attribute_name` - (Optional) - One or more substitution tokens for attribute names in an expression. Use the `#` character in an expression to dereference an attribute name.
* `output_limit` - (Optional) - The total number of items to include in the result. This data source handles pagination internally (using the ExclusiveStartKey and LastEvaluatedKey parameters). It will stop performing additional queries once this `output_limit` is reached. When not specified, it will keep querying until the LastEvaluatedKey is null.
* `projection_expression` - (Optional) A string that identifies one or more attributes to retrieve from the table. These attributes can include scalars, sets, or elements of a JSON document. The attributes in the expression must be separated by commas.
If no attribute names are specified, then all attributes are returned. If any of the requested attributes are not found, they do not appear in the result.
* `consistent_read` - (Optional) - If set to true, then the operation uses strongly consistent reads; otherwise, the operation uses eventually consistent reads.
* `filter_expression` - (Optional) - A string that contains conditions that DynamoDB applies after the Query operation, but before the data is returned to you. Items that do not satisfy the FilterExpression criteria are not returned.
* `index_name` - (Optional) - The name of an index to query. This index can be any local secondary index or global secondary index on the table.
* `projection_expression` - (Optional) - A string that identifies one or more attributes to retrieve from the table. These attributes can include scalars, sets, or elements of a JSON document. The attributes in the expression must be separated by commas. If no attribute names are specified, then all attributes will be returned. If any of the requested attributes are not found, they will not appear in the result.
* `scan_index_forward` - (Optional) - Specifies the order for index traversal: If true (default), the traversal is performed in ascending order; if false, the traversal is performed in descending order. If ScanIndexForward is true, DynamoDB returns the results in the order in which they are stored (by sort key value). This is the default behavior.
* `select` - (Optional) - The attributes to be returned in the result. You can retrieve all item attributes, specific item attributes, the count of matching items, or in the case of an index, some or all of the attributes projected into the index with `ALL_ATTRIBUTES`, `ALL_PROJECTED_ATTRIBUTES`, `COUNT`, `SPECIFIC_ATTRIBUTES`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:
* `items` - List of JSON representations of maps of attribute names to [AttributeValue](https://docs.aws.amazon.com/amazondynamodb/latest/APIReference/API_AttributeValue.html) objects.
* `item_count` - The number of items in the `items` list.
* `scanned_count` - The number of items evaluated, before any QueryFilter is applied. A high ScannedCount value with few, or no, Count results indicates an inefficient Query operation.
* `query_count` - The number of queries that were performed. As discussed earlier, the data source handles pagination internally, hence this number represents the number of query calls that were done against Dynamo until the output_limit, if specified, was reached or until the LastEvaluatedKey returns null.
