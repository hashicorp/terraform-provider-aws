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
* `topic_configuration` - (Required) Configuration of the Apache Kafka topic that feeds the channel. Changing this forces a new resource to be created. [See below](#topic_configuration).

Exactly one of the following destination blocks is required:

* `iceberg_destination` - (Optional) Apache Iceberg destination for the channel. Changing this forces a new resource to be created. [See below](#iceberg_destination).
* `s3_destination` - (Optional) Amazon S3 destination for the channel. Changing this forces a new resource to be created. [See below](#s3_destination).

The following arguments are optional:

* `encryption_configuration` - (Optional) AWS KMS encryption configuration applied to data at rest. Changing this forces a new resource to be created. [See below](#encryption_configuration).
* `logging_info` - (Optional) Destinations to which the channel publishes operational logs. Changing this forces a new resource to be created. [See below](#logging_info).
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Map of tags to assign to the channel. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### topic_configuration

* `record_converter` - (Required) Configuration that controls how Apache Kafka record values are deserialized for the destination. [See below](#record_converter).
* `topic_arn` - (Required) Amazon Resource Name (ARN) that uniquely identifies the topic.
* `record_schema` - (Optional) Schema used to validate records when the value converter requires one. [See below](#record_schema).

### record_converter

* `value_converter` - (Required) Deserialization format applied to Apache Kafka record values. Valid values are `BYTE_ARRAY`, `STRING`, `JSON`, and `JSON_SCHEMA_GSR`. The `iceberg_destination` accepts only `JSON` or `JSON_SCHEMA_GSR`; the `s3_destination` accepts `BYTE_ARRAY`, `STRING`, or `JSON`.

### record_schema

* `gsr_arn` - (Required) Amazon Resource Name (ARN) of the AWS Glue Schema Registry schema used to validate records for the destination Apache Iceberg table.

### s3_destination

* `dead_letter_queue_s3` - (Required) Amazon S3 bucket and prefix where MSK writes records that fail to deliver. [See below](#dead_letter_queue_s3).
* `service_execution_role_arn` - (Required) Amazon Resource Name (ARN) of the IAM role that MSK assumes to write to the destination Amazon S3 bucket and the dead-letter bucket.
* `storage` - (Required) Amazon S3 bucket, prefix, and storage class for delivered records. [See below](#storage).
* `data_freshness_in_seconds` - (Optional) Maximum time, in seconds, that records buffer in MSK before being flushed to the destination. Valid values are between `300` and `900`. Defaults to `600`.

### storage

* `bucket_arn` - (Required) Amazon Resource Name (ARN) of the destination Amazon S3 bucket.
* `compression_type` - (Required) Compression codec applied to delivered Amazon S3 objects.
* `storage_class` - (Required) Amazon S3 storage class for delivered objects.
* `expected_bucket_owner` - (Optional) 12-digit AWS account ID expected to own the Amazon S3 bucket.
* `output_key_template` - (Optional) Template that controls the Amazon S3 object key for each delivered record.
* `output_prefix` - (Optional) Prefix prepended to every Amazon S3 object key written by the channel.

### iceberg_destination

* `append_only` - (Required) Whether the destination is append-only. Must be `true`; updates and deletes are not supported.
* `dead_letter_queue_s3` - (Required) Amazon S3 bucket and prefix where MSK writes records that fail to deliver. [See below](#dead_letter_queue_s3).
* `destination_table` - (Required) Destination Iceberg table. [See below](#destination_table).
* `schema_evolution` - (Required) Configuration controlling whether the destination table's schema is evolved to match incoming records. [See below](#schema_evolution).
* `service_execution_role_arn` - (Required) Amazon Resource Name (ARN) of the IAM role that MSK assumes to access the destination table, the AWS Glue Data Catalog, and the dead-letter Amazon S3 bucket.
* `table_creation` - (Required) Configuration controlling whether MSK creates the destination table if it does not already exist. [See below](#table_creation).
* `catalog` - (Optional) AWS Glue Data Catalog and S3 Tables warehouse used by the destination. [See below](#catalog).
* `compression_type` - (Optional) Compression codec for Iceberg table data files. Defaults to `ZSTD`.
* `data_freshness_in_seconds` - (Optional) Maximum time, in seconds, that records buffer in MSK before being flushed to the destination. Valid values are between `300` and `900`. Defaults to `600`.

### destination_table

* `destination_database_name` - (Optional) Name of the destination namespace (database) in the AWS Glue Data Catalog.
* `destination_table_name` - (Optional) Name of the destination Apache Iceberg table.
* `partition_spec` - (Optional) Partition specification for the destination table. [See below](#partition_spec).

### partition_spec

* `partition_strategy` - (Required) Partitioning strategy applied to records written to the table. `TIME_HOUR` partitions by hour using a timestamp source column.
* `source` - (Required) Source column used by the partitioning strategy. For `TIME_HOUR`, exactly one source must be specified and its column must be a timestamp. [See below](#source).

### source

* `source_name` - (Required) Name of the source column. For `TIME_HOUR` partitioning this must be a timestamp column defined in the Glue Schema Registry schema.

### schema_evolution

* `enable_schema_evolution` - (Optional) Whether to allow MSK to evolve the destination table's schema.

### table_creation

* `enable_table_creation` - (Optional) Whether MSK creates the destination table on the customer's behalf.

### catalog

* `catalog_arn` - (Optional) Amazon Resource Name (ARN) of the federated AWS Glue Data Catalog that projects the S3 Tables bucket.
* `warehouse_location` - (Optional) Amazon Resource Name (ARN) of the S3 Tables bucket that backs the Apache Iceberg warehouse.

### dead_letter_queue_s3

* `bucket_arn` - (Required) Amazon Resource Name (ARN) of the dead-letter Amazon S3 bucket.
* `error_output_prefix` - (Optional) Prefix prepended to every dead-letter Amazon S3 object key.
* `expected_bucket_owner` - (Optional) 12-digit AWS account ID expected to own the dead-letter Amazon S3 bucket.

### encryption_configuration

* `kms_key_arn` - (Required) Amazon Resource Name (ARN) of the AWS KMS key used to encrypt the data.

### logging_info

* `cloudwatch_logs` - (Optional) CloudWatch Logs destination for channel logs. [See below](#cloudwatch_logs).
* `firehose` - (Optional) Kinesis Data Firehose delivery stream destination for channel logs. [See below](#firehose).
* `s3` - (Optional) Amazon S3 destination for channel logs. [See below](#s3).

### cloudwatch_logs

* `enabled` - (Required) Whether the CloudWatch Logs destination is enabled.
* `log_group` - (Optional) Name of the CloudWatch log group that receives the logs.

### firehose

* `enabled` - (Required) Whether the Firehose destination is enabled.
* `delivery_stream` - (Optional) Name of the Kinesis Data Firehose delivery stream that receives the logs.

### s3

* `enabled` - (Required) Whether the Amazon S3 destination is enabled.
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
