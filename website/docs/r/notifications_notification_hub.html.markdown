---
subcategory: "User Notifications"
layout: "aws"
page_title: "AWS: aws_notifications_notification_hub"
description: |-
  Terraform resource for managing an AWS User Notifications Notification Hub.
---
# Resource: aws_notifications_notification_hub

Terraform resource for managing an AWS User Notifications Notification Hub.

## Example Usage

### Basic Usage

```terraform
resource "aws_notifications_notification_hub" "example" {
  notification_hub_region = "us-west-2"
}
```

## Argument Reference

The following arguments are required:

* `notification_hub_region` - Notification Hub region.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import User Notifications Notification Hub using the `notification_hub_region`. For example:

```terraform
import {
  to = aws_notifications_notification_hub.example
  id = "us-west-2"
}
```

Using `terraform import`, import User Notifications Notification Hub using the `notification_hub_region `. For example:

```console
% terraform import aws_notifications_notification_hub.example us-west-2
```
