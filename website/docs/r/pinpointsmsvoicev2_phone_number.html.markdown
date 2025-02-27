---
subcategory: "End User Messaging SMS"
layout: "aws"
page_title: "AWS: aws_pinpointsmsvoicev2_phone_number"
description: |-
  Manages an AWS End User Messaging SMS phone number.
---

# Resource: aws_pinpointsmsvoicev2_phone_number

Manages an AWS End User Messaging SMS phone number.

## Example Usage

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

This resource supports the following arguments:

* `deletion_protection_enabled` - (Optional) By default this is set to `false`. When set to true the phone number can’t be deleted.
* `iso_country_code` - (Required) The two-character code, in ISO 3166-1 alpha-2 format, for the country or region.
* `message_type` - (Required) The type of message. Valid values are `TRANSACTIONAL` for messages that are critical or time-sensitive and `PROMOTIONAL` for messages that aren’t critical or time-sensitive.
* `number_capabilities` - (Required) Describes if the origination identity can be used for text messages, voice calls or both. valid values are `SMS` and `VOICE`.
* `number_type` - (Required) The type of phone number to request. Possible values are `LONG_CODE`, `TOLL_FREE`, `TEN_DLC`, or `SIMULATOR`.
* `opt_out_list_name` - (Optional) The name of the opt-out list to associate with the phone number.
* `registration_id` - (Optional) Use this field to attach your phone number for an external registration process.
* `self_managed_opt_outs_enabled` - (Optional) When set to `false` an end recipient sends a message that begins with HELP or STOP to one of your dedicated numbers, AWS End User Messaging SMS and Voice automatically replies with a customizable message and adds the end recipient to the opt-out list. When set to true you’re responsible for responding to HELP and STOP requests. You’re also responsible for tracking and honoring opt-out request.
* `two_way_channel_arn` - (Optional) The Amazon Resource Name (ARN) of the two way channel.
* `two_way_channel_enabled` - (Optional) By default this is set to `false`. When set to `true` you can receive incoming text messages from your end recipients.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the phone number.
* `id` - ID of the phone number.
* `monthly_leasing_price` - The monthly price, in US dollars, to lease the phone number.
* `phone_number` - The new phone number that was requested.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import phone numbers using the `id`. For example:

```terraform
import {
  to = aws_pinpointsmsvoicev2_phone_number.example
  id = "phone-abcdef0123456789abcdef0123456789"
}
```

Using `terraform import`, import phone numbers using the `id`. For example:

```console
% terraform import aws_pinpointsmsvoicev2_phone_number.example phone-abcdef0123456789abcdef0123456789
```
