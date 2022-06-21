---
subcategory: "CE (Cost Explorer)"
layout: "aws"
page_title: "AWS: aws_ce_anomaly_subscription"
description: |-
  Provides a CE Cost Anomaly Subscription
---

# Resource: aws_ce_anomaly_subscription

Provides a CE Anomaly Subscription.

## Example Usage

### Basic Example

```terraform
resource "aws_ce_anomaly_monitor" "test" {
  name      = "AWSServiceMonitor"
  type      = "DIMENSIONAL"
  dimension = "SERVICE"
}

resource "aws_ce_anomaly_subscription" "test" {
  name      = "DAILYSUBSCRIPTION"
  threshold = 100
  frequency = "DAILY"

  monitor_arn_list = [
    aws_ce_anomaly_monitor.test.arn,
  ]

  subscriber {
    type    = "EMAIL"
    address = "abc@example.com"
  }
}
```

### SNS Example

```terraform
resource "aws_sns_topic" "cost_anomaly_updates" {
  name = "CostAnomalyUpdates"
}

data "aws_iam_policy_document" "sns_topic_policy" {
  policy_id = "__default_policy_ID"

  statement {
    sid = "AWSAnomalyDetectionSNSPublishingPermissions"

    actions = [
      "SNS:Publish",
    ]

    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["costalerts.amazonaws.com"]
    }

    resources = [
      aws_sns_topic.cost_anomaly_updates.arn,
    ]
  }

  statement {
    sid = "__default_statement_ID"

    actions = [
      "SNS:Subscribe",
      "SNS:SetTopicAttributes",
      "SNS:RemovePermission",
      "SNS:Receive",
      "SNS:Publish",
      "SNS:ListSubscriptionsByTopic",
      "SNS:GetTopicAttributes",
      "SNS:DeleteTopic",
      "SNS:AddPermission",
    ]

    condition {
      test     = "StringEquals"
      variable = "AWS:SourceOwner"

      values = [
        var.account-id,
      ]
    }

    effect = "Allow"

    principals {
      type        = "AWS"
      identifiers = ["*"]
    }

    resources = [
      aws_sns_topic.cost_anomaly_updates.arn,
    ]
  }
}

resource "aws_sns_topic_policy" "default" {
  arn = aws_sns_topic.cost_anomaly_updates.arn

  policy = data.aws_iam_policy_document.sns_topic_policy.json
}

resource "aws_ce_anomaly_monitor" "anomaly_monitor" {
  name      = "AWSServiceMonitor"
  type      = "DIMENSIONAL"
  dimension = "SERVICE"
}

resource "aws_ce_anomaly_subscription" "realtime_subscription" {
  name      = "RealtimeAnomalySubscription"
  threshold = 0
  frequency = "IMMEDIATE"

  monitor_arn_list = [
    aws_ce_anomaly_monitor.anomaly_monitor.arn,
  ]

  subscriber {
    type    = "SNS"
    address = aws_sns_topic.cost_anomaly_updates.arn
  }

  depends_on = [
    aws_sns_topic_policy.default,
  ]
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) The name for the subscription.
* `frequency` - (Required) The frequency that anomaly reports are sent. Valid Values: `DAILY` | `IMMEDIATE` | `WEEKLY`.
* `monitor_arn_list` - (Required) A list of cost anomaly monitors.
* `subscriber` - (Required) A subscriber configuration. Multiple subscribers can be defined.
    * `type` - (Required) The type of subscription. Valid Values: `SNS` | `EMAIL`.
    * `address` - (Required) The address of the subscriber. If type is `SNS`, this will be the arn of the sns topic. If type is `EMAIL`, this will be the destination email address.
* `threshold` - (Required) The dollar value that triggers a notification if the threshold is exceeded.
* `account_id` - (Optional) The unique identifier for the AWS account in which the anomaly subscription ought to be created.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of the anomaly subscription.
* `id` - Unique ID of the anomaly subscription. Same as `arn`.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

`aws_ce_anomaly_subscription` can be imported using the `id`, e.g.

```
$ terraform import aws_ce_anomaly_subscription.example AnomalySubscriptionARN
```
