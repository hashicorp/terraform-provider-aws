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

### Basic Usage

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

### Logging Custom Data From a CloudFront Function

A viewer request or viewer response [CloudFront Function](cloudfront_function.html) can write arbitrary data into the log record for the request by calling the `cf.logCustomData()` helper method, which requires JavaScript runtime 2.0 (`cloudfront-js-2.0`) and the `cloudfront` module. The data is surfaced in the `viewer-request-log-data` and `viewer-response-log-data` fields, which must be selected in the real-time log configuration for them to be delivered.

```terraform
resource "aws_cloudfront_function" "example" {
  name    = "tag-request"
  runtime = "cloudfront-js-2.0"
  publish = true

  code = <<-EOT
    import cf from 'cloudfront';

    function handler(event) {
      var variant = event.request.uri.indexOf("/beta") === 0 ? "b" : "a";
      cf.logCustomData("variant=" + variant);
      return event.request;
    }
  EOT
}

resource "aws_cloudfront_realtime_log_config" "example" {
  name          = "example"
  sampling_rate = 100

  fields = [
    "timestamp",
    "c-ip",
    "sc-status",
    "viewer-request-log-data",
    "viewer-response-log-data",
  ]

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

Associate the function with a cache behavior on the distribution (`function_association` with `event_type = "viewer-request"`) and attach the real-time log configuration to the same cache behavior via `realtime_log_config_arn`. The field is `-` for any request whose viewer request or viewer response function did not call `cf.logCustomData()`.

## Argument Reference

This resource supports the following arguments:

* `endpoint` - (Required) The Amazon Kinesis data streams where real-time log data is sent.
* `fields` - (Required) The fields that are included in each real-time log record. See the [AWS documentation](https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/real-time-logs.html#understand-real-time-log-config-fields) for supported values. This includes `viewer-request-log-data` and `viewer-response-log-data`, which carry the custom data that a CloudFront Function logs with `cf.logCustomData()`.
* `name` - (Required) The unique name to identify this real-time log configuration.
* `sampling_rate` - (Required) The sampling rate for this real-time log configuration. The sampling rate determines the percentage of viewer requests that are represented in the real-time log data. An integer between `1` and `100`, inclusive.

The `endpoint` object supports the following:

* `kinesis_stream_config` - (Required) The Amazon Kinesis data stream configuration.
* `stream_type` - (Required) The type of data stream where real-time log data is sent. The only valid value is `Kinesis`.

The `kinesis_stream_config` object supports the following:

* `role_arn` - (Required) The ARN of an [IAM role](iam_role.html) that CloudFront can use to send real-time log data to the Kinesis data stream.
See the [AWS documentation](https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/real-time-logs.html#understand-real-time-log-config-iam-role) for more information.
* `stream_arn` - (Required) The ARN of the [Kinesis data stream](kinesis_stream.html).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The ID of the CloudFront real-time log configuration.
* `arn` - The ARN (Amazon Resource Name) of the CloudFront real-time log configuration.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_cloudfront_realtime_log_config.example
  identity = {
    "arn" = "arn:aws:cloudfront::123456789012:realtime-log-config/ExampleNameForRealtimeLogConfig"
  }
}

resource "aws_cloudfront_realtime_log_config" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

- `arn` (String) Amazon Resource Name (ARN) of the CloudFront real-time log configuration.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import CloudFront real-time log configurations using the ARN. For example:

```terraform
import {
  to = aws_cloudfront_realtime_log_config.example
  id = "arn:aws:cloudfront::111122223333:realtime-log-config/ExampleNameForRealtimeLogConfig"
}
```

Using `terraform import`, import CloudFront real-time log configurations using the ARN. For example:

```console
% terraform import aws_cloudfront_realtime_log_config.example arn:aws:cloudfront::111122223333:realtime-log-config/ExampleNameForRealtimeLogConfig
```
