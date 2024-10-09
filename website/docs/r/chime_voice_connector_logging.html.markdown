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
  enable_sip_logs          = true
  enable_media_metric_logs = true
  voice_connector_id       = aws_chime_voice_connector.default.id
}
```

## Argument Reference

This resource supports the following arguments:

* `voice_connector_id` - (Required) The Amazon Chime Voice Connector ID.
* `enable_sip_logs` - (Optional) When true, enables SIP message logs for sending to Amazon CloudWatch Logs.
* `enable_media_metric_logs` - (Optional) When true, enables logging of detailed media metrics for Voice Connectors to Amazon CloudWatch logs.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The Amazon Chime Voice Connector ID.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Chime Voice Connector Logging using the `voice_connector_id`. For example:

```terraform
import {
  to = aws_chime_voice_connector_logging.default
  id = "abcdef1ghij2klmno3pqr4"
}
```

Using `terraform import`, import Chime Voice Connector Logging using the `voice_connector_id`. For example:

```console
% terraform import aws_chime_voice_connector_logging.default abcdef1ghij2klmno3pqr4
```
