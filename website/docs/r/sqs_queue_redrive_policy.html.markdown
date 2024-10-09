---
subcategory: "SQS (Simple Queue)"
layout: "aws"
page_title: "AWS: aws_sqs_queue_redrive_policy"
description: |-
  Provides a SQS Queue Redrive Policy resource.
---

# Resource: aws_sqs_queue_redrive_policy

Allows you to set a redrive policy of an SQS Queue
while referencing ARN of the dead letter queue inside the redrive policy.

This is useful when you want to set a dedicated
dead letter queue for a standard or FIFO queue, but need
the dead letter queue to exist before setting the redrive policy.

## Example Usage

```terraform
resource "aws_sqs_queue" "q" {
  name = "examplequeue"
}

resource "aws_sqs_queue" "ddl" {
  name = "examplequeue-ddl"
  redrive_allow_policy = jsonencode({
    redrivePermission = "byQueue",
    sourceQueueArns   = [aws_sqs_queue.q.arn]
  })
}

resource "aws_sqs_queue_redrive_policy" "q" {
  queue_url = aws_sqs_queue.q.id
  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.ddl.arn
    maxReceiveCount     = 4
  })
}
```

## Argument Reference

This resource supports the following arguments:

* `queue_url` - (Required) The URL of the SQS Queue to which to attach the policy
* `redrive_policy` - (Required) The JSON redrive policy for the SQS queue. Accepts two key/val pairs: `deadLetterTargetArn` and `maxReceiveCount`. Learn more in the [Amazon SQS dead-letter queues documentation](https://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/sqs-dead-letter-queues.html).

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import SQS Queue Redrive Policies using the queue URL. For example:

```terraform
import {
  to = aws_sqs_queue_redrive_policy.test
  id = "https://queue.amazonaws.com/0123456789012/myqueue"
}
```

Using `terraform import`, import SQS Queue Redrive Policies using the queue URL. For example:

```console
% terraform import aws_sqs_queue_redrive_policy.test https://queue.amazonaws.com/0123456789012/myqueue
```
