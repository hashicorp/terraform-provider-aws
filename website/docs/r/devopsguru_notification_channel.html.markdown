---
subcategory: "DevOps Guru"
layout: "aws"
page_title: "AWS: aws_devopsguru_notification_channel"
description: |-
  Terraform resource for managing an AWS DevOps Guru Notification Channel.
---
# Resource: aws_devopsguru_notification_channel

Terraform resource for managing an AWS DevOps Guru Notification Channel.

## Example Usage

### Basic Usage

```terraform
resource "aws_devopsguru_notification_channel" "example" {
  sns {
    topic_arn = aws_sns_topic.example.arn
  }
}
```

### Filters

```terraform
resource "aws_devopsguru_notification_channel" "example" {
  sns {
    topic_arn = aws_sns_topic.example.arn
  }

  filters {
    message_types = ["NEW_INSIGHT"]
    severities    = ["HIGH"]
  }
}
```

## Argument Reference

The following arguments are required:

* `sns` - (Required) SNS noficiation channel configurations. See the [`sns` argument reference](#sns-argument-reference) below.

The following arguments are optional:

* `filters` - (Optional) Filter configurations for the Amazon SNS notification topic. See the [`filters` argument reference](#filters-argument-reference) below.

### `sns` Argument Reference

* `topic_arn` - (Required) Amazon Resource Name (ARN) of an Amazon Simple Notification Service topic.

### `filters` Argument Reference

* `message_types` - (Optional) Events to receive notifications for. Valid values are `NEW_INSIGHT`, `CLOSED_INSIGHT`, `NEW_ASSOCIATION`, `SEVERITY_UPGRADED`, and `NEW_RECOMMENDATION`.
* `severities` - (Optional) Severity levels to receive notifications for. Valid values are `LOW`, `MEDIUM`, and `HIGH`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Unique identifier for the notification channel.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import DevOps Guru Notification Channel using the `id`. For example:

```terraform
import {
  to = aws_devopsguru_notification_channel.example
  id = "id-12345678"
}
```

Using `terraform import`, import DevOps Guru Notification Channel using the `id`. For example:

```console
% terraform import aws_devopsguru_notification_channel.example id-12345678
```
