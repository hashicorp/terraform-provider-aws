---
subcategory: "DynamoDB"
layout: "aws"
page_title: "AWS: aws_dynamodb_kinesis_streaming_destination"
description: |-
  Enables a Kinesis streaming destination for a DynamoDB table
---

# Resource: aws_dynamodb_kinesis_streaming_destination

Enables a [Kinesis streaming destination](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/kds.html) for data replication of a DynamoDB table.

## Example Usage

```terraform
resource "aws_dynamodb_table" "example" {
  name     = "orders"
  hash_key = "id"

  attribute {
    name = "id"
    type = "S"
  }
}

resource "aws_kinesis_stream" "example" {
  name        = "order_item_changes"
  shard_count = 1
}

resource "aws_dynamodb_kinesis_streaming_destination" "example" {
  stream_arn = aws_kinesis_stream.example.arn
  table_name = aws_dynamodb_table.example.name
}
```

## Argument Reference

This resource supports the following arguments:

* `stream_arn` - (Required) The ARN for a Kinesis data stream. This must exist in the same account and region as the DynamoDB table.
  
* `table_name` - (Required) The name of the DynamoDB table. There
  can only be one Kinesis streaming destination for a given DynamoDB table.
  
## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The `table_name` and `stream_arn` separated by a comma (`,`).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import DynamoDB Kinesis Streaming Destinations using the `table_name` and `stream_arn` separated by `,`. For example:

```terraform
import {
  to = aws_dynamodb_kinesis_streaming_destination.example
  id = "example,arn:aws:kinesis:us-east-1:111122223333:exampleStreamName"
}
```

Using `terraform import`, import DynamoDB Kinesis Streaming Destinations using the `table_name` and `stream_arn` separated by `,`. For example:

```console
% terraform import aws_dynamodb_kinesis_streaming_destination.example example,arn:aws:kinesis:us-east-1:111122223333:exampleStreamName
```
