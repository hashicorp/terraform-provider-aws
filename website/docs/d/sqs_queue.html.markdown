---
subcategory: "SQS (Simple Queue)"
layout: "aws"
page_title: "AWS: aws_sqs_queue"
description: |-
  Get information on an Amazon Simple Queue Service (SQS) Queue
---

# Data Source: aws_sqs_queue

Use this data source to get the ARN and URL of queue in AWS Simple Queue Service (SQS).
By using this data source, you can reference SQS queues without having to hardcode
the ARNs as input.

## Example Usage

```terraform
data "aws_sqs_queue" "example" {
  name = "queue"
}
```

## Argument Reference

* `name` - (Required) Name of the queue to match.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the queue.
* `url` - URL of the queue.
* `tags` - Map of tags for the resource.
