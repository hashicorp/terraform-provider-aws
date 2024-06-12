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
  name              = "AWSServiceMonitor"
  monitor_type      = "DIMENSIONAL"
  monitor_dimension = "SERVICE"
}

resource "aws_ce_anomaly_subscription" "test" {
  name      = "DAILYSUBSCRIPTION"
  frequency = "DAILY"

  monitor_arn_list = [
    aws_ce_anomaly_monitor.test.arn
  ]

  subscriber {
    type    = "EMAIL"
    address = "abc@example.com"
  }

  threshold_expression {
    dimension {
      key           = "ANOMALY_TOTAL_IMPACT_ABSOLUTE"
      match_options = ["GREATER_THAN_OR_EQUAL"]
      values        = ["100"]
    }
  }
}
```

### Threshold Expression Example

#### Using a Percentage Threshold

```terraform
resource "aws_ce_anomaly_subscription" "test" {
  name      = "AWSServiceMonitor"
  frequency = "DAILY"

  monitor_arn_list = [
    aws_ce_anomaly_monitor.test.arn
  ]

  subscriber {
    type    = "EMAIL"
    address = "abc@example.com"
  }

  threshold_expression {
    dimension {
      key           = "ANOMALY_TOTAL_IMPACT_PERCENTAGE"
      match_options = ["GREATER_THAN_OR_EQUAL"]
      values        = ["100"]
    }
  }
}
```

#### Using an `and` Expression

```terraform
resource "aws_ce_anomaly_subscription" "test" {
  name      = "AWSServiceMonitor"
  frequency = "DAILY"

  monitor_arn_list = [
    aws_ce_anomaly_monitor.test.arn
  ]

  subscriber {
    type    = "EMAIL"
    address = "abc@example.com"
  }

  threshold_expression {
    and {
      dimension {
        key           = "ANOMALY_TOTAL_IMPACT_ABSOLUTE"
        match_options = ["GREATER_THAN_OR_EQUAL"]
        values        = ["100"]
      }
    }
    and {
      dimension {
        key           = "ANOMALY_TOTAL_IMPACT_PERCENTAGE"
        match_options = ["GREATER_THAN_OR_EQUAL"]
        values        = ["50"]
      }
    }
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
      "SNS:Publish"
    ]

    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["costalerts.amazonaws.com"]
    }

    resources = [
      aws_sns_topic.cost_anomaly_updates.arn
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
      "SNS:AddPermission"
    ]

    condition {
      test     = "StringEquals"
      variable = "AWS:SourceOwner"

      values = [
        var.account_id
      ]
    }

    effect = "Allow"

    principals {
      type        = "AWS"
      identifiers = ["*"]
    }

    resources = [
      aws_sns_topic.cost_anomaly_updates.arn
    ]
  }
}

resource "aws_sns_topic_policy" "default" {
  arn = aws_sns_topic.cost_anomaly_updates.arn

  policy = data.aws_iam_policy_document.sns_topic_policy.json
}

resource "aws_ce_anomaly_monitor" "anomaly_monitor" {
  name              = "AWSServiceMonitor"
  monitor_type      = "DIMENSIONAL"
  monitor_dimension = "SERVICE"
}

resource "aws_ce_anomaly_subscription" "realtime_subscription" {
  name      = "RealtimeAnomalySubscription"
  frequency = "IMMEDIATE"

  monitor_arn_list = [
    aws_ce_anomaly_monitor.anomaly_monitor.arn
  ]

  subscriber {
    type    = "SNS"
    address = aws_sns_topic.cost_anomaly_updates.arn
  }

  depends_on = [
    aws_sns_topic_policy.default
  ]
}
```

## Argument Reference

The following arguments are required:

* `account_id` - (Optional) The unique identifier for the AWS account in which the anomaly subscription ought to be created.
* `frequency` - (Required) The frequency that anomaly reports are sent. Valid Values: `DAILY` | `IMMEDIATE` | `WEEKLY`.
* `monitor_arn_list` - (Required) A list of cost anomaly monitors.
* `name` - (Required) The name for the subscription.
* `subscriber` - (Required) A subscriber configuration. Multiple subscribers can be defined.
    * `type` - (Required) The type of subscription. Valid Values: `SNS` | `EMAIL`.
    * `address` - (Required) The address of the subscriber. If type is `SNS`, this will be the arn of the sns topic. If type is `EMAIL`, this will be the destination email address.
* `threshold_expression` - (Optional) An Expression object used to specify the anomalies that you want to generate alerts for. See [Threshold Expression](#threshold-expression).
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### Threshold Expression

* `and` - (Optional) Return results that match both [Dimension](#dimension) objects.
* `cost_category` - (Optional) Configuration block for the filter that's based on  values. See [Cost Category](#cost-category) below.
* `dimension` - (Optional) Configuration block for the specific [Dimension](#dimension) to use for.
* `not` - (Optional) Return results that match both [Dimension](#dimension) object.
* `or` - (Optional) Return results that match both [Dimension](#dimension) object.
* `tags` - (Optional) Configuration block for the specific Tag to use for. See [Tags](#tags) below.

### Cost Category

* `key` - (Optional) Unique name of the Cost Category.
* `match_options` - (Optional) Match options that you can use to filter your results. MatchOptions is only applicable for actions related to cost category. The default values for MatchOptions is `EQUALS` and `CASE_SENSITIVE`. Valid values are: `EQUALS`,  `ABSENT`, `STARTS_WITH`, `ENDS_WITH`, `CONTAINS`, `CASE_SENSITIVE`, `CASE_INSENSITIVE`.
* `values` - (Optional) Specific value of the Cost Category.

### Dimension

* `key` - (Optional) Unique name of the Cost Category.
* `match_options` - (Optional) Match options that you can use to filter your results. MatchOptions is only applicable for actions related to cost category. The default values for MatchOptions is `EQUALS` and `CASE_SENSITIVE`. Valid values are: `EQUALS`,  `ABSENT`, `STARTS_WITH`, `ENDS_WITH`, `CONTAINS`, `CASE_SENSITIVE`, `CASE_INSENSITIVE`.
* `values` - (Optional) Specific value of the Cost Category.

### Tags

* `key` - (Optional) Key for the tag.
* `match_options` - (Optional) Match options that you can use to filter your results. MatchOptions is only applicable for actions related to cost category. The default values for MatchOptions is `EQUALS` and `CASE_SENSITIVE`. Valid values are: `EQUALS`,  `ABSENT`, `STARTS_WITH`, `ENDS_WITH`, `CONTAINS`, `CASE_SENSITIVE`, `CASE_INSENSITIVE`.
* `values` - (Optional) Specific value of the Cost Category.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the anomaly subscription.
* `id` - Unique ID of the anomaly subscription. Same as `arn`.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_ce_anomaly_subscription` using the `id`. For example:

```terraform
import {
  to = aws_ce_anomaly_subscription.example
  id = "AnomalySubscriptionARN"
}
```

Using `terraform import`, import `aws_ce_anomaly_subscription` using the `id`. For example:

```console
% terraform import aws_ce_anomaly_subscription.example AnomalySubscriptionARN
```
