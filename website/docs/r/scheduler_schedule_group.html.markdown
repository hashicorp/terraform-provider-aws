---
subcategory: "EventBridge Scheduler"
layout: "aws"
page_title: "AWS: aws_scheduler_schedule_group"
description: |-
  Provides an EventBridge Scheduler Schedule Group resource.
---

# Resource: aws_scheduler_schedule_group

Provides an EventBridge Scheduler Schedule Group resource.

You can find out more about EventBridge Scheduler in the [User Guide](https://docs.aws.amazon.com/scheduler/latest/UserGuide/what-is-scheduler.html).

~> **Note:** EventBridge was formerly known as CloudWatch Events. The functionality is identical.

## Example Usage

```terraform
resource "aws_scheduler_schedule_group" "example" {
  name = "my-schedule-group"
}
```

## Argument Reference

The following arguments are optional:

* `name` - (Optional, Forces new resource) Name of the schedule group. If omitted, Terraform will assign a random, unique name. Conflicts with `name_prefix`.
* `name_prefix` - (Optional, Forces new resource) Creates a unique name beginning with the specified prefix. Conflicts with `name`.
* `tags` - (Optional) Key-value mapping of resource tags. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Name of the schedule group.
* `arn` - ARN of the schedule group.
* `creation_date` - Time at which the schedule group was created.
* `last_modification_date` - Time at which the schedule group was last modified.
* `state` - State of the schedule group. Can be `ACTIVE` or `DELETING`.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `5m`)
- `delete` - (Default `5m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import schedule groups using the `name`. For example:

```terraform
import {
  to = aws_scheduler_schedule_group.example
  id = "my-schedule-group"
}
```

Using `terraform import`, import schedule groups using the `name`. For example:

```console
% terraform import aws_scheduler_schedule_group.example my-schedule-group
```
