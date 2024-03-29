---
subcategory: "Auto Scaling"
layout: "aws"
page_title: "AWS: aws_autoscaling_notification"
description: |-
  Provides an AutoScaling Group with Notification support
---

# Resource: aws_autoscaling_notification

Provides an AutoScaling Group with Notification support, via SNS Topics. Each of
the `notifications` map to a [Notification Configuration][2] inside Amazon Web
Services, and are applied to each AutoScaling Group you supply.

## Example Usage

Basic usage:

```terraform
resource "aws_autoscaling_notification" "example_notifications" {
  group_names = [
    aws_autoscaling_group.bar.name,
    aws_autoscaling_group.foo.name,
  ]

  notifications = [
    "autoscaling:EC2_INSTANCE_LAUNCH",
    "autoscaling:EC2_INSTANCE_TERMINATE",
    "autoscaling:EC2_INSTANCE_LAUNCH_ERROR",
    "autoscaling:EC2_INSTANCE_TERMINATE_ERROR",
  ]

  topic_arn = aws_sns_topic.example.arn
}

resource "aws_sns_topic" "example" {
  name = "example-topic"

  # arn is an exported attribute
}

resource "aws_autoscaling_group" "bar" {
  name = "foobar1-terraform-test"

  # ...
}

resource "aws_autoscaling_group" "foo" {
  name = "barfoo-terraform-test"

  # ...
}
```

## Argument Reference

This resource supports the following arguments:

* `group_names` - (Required) List of AutoScaling Group Names
* `notifications` - (Required) List of Notification Types that trigger
notifications. Acceptable values are documented [in the AWS documentation here][1]
* `topic_arn` - (Required) Topic ARN for notifications to be sent through

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `group_names`
* `notifications`
* `topic_arn`

[1]: https://docs.aws.amazon.com/AutoScaling/latest/APIReference/API_NotificationConfiguration.html
[2]: https://docs.aws.amazon.com/AutoScaling/latest/APIReference/API_DescribeNotificationConfigurations.html
