---
subcategory: "EC2 (Elastic Compute Cloud)"
layout: "aws"
page_title: "AWS: aws_ec2_instance_event_window"
description: |-
  Provides an EC2 Instance Event Window resource.
---

# Resource: aws_ec2_instance_event_window

Provides an EC2 Instance Event Window resource. Event windows define a schedule of time ranges in which AWS can schedule maintenance events (such as reboot, stop, or terminate) for associated EC2 instances or Dedicated Hosts.

Use the [`aws_ec2_instance_event_window_association` resource](ec2_instance_event_window_association.html.markdown) to associate instances, instance tags, or Dedicated Hosts with an event window.

## Example Usage

### Using Time Ranges

```terraform
resource "aws_ec2_instance_event_window" "example" {
  name = "example"

  time_ranges {
    start_week_day = "sunday"
    start_hour     = 2
    end_week_day   = "sunday"
    end_hour       = 6
  }

  tags = {
    Name = "example"
  }
}
```

### Using a Cron Expression

```terraform
resource "aws_ec2_instance_event_window" "example" {
  name            = "example"
  cron_expression = "0 4-6 ? * SUN *"
}
```

## Argument Reference

This resource supports the following arguments:

* `cron_expression` - (Optional) Cron expression for the event window, defining a recurring schedule. Conflicts with `time_ranges`. Exactly one of `cron_expression` or `time_ranges` must be specified.
* `name` - (Optional) Name of the event window.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `time_ranges` - (Optional) One or more time ranges defining the event window. Conflicts with `cron_expression`. Exactly one of `cron_expression` or `time_ranges` must be specified. All times are in UTC. Each individual time range must be at least 2 hours, and the combined duration of all time ranges must total at least 4 hours. See [`time_ranges`](#time_ranges) below.

### `time_ranges` Block

* `end_hour` - (Required) Hour (in UTC) at which the time range ends. Valid values: `0` to `23`.
* `end_week_day` - (Required) Day of the week on which the time range ends. Valid values: `sunday`, `monday`, `tuesday`, `wednesday`, `thursday`, `friday`, `saturday`.
* `start_hour` - (Required) Hour (in UTC) at which the time range begins. Valid values: `0` to `23`.
* `start_week_day` - (Required) Day of the week on which the time range begins. Valid values: `sunday`, `monday`, `tuesday`, `wednesday`, `thursday`, `friday`, `saturday`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - ID of the event window.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_ec2_instance_event_window.example
  identity = {
    id = "iew-0123456789abcdef0"
  }
}

resource "aws_ec2_instance_event_window" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `id` - (String) ID of the event window.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import EC2 Instance Event Windows using the `id`. For example:

```terraform
import {
  to = aws_ec2_instance_event_window.example
  id = "iew-0123456789abcdef0"
}
```

Using `terraform import`, import EC2 Instance Event Windows using the `id`. For example:

```console
% terraform import aws_ec2_instance_event_window.example iew-0123456789abcdef0
```
