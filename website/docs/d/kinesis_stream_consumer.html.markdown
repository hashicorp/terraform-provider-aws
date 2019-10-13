---
layout: "aws"
page_title: "AWS: aws_kinesis_stream_consumer"
sidebar_current: "docs-aws-datasource-kinesis-stream-consumer"
description: |-
  Provides a Kinesis Stream Consumer data source.
---

# Data Source: aws_kinesis_stream_consumer

Use this data source to get information about a Kinesis Stream Consumer for use in other
resources.

For more details, see the [Amazon Kinesis Stream Consumer Documentation][1].

## Example Usage

```hcl
data "aws_kinesis_stream_consumer" "stream_consumer" {
  name = "stream-consumer-name"
  stream_arn = "${aws_kinesis_stream.stream.arn}"
}
```

## Argument Reference

* `name` - (Required) The name of the Kinesis Stream Consumer.
* `stream_arn` - (Required) The Amazon Resource Name (ARN) of the Kinesis Stream.

## Attributes Reference

`id` is set to the Amazon Resource Name (ARN) of the Kinesis Stream Consumer. In addition, the following attributes
are exported:

* `arn` - The Amazon Resource Name (ARN) of the Kinesis Stream Consumer (same as id).
* `name` - The name of the Kinesis Stream Consumer.
* `creation_timestamp` - The approximate UNIX timestamp that the stream consumer was created.
* `status` - The current status of the stream consumer. The stream status is one of CREATING, DELETING, ACTIVE, or UPDATING.
* `stream_arn` - The Amazon Resource Name (ARN) of the Kinesis Stream.

[1]: https://docs.aws.amazon.com/streams/latest/dev/introduction-to-enhanced-consumers.html