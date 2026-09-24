---
subcategory: "Kinesis"
layout: "aws"
page_title: "AWS: aws_kinesis_channel"
description: |-
  Manages an AWS Kinesis Channel.
---

# Resource: aws_kinesis_channel

Manages an AWS Kinesis Channel.

Provides a fully managed way to deliver streaming data from your Kinesis stream to general-purpose Amazon S3 buckets or Apache Iceberg streaming tables, with no infrastructure to operate.

## Example Usage

### Basic Usage

```terraform
resource "aws_kinesis_channel" "example" {
  channel_name               = "example_channel"
  service_execution_role_arn = aws_iam_role.example.arn

  stream_configuration_list {
    stream_arn = aws_kinesis_stream.example.arn

    record_configuration {
      record_format_type = "JSON"
    }
  }

  s3_destination_configuration {
    storage_configuration {
      bucket_arn            = aws_s3_bucket.example.arn
      expected_bucket_owner = data.aws_caller_identity.current.account_id
      compression_type      = "NONE"
    }
  }

  depends_on = [
    aws_iam_role_policy.example,
  ]
}

resource "aws_iam_role" "example" {
  name = "example_role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = [
            "kinesis.amazonaws.com",
          ]
        }
        Action = "sts:AssumeRole"
      }
    ]
  })
}

resource "aws_iam_role_policy" "example" {
  name = "example_policy"
  role = aws_iam_role.example.id

  policy = jsonencode({
    Version = "2012-10-17"

    Statement = [
      {
        Effect = "Allow"
        Action = [
          "kinesis:*"
        ]
        Resource = aws_kinesis_stream.example.arn
      },
      {
        Effect = "Allow"
        Action = [
          "logs:CreateLogStream",
          "logs:PutLogEvents",
          "logs:DescribeLogStreams"
        ]
        Resource = "${aws_cloudwatch_log_group.example.arn}:*"
      },
      {
        Effect = "Allow"
        Action = ["s3:*"]
        Resource = [
          "${aws_s3_bucket.example.arn}",
          "${aws_s3_bucket.example.arn}/*"
        ]
      }
    ]
  })
}

resource "aws_s3_bucket" "example" {
  bucket        = "example-s3-bucket"
  force_destroy = true
}

resource "aws_kinesis_stream" "example" {
  name                      = "example_stream"
  encryption_type           = "NONE"
  enforce_consumer_deletion = false
  max_record_size_in_kib    = 1024
  retention_period          = 24
  stream_mode_details {
    stream_mode = "ON_DEMAND"
  }
}

resource "aws_cloudwatch_log_group" "example" {
  name = "/aws/kinesis/example_logs"
}

data "aws_caller_identity" "current" {}
```

### Streaming Table

```terraform
resource "aws_kinesis_channel" "example" {
  channel_name               = "example_channel"
  service_execution_role_arn = aws_iam_role.example.arn

  stream_configuration_list {
    stream_arn = aws_kinesis_stream.example.arn

    record_configuration {
      record_format_type = "GSR_JSON"
      gsr_schema_arn     = aws_glue_schema.example.arn
    }
  }

  s3_tables_destination_configuration {
    dead_letter_queue_s3_configuration {
      bucket_arn            = aws_s3_bucket.example.arn
      expected_bucket_owner = data.aws_caller_identity.current.account_id
      error_output_prefix   = "errors/"
    }

    s3_tables_configuration_list {
      table_bucket_arn = aws_s3tables_table_bucket.example.arn
      namespace        = aws_s3tables_namespace.example.namespace
      table_name       = "example_table"
      compression_type = "ZSTD"

      partition_spec {
        partition_fields {
          source_name = "event_time"
          transform   = "TIME_HOUR"
        }
      }
    }
  }

  depends_on = [
    aws_iam_role_policy.example,
  ]
}

resource "aws_s3tables_namespace" "example" {
  namespace        = "example_namespace"
  table_bucket_arn = aws_s3tables_table_bucket.example.arn
}

resource "aws_s3tables_table_bucket" "example" {
  name = "example-s3table-bucket"
}

resource "aws_glue_registry" "example" {
  registry_name = "example_registry"
}

resource "aws_glue_schema" "example" {
  schema_name   = "example_schema"
  registry_arn  = aws_glue_registry.example.arn
  data_format   = "JSON"
  compatibility = "NONE"
  schema_definition = jsonencode({
    "$id" : "https://example.com/record.schema.json",
    "$schema" : "http://json-schema.org/draft-07/schema#",
    "title" : "Record",
    "type" : "object",
    "properties" : {
      "event_time" : {
        "type" : "string",
        "format" : "date-time"
      },
      "data" : {
        "type" : "string"
      }
    },
    "required" : ["event_time"]
  })
}

resource "aws_iam_role" "example" {
  name = "example_role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = [
            "kinesis.amazonaws.com",
          ]
        }
        Action = "sts:AssumeRole"
      }
    ]
  })
}

resource "aws_iam_role_policy" "example" {
  name = "example_policy"
  role = aws_iam_role.example.id

  policy = jsonencode({
    Version = "2012-10-17"

    Statement = [
      {
        Effect = "Allow"
        Action = [
          "kinesis:*"
        ]
        Resource = aws_kinesis_stream.example.arn
      },
      {
        Effect = "Allow"
        Action = [
          "glue:*"
        ]
        Resource = ["*"]
      },
      {
        Effect = "Allow"
        Action = [
          "logs:CreateLogStream",
          "logs:PutLogEvents",
          "logs:DescribeLogStreams"
        ]
        Resource = "${aws_cloudwatch_log_group.example.arn}:*"
      },
      {
        Effect   = "Allow"
        Action   = ["s3:*"]
        Resource = [
          "${aws_s3_bucket.example.arn}",
          "${aws_s3_bucket.example.arn}/*"
        ]
      }
    ]
  })
}

resource "aws_s3_bucket" "example" {
  bucket        = "example-s3-bucket"
  force_destroy = true
}

resource "aws_kinesis_stream" "example" {
  name                      = "example_stream"
  encryption_type           = "NONE"
  enforce_consumer_deletion = false
  max_record_size_in_kib    = 1024
  retention_period          = 24
  stream_mode_details {
    stream_mode = "ON_DEMAND"
  }
}

resource "aws_cloudwatch_log_group" "example" {
  name = "/aws/kinesis/example_logs"
}

data "aws_caller_identity" "current" {}
```

## Argument Reference

The following arguments are required:

* `channel_name` - (Required) Name of the channel. Unique within the AWS account and Region.
* `service_execution_role_arn` - (Required) ARN of the IAM role that Amazon Kinesis Data Streams assumes to write records to the destination.
* `stream_configuration_list` - (Required) Source stream configuration for the channel. See [`stream_configuration_list`](#stream_configuration_list) below.

The following arguments are optional:

* `encryption_configuration` - (Optional) Server-side encryption configuration for the channel. See [`encryption_configuration`](#encryption_configuration) below.
* `logging_configuration` - (Optional) CloudWatch Logs configuration for the channel. See [`logging_configuration`](#logging_configuration) below.
* `s3_destination_configuration` - (Optional) Configuration for delivery to a general purpose Amazon S3 bucket. Present only when the channel destination is a general purpose Amazon S3 bucket. Conflicts with `s3_tables_destination_configuration`. Exactly one of `s3_destination_configuration` or `s3_tables_destination_configuration` must be specified. See [`s3_destination_configuration`](#s3_destination_configuration) below.
* `s3_tables_destination_configuration` - (Optional) Configuration for delivery to streaming tables on Apache Iceberg in Amazon S3 Tables. Present only when the channel destination is a streaming table. Conflicts with `s3_destination_configuration`. Exactly one of `s3_destination_configuration` or `s3_tables_destination_configuration` must be specified. See [`s3_tables_destination_configuration`](#s3_tables_destination_configuration) below.
* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### stream_configuration_list

The following arguments are required:

* `record_configuration` - (Required) Record format configuration for the source stream. See [`record_configuration`](#record_configuration) below.
* `stream_arn` - (Required) ARN of the source Kinesis data stream.

### record_configuration

The following arguments are required:

* `record_format_type` - (Required) Format of records on the source stream. Valid values are `GSR_JSON`, `JSON`, `STRING`, or `BYTE_ARRAY`. Note that `GSR_JSON` is supported only for S3 Tables destinations, while `STRING` and `BYTE_ARRAY` are supported only for general-purpose S3 destinations.

The following arguments are optional:

* `gsr_schema_arn` - (Optional) ARN of the AWS Glue Schema Registry schema used to validate records. Required when the channel destination is a streaming table (Amazon S3 Tables), for both `JSON` and `GSR_JSON` record formats.

### s3_destination_configuration

The following arguments are required:

* `storage_configuration` - (Required) S3 storage configuration for the channel. See [`storage_configuration`](#storage_configuration) below.

The following arguments are optional:

* `data_freshness_in_seconds` - (Optional) Maximum age, in seconds, of undelivered data. Valid values are between `300` and `900`. Defaults to `300`.
* `dead_letter_queue_s3_configuration` - (Optional) Dead-letter queue configuration for records that cannot be delivered. If not specified, defaults to the destination bucket with an error prefix. See [`dead_letter_queue_s3_configuration`](#dead_letter_queue_s3_configuration) below.

### storage_configuration

The following arguments are required:

* `bucket_arn` - (Required) ARN of the destination S3 bucket.
* `compression_type` - (Required) Compression applied to delivered objects. Valid values are `NONE`, `GZIP`, `SNAPPY`, or `ZSTD`.
* `expected_bucket_owner` - (Required) AWS account ID of the expected owner of the destination bucket. Helps prevent delivery to an unintended bucket if ownership changes.

The following arguments are optional:

* `output_key_template` - (Optional) Template used to construct the S3 object key for delivered objects. If not specified, a default template is used.
* `storage_class` - (Optional) S3 storage class for delivered objects. Valid values are `STANDARD`, `INTELLIGENT_TIERING`, `STANDARD_IA`, or `GLACIER_IR`. Defaults to `STANDARD`.

### dead_letter_queue_s3_configuration

The following arguments are required:

* `bucket_arn` - (Required) ARN of the dead-letter queue S3 bucket.
* `expected_bucket_owner` - (Required) AWS account ID of the expected owner of the dead-letter queue bucket.

The following arguments are optional:

* `error_output_prefix` - (Optional) S3 key prefix for error records.

### s3_tables_destination_configuration

The following arguments are required:

* `dead_letter_queue_s3_configuration` - (Required) Dead-letter queue configuration for records that cannot be delivered. See [`dead_letter_queue_s3_configuration`](#dead_letter_queue_s3_configuration) above.
* `s3_tables_configuration_list` - (Required) List of streaming table configurations. See [`s3_tables_configuration_list`](#s3_tables_configuration_list) below.

The following arguments are optional:

* `data_freshness_in_seconds` - (Optional) Maximum age, in seconds, of undelivered data. Valid values are between `300` and `900`. Defaults to `300`.

### s3_tables_configuration_list

The following arguments are required:

* `compression_type` - (Required) Compression applied to Parquet data files. Valid values are `NONE`, `GZIP`, `SNAPPY`, or `ZSTD`.
* `namespace` - (Required) Namespace (database) of the destination table.
* `table_bucket_arn` - (Required) ARN of the Amazon S3 table bucket.
* `table_name` - (Required) Name of the destination table.

The following arguments are optional:

* `partition_spec` - (Optional) Partitioning specification for the destination table. See [`partition_spec`](#partition_spec) below.

### partition_spec

The following arguments are required:

* `partition_fields` - (Required) List of partition fields. See [`partition_fields`](#partition_fields) below.

### partition_fields

The following arguments are required:

* `source_name` - (Required) Name of the source column used for partitioning. Must be of the `timestamptz` type.
* `transform` - (Required) Partition transform to apply. The only valid value is `TIME_HOUR`.

### encryption_configuration

The following arguments are required:

* `encryption_type` - (Required) Encryption type. The only valid value is `KMS`.
* `key_id` - (Required) Identifier of the customer managed AWS KMS key. Cannot use the Amazon Kinesis Data Streams service key (`aws/kinesis`).

### logging_configuration

The following arguments are required:

* `cloudwatch_logs` - (Required) CloudWatch Logs settings for the channel. See [`cloudwatch_logs`](#cloudwatch_logs) below.

### cloudwatch_logs

The following arguments are required:

* `enabled` - (Required) Whether to enable logging to CloudWatch Logs.

The following arguments are optional:

* `log_group_name` - (Optional) Name of the CloudWatch Logs log group. Defaults to `/aws/kinesis/{channelName}/{channelId}`.
* `log_stream_name` - (Optional) Name of the CloudWatch Logs log stream. Defaults to `DestinationDelivery`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `channel_arn` - ARN of the channel.
* `channel_creation_timestamp` - Time at which the channel was created.
* `channel_id` - Unique identifier of the channel.
* `channel_status` - Current status of the channel.
* `channel_status_reason` - Message describing the reason for a `FAILED` status.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `5m`)
* `update` - (Default `5m`)
* `delete` - (Default `5m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_kinesis_channel.example
  identity = {
    arn = "arn:aws:kinesis:us-east-1:123456789012:channel/pw203t4lou5vu3p76"
  }
}

resource "aws_kinesis_channel" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `arn` - ARN of the Channel.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Kinesis Channel using the `channel_arn`. For example:

```terraform
import {
  to = aws_kinesis_channel.example
  id = "arn:aws:kinesis:us-east-1:123456789012:channel/pw203t4lou5vu3p76"
}
```

Using `terraform import`, import Kinesis Channel using the `channel_arn`. For example:

```console
% terraform import aws_kinesis_channel.example arn:aws:kinesis:us-east-1:123456789012:channel/pw203t4lou5vu3p76
```
