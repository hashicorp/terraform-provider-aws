---
subcategory: "Chime"
layout: "aws"
page_title: "AWS: aws_chime_voice_connector_logging"
description: |-
    Adds a logging configuration for the specified Amazon Chime Voice Connector. The logging configuration specifies whether SIP message logs are enabled for sending to Amazon CloudWatch Logs.
---

# Resource: aws_chime_voice_connector_logging

Adds a logging configuration for the specified Amazon Chime Voice Connector. The logging configuration specifies whether SIP message logs are enabled for sending to Amazon CloudWatch Logs.

## Example Usage

```terraform
resource "aws_chime_voice_connector" "default" {
  name               = "vc-name-test"
  require_encryption = true
}

resource "aws_chime_voice_connector_logging" "default" {
  enable_sip_logs    = true
  voice_connector_id = aws_chime_voice_connector.default.id
}
```

## Argument Reference

The following arguments are supported:

* `voice_connector_id` - (Required) The Amazon Chime Voice Connector ID.
* `enable_sip_logs` - (Optional) When true, enables SIP message logs for sending to Amazon CloudWatch Logs.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The Amazon Chime Voice Connector ID.

## Import

Chime Voice Connector Logging can be imported using the `voice_connector_id`, e.g.,

```
$ terraform import aws_chime_voice_connector_logging.default abcdef1ghij2klmno3pqr4
```
