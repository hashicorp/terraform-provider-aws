---
subcategory: "DynamoDB"
layout: "aws"
page_title: "AWS: aws_dynamodb_kinesis_streaming_destination"
description: |-
  Configures a Kinesis streaming destination for item level changes to a DynamoDB table
---

# Resource: aws_dynamodb_kinesis_streaming_destination

Configures a [Kinesis streaming destination](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/kds.html) for item level changes to a DynamoDB table.

## Example Usage

```hcl
resource "aws_dynamodb_table" "orders" {
  name           = "orders"
  hash_key       = "id"

  attribute {
    name = "id"
    type = "S"
  }
}

resource "aws_kinesis_stream" "order_item_changes" {
  name        = "order_item_changes"
  shard_count = 1
}

resource "aws_dynamodb_kinesis_streaming_destination" "order_changes" {
  table_name = aws_dynamodb_table.orders.name
  stream_arn = aws_kinesis_stream.order_item_changes.arn
}
```

## Argument Reference

The following arguments are supported:

* `table_name` - (Required) The name of the DynamoDB table to capture changes from. There 
  can only be one Kinesis streaming destination for a given DynamoDB table.
* `stream_arn` - (Required) The arn of the Kinesis stream to capture changes into. This 
must exist in the same account and region as the DynamoDB table.
