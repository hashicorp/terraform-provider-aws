---
subcategory: "PinpointSMSVoiceV2"
layout: "aws"
page_title: "AWS: aws_pinpointsmsvoicev2_phone_number"
description: |-
  Terraform resource for managing an AWS Pinpoint SMS Voice V2 Phone Number.
---

# Resource: aws_pinpointsmsvoicev2_phone_number

Terraform resource for managing an AWS Pinpoint SMS Voice V2 PhoneNumber.

## Example Usage

### Basic Usage

```terraform
resource "aws_pinpointsmsvoicev2_phone_number" "example" {
  iso_country_code = "US"
  message_type     = "TRANSACTIONAL"
  number_type      = "TOLL_FREE"

  number_capabilities = [
    "SMS"
  ]
}
```

## Argument Reference

The following arguments are required:

* `iso_country_code` - (Required) The two-character code, in ISO 3166-1 alpha-2 format, for the country or region.
* `message_type` - (Required) The type of message. Valid values are TRANSACTIONAL for messages that are critical or time-sensitive and PROMOTIONAL for messages that aren’t critical or time-sensitive.
* `number_capabilities` - (Required) Describes if the origination identity can be used for text messages, voice calls or both. valid values are SMS and VOICE
* `number_type` - (Required) The type of phone number to request. Possible values are LONG_CODE, TOLL_FREE or TEN_DLC.

The following arguments are optional:

* `deletion_protection_enabled` - (Optional) By default this is set to false. When set to true the phone number can’t be deleted.
* `opt_out_list_name` - (Optional) The name of the OptOutList to associate with the phone number.
* `registration_id` - (Optional) Use this field to attach your phone number for an external registration process.
* `self_managed_opt_outs_enabled` - (Optional) When set to false an end recipient sends a message that begins with HELP or STOP to one of your dedicated numbers, Amazon Pinpoint automatically replies with a customizable message and adds the end recipient to the OptOutList. When set to true you’re responsible for responding to HELP and STOP requests. You’re also responsible for tracking and honoring opt-out request.
* `two_way_channel_arn` - (Optional) The Amazon Resource Name (ARN) of the two way channel.
* `two_way_channel_enabled` - (Optional) By default this is set to false. When set to true you can receive incoming text messages from your end recipients.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The Amazon Resource Name (ARN) of the requested phone number.
* `phone_number` - The new phone number that was requested.
* `monthly_leasing_price` - The monthly price, in US dollars, to lease the phone number.

## Import

Pinpoint SMS Voice V2 PhoneNumber can be imported using the `id`, e.g.,

```
$ terraform import aws_pinpointsmsvoicev2_phone_number.example phone-abcdef0123456789abcdef0123456789
```
