---
subcategory: "User Notifications"
layout: "aws"
page_title: "AWS: aws_notifications_channel_association"
description: |-
  Terraform resource for managing an AWS User Notifications Channel Association.
---
# Resource: aws_notifications_channel_association

Terraform resource for managing an AWS User Notifications Channel Association. This resource associates a channel (such as an email contact) with a notification configuration.

## Example Usage

### Basic Usage

```terraform
resource "aws_notifications_notification_configuration" "example" {
  name        = "example-notification-config"
  description = "Example notification configuration"
}

resource "aws_notificationscontacts_email_contact" "example" {
  name          = "example-contact"
  email_address = "example@example.com"
}

resource "aws_notifications_channel_association" "example" {
  arn                            = aws_notificationscontacts_email_contact.example.arn
  notification_configuration_arn = aws_notifications_notification_configuration.example.arn
}
```

## Argument Reference

The following arguments are required:

* `arn` - (Required) ARN of the channel to associate with the notification configuration. This can be an email contact ARN.
* `notification_configuration_arn` - (Required) ARN of the notification configuration to associate the channel with.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import User Notifications Channel Association using the `notification_configuration_arn,channel_arn` format. For example:

```terraform
import {
  to = aws_notifications_channel_association.example
  id = "arn:aws:notifications:us-west-2:123456789012:configuration:example-notification-config,arn:aws:notificationscontacts:us-west-2:123456789012:emailcontact:example-contact"
}
```

Using `terraform import`, import User Notifications Channel Association using the `notification_configuration_arn,channel_arn` format. For example:

```console
% terraform import aws_notifications_channel_association.example arn:aws:notifications:us-west-2:123456789012:configuration:example-notification-config,arn:aws:notificationscontacts:us-west-2:123456789012:emailcontact:example-contact
```
