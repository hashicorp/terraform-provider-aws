---
subcategory: "Kinesis"
layout: "aws"
page_title: "AWS: aws_kinesis_stream"
description: |-
  Provides a AWS Kinesis Stream
---

# Resource: aws_kinesis_stream

Provides a Kinesis Stream resource. Amazon Kinesis is a managed service that
scales elastically for real-time processing of streaming big data.

For more details, see the [Amazon Kinesis Documentation](https://aws.amazon.com/documentation/kinesis/).

## Example Usage

```terraform
resource "aws_kinesis_stream" "test_stream" {
  name             = "terraform-kinesis-test"
  shard_count      = 1
  retention_period = 48

  shard_level_metrics = [
    "IncomingBytes",
    "OutgoingBytes",
  ]

  stream_mode_details {
    stream_mode = "PROVISIONED"
  }

  tags = {
    Environment = "test"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `name` - (Required) A name to identify the stream. This is unique to the AWS account and region the Stream is created in.
* `encryption_type` - (Optional) The encryption type to use. The only acceptable values are `NONE` or `KMS`. The default value is `NONE`.
* `enforce_consumer_deletion` - (Optional) A boolean that indicates all registered consumers should be deregistered from the stream so that the stream can be destroyed without error. The default value is `false`.
* `kms_key_id` - (Optional) The identifier for the customer-managed KMS key to use for encryption. This can be a Key ID (UUID), a Key ARN, an Alias Name (prefixed with `alias/`), or an Alias ARN. You can also use a master key owned by Kinesis Data Streams by specifying the alias `aws/kinesis`.
* `max_record_size_in_kib` - (Optional) The maximum size for a single data record in KiB. The minimum value is 1024. The maximum value is 10240.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `retention_period` - (Optional) Length of time data records are accessible after they are added to the stream. The maximum value of a stream's retention period is 8760 hours. Minimum value is 24. Default is 24.
* `shard_count` - (Optional) The number of shards that the stream will use. If the `stream_mode` is `PROVISIONED`, this field is required. Amazon has guidelines for specifying the Stream size that should be referenced when creating a Kinesis stream. See [Amazon Kinesis Streams](https://docs.aws.amazon.com/kinesis/latest/dev/amazon-kinesis-streams.html) for more.
* `shard_level_metrics` - (Optional) A list of shard-level CloudWatch metrics which can be enabled for the stream. See [Monitoring with CloudWatch](https://docs.aws.amazon.com/streams/latest/dev/monitoring-with-cloudwatch.html) for more. Note that the value ALL should not be used; instead you should provide an explicit list of metrics you wish to enable.
* `stream_mode_details` - (Optional) Indicates the [capacity mode](https://docs.aws.amazon.com/streams/latest/dev/how-do-i-size-a-stream.html) of the data stream. Detailed below.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `warm_throughput_mib_ps` - (Optional) Target warm throughput in MB/s that the stream should be scaled to handle.

### stream_mode_details Configuration Block

* `stream_mode` - (Required) Specifies the capacity mode of the stream. Must be either `PROVISIONED` or `ON_DEMAND`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The unique stream ID.
* `arn` - The Amazon Resource Name (ARN) specifying the stream (same as `id`).
* `name` - The unique stream name.
* `shard_count` - The count of shards for this stream.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `5m`)
- `update` - (Default `120m`)
- `delete` - (Default `120m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_kinesis_stream.example
  identity = {
    name = "example-stream"
  }
}

resource "aws_kinesis_stream" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `name` (String) Name of the stream.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Kinesis Streams using `name`. For example:

```terraform
import {
  to = aws_kinesis_stream.example
  id = "example-stream"
}
```

Using `terraform import`, import Kinesis Streams using `name`. For example:

```console
% terraform import aws_kinesis_stream.example example-stream
```
