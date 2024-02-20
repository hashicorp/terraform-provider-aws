---
subcategory: "Chime SDK Voice"
layout: "aws"
page_title: "AWS: aws_chimesdkvoice_sip_rule"
description: |-
    A SIP rule associates your SIP media application with a phone number or a Request URI hostname. You can associate a SIP rule with more than one SIP media application. Each application then runs only that rule.
---
# Resource: aws_chimesdkvoice_sip_rule

A SIP rule associates your SIP media application with a phone number or a Request URI hostname. You can associate a SIP rule with more than one SIP media application. Each application then runs only that rule.

## Example Usage

### Basic Usage

```terraform
resource "aws_chimesdkvoice_sip_rule" "example" {
  name          = "example-sip-rule"
  trigger_type  = "RequestUriHostname"
  trigger_value = aws_chime_voice_connector.example-voice-connector.outbound_host_name
  target_applications {
    priority                 = 1
    sip_media_application_id = aws_chimesdkvoice_sip_media_application.example-sma.id
    aws_region               = "us-east-1"
  }
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) The name of the SIP rule.
* `target_applications` - (Required) List of SIP media applications with priority and AWS Region. Only one SIP application per AWS Region can be used. See [`target_applications`](#target_applications).
* `trigger_type` - (Required) The type of trigger assigned to the SIP rule in `trigger_value`. Valid values are `RequestUriHostname` or `ToPhoneNumber`.
* `trigger_value` - (Required) If `trigger_type` is `RequestUriHostname`, the value can be the outbound host name of an Amazon Chime Voice Connector. If `trigger_type` is `ToPhoneNumber`, the value can be a customer-owned phone number in the E164 format. The Sip Media Application specified in the Sip Rule is triggered if the request URI in an incoming SIP request matches the `RequestUriHostname`, or if the "To" header in the incoming SIP request matches the `ToPhoneNumber` value.

The following arguments are optional:

* `disabled` - (Optional) Enables or disables a rule. You must disable rules before you can delete them.

### `target_applications`

List of SIP media applications with priority and AWS Region. Only one SIP application per AWS Region can be used.

* `aws_region` - (Required) The AWS Region of the target application.
* `priority` - (Required) Priority of the SIP media application in the target list.
* `sip_media_application_id` - (Required) The SIP media application ID.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The SIP rule ID.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import a ChimeSDKVoice SIP Rule using the `id`. For example:

```terraform
import {
  to = aws_chimesdkvoice_sip_rule.example
  id = "abcdef123456"
}
```

Using `terraform import`, import a ChimeSDKVoice SIP Rule using the `id`. For example:

```console
% terraform import aws_chimesdkvoice_sip_rule.example abcdef123456
```
