---
subcategory: "End User Messaging SMS"
layout: "aws"
page_title: "AWS: aws_pinpointsmsvoicev2_protect_configuration"
description: |-
  Manages an AWS End User Messaging SMS Protect Configuration.
---

# Resource: aws_pinpointsmsvoicev2_protect_configuration

Manages an AWS End User Messaging SMS Protect Configuration. A protect configuration sets country rules that allow, block, monitor, or filter messages sent to each destination country.

## Example Usage

### Basic Usage

```terraform
resource "aws_pinpointsmsvoicev2_protect_configuration" "example" {}
```

### Country Rules

```terraform
resource "aws_pinpointsmsvoicev2_protect_configuration" "example" {
  deletion_protection_enabled = true

  sms_country_rules = {
    CA = "FILTER"
    DE = "MONITOR"
    RU = "BLOCK"
  }

  voice_country_rules = {
    RU = "BLOCK"
  }

  tags = {
    Name = "example"
  }
}
```

### Default Status

Block SMS messages to every country except the United States and Canada:

```terraform
resource "aws_pinpointsmsvoicev2_protect_configuration" "example" {
  sms_default_status = "BLOCK"

  sms_country_rules = {
    CA = "ALLOW"
    US = "ALLOW"
  }
}
```

## Argument Reference

The following arguments are optional:

* `deletion_protection_enabled` - (Optional) Whether deletion protection is enabled. When `true`, the protect configuration can't be deleted.
* `mms_country_rules` - (Optional) Map of two-character ISO 3166-1 alpha-2 country codes to protect status for MMS messages. Overrides `mms_default_status` for the listed countries. Valid values are `ALLOW` and `BLOCK`.
* `mms_default_status` - (Optional) Protect status for every country not listed in `mms_country_rules`. Valid values are `ALLOW` and `BLOCK`. Defaults to `ALLOW`.
* `rcs_country_rules` - (Optional) Map of two-character ISO 3166-1 alpha-2 country codes to protect status for RCS messages. Overrides `rcs_default_status` for the listed countries. Valid values are `ALLOW` and `BLOCK`.
* `rcs_default_status` - (Optional) Protect status for every country not listed in `rcs_country_rules`. Valid values are `ALLOW` and `BLOCK`. Defaults to `ALLOW`.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `sms_country_rules` - (Optional) Map of two-character ISO 3166-1 alpha-2 country codes to protect status for SMS messages. Overrides `sms_default_status` for the listed countries. Valid values are `ALLOW`, `BLOCK`, `FILTER`, and `MONITOR`.
* `sms_default_status` - (Optional) Protect status for every country not listed in `sms_country_rules`. Valid values are `ALLOW`, `BLOCK`, `FILTER`, and `MONITOR`. Defaults to `ALLOW`.
* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `voice_country_rules` - (Optional) Map of two-character ISO 3166-1 alpha-2 country codes to protect status for voice messages. Overrides `voice_default_status` for the listed countries. Valid values are `ALLOW` and `BLOCK`.
* `voice_default_status` - (Optional) Protect status for every country not listed in `voice_country_rules`. Valid values are `ALLOW` and `BLOCK`. Defaults to `ALLOW`.

~> **Note:** Terraform manages the status of every country. Each `*_country_rules` map lists only the countries whose status differs from the default status, so a status equal to the default is not valid. Removing a country from a map resets it to the default status.

-> **Note:** The countries each message type supports differ and are validated by AWS. See [Supported countries and regions](https://docs.aws.amazon.com/sms-voice/latest/userguide/phone-numbers-sms-by-country.html) in the AWS End User Messaging SMS User Guide. `FILTER` and `MONITOR` incur additional charges for AIT risk evaluation. See [Country rule modes](https://docs.aws.amazon.com/sms-voice/latest/userguide/filter-and-monitor-messages.html) for details.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `account_default` - Whether the protect configuration is the account default.
* `arn` - ARN of the protect configuration.
* `created_timestamp` - Time when the protect configuration was created, in [RFC3339 format](https://datatracker.ietf.org/doc/html/rfc3339#section-5.8).
* `id` - ID of the protect configuration.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

~> **Note:** AWS doesn't allow deleting a protect configuration while it is associated with a configuration set, is the account default, or has deletion protection enabled.

~> **Note:** AWS doesn't store a default status. After import, the `*_default_status` arguments are unset, and every country whose status isn't `ALLOW` appears in the `*_country_rules` maps. If the configuration sets a default status, the first plan after import shows a difference. Applying it changes no country only if the configured default status and country rules describe the imported rules. Otherwise, Terraform updates the countries to match the configuration.

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_pinpointsmsvoicev2_protect_configuration.example
  identity = {
    id = "protect-abcdef0123456789abcdef0123456789"
  }
}

resource "aws_pinpointsmsvoicev2_protect_configuration" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `id` - (String) ID of the protect configuration.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import an AWS End User Messaging SMS Protect Configuration using the `id`. For example:

```terraform
import {
  to = aws_pinpointsmsvoicev2_protect_configuration.example
  id = "protect-abcdef0123456789abcdef0123456789"
}
```

Using `terraform import`, import an AWS End User Messaging SMS Protect Configuration using the `id`. For example:

```console
% terraform import aws_pinpointsmsvoicev2_protect_configuration.example protect-abcdef0123456789abcdef0123456789
```
