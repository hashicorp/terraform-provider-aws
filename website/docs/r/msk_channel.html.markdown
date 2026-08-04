---
subcategory: "Managed Streaming for Kafka"
layout: "aws"
page_title: "AWS: aws_msk_channel"
description: |-
  Manages an Amazon Managed Streaming for Apache Kafka (MSK) Channel.
---

# Resource: aws_msk_channel

Manages an Amazon Managed Streaming for Apache Kafka (MSK) Channel.

With [Amazon MSK Data Delivery](https://docs.aws.amazon.com/msk/latest/developerguide/msk-data-delivery.html), a channel delivers Apache Kafka data from an Amazon MSK Provisioned cluster that uses Express brokers directly to Amazon S3, without connectors or additional infrastructure to manage. A channel reads records from a Kafka topic and delivers them to one of two destinations:

* **Amazon S3 general purpose buckets** — records are written in their source format (`JSON`, `ByteArray`, or `String`) as objects, for use cases such as log archival, compliance retention, and Kafka replay.
* **Apache Iceberg tables on Amazon S3 Tables** — `JSON` records, validated against an AWS Glue Schema Registry schema, are materialized as Apache Iceberg tables.

Records that cannot be processed are routed to a required dead-letter queue (DLQ) Amazon S3 bucket. Channels are supported only on Amazon MSK Provisioned clusters that use Express brokers; Standard brokers and Amazon MSK Serverless are not supported.

~> **Note:** A channel does not backfill previously produced data; only data produced after creation is delivered. For the Iceberg destination, a channel creates a new Iceberg table for each configuration.

## Example Usage

### Amazon S3 Destination

```terraform
resource "aws_msk_channel" "example" {
  channel_name = "example"
  cluster_arn  = aws_msk_cluster.example.arn

  topic_configuration {
    topic_arn = aws_msk_topic.example.arn

    record_converter {
      value_converter = "BYTE_ARRAY"
    }
  }

  s3_destination {
    service_execution_role_arn = aws_iam_role.example.arn

    dead_letter_queue_s3 {
      bucket_arn = aws_s3_bucket.dlq.arn
    }

    storage {
      bucket_arn       = aws_s3_bucket.example.arn
      compression_type = "NONE"
      storage_class    = "STANDARD"
    }
  }
}
```

### Apache Iceberg Destination

```terraform
resource "aws_msk_channel" "example" {
  channel_name = "example"
  cluster_arn  = aws_msk_cluster.example.arn

  topic_configuration {
    topic_arn = aws_msk_topic.example.arn

    # The Iceberg destination accepts only JSON or JSON_SCHEMA_GSR records.
    record_converter {
      value_converter = "JSON"
    }

    record_schema {
      gsr_arn = aws_glue_schema.example.arn
    }
  }

  iceberg_destination {
    append_only                = true
    service_execution_role_arn = aws_iam_role.example.arn

    catalog {
      warehouse_location = aws_s3tables_table_bucket.example.arn
    }

    dead_letter_queue_s3 {
      bucket_arn = aws_s3_bucket.dlq.arn
    }

    destination_table {
      destination_database_name = "example_namespace"
      destination_table_name    = "example_table"

      # Iceberg tables use time-based partitioning. `source` must reference exactly one
      # timestamp column defined in the Glue Schema Registry schema.
      partition_spec {
        partition_strategy = "TIME_HOUR"

        source {
          source_name = "event_time"
        }
      }
    }

    schema_evolution {
      enable_schema_evolution = false
    }

    table_creation {
      enable_table_creation = true
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `channel_name` - (Required) Name of the channel. Must be unique within the cluster. Changing this forces a new resource to be created.
* `cluster_arn` - (Required) Amazon Resource Name (ARN) that uniquely identifies the cluster. Changing this forces a new resource to be created.
* `topic_configuration` - (Required) Configuration of the Apache Kafka topic that feeds the channel. Changing this forces a new resource to be created. See [`topic_configuration` Block](#topic_configuration-block) below.

The following arguments are optional:

* `encryption_configuration` - (Optional) AWS KMS encryption configuration applied to data at rest. Changing this forces a new resource to be created. See [`encryption_configuration` Block](#encryption_configuration-block) below.
* `iceberg_destination` - (Optional) Apache Iceberg destination for the channel. Exactly one of `iceberg_destination` or `s3_destination` is required. With the exception of `data_freshness_in_seconds`, changing an argument in this block forces a new resource to be created. See [`iceberg_destination` Block](#iceberg_destination-block) below.
* `logging_info` - (Optional) Destinations to which the channel publishes operational logs. Changing this forces a new resource to be created. See [`logging_info` Block](#logging_info-block) below.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `s3_destination` - (Optional) Amazon S3 destination for the channel. Exactly one of `iceberg_destination` or `s3_destination` is required. With the exception of `data_freshness_in_seconds`, changing an argument in this block forces a new resource to be created. See [`s3_destination` Block](#s3_destination-block) below.
* `tags` - (Optional) Map of tags to assign to the channel. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `topic_configuration` Block

The following arguments are required:

* `record_converter` - (Required) Configuration that controls how Apache Kafka record values are deserialized for the destination. See [`record_converter` Block](#record_converter-block) below.
* `topic_arn` - (Required) Amazon Resource Name (ARN) that uniquely identifies the topic.

The following arguments are optional:

* `record_schema` - (Optional) Schema used to validate records when the value converter requires one. See [`record_schema` Block](#record_schema-block) below.

### `record_converter` Block

* `value_converter` - (Required) Deserialization format applied to Apache Kafka record values. Valid values are `BYTE_ARRAY`, `STRING`, `JSON`, and `JSON_SCHEMA_GSR`. The `iceberg_destination` accepts only `JSON` or `JSON_SCHEMA_GSR`; the `s3_destination` accepts `BYTE_ARRAY`, `STRING`, or `JSON`.

### `record_schema` Block

* `gsr_arn` - (Required) Amazon Resource Name (ARN) of the AWS Glue Schema Registry schema used to validate records for the destination Apache Iceberg table.

### `s3_destination` Block

The following arguments are required:

* `dead_letter_queue_s3` - (Required) Amazon S3 bucket and prefix where MSK writes records that fail to deliver. See [`dead_letter_queue_s3` Block](#dead_letter_queue_s3-block) below.
* `service_execution_role_arn` - (Required) Amazon Resource Name (ARN) of the IAM role that MSK assumes to write to the destination Amazon S3 bucket and the dead-letter bucket.
* `storage` - (Required) Amazon S3 bucket, prefix, and storage class for delivered records. See [`storage` Block](#storage-block) below.

The following arguments are optional:

* `data_freshness_in_seconds` - (Optional) Maximum time, in seconds, that records buffer in MSK before being flushed to the destination. Valid values are between `300` and `900`. Defaults to `600`. Can be updated in place without recreating the channel.

### `storage` Block

The following arguments are required:

* `bucket_arn` - (Required) Amazon Resource Name (ARN) of the destination Amazon S3 bucket.
* `compression_type` - (Required) Compression codec applied to delivered Amazon S3 objects.
* `storage_class` - (Required) Amazon S3 storage class for delivered objects.

The following arguments are optional:

* `expected_bucket_owner` - (Optional) 12-digit AWS account ID expected to own the Amazon S3 bucket.
* `output_key_template` - (Optional) Template that controls the Amazon S3 object key for each delivered record.
* `output_prefix` - (Optional) Prefix prepended to every Amazon S3 object key written by the channel.

### `iceberg_destination` Block

The following arguments are required:

* `append_only` - (Required) Whether the destination is append-only. Must be `true`; updates and deletes are not supported.
* `dead_letter_queue_s3` - (Required) Amazon S3 bucket and prefix where MSK writes records that fail to deliver. See [`dead_letter_queue_s3` Block](#dead_letter_queue_s3-block) below.
* `destination_table` - (Required) Destination Iceberg table. See [`destination_table` Block](#destination_table-block) below.
* `schema_evolution` - (Required) Configuration controlling whether the destination table's schema is evolved to match incoming records. See [`schema_evolution` Block](#schema_evolution-block) below.
* `service_execution_role_arn` - (Required) Amazon Resource Name (ARN) of the IAM role that MSK assumes to access the destination table, the AWS Glue Data Catalog, and the dead-letter Amazon S3 bucket.
* `table_creation` - (Required) Configuration controlling whether MSK creates the destination table if it does not already exist. See [`table_creation` Block](#table_creation-block) below.

The following arguments are optional:

* `catalog` - (Optional) AWS Glue Data Catalog and S3 Tables warehouse used by the destination. See [`catalog` Block](#catalog-block) below.
* `compression_type` - (Optional) Compression codec for Iceberg table data files. Defaults to `ZSTD`.
* `data_freshness_in_seconds` - (Optional) Maximum time, in seconds, that records buffer in MSK before being flushed to the destination. Valid values are between `300` and `900`. Defaults to `600`. Can be updated in place without recreating the channel.

### `destination_table` Block

* `destination_database_name` - (Optional) Name of the destination namespace (database) in the AWS Glue Data Catalog.
* `destination_table_name` - (Optional) Name of the destination Apache Iceberg table.
* `partition_spec` - (Optional) Partition specification for the destination table. See [`partition_spec` Block](#partition_spec-block) below.

### `partition_spec` Block

* `partition_strategy` - (Required) Partitioning strategy applied to records written to the table. `TIME_HOUR` partitions by hour using a timestamp source column.
* `source` - (Required) Source column used by the partitioning strategy. For `TIME_HOUR`, exactly one source must be specified and its column must be a timestamp. See [`source` Block](#source-block) below.

### `source` Block

* `source_name` - (Required) Name of the source column. For `TIME_HOUR` partitioning this must be a timestamp column defined in the Glue Schema Registry schema.

### `schema_evolution` Block

* `enable_schema_evolution` - (Optional) Whether to allow MSK to evolve the destination table's schema.

### `table_creation` Block

* `enable_table_creation` - (Optional) Whether MSK creates the destination table on the customer's behalf.

### `catalog` Block

* `catalog_arn` - (Optional) Amazon Resource Name (ARN) of the federated AWS Glue Data Catalog that projects the S3 Tables bucket.
* `warehouse_location` - (Optional) Amazon Resource Name (ARN) of the S3 Tables bucket that backs the Apache Iceberg warehouse.

### `dead_letter_queue_s3` Block

The following arguments are required:

* `bucket_arn` - (Required) Amazon Resource Name (ARN) of the dead-letter Amazon S3 bucket.

The following arguments are optional:

* `error_output_prefix` - (Optional) Prefix prepended to every dead-letter Amazon S3 object key.
* `expected_bucket_owner` - (Optional) 12-digit AWS account ID expected to own the dead-letter Amazon S3 bucket.

### `encryption_configuration` Block

* `kms_key_arn` - (Required) Amazon Resource Name (ARN) of the AWS KMS key used to encrypt the data.

### `logging_info` Block

* `cloudwatch_logs` - (Optional) CloudWatch Logs destination for channel logs. See [`cloudwatch_logs` Block](#cloudwatch_logs-block) below.
* `firehose` - (Optional) Kinesis Data Firehose delivery stream destination for channel logs. See [`firehose` Block](#firehose-block) below.
* `s3` - (Optional) Amazon S3 destination for channel logs. See [`s3` Block](#s3-block) below.

### `cloudwatch_logs` Block

The following arguments are required:

* `enabled` - (Required) Whether the CloudWatch Logs destination is enabled.

The following arguments are optional:

* `log_group` - (Optional) Name of the CloudWatch log group that receives the logs.

### `firehose` Block

The following arguments are required:

* `enabled` - (Required) Whether the Firehose destination is enabled.

The following arguments are optional:

* `delivery_stream` - (Optional) Name of the Kinesis Data Firehose delivery stream that receives the logs.

### `s3` Block

The following arguments are required:

* `enabled` - (Required) Whether the Amazon S3 destination is enabled.

The following arguments are optional:

* `bucket` - (Optional) Name of the Amazon S3 bucket that receives the logs.
* `prefix` - (Optional) Prefix applied to the Amazon S3 log object keys.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the channel.
* `cluster_operation_arn` - ARN of the in-flight cluster operation.
* `creation_time` - Time when the channel was created.
* `destination_type` - Type of destination configured for the channel.
* `status` - Current lifecycle state of the channel.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_msk_channel.example
  identity = {
    arn         = "arn:aws:kafka:us-west-2:123456789012:channel/example/279c0212-d057-4dba-9aa9-1c4e5a25bfc7-3/a1b2c3d4"
    cluster_arn = "arn:aws:kafka:us-west-2:123456789012:cluster/example/279c0212-d057-4dba-9aa9-1c4e5a25bfc7-3"
  }
}

resource "aws_msk_channel" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `arn` (String) ARN of the channel.
* `cluster_arn` (String) Amazon Resource Name (ARN) that uniquely identifies the cluster.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Managed Streaming for Kafka Channel using the `arn` and `cluster_arn`. For example:

```terraform
import {
  to = aws_msk_channel.example
  id = "arn:aws:kafka:us-west-2:123456789012:channel/example/279c0212-d057-4dba-9aa9-1c4e5a25bfc7-3/a1b2c3d4,arn:aws:kafka:us-west-2:123456789012:cluster/example/279c0212-d057-4dba-9aa9-1c4e5a25bfc7-3"
}
```

Using `terraform import`, import Managed Streaming for Kafka Channel using the `arn` and `cluster_arn`. For example:

```console
% terraform import aws_msk_channel.example arn:aws:kafka:us-west-2:123456789012:channel/example/279c0212-d057-4dba-9aa9-1c4e5a25bfc7-3/a1b2c3d4,arn:aws:kafka:us-west-2:123456789012:cluster/example/279c0212-d057-4dba-9aa9-1c4e5a25bfc7-3
```
