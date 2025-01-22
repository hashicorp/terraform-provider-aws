---
subcategory: "Connect"
layout: "aws"
page_title: "AWS: aws_connect_instance_storage_config"
description: |-
  Provides details about a specific Amazon Connect Instance Storage Config.
---

# Data Source: aws_connect_instance_storage_config

Provides details about a specific Amazon Connect Instance Storage Config.

## Example Usage

```terraform
data "aws_connect_instance_storage_config" "example" {
  association_id = "1234567891234567890122345678912345678901223456789123456789012234"
  instance_id    = "aaaaaaaa-bbbb-cccc-dddd-111111111111"
  resource_type  = "CONTACT_TRACE_RECORDS"
}
```

## Argument Reference

This data source supports the following arguments:

* `association_id` - (Required) The existing association identifier that uniquely identifies the resource type and storage config for the given instance ID.
* `instance_id` - (Required) Reference to the hosting Amazon Connect Instance
* `resource_type` - (Required) A valid resource type. Valid Values: `AGENT_EVENTS` | `ATTACHMENTS` | `CALL_RECORDINGS` | `CHAT_TRANSCRIPTS` | `CONTACT_EVALUATIONS` | `CONTACT_TRACE_RECORDS` | `MEDIA_STREAMS` | `REAL_TIME_CONTACT_ANALYSIS_SEGMENTS` | `SCHEDULED_REPORTS` |  `SCREEN_RECORDINGS`.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - The identifier of the hosting Amazon Connect Instance, `association_id`, and `resource_type` separated by a colon (`:`).
* `storage_config` - Specifies the storage configuration options for the Connect Instance. [Documented below](#storage_config).

### `storage_config`

The `storage_config` configuration block supports the following arguments:

* `kinesis_firehose_config` - A block that specifies the configuration of the Kinesis Firehose delivery stream. [Documented below](#kinesis_firehose_config).
* `kinesis_stream_config` - A block that specifies the configuration of the Kinesis data stream. [Documented below](#kinesis_stream_config).
* `kinesis_video_stream_config` - A block that specifies the configuration of the Kinesis video stream. [Documented below](#kinesis_video_stream_config).
* `s3_config` - A block that specifies the configuration of S3 Bucket. [Documented below](#s3_config).
* `storage_type` - A valid storage type. Valid Values: `S3` | `KINESIS_VIDEO_STREAM` | `KINESIS_STREAM` | `KINESIS_FIREHOSE`.

#### `kinesis_firehose_config`

The `kinesis_firehose_config` configuration block supports the following arguments:

* `firehose_arn` - The Amazon Resource Name (ARN) of the delivery stream.

#### `kinesis_stream_config`

The `kinesis_stream_config` configuration block supports the following arguments:

* `stream_arn` - The Amazon Resource Name (ARN) of the data stream.

#### `kinesis_video_stream_config`

The `kinesis_video_stream_config` configuration block supports the following arguments:

* `encryption_config` - The encryption configuration. [Documented below](#encryption_config).
* `prefix` - The prefix of the video stream. Minimum length of `1`. Maximum length of `128`. When read from the state, the value returned is `<prefix>-connect-<connect_instance_alias>-contact-` since the API appends additional details to the `prefix`.
* `retention_period_hours` - The number of hours to retain the data in a data store associated with the stream. Minimum value of `0`. Maximum value of `87600`. A value of `0` indicates that the stream does not persist data.

#### `s3_config`

The `s3_config` configuration block supports the following arguments:

* `bucket_name` - The S3 bucket name.
* `bucket_prefix` - The S3 bucket prefix.
* `encryption_config` - The encryption configuration. [Documented below](#encryption_config).

#### `encryption_config`

The `encryption_config` configuration block supports the following arguments:

* `encryption_type` - The type of encryption. Valid Values: `KMS`.
* `key_id` - The full ARN of the encryption key. Be sure to provide the full ARN of the encryption key, not just the ID.
