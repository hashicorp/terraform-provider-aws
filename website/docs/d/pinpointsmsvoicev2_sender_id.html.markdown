---
subcategory: "End User Messaging SMS"
layout: "aws"
page_title: "AWS: aws_pinpointsmsvoicev2_sender_id"
description: |-
  Provides details about an End User Messaging SMS Sender ID.
---

# Data Source: aws_pinpointsmsvoicev2_sender_id

Provides details about an End User Messaging SMS Sender ID.

## Example Usage

```terraform
data "aws_pinpointsmsvoicev2_sender_id" "example" {
  sender_id        = "MYCOMPANY"
  iso_country_code = "GB"
}
```

## Argument Reference

The following arguments are required:

* `iso_country_code` - (Required) Two-character code, in ISO 3166-1 alpha-2 format, for the country or region.
* `sender_id` - (Required) Alphanumeric sender ID to look up. Must be between 3 and 11 characters long, contain only upper case letters, numbers, and dashes, and cannot be numeric-only.

The following arguments are optional:

* `region` - (Optional) Region where this data source will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Sender ID.
* `deletion_protection_enabled` - Whether deletion protection is enabled.
* `message_types` - Message types supported by the Sender ID. Valid values are `TRANSACTIONAL` and `PROMOTIONAL`.
* `monthly_leasing_price` - Monthly leasing price, in US dollars.
* `registered` - Whether the Sender ID is registered.
* `registration_id` - Unique identifier for the registration.
* `tags` - Map of tags assigned to the Sender ID.
