---
subcategory: "End User Messaging SMS"
layout: "aws"
page_title: "AWS: aws_pinpointsmsvoicev2_notify_configuration"
description: |-
  Provides details about an AWS End User Messaging SMS Notify Configuration.
---

# Data Source: aws_pinpointsmsvoicev2_notify_configuration

Provides details about an AWS End User Messaging SMS (Pinpoint SMS Voice v2) Notify Configuration.

## Example Usage

```terraform
data "aws_pinpointsmsvoicev2_notify_configuration" "example" {
  id = "notify-0123456789abcdef0"
}
```

## Argument Reference

The following arguments are required:

* `id` - (Required) Unique identifier of the notify configuration.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the notify configuration.
* `created_timestamp` - RFC3339 creation timestamp of the notify configuration.
* `default_template_id` - Default template ID if not specified at send time.
* `deletion_protection_enabled` - Whether deletion protection is enabled.
* `display_name` - Brand or display name.
* `enabled_channels` - Set of communication channels enabled for this config.
* `enabled_countries` - Set of two-letter ISO 3166-1 alpha-2 country codes where messaging is enabled.
* `pool_id` - ID or ARN of the associated phone number pool.
* `tags` - Key-value map of resource tags.
* `use_case` - Messaging use case (e.g., `CODE_VERIFICATION`).
