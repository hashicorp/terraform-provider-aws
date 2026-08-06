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
  sender_id        = "MYCOMPANY"
  iso_country_code = "GB"
  message_types    = ["TRANSACTIONAL"]
}
```

### With Deletion Protection

```terraform
resource "aws_pinpointsmsvoicev2_sender_id" "example" {
  sender_id                   = "MYCOMPANY"
  iso_country_code            = "GB"
  message_types               = ["TRANSACTIONAL"]
  deletion_protection_enabled = true
}
```

## Argument Reference

The following arguments are required:

* `iso_country_code` - (Required, Forces new resource) Two-character code, in ISO 3166-1 alpha-2 format, for the country or region.
* `sender_id` - (Required, Forces new resource) Alphanumeric sender ID to request. Must be between 3 and 11 characters long, contain only upper case letters, numbers, and dashes, and cannot be numeric-only. AWS stores sender IDs in upper case, so upper case is required here to keep configuration, state, and resource identity consistent.

The following arguments are optional:

* `deletion_protection_enabled` - (Optional) Whether deletion protection is enabled. When set to `true`, the sender ID cannot be deleted. Defaults to `false`.
* `message_types` - (Optional, Forces new resource) Type of message. Valid values are `TRANSACTIONAL` and `PROMOTIONAL`. Defaults to `["TRANSACTIONAL"]` if not specified.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the sender ID.
* `monthly_leasing_price` - Monthly leasing price, in US dollars.
* `registered` - Whether the sender ID is registered.
* `registration_id` - Unique identifier for the registration.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_pinpointsmsvoicev2_sender_id.example
  identity = {
    sender_id        = "MYSENDERID"
    iso_country_code = "US"
  }
}

resource "aws_pinpointsmsvoicev2_sender_id" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `sender_id` (String) Sender ID.
* `iso_country_code` (String) Two-character code, in ISO 3166-1 alpha-2 format, for the country or region.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import an End User Messaging SMS Sender ID using the `sender_id` and `iso_country_code`, separated by a comma. For example:

```terraform
import {
  to = aws_pinpointsmsvoicev2_sender_id.example
  id = "MYSENDERID,US"
}
```

Using `terraform import`, import an End User Messaging SMS Sender ID using the `sender_id` and `iso_country_code`, separated by a comma. For example:

```console
% terraform import aws_pinpointsmsvoicev2_sender_id.example "MYSENDERID,US"
```
