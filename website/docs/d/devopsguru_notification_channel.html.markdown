---
subcategory: "DevOps Guru"
layout: "aws"
page_title: "AWS: aws_devopsguru_notification_channel"
description: |-
  Terraform data source for managing an AWS DevOps Guru Notification Channel.
---

# Data Source: aws_devopsguru_notification_channel

Terraform data source for managing an AWS DevOps Guru Notification Channel.

## Example Usage

### Basic Usage

```terraform
data "aws_devopsguru_notification_channel" "example" {
  id = "channel-1234"
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `id` - (Required) Unique identifier for the notification channel.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `filters` - Filter configurations for the Amazon SNS notification topic. See the [`filters` attribute reference](#filters-attribute-reference) below.
* `sns` - SNS noficiation channel configurations. See the [`sns` attribute reference](#sns-attribute-reference) below.

### `sns` Attribute Reference

* `topic_arn` - Amazon Resource Name (ARN) of an Amazon Simple Notification Service topic.

### `filters` Attribute Reference

* `message_types` - Events to receive notifications for.
* `severities` - Severity levels to receive notifications for.
