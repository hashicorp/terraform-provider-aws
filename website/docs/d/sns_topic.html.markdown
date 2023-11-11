---
subcategory: "SNS (Simple Notification)"
layout: "aws"
page_title: "AWS: aws_sns_topic"
description: |-
  Get information on a Amazon Simple Notification Service (SNS) Topic
---

# Data Source: aws_sns_topic

Use this data source to get the ARN of a topic in AWS Simple Notification
Service (SNS). By using this data source, you can reference SNS topics
without having to hard code the ARNs as input.

## Example Usage

```terraform
data "aws_sns_topic" "example" {
  name = "an_example_topic"
}
```

## Argument Reference

* `name` - (Required) Friendly name of the topic to match.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the found topic, suitable for referencing in other resources that support SNS topics.
* `id` - ARN of the found topic, suitable for referencing in other resources that support SNS topics.
