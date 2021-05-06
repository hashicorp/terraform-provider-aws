---
subcategory: "Chime"
layout: "aws"
page_title: "AWS: aws_chime_voice_connector_streaming"
description: |-
The streaming configuration associated with an Amazon Chime Voice Connector. Specifies whether media streaming is enabled for sending to Amazon Kinesis, and shows the retention period for the Amazon Kinesis data, in hours.
---

# Resource: aws_chime_voice_connector_streaming

Adds a streaming configuration for the specified Amazon Chime Voice Connector. The streaming configuration specifies whether media streaming is enabled for sending to Amazon Kinesis. It also sets the retention period, in hours, for the Amazon Kinesis data.

```terraform
resource "aws_chime_voice_connector" "default" {
  name               = "test"
  require_encryption = true
}

resource "aws_chime_voice_connector_streaming" "default" {
  disabled                       = false
  voice_connector_id             = aws_chime_voice_connector.default.id
  data_retention                 = 7
  streaming_notification_targets = ["SQS"]
}
```

## Argument Reference

The following arguments are supported:

* `voice_connector_id` - (Required) The Amazon Chime Voice Connector ID.
* `data_retention`  - (Required) The retention period, in hours, for the Amazon Kinesis data.
* `disabled` - (Optional) When true, media streaming to Amazon Kinesis is turned off. Default: `false`
* `streaming_notification_targets` - (Optional) The streaming notification targets. Valid Values: `EventBridge | SNS | SQS`

## Import

Configuration Recorder can be imported using the name, e.g.

```
$ terraform import aws_chime_voice_connector_streaming.default example
```