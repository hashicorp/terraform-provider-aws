---
subcategory: "End User Messaging SMS"
layout: "aws"
page_title: "AWS: aws_pinpointsmsvoicev2_keyword"
description: |-
  Manages an AWS End User Messaging SMS Keyword.
---

# Resource: aws_pinpointsmsvoicev2_keyword

Manages an AWS End User Messaging SMS Keyword.

~> **Note:** The mandatory keywords `HELP` and `STOP` exist on every origination identity and cannot be created or deleted independently of it. This resource adopts and manages their `keyword_message` in place, while `keyword_action` is managed by AWS and cannot be set. Destroying the resource does not delete or reset the keyword; it remains on the origination identity with its last-applied message.

## Example Usage

### Phone Number

```terraform
resource "aws_pinpointsmsvoicev2_phone_number" "example" {
  iso_country_code    = "US"
  message_type        = "TRANSACTIONAL"
  number_type         = "SIMULATOR"
  number_capabilities = ["SMS"]
}

resource "aws_pinpointsmsvoicev2_keyword" "example" {
  origination_identity = aws_pinpointsmsvoicev2_phone_number.example.id
  keyword              = "EXAMPLE"
  keyword_message      = "Thanks for messaging our example number."
}
```

### Pool

```terraform
resource "aws_pinpointsmsvoicev2_phone_number" "example" {
  iso_country_code    = "US"
  message_type        = "TRANSACTIONAL"
  number_type         = "SIMULATOR"
  number_capabilities = ["SMS"]
}

resource "aws_pinpointsmsvoicev2_pool" "example" {
  iso_country_code       = "US"
  message_type           = "TRANSACTIONAL"
  origination_identities = [aws_pinpointsmsvoicev2_phone_number.example.arn]
}

resource "aws_pinpointsmsvoicev2_keyword" "example" {
  origination_identity = aws_pinpointsmsvoicev2_pool.example.id
  keyword              = "OPTOUT"
  keyword_message      = "You have been unsubscribed."
  keyword_action       = "OPT_OUT"
}
```

### Mandatory Keyword

The mandatory `HELP` and `STOP` keywords are adopted rather than created. Omit `keyword_action`; AWS manages it.

```terraform
resource "aws_pinpointsmsvoicev2_phone_number" "example" {
  iso_country_code    = "US"
  message_type        = "TRANSACTIONAL"
  number_type         = "SIMULATOR"
  number_capabilities = ["SMS"]
}

resource "aws_pinpointsmsvoicev2_keyword" "help" {
  origination_identity = aws_pinpointsmsvoicev2_phone_number.example.id
  keyword              = "HELP"
  keyword_message      = "Reply STOP to unsubscribe. Message and data rates may apply."
}
```

## Argument Reference

The following arguments are required:

* `keyword` - (Required) Keyword to configure. Changing this forces a new resource.
* `keyword_message` - (Required) Message to send when the keyword is received.
* `origination_identity` - (Required) Origination identity to attach the keyword to. Value is the ID or ARN of a phone number, pool, or sender ID. Changing this forces a new resource.

The following arguments are optional:

* `keyword_action` - (Optional) Action to perform when the keyword is received. Valid values: `AUTOMATIC_RESPONSE`, `OPT_OUT`, `OPT_IN`. Defaults to `AUTOMATIC_RESPONSE`. Must not be set for mandatory keywords, whose action is managed by AWS.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `origination_identity_arn` - ARN of the origination identity the keyword is attached to.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_pinpointsmsvoicev2_keyword.example
  identity = {
    origination_identity = "phone-abcdef0123456789abcdef0123456789"
    keyword              = "EXAMPLE"
  }
}

resource "aws_pinpointsmsvoicev2_keyword" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `origination_identity` (String) Origination identity the keyword is attached to.
* `keyword` (String) Keyword text.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import a keyword using the `origination_identity` and `keyword`, separated by a comma. For example:

```terraform
import {
  to = aws_pinpointsmsvoicev2_keyword.example
  id = "phone-abcdef0123456789abcdef0123456789,EXAMPLE"
}
```

Using `terraform import`, import a keyword using the `origination_identity` and `keyword`, separated by a comma. For example:

```console
% terraform import aws_pinpointsmsvoicev2_keyword.example "phone-abcdef0123456789abcdef0123456789,EXAMPLE"
```
