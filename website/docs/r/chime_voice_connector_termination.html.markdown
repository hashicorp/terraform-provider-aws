---
subcategory: "Chime"
layout: "aws"
page_title: "AWS: aws_chime_voice_connector_termination"
description: |-
    Enable Termination settings to control outbound calling from your SIP infrastructure.
---

# Resource: aws_chime_voice_connector_termination

Enable Termination settings to control outbound calling from your SIP infrastructure.

## Example Usage

```terraform
resource "aws_chime_voice_connector" "default" {
  name               = "vc-name-test"
  require_encryption = true
}

resource "aws_chime_voice_connector_termination" "default" {
  disabled           = false
  cps_limit          = 1
  cidr_allow_list    = ["50.35.78.96/31"]
  calling_regions    = ["US", "CA"]
  voice_connector_id = aws_chime_voice_connector.default.id
}
```

## Argument Reference

This resource supports the following arguments:

* `voice_connector_id` - (Required) The Amazon Chime Voice Connector ID.
* `cidr_allow_list` - (Required) The IP addresses allowed to make calls, in CIDR format.
* `calling_regions` - (Required) The countries to which calls are allowed, in ISO 3166-1 alpha-2 format.
* `disabled` - (Optional) When termination settings are disabled, outbound calls can not be made.
* `default_phone_number` - (Optional) The default caller ID phone number.
* `cps_limit` - (Optional) The limit on calls per second. Max value based on account service quota. Default value of `1`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The Amazon Chime Voice Connector ID.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Chime Voice Connector Termination using the `voice_connector_id`. For example:

```terraform
import {
  to = aws_chime_voice_connector_termination.default
  id = "abcdef1ghij2klmno3pqr4"
}
```

Using `terraform import`, import Chime Voice Connector Termination using the `voice_connector_id`. For example:

```console
% terraform import aws_chime_voice_connector_termination.default abcdef1ghij2klmno3pqr4
```
