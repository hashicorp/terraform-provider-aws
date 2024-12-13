---
subcategory: "Chime"
layout: "aws"
page_title: "AWS: aws_chime_voice_connector_streaming"
description: |-
    The streaming configuration associated with an Amazon Chime Voice Connector. Specifies whether media streaming is enabled for sending to Amazon Kinesis, and shows the retention period for the Amazon Kinesis data, in hours.
---

# Resource: aws_chime_voice_connector_streaming

Adds a streaming configuration for the specified Amazon Chime Voice Connector. The streaming configuration specifies whether media streaming is enabled for sending to Amazon Kinesis.
It also sets the retention period, in hours, for the Amazon Kinesis data.

## Example Usage

```terraform
resource "aws_chime_voice_connector" "default" {
  name               = "vc-name-test"
  require_encryption = true
}

resource "aws_chime_voice_connector_streaming" "default" {
  disabled                       = false
  voice_connector_id             = aws_chime_voice_connector.default.id
  data_retention                 = 7
  streaming_notification_targets = ["SQS"]
}
```

### Example Usage With Media Insights

```terraform
resource "aws_chime_voice_connector" "default" {
  name               = "vc-name-test"
  require_encryption = true
}

resource "aws_chime_voice_connector_streaming" "default" {
  disabled                       = false
  voice_connector_id             = aws_chime_voice_connector.default.id
  data_retention                 = 7
  streaming_notification_targets = ["SQS"]
  media_insights_configuration {
    disabled          = false
    configuration_arn = aws_chimesdkmediapipelines_media_insights_pipeline_configuration.example.arn
  }
}

resource "aws_chimesdkmediapipelines_media_insights_pipeline_configuration" "example" {
  name                     = "ExampleConfig"
  resource_access_role_arn = aws_iam_role.example.arn
  elements {
    type = "AmazonTranscribeCallAnalyticsProcessor"
    amazon_transcribe_call_analytics_processor_configuration {
      language_code = "en-US"
    }
  }

  elements {
    type = "KinesisDataStreamSink"
    kinesis_data_stream_sink_configuration {
      insights_target = aws_kinesis_stream.example.arn
    }
  }
}

data "aws_iam_policy_document" "assume_role" {
  statement {
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["mediapipelines.chime.amazonaws.com"]
    }

    actions = ["sts:AssumeRole"]
  }
}

resource "aws_iam_role" "example" {
  name               = "ExampleResourceAccessRole"
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

resource "aws_kinesis_stream" "example" {
  name        = "ExampleStream"
  shard_count = 2
}
```

## Argument Reference

This resource supports the following arguments:

* `voice_connector_id` - (Required) The Amazon Chime Voice Connector ID.
* `data_retention`  - (Required) The retention period, in hours, for the Amazon Kinesis data.
* `disabled` - (Optional) When true, media streaming to Amazon Kinesis is turned off. Default: `false`
* `streaming_notification_targets` - (Optional) The streaming notification targets. Valid Values: `EventBridge | SNS | SQS`
* `media_insights_configuration` - (Optional) The media insights configuration. See [`media_insights_configuration`](#media_insights_configuration).

### media_insights_configuration

* `disabled` - (Optional) When `true`, the media insights configuration is not enabled. Defaults to `false`.
* `configuration_arn` - (Optional) The media insights configuration that will be invoked by the Voice Connector.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The Amazon Chime Voice Connector ID.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Chime Voice Connector Streaming using the `voice_connector_id`. For example:

```terraform
import {
  to = aws_chime_voice_connector_streaming.default
  id = "abcdef1ghij2klmno3pqr4"
}
```

Using `terraform import`, import Chime Voice Connector Streaming using the `voice_connector_id`. For example:

```console
% terraform import aws_chime_voice_connector_streaming.default abcdef1ghij2klmno3pqr4
```
