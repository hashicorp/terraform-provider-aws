---
subcategory: "Chime SDK Voice"
layout: "aws"
page_title: "AWS: aws_chimesdkvoice_global_settings"
description: |-
  Updates global settings for the Amazon Chime SDK Voice Connectors in an AWS account.
---

# Resource: aws_chimesdkvoice_global_settings

Updates global settings for the Amazon Chime SDK Voice Connectors in an AWS account.

```terraform
resource "aws_chimesdkvoice_global_settings" "test" {
	voice_connector {
		cdr_bucket = "test-bucket-name"
    }
}
```

## Argument Reference

The following arguments are supported:

* `voice_connector` - (Required) The Voice Connector settings. See [voice_connector](#voice_connector) below for more details

### `voice_connector`

The Amazon Chime SDK Voice Connector settings. Includes any Amazon S3 buckets designated for storing call detail records.

* `cdr_bucket` - (Optional) The S3 bucket that stores the Voice Connector's call detail records.