---
subcategory: "EC2 (Elastic Compute Cloud)"
layout: "aws"
page_title: "AWS: aws_ec2_instance_event_window"
description: |-
  Get information on an EC2 Instance Event Window.
---

# Data Source: aws_ec2_instance_event_window

Use this data source to get information on an EC2 Instance Event Window.

## Example Usage

### Filter By ID

```terraform
data "aws_ec2_instance_event_window" "example" {
  id = "iew-0123456789abcdef0"
}
```

### Filter By Name

```terraform
data "aws_ec2_instance_event_window" "example" {
  filter {
    name   = "event-window-name"
    values = ["example"]
  }
}
```

## Argument Reference

This data source supports the following arguments:

* `filter` - (Optional) One or more configuration blocks used to filter results by specific criteria. Conflicts with `id`. Exactly one of `id` or `filter` must be specified. See the [`describe-instance-event-windows` AWS CLI reference](https://docs.aws.amazon.com/cli/latest/reference/ec2/describe-instance-event-windows.html) for supported filters. See [Filter](#filter) below.
* `id` - (Optional) ID of the event window. Conflicts with `filter`. Exactly one of `id` or `filter` must be specified.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

### `filter` Block

* `name` - (Required) Name of the filter.
* `values` - (Required) List of one or more values for the filter.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `cron_expression` - Cron expression defined for the event window.
* `name` - Name of the event window.
* `tags` - Map of tags assigned to the event window.
* `time_ranges` - One or more time ranges defined for the event window. See [`time_ranges`](#time_ranges) below.

### `time_ranges` Block

* `end_hour` - Hour (in UTC) at which the time range ends.
* `end_week_day` - Day of the week on which the time range ends.
* `start_hour` - Hour (in UTC) at which the time range begins.
* `start_week_day` - Day of the week on which the time range begins.
