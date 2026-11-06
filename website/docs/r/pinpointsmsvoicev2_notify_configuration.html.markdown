---
subcategory: "End User Messaging SMS"
layout: "aws"
page_title: "AWS: aws_pinpointsmsvoicev2_notify_configuration"
description: |-
  Manages an AWS End User Messaging SMS Notify Configuration.
---

# Resource: aws_pinpointsmsvoicev2_notify_configuration

Manages an AWS End User Messaging SMS (Pinpoint SMS Voice v2) Notify Configuration.

## Example Usage

### Basic Usage

```terraform
resource "aws_pinpointsmsvoicev2_notify_configuration" "example" {
  display_name     = "MyBrand"
  use_case         = "CODE_VERIFICATION"
  enabled_channels = ["SMS", "VOICE"]
}
```

### With Optional Parameters

```terraform
resource "aws_pinpointsmsvoicev2_notify_configuration" "example" {
  display_name                = "MyBrand"
  use_case                    = "CODE_VERIFICATION"
  enabled_channels            = ["SMS", "VOICE"]
  enabled_countries           = ["US", "CA"]
  default_template_id         = "template-0123456789abcdef0"
  pool_id                     = aws_pinpointsmsvoicev2_pool.example.id
  deletion_protection_enabled = true

  tags = {
    Environment = "production"
  }
}
```

## Argument Reference

The following arguments are required:

* `display_name` - (Required) Display name representing your brand (1–15 characters). Forces new resource.
* `enabled_channels` - (Required) Communication channels enabled for this config. Valid values: `SMS`, `VOICE`, `MMS`, `RCS`.
* `use_case` - (Required) Use case for the notify configuration (e.g., `CODE_VERIFICATION`). Forces new resource.

The following arguments are optional:

* `default_template_id` - (Optional) Default template ID if not specified at send time.
* `deletion_protection_enabled` - (Optional) Whether to protect against accidental deletion. Defaults to `false`.
* `enabled_countries` - (Optional) Set of two-letter ISO 3166-1 alpha-2 country codes where messaging is enabled (e.g. `["US", "AU"]`). If not specified, defaults to the AWS default country list.
* `pool_id` - (Optional) ID or ARN of phone number pool to associate.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the notify configuration.
* `created_timestamp` - RFC3339 creation timestamp of the notify configuration.
* `id` - Notify configuration ID (e.g., `notify-0123456789abcdef0`).
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_pinpointsmsvoicev2_notify_configuration.example
  identity = {
    id = "notify-0123456789abcdef0"
  }
}

resource "aws_pinpointsmsvoicev2_notify_configuration" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `id` (String) Notify configuration ID.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import an AWS End User Messaging SMS Notify Configuration using the `id`. For example:

```terraform
import {
  to = aws_pinpointsmsvoicev2_notify_configuration.example
  id = "notify-0123456789abcdef0"
}
```

Using `terraform import`, import an AWS End User Messaging SMS Notify Configuration using the `id`. For example:

```console
% terraform import aws_pinpointsmsvoicev2_notify_configuration.example notify-0123456789abcdef0
```
