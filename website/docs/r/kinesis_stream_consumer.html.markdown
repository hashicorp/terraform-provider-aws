---
layout: "aws"
page_title: "AWS: aws_kinesis_stream_consumer"
sidebar_current: "docs-aws-resource-kinesis-stream-consumer"
description: |-
  Provides a AWS Kinesis Stream Consumer
---

# aws_kinesis_stream_consumer

Provides a Kinesis Stream Consumer resource. A consumer is an application that processes all data from a Kinesis data stream. 
When a consumer uses enhanced fan-out, it gets its own 2 MiB/sec allotment of read throughput, allowing multiple consumers to 
read data from the same stream in parallel, without contending for read throughput with other consumers. [Reading Data from Kinesis][1]

You can register up to 20 consumers per stream. However, you can request a limit increase using the [Kinesis Data Streams limits form][2]. A given consumer can only be registered with one stream at a time.

For more details, see the [Amazon Kinesis Stream Consumer Documentation][3].

## Example Usage

```hcl
resource "aws_kinesis_stream" "test_stream" {
  name             = "terraform-kinesis-test"
  shard_count      = 1
}

resource "aws_lambda_event_source_mapping" "example" {
  event_source_arn  = "${aws_kinesis_stream_consumer.test_stream_consumer.arn}"
  function_name     = "${aws_lambda_function.example.arn}"
  starting_position = "LATEST"
}

resource "aws_kinesis_stream_consumer" "test_stream_consumer" {
  name             = "terraform-kinesis-stream-consumer-test"
  stream_arn       = "${aws_kinesis_stream.test_stream.arn}"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A name to identify the stream. This is unique to the
AWS account and region the Stream Consumer is created in.
* `stream_arn` – (Required) The Amazon Resource Name (ARN) of the Kinesis Stream, the Consumer is connected to.

## Attributes Reference

* `id` - The unique Stream Consumer id
* `name` - The unique Stream Consumer name
* `arn` - The Amazon Resource Name (ARN) specifying the Stream Consumer (same as `id`)

## Timeouts

`aws_kinesis_stream_consumer` provides the following [Timeouts](/docs/configuration/resources.html#timeouts) configuration options:

- `create` - (Default `5 minutes`)  Used for Creating a Kinesis Stream Consumer
- `delete` - (Default `120 minutes`) Used for Destroying a Kinesis Stream Consumer

Kinesis Streams can be imported using the `name`, e.g.

```
$ terraform import aws_kinesis_stream_consumer.test_stream_consumer terraform-kinesis-stream-consumer-test
```

[1]: https://docs.aws.amazon.com/streams/latest/dev/building-consumers.html
[2]: https://console.aws.amazon.com/support/v1?#/
[3]: https://docs.aws.amazon.com/streams/latest/dev/introduction-to-enhanced-consumers.html