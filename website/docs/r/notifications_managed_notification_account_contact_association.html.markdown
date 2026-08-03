---
subcategory: "User Notifications"
layout: "aws"
page_title: "AWS: aws_notifications_managed_notification_account_contact_association"
description: |-
  Terraform resource for managing an AWS User Notifications Managed Notification Account Contact Association.
---
# Resource: aws_notifications_managed_notification_account_contact_association

Terraform resource for managing an AWS User Notifications Managed Notification Account Contact Association. This resource associates an account contact with a managed notification configuration.

## Example Usage

### Basic Usage

```terraform
resource "aws_notifications_managed_notification_account_contact_association" "example" {
  contact_identifier                     = "ACCOUNT_PRIMARY"
  managed_notification_configuration_arn = "arn:aws:notifications::123456789012:managed-notification-configuration/category/AWS-Health/sub-category/Security"
}
```

## Argument Reference

The following arguments are required:

* `contact_identifier` - (Required) A unique value of an Account Contact Type to associate with the ManagedNotificationConfiguration. Valid values: `ACCOUNT_PRIMARY`, `ACCOUNT_ALTERNATE_BILLING`, `ACCOUNT_ALTERNATE_OPERATIONS`, `ACCOUNT_ALTERNATE_SECURITY`.
* `managed_notification_configuration_arn` - (Required) ARN of the managed notification configuration to associate the account contact with.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import User Notifications Managed Notification Account Contact Association using the `managed_notification_configuration_arn,contact_identifier` format. For example:

```terraform
import {
  to = aws_notifications_managed_notification_account_contact_association.example
  id = "arn:aws:notifications::123456789012:managed-notification-configuration/category/AWS-Health/sub-category/Security,ACCOUNT_PRIMARY"
}
```

Using `terraform import`, import User Notifications Managed Notification Account Contact Association using the `managed_notification_configuration_arn,contact_identifier` format. For example:

```console
% terraform import aws_notifications_managed_notification_account_contact_association.example arn:aws:notifications::123456789012:managed-notification-configuration/category/AWS-Health/sub-category/Security,ACCOUNT_PRIMARY
```
