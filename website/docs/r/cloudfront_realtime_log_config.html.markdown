---
subcategory: "CloudFront"
layout: "aws"
page_title: "AWS: aws_cloudfront_realtime_log_config"
description: |-
  Provides a CloudFront real-time log configuration resource.
---

# Resource: aws_cloudfront_realtime_log_config

Provides a CloudFront real-time log configuration resource.

## Example Usage

```terraform
data "aws_iam_policy_document" "assume_role" {
  statement {
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["cloudfront.amazonaws.com"]
    }

    actions = ["sts:AssumeRole"]
  }
}

resource "aws_iam_role" "example" {
  name               = "cloudfront-realtime-log-config-example"
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

data "aws_iam_policy_document" "example" {
  statement {
    effect = "Allow"

    actions = [
      "kinesis:DescribeStreamSummary",
      "kinesis:DescribeStream",
      "kinesis:PutRecord",
      "kinesis:PutRecords",
    ]

    resources = [aws_kinesis_stream.example.arn]
  }
}

resource "aws_iam_role_policy" "example" {
  name   = "cloudfront-realtime-log-config-example"
  role   = aws_iam_role.example.id
  policy = data.aws_iam_policy_document.example.json
}

resource "aws_cloudfront_realtime_log_config" "example" {
  name          = "example"
  sampling_rate = 75
  fields        = ["timestamp", "c-ip"]

  endpoint {
    stream_type = "Kinesis"

    kinesis_stream_config {
      role_arn   = aws_iam_role.example.arn
      stream_arn = aws_kinesis_stream.example.arn
    }
  }

  depends_on = [aws_iam_role_policy.example]
}
```

## Argument Reference

The following arguments are supported:

* `endpoint` - (Required) The Amazon Kinesis data streams where real-time log data is sent.
* `fields` - (Required) The fields that are included in each real-time log record. See the [AWS documentation](https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/real-time-logs.html#understand-real-time-log-config-fields) for supported values.
* `name` - (Required) The unique name to identify this real-time log configuration.
* `sampling_rate` - (Required) The sampling rate for this real-time log configuration. The sampling rate determines the percentage of viewer requests that are represented in the real-time log data. An integer between `1` and `100`, inclusive.

The `endpoint` object supports the following:

* `kinesis_stream_config` - (Required) The Amazon Kinesis data stream configuration.
* `stream_type` - (Required) The type of data stream where real-time log data is sent. The only valid value is `Kinesis`.

The `kinesis_stream_config` object supports the following:

* `role_arn` - (Required) The ARN of an [IAM role](iam_role.html) that CloudFront can use to send real-time log data to the Kinesis data stream.
See the [AWS documentation](https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/real-time-logs.html#understand-real-time-log-config-iam-role) for more information.
* `stream_arn` - (Required) The ARN of the [Kinesis data stream](kinesis_stream.html).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the CloudFront real-time log configuration.
* `arn` - The ARN (Amazon Resource Name) of the CloudFront real-time log configuration.

## Import

CloudFront real-time log configurations can be imported using the ARN, e.g.,

```
$ terraform import aws_cloudfront_realtime_log_config.example arn:aws:cloudfront::111122223333:realtime-log-config/ExampleNameForRealtimeLogConfig
```
