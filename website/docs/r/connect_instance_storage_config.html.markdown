---
subcategory: "Connect"
layout: "aws"
page_title: "AWS: aws_connect_instance_storage_config"
description: |-
  Provides details about a specific Amazon Connect Instance Storage Config.
---

# Resource: aws_connect_instance_storage_config

Provides an Amazon Connect Instance Storage Config resource. For more information see
[Amazon Connect: Getting Started](https://docs.aws.amazon.com/connect/latest/adminguide/amazon-connect-get-started.html)

## Example Usage

### Storage Config Kinesis Firehose Config

```terraform
resource "aws_connect_instance_storage_config" "example" {
  instance_id   = aws_connect_instance.example.id
  resource_type = "CONTACT_TRACE_RECORDS"

  storage_config {
    kinesis_firehose_config {
      firehose_arn = aws_kinesis_firehose_delivery_stream.example.arn
    }
    storage_type = "KINESIS_FIREHOSE"
  }
}
```

### Storage Config Kinesis Stream Config

```terraform
resource "aws_connect_instance_storage_config" "example" {
  instance_id   = aws_connect_instance.example.id
  resource_type = "CONTACT_TRACE_RECORDS"

  storage_config {
    kinesis_stream_config {
      stream_arn = aws_kinesis_stream.example.arn
    }
    storage_type = "KINESIS_STREAM"
  }
}
```

### Storage Config Kinesis Video Stream Config

```terraform
resource "aws_connect_instance_storage_config" "example" {
  instance_id   = aws_connect_instance.example.id
  resource_type = "MEDIA_STREAMS"

  storage_config {
    kinesis_video_stream_config {
      prefix                 = "example"
      retention_period_hours = 3

      encryption_config {
        encryption_type = "KMS"
        key_id          = aws_kms_key.example.arn
      }
    }
    storage_type = "KINESIS_VIDEO_STREAM"
  }
}
```

### Storage Config S3 Config

```terraform
resource "aws_connect_instance_storage_config" "example" {
  instance_id   = aws_connect_instance.example.id
  resource_type = "CHAT_TRANSCRIPTS"

  storage_config {
    s3_config {
      bucket_name   = aws_s3_bucket.example.id
      bucket_prefix = "example"
    }
    storage_type = "S3"
  }
}
```

### Storage Config S3 Config with Encryption Config

```terraform
resource "aws_connect_instance_storage_config" "example" {
  instance_id   = aws_connect_instance.example.id
  resource_type = "CHAT_TRANSCRIPTS"

  storage_config {
    s3_config {
      bucket_name   = aws_s3_bucket.example.id
      bucket_prefix = "example"

      encryption_config {
        encryption_type = "KMS"
        key_id          = aws_kms_key.example.arn
      }
    }
    storage_type = "S3"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `instance_id` - (Required) Specifies the identifier of the hosting Amazon Connect Instance.
* `resource_type` - (Required) A valid resource type. Valid Values: `AGENT_EVENTS` | `ATTACHMENTS` | `CALL_RECORDINGS` | `CHAT_TRANSCRIPTS` | `CONTACT_EVALUATIONS` | `CONTACT_TRACE_RECORDS` | `MEDIA_STREAMS` | `REAL_TIME_CONTACT_ANALYSIS_SEGMENTS` | `SCHEDULED_REPORTS`.
* `storage_config` - (Required) Specifies the storage configuration options for the Connect Instance. [Documented below](#storage_config).

### `storage_config`

The `storage_config` configuration block supports the following arguments:

* `kinesis_firehose_config` - (Required if `type` is set to `KINESIS_FIREHOSE`) A block that specifies the configuration of the Kinesis Firehose delivery stream. [Documented below](#kinesis_firehose_config).
* `kinesis_stream_config` - (Required if `type` is set to `KINESIS_STREAM`) A block that specifies the configuration of the Kinesis data stream. [Documented below](#kinesis_stream_config).
* `kinesis_video_stream_config` - (Required if `type` is set to `KINESIS_VIDEO_STREAM`) A block that specifies the configuration of the Kinesis video stream. [Documented below](#kinesis_video_stream_config).
* `s3_config` - (Required if `type` is set to `S3`) A block that specifies the configuration of S3 Bucket. [Documented below](#s3_config).
* `storage_type` - (Required) A valid storage type. Valid Values: `S3` | `KINESIS_VIDEO_STREAM` | `KINESIS_STREAM` | `KINESIS_FIREHOSE`.

#### `kinesis_firehose_config`

The `kinesis_firehose_config` configuration block supports the following arguments:

* `firehose_arn` - (Required) The Amazon Resource Name (ARN) of the delivery stream.

#### `kinesis_stream_config`

The `kinesis_stream_config` configuration block supports the following arguments:

* `stream_arn` - (Required) The Amazon Resource Name (ARN) of the data stream.

#### `kinesis_video_stream_config`

The `kinesis_video_stream_config` configuration block supports the following arguments:

* `encryption_config` - (Required) The encryption configuration. [Documented below](#encryption_config).
* `prefix` - (Required) The prefix of the video stream. Minimum length of `1`. Maximum length of `128`. When read from the state, the value returned is `<prefix>-connect-<connect_instance_alias>-contact-` since the API appends additional details to the `prefix`.
* `retention_period_hours` - (Required) The number of hours data is retained in the stream. Kinesis Video Streams retains the data in a data store that is associated with the stream. Minimum value of `0`. Maximum value of `87600`. A value of `0`, indicates that the stream does not persist data.

#### `s3_config`

The `s3_config` configuration block supports the following arguments:

* `bucket_name` - (Required) The S3 bucket name.
* `bucket_prefix` - (Required) The S3 bucket prefix.
* `encryption_config` - (Optional) The encryption configuration. [Documented below](#encryption_config).

#### `encryption_config`

The `encryption_config` configuration block supports the following arguments:

* `encryption_type` - (Required) The type of encryption. Valid Values: `KMS`.
* `key_id` - (Required) The full ARN of the encryption key. Be sure to provide the full ARN of the encryption key, not just the ID.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `association_id` - The existing association identifier that uniquely identifies the resource type and storage config for the given instance ID.
* `id` - The identifier of the hosting Amazon Connect Instance, `association_id`, and `resource_type` separated by a colon (`:`).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Amazon Connect Instance Storage Configs using the `instance_id`, `association_id`, and `resource_type` separated by a colon (`:`). For example:

```terraform
import {
  to = aws_connect_instance_storage_config.example
  id = "f1288a1f-6193-445a-b47e-af739b2:c1d4e5f6-1b3c-1b3c-1b3c-c1d4e5f6c1d4e5:CHAT_TRANSCRIPTS"
}
```

Using `terraform import`, import Amazon Connect Instance Storage Configs using the `instance_id`, `association_id`, and `resource_type` separated by a colon (`:`). For example:

```console
% terraform import aws_connect_instance_storage_config.example f1288a1f-6193-445a-b47e-af739b2:c1d4e5f6-1b3c-1b3c-1b3c-c1d4e5f6c1d4e5:CHAT_TRANSCRIPTS
```
