---
subcategory: "Auto Scaling"
layout: "aws"
page_title: "AWS: aws_autoscaling_groups"
description: |-
    Provides a list of Autoscaling Groups within a specific region.
---

# Data Source: aws_autoscaling_groups

The Autoscaling Groups data source allows access to the list of AWS
ASGs within a specific region. This will allow you to pass a list of AutoScaling Groups to other resources.

## Example Usage

```terraform
data "aws_autoscaling_groups" "groups" {
  filter {
    name   = "tag:Team"
    values = ["Pets"]
  }

  filter {
    name   = "tag-key"
    values = ["Environment"]
  }
}

resource "aws_autoscaling_notification" "slack_notifications" {
  group_names = data.aws_autoscaling_groups.groups.names

  notifications = [
    "autoscaling:EC2_INSTANCE_LAUNCH",
    "autoscaling:EC2_INSTANCE_TERMINATE",
    "autoscaling:EC2_INSTANCE_LAUNCH_ERROR",
    "autoscaling:EC2_INSTANCE_TERMINATE_ERROR",
  ]

  topic_arn = "TOPIC ARN"
}
```

## Argument Reference

* `names` - (Optional) List of autoscaling group names
* `filter` - (Optional) Filter used to scope the list e.g., by tags. See [related docs](http://docs.aws.amazon.com/AutoScaling/latest/APIReference/API_Filter.html).
    * `name` - (Required) Name of the DescribeAutoScalingGroup filter. The recommended values are: `tag-key`, `tag-value`, and `tag:<tag name>`
    * `values` - (Required) Value of the filter.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arns` - List of the Autoscaling Groups Arns in the current region.
* `id` - AWS Region.
* `names` - List of the Autoscaling Groups in the current region.
