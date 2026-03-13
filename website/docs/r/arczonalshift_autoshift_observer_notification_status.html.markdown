---
subcategory: "Application Recovery Controller Zonal Shift"
layout: "aws"
page_title: "AWS: aws_arczonalshift_autoshift_observer_notification_status"
description: |-
  Manages the autoshift observer notification status for AWS Application Recovery Controller Zonal Shift.
---

# Resource: aws_arczonalshift_autoshift_observer_notification_status

Manages the autoshift observer notification status for AWS Application Recovery Controller Zonal Shift. This is an account-level singleton setting that controls whether autoshift observer notifications are enabled or disabled.

~> NOTE: Removing this Terraform resource disables zonal autoshift observer notifications.

## Example Usage

```terraform
resource "aws_arczonalshift_autoshift_observer_notification_status" "example" {
  status = "ENABLED"
}
```

## Argument Reference

The following arguments are required:

* `status` - (Required) Autoshift observer notification status. Valid values are `ENABLED` or `DISABLED`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The resource identifier.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

ARC Zonal Shift Autoshift Observer Notification Status can be imported using the account identifier:

```shell
terraform import aws_arczonalshift_autoshift_observer_notification_status.example autoshift-observer-notification-status
```
