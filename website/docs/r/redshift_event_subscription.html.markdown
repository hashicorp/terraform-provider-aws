---
subcategory: "Redshift"
layout: "aws"
page_title: "AWS: aws_redshift_event_subscription"
description: |-
  Provides a Redshift event subscription resource.
---

# Resource: aws_redshift_event_subscription

Provides a Redshift event subscription resource.

## Example Usage

```terraform
resource "aws_redshift_cluster" "default" {
  cluster_identifier = "default"
  database_name      = "default"

  # ...
}

resource "aws_sns_topic" "default" {
  name = "redshift-events"
}

resource "aws_redshift_event_subscription" "default" {
  name          = "redshift-event-sub"
  sns_topic_arn = aws_sns_topic.default.arn

  source_type = "cluster"
  source_ids  = [aws_redshift_cluster.default.id]

  severity = "INFO"

  event_categories = [
    "configuration",
    "management",
    "monitoring",
    "security",
  ]

  tags = {
    Name = "default"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `name` - (Required) The name of the Redshift event subscription.
* `sns_topic_arn` - (Required) The ARN of the SNS topic to send events to.
* `source_ids` - (Optional) A list of identifiers of the event sources for which events will be returned. If not specified, then all sources are included in the response. If specified, a `source_type` must also be specified.
* `source_type` - (Optional) The type of source that will be generating the events. Valid options are `cluster`, `cluster-parameter-group`, `cluster-security-group`, `cluster-snapshot`, or `scheduled-action`. If not set, all sources will be subscribed to.
* `severity` - (Optional) The event severity to be published by the notification subscription. Valid options are `INFO` or `ERROR`. Default value of `INFO`.
* `event_categories` - (Optional) A list of event categories for a SourceType that you want to subscribe to. See https://docs.aws.amazon.com/redshift/latest/mgmt/working-with-event-notifications.html or run `aws redshift describe-event-categories`.
* `enabled` - (Optional) A boolean flag to enable/disable the subscription. Defaults to `true`.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Amazon Resource Name (ARN) of the Redshift event notification subscription
* `id` - The name of the Redshift event notification subscription
* `customer_aws_id` - The AWS customer account associated with the Redshift event notification subscription
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Redshift Event Subscriptions using the `name`. For example:

```terraform
import {
  to = aws_redshift_event_subscription.default
  id = "redshift-event-sub"
}
```

Using `terraform import`, import Redshift Event Subscriptions using the `name`. For example:

```console
% terraform import aws_redshift_event_subscription.default redshift-event-sub
```
