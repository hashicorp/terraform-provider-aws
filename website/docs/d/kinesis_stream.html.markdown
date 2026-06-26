---
subcategory: "Kinesis"
layout: "aws"
page_title: "AWS: aws_kinesis_stream"
description: |-
  Provides a Kinesis Stream data source.
---

# Data Source: aws_kinesis_stream

Use this data source to get information about a Kinesis Stream for use in other
resources.

For more details, see the [Amazon Kinesis Documentation](https://aws.amazon.com/documentation/kinesis/).

## Example Usage

```terraform
data "aws_kinesis_stream" "stream" {
  name = "stream-name"
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `name` - (Required) Name of the Kinesis Stream.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - ARN of the Kinesis Stream.
* `arn` - ARN of the Kinesis Stream (same as `id`).
* `closed_shards` - List of shard ids in the CLOSED state. See [Shard State](https://docs.aws.amazon.com/streams/latest/dev/kinesis-using-sdk-java-after-resharding.html#kinesis-using-sdk-java-resharding-data-routing) for more.
* `creation_timestamp` - Approximate UNIX timestamp that the stream was created.
* `encryption_type` - Encryption type used.
* `kms_key_id` - The identifier for the customer-managed KMS key to use for encryption. This can be a Key ID (UUID), a Key ARN, an Alias Name (prefixed with `alias/`), or an Alias ARN.
* `max_record_size_in_kib` - The maximum size for a single data record in KiB.
* `name` - Name of the Kinesis Stream.
* `open_shards` - List of shard ids in the OPEN state. See [Shard State](https://docs.aws.amazon.com/streams/latest/dev/kinesis-using-sdk-java-after-resharding.html#kinesis-using-sdk-java-resharding-data-routing) for more.
* `retention_period` - Length of time (in hours) data records are accessible after they are added to the stream.
* `shard_level_metrics` - List of shard-level CloudWatch metrics which are enabled for the stream. See [Monitoring with CloudWatch](https://docs.aws.amazon.com/streams/latest/dev/monitoring-with-cloudwatch.html) for more.
* `status` - Current status of the stream. The stream status is one of CREATING, DELETING, ACTIVE, or UPDATING.
* `stream_mode_details` - [Capacity mode](https://docs.aws.amazon.com/streams/latest/dev/how-do-i-size-a-stream.html) of the data stream. Detailed below.
* `tags` - Map of tags to assigned to the stream.
* `warm_throughput` - Warm throughput in MB/s for the stream. Detailed below.

### stream_mode_details Configuration Block

* `stream_mode` - Capacity mode of the stream. Either `ON_DEMAND` or `PROVISIONED`.

### warm_throughput Configuration Block

* `current_mib_ps` - Current warm throughput value on the stream.
* `target_mib_ps` - Target warm throughput value on the stream.
