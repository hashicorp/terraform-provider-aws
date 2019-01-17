---
layout: "aws"
page_title: "AWS: aws_dms_event_subscription"
sidebar_current: "docs-aws-resource-dms-event-subscription"
description: |-
  Provides a DMS (Data Migration Service) event subscription resource.
---

# aws_dms_event_subscription

Provides a DMS (Data Migration Service) event subscription resource. DMS event subscriptions can be created, updated, deleted, and imported.

## Example Usage

```hcl
# Create a new DMS event subscription
resource "aws_dms_event_subscription" "test" {
  name = "my-favorite-event-subscription"
  enabled = true
  event_categories = ["creation", "failure"]
  source_type = "replication-task"
  source_ids = ["${aws_dms_replication_task.dms_replication_task.replication_task_id}"]
  sns_topic_arn    = "${aws_sns_topic.topic.arn}"
}

  tags = {
    Name = "test"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of event subscription.
* `enabled` - (Optional, Default: true) Whether the event subscription should be enabled.
* `event_categories` - (Optional) List of event categories to listen for, see `DescribeEventCategories` for a canonical list.
* `source_type` - (Optional, Default: all events) If specificed may be either `replication-instance` or `migration-task`
* `source_ids` - (Required) Ids of sources to listen to.
* `sns_topic_arn` - (Required) SNS topic arn to send events on.

<a id="timeouts"></a>
## Timeouts

`aws_dms_event_subscription` provides the following
[Timeouts](/docs/configuration/resources.html#timeouts) configuration options:

- `create` - (Default `30 minutes`) Used for creating event subscriptions.
- `update` - (Default `30 minutes`) Used for event subscription modifications.
- `delete` - (Default `30 minutes`) Used for destroying event descriptions.

## Import

Event subscriptions can be imported using the `name`, e.g.

```
$ terraform import aws_dms_event_subscription.test my-awesome-event-subscription
```
