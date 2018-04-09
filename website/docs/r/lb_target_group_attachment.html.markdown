---
layout: "aws"
page_title: "AWS: aws_lb_target_group_attachment"
sidebar_current: "docs-aws-resource-elbv2-target-group-attachment"
description: |-
  Provides the ability to register instances and containers with a LB
  target group
---

# aws_lb_target_group_attachment

Provides the ability to register instances and containers with a LB
target group

~> **Note:** `aws_alb_target_group_attachment` is known as `aws_lb_target_group_attachment`. The functionality is identical.

## Example Usage

```hcl
resource "aws_lb_target_group_attachment" "test" {
  target_group_arn = "${aws_lb_target_group.test.arn}"
  target_id        = "${aws_instance.test.id}"
  port             = 80
}

resource "aws_lb_target_group" "test" {
  // Other arguments
}

resource "aws_instance" "test" {
  // Other arguments
}
```

## Argument Reference

The following arguments are supported:

* `target_group_arn` - (Required) The ARN of the target group with which to register targets
* `target_id` (Required) The ID of the target. This is the Instance ID for an instance, or the container ID for an ECS container. If the target type is ip, specify an IP address.
* `port` - (Optional) The port on which targets receive traffic.
* `availability_zone` - (Optional) The Availability Zone where the IP address of the target is to be registered.

## Attributes Reference

The following attributes are exported in addition to the arguments listed above:

* `id` - A unique identifier for the attachment

## Import

Target Group Attachments cannot be imported.

