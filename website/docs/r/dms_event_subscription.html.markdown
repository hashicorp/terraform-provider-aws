---
subcategory: "DMS (Database Migration)"
layout: "aws"
page_title: "AWS: aws_dms_event_subscription"
description: |-
  Provides a DMS (Data Migration Service) event subscription resource.
---

# Resource: aws_dms_event_subscription

Provides a DMS (Data Migration Service) event subscription resource.

## Example Usage

```terraform
resource "aws_dms_event_subscription" "example" {
  enabled          = true
  event_categories = ["creation", "failure"]
  name             = "my-favorite-event-subscription"
  sns_topic_arn    = aws_sns_topic.example.arn
  source_ids       = [aws_dms_replication_task.example.replication_task_id]
  source_type      = "replication-task"

  tags = {
    Name = "example"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `name` - (Required) Name of event subscription.
* `enabled` - (Optional, Default: true) Whether the event subscription should be enabled.
* `event_categories` - (Optional) List of event categories to listen for, see `DescribeEventCategories` for a canonical list.
* `sns_topic_arn` - (Required) SNS topic arn to send events on.
* `source_ids` - (Optional) Ids of sources to listen to. If you don't specify a value, notifications are provided for all sources.
* `source_type` - (Required) Type of source for events. Valid values: `replication-instance` or `replication-task`
* `tags` - (Optional) Map of resource tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Amazon Resource Name (ARN) of the DMS Event Subscription.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `10m`)
- `update` - (Default `10m`)
- `delete` - (Default `10m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import event subscriptions using the `name`. For example:

```terraform
import {
  to = aws_dms_event_subscription.test
  id = "my-awesome-event-subscription"
}
```

Using `terraform import`, import event subscriptions using the `name`. For example:

```console
% terraform import aws_dms_event_subscription.test my-awesome-event-subscription
```
