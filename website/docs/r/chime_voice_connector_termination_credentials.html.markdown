---
subcategory: "Chime"
layout: "aws"
page_title: "AWS: aws_chime_voice_connector_termination_credentials"
description: |-
    Adds termination SIP credentials for the specified Amazon Chime Voice Connector.
---

# Resource: aws_chime_voice_connector_termination_credentials

Adds termination SIP credentials for the specified Amazon Chime Voice Connector.

~> **Note:** Voice Connector Termination Credentials requires a [Voice Connector Termination](/docs/providers/aws/r/chime_voice_connector_termination.html) to be present. Use of `depends_on` (as shown below) is recommended to avoid race conditions.

## Example Usage

```terraform
resource "aws_chime_voice_connector" "default" {
  name               = "test"
  require_encryption = true
}

resource "aws_chime_voice_connector_termination" "default" {
  disabled           = true
  cps_limit          = 1
  cidr_allow_list    = ["50.35.78.96/31"]
  calling_regions    = ["US", "CA"]
  voice_connector_id = aws_chime_voice_connector.default.id
}

resource "aws_chime_voice_connector_termination_credentials" "default" {
  voice_connector_id = aws_chime_voice_connector.default.id

  credentials {
    username = "test"
    password = "test!"
  }

  depends_on = [aws_chime_voice_connector_termination.default]
}
```

## Argument Reference

This resource supports the following arguments:

* `voice_connector_id` - (Required) Amazon Chime Voice Connector ID.
* `credentials` - (Required) List of termination SIP credentials.

### `credentials`

The SIP credentials used to authenticate requests to your Amazon Chime Voice Connector.

* `username` - (Required) RFC2617 compliant username associated with the SIP credentials.
* `password` - (Required) RFC2617 compliant password associated with the SIP credentials.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Amazon Chime Voice Connector ID.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Chime Voice Connector Termination Credentials using the `voice_connector_id`. For example:

```terraform
import {
  to = aws_chime_voice_connector_termination_credentials.default
  id = "abcdef1ghij2klmno3pqr4"
}
```

Using `terraform import`, import Chime Voice Connector Termination Credentials using the `voice_connector_id`. For example:

```console
% terraform import aws_chime_voice_connector_termination_credentials.default abcdef1ghij2klmno3pqr4
```
