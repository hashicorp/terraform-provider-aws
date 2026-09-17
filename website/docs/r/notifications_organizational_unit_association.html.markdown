---
subcategory: "User Notifications"
layout: "aws"
page_title: "AWS: aws_notifications_organizational_unit_association"
description: |-
  Terraform resource for managing an AWS User Notifications Organizational Unit Association.
---
# Resource: aws_notifications_organizational_unit_association

Terraform resource for managing an AWS User Notifications Organizational Unit Association. This resource associates an organizational unit with a notification configuration.

## Example Usage

### Basic Usage

```terraform
data "aws_organizations_organization" "example" {}

resource "aws_notifications_notification_configuration" "example" {
  name        = "example-notification-config"
  description = "Example notification configuration"
}

resource "aws_organizations_organizational_unit" "example" {
  name      = "example-ou"
  parent_id = data.aws_organizations_organization.example.roots[0].id
}

# Allow time for organizational unit creation to propagate
resource "time_sleep" "wait" {
  depends_on = [
    aws_organizations_organizational_unit.example,
    aws_notifications_notification_configuration.example,
  ]

  create_duration = "5s"
}

resource "aws_notifications_organizational_unit_association" "example" {
  depends_on = [time_sleep.wait]

  organizational_unit_id         = aws_organizations_organizational_unit.example.id
  notification_configuration_arn = aws_notifications_notification_configuration.example.arn
}
```

### Associate with Organization Root

```terraform
data "aws_organizations_organization" "example" {}

resource "aws_notifications_notification_configuration" "example" {
  name        = "example-notification-config"
  description = "Example notification configuration"
}

resource "aws_notifications_organizational_unit_association" "example" {
  organizational_unit_id         = data.aws_organizations_organization.example.roots[0].id
  notification_configuration_arn = aws_notifications_notification_configuration.example.arn
}
```

## Argument Reference

The following arguments are required:

* `organizational_unit_id` - (Required) ID of the organizational unit or ID of the root to associate with the notification configuration. Can be a root ID (e.g., `r-1234`), or an organization ID (e.g., `o-1234567890`).
* `notification_configuration_arn` - (Required) ARN of the notification configuration to associate the organizational unit with.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import User Notifications Organizational Unit Association using the `notification_configuration_arn,organizational_unit_id` format. For example:

```terraform
import {
  to = aws_notifications_organizational_unit_association.example
  id = "arn:aws:notifications:us-west-2:123456789012:configuration:example-notification-config,ou-1234-12345678"
}
```

Using `terraform import`, import User Notifications Organizational Unit Association using the `notification_configuration_arn,organizational_unit_id` format. For example:

```console
% terraform import aws_notifications_organizational_unit_association.example arn:aws:notifications:us-west-2:123456789012:configuration:example-notification-config,ou-1234-12345678
```
