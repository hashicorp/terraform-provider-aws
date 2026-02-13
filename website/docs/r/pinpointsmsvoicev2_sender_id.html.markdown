---
subcategory: "End User Messaging SMS"
layout: "aws"
page_title: "AWS: aws_pinpointsmsvoicev2_sender_id"
description: |-
  Manages an End User Messaging SMS Sender ID.
---

# Resource: aws_pinpointsmsvoicev2_sender_id

Manages an End User Messaging SMS Sender ID.

## Example Usage

### Basic Usage

```terraform
resource "aws_pinpointsmsvoicev2_sender_id" "example" {
  sender_id        = "MyCompany"
  iso_country_code = "GB"
  message_types    = ["TRANSACTIONAL"]
}
```

### With Deletion Protection

```terraform
resource "aws_pinpointsmsvoicev2_sender_id" "example" {
  sender_id                   = "MyCompany"
  iso_country_code            = "GB"
  message_types               = ["TRANSACTIONAL"]
  deletion_protection_enabled = true
}
```

## Argument Reference

The following arguments are required:

* `sender_id` - (Required) The alphanumeric sender ID to request. Must be between 3 and 11 characters long, contain only letters, numbers, and dashes, and cannot be numeric-only.
* `iso_country_code` - (Required) The two-character code, in ISO 3166-1 alpha-2 format, for the country or region.

The following arguments are optional:

* `deletion_protection_enabled` - (Optional) Whether deletion protection is enabled. When set to `true`, the sender ID cannot be deleted. Defaults to `false`.
* `message_types` - (Optional) The type of message. Valid values are `TRANSACTIONAL` and `PROMOTIONAL`. Defaults to `["TRANSACTIONAL"]` if not specified.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the sender ID.
* `id` - The sender ID and ISO country code separated by a comma (`,`).
* `monthly_leasing_price` - The monthly leasing price, in US dollars.
* `registered` - Whether the sender ID is registered.
* `registration_id` - The unique identifier for the registration.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

End User Messaging SMS Sender IDs can be imported using the sender ID and ISO country code separated by a comma (`,`):

```
$ terraform import aws_pinpointsmsvoicev2_sender_id.example MySenderId,US
```
