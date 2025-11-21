---
subcategory: "User Notifications"
layout: "aws"
page_title: "AWS: aws_notifications_managed_notification_additional_channel_association"
description: |-
  Terraform resource for managing an AWS User Notifications Managed Notification Additional Channel Association.
---
# Resource: aws_notifications_managed_notification_additional_channel_association

Terraform resource for managing an AWS User Notifications Managed Notification Additional Channel Association. This resource associates a channel (such as an email contact, mobile device, or chat channel) with a managed notification.

## Example Usage

### Basic Usage

```terraform
resource "aws_notificationscontacts_email_contact" "example" {
  name          = "example-contact"
  email_address = "example@example.com"
}

resource "aws_notifications_managed_notification_additional_channel_association" "example" {
  arn                      = aws_notificationscontacts_email_contact.example.arn
  managed_notification_arn = "arn:aws:notifications::123456789012:managed-notification-configuration/category/AWS-Health/sub-category/Security"
}
```

## Argument Reference

The following arguments are required:

* `arn` - (Required) ARN of the channel to associate with the managed notification.
* `managed_notification_arn` - (Required) ARN of the managed notification to associate the channel with.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import User Notifications Managed Notification Additional Channel Association using the `managed_notification_arn,channel_arn` format. For example:

```terraform
import {
  to = aws_notifications_managed_notification_additional_channel_association.example
  id = "arn:aws:notifications::123456789012:managed-notification-configuration/category/AWS-Health/sub-category/Security,arn:aws:notificationscontacts:us-west-2:123456789012:emailcontact:example-contact"
}
```

Using `terraform import`, import User Notifications Managed Notification Additional Channel Association using the `managed_notification_arn,channel_arn` format. For example:

```console
% terraform import aws_notifications_managed_notification_additional_channel_association.example arn:aws:notifications::123456789012:managed-notification-configuration/category/AWS-Health/sub-category/Security,arn:aws:notificationscontacts:us-west-2:123456789012:emailcontact:example-contact
```
