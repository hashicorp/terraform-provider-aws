---
subcategory: "User Notifications"
layout: "aws"
page_title: "AWS: aws_notifications_notification_configuration"
description: |-
  Terraform resource for managing an AWS User Notifications Notification Configuration.
---

# Resource: aws_notifications_notification_configuration

Terraform resource for managing an AWS User Notifications Notification Configuration.

## Example Usage

### Basic Usage

```terraform
resource "aws_notifications_notification_configuration" "example" {
  name        = "example"
  description = "Example notification configuration"

  tags = {
    Environment = "production"
    Project     = "example"
  }
}
```

### With Aggregation Duration

```terraform
resource "aws_notifications_notification_configuration" "example" {
  name                 = "example-aggregation"
  description          = "Example notification configuration with aggregation"
  aggregation_duration = "SHORT"

  tags = {
    Environment = "production"
    Project     = "example"
  }
}
```

## Argument Reference

The following arguments are required:

* `description` - (Required) Description of the NotificationConfiguration. Length constraints: Minimum length of 0,
  maximum length of 256.
* `name` - (Required) Name of the NotificationConfiguration. Supports RFC 3986's unreserved characters. Length
  constraints: Minimum length of 1, maximum length of 64. Pattern: `[A-Za-z0-9_\-]+`.

The following arguments are optional:

* `aggregation_duration` - (Optional) Aggregation preference of the NotificationConfiguration. Valid values: `LONG` (
  aggregate notifications for 12 hours), `SHORT` (aggregate notifications for 5 minutes), `NONE` (don't aggregate
  notifications). Default: `NONE`.
* `tags` - (Optional) Map of tags to assign to the resource. A tag is a string-to-string map of key-value pairs. If
  configured with a provider `default_tags` configuration block present, tags with matching keys will overwrite those
  defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Amazon Resource Name (ARN) of the NotificationConfiguration.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider `default_tags`
  configuration block.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to
import User Notifications Notification Configuration using the `arn`. For example:

```terraform
import {
  to = aws_notifications_notification_configuration.example
  id = "arn:aws:notifications::123456789012:configuration/abcdef1234567890abcdef1234567890"
}
```

Using `terraform import`, import User Notifications Notification Configuration using the `arn`. For example:

```console
% terraform import aws_notifications_notification_configuration.example arn:aws:notifications::123456789012:configuration/abcdef1234567890abcdef1234567890
```
