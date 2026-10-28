---
subcategory: "End User Messaging SMS"
layout: "aws"
page_title: "AWS: aws_pinpointsmsvoicev2_phone_number"
description: |-
  Provides details about an AWS End User Messaging SMS Phone Number.
---

# Data Source: aws_pinpointsmsvoicev2_phone_number

Provides details about an AWS End User Messaging SMS Phone Number.

## Example Usage

### Basic Usage

```terraform
data "aws_pinpointsmsvoicev2_phone_number" "example" {
  id = "phone-abcdef0123456789abcdef0123456789"
}
```

## Argument Reference

The following arguments are required:

* `id` - (Required) ID of the phone number.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the phone number.
* `created_timestamp` - Time when the phone number was created, in [RFC3339 format](https://tools.ietf.org/html/rfc3339#section-5.8).
* `deletion_protection_enabled` - Whether deletion protection is enabled. When `true`, the phone number cannot be deleted.
* `iso_country_code` - Two-character code, in ISO 3166-1 alpha-2 format, for the country or region.
* `message_type` - Type of message. Possible values are `TRANSACTIONAL` for messages that are critical or time-sensitive and `PROMOTIONAL` for messages that aren't critical or time-sensitive.
* `monthly_leasing_price` - Monthly price, in US dollars, to lease the phone number.
* `number_capabilities` - Set describing whether the origination identity can be used for text messages, voice calls, or both. Possible values are `SMS`, `VOICE`, `MMS`, and `RCS`.
* `number_type` - Type of phone number. Possible values are `SHORT_CODE`, `LONG_CODE`, `TOLL_FREE`, `TEN_DLC`, and `SIMULATOR`.
* `opt_out_list_name` - Name of the opt-out list associated with the phone number.
* `phone_number` - Phone number in E.164 format.
* `pool_id` - ID of the pool associated with the phone number, if any.
* `registration_id` - ID of the registration attached to the phone number for an external registration process.
* `self_managed_opt_outs_enabled` - Whether the phone number relies on self-managed opt-out handling. When `false`, AWS auto-replies to HELP/STOP requests and manages the opt-out list.
* `status` - Status of the phone number. Possible values are `PENDING`, `ACTIVE`, `ASSOCIATING`, `DISASSOCIATING`, and `DELETED`.
* `tags` - Map of tags assigned to the phone number.
* `two_way_channel_arn` - Destination for incoming messages.
* `two_way_channel_role` - ARN of the IAM role that End User Messaging SMS assumes to publish inbound messages to the two-way channel.
* `two_way_channel_enabled` - Whether inbound message reception is enabled for the phone number.
