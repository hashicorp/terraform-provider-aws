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
  name = "example-sip-rule"
  trigger_type = "RequestUriHostname"
  trigger_value = aws_chime_voice_connector.example-voice-connector.outbound_host_name
  target_applications {
	priority = 1
	sip_media_application_id = aws_chimesdkvoice_sip_media_application.example-sma.id
	aws_region = "us-east-1"
  }
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) The name of the SIP rule.
* `target_applications` - (Required) List of SIP media applications with priority and AWS Region. Only one SIP application per AWS Region can be used.
* `trigger_type` - (Required) The type of trigger assigned to the SIP rule in TriggerValue, currently RequestUriHostname or ToPhoneNumber.
* `trigger_value` - (Required) If TriggerType is RequestUriHostname, the value can be the outbound host name of an Amazon Chime Voice Connector. If TriggerType is ToPhoneNumber, the value can be a customer-owned phone number in the E164 format. The SipMediaApplication specified in the SipRule is triggered if the request URI in an incoming SIP request matches the RequestUriHostname, or if the To header in the incoming SIP request matches the ToPhoneNumber value.

The following arguments are optional:

* `disabled` - (Optional) Enables or disables a rule. You must disable rules before you can delete them.

### `target_applications`

List of SIP media applications with priority and AWS Region. Only one SIP application per AWS Region can be used.

* `AwsRegion` - (Required) The AWS Region of the target application.
* `Priority` - (Required) Priority of the SIP media application in the target list.
* `SipMediaApplicationId` - (Required) The SIP media application ID.


## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The SIP rule ID.

## Import

A ChimeSDKVoice SIP Rule can be imported using the `id`, e.g.,

```
$ terraform import aws_chimesdkvoice_sip_rule.example abcdef123456
```
