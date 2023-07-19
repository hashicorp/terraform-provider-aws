---
subcategory: "SQS (Simple Queue)"
layout: "aws"
page_title: "AWS: aws_sqs_queue_redrive_allow_policy"
description: |-
  Provides a SQS Queue Redrive Allow Policy resource.
---

# Resource: aws_sqs_queue_redrive_allow_policy

Provides a SQS Queue Redrive Allow Policy resource.

## Example Usage

```terraform
resource "aws_sqs_queue" "src" {
  name = "srcqueue"

  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.example.arn
    maxReceiveCount     = 4
  })
}

resource "aws_sqs_queue" "example" {
  name = "examplequeue"
}

resource "aws_sqs_queue_redrive_allow_policy" "example" {
  queue_url = aws_sqs_queue.example.id

  redrive_allow_policy = jsonencode({
    redrivePermission = "byQueue",
    sourceQueueArns   = [aws_sqs_queue.src.arn]
  })
}
```

## Argument Reference

This resource supports the following arguments:

* `queue_url` - (Required) The URL of the SQS Queue to which to attach the policy
* `redrive_allow_policy` - (Required) The JSON redrive allow policy for the SQS queue. Learn more in the [Amazon SQS dead-letter queues documentation](https://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/sqs-dead-letter-queues.html).

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import SQS Queue Redrive Allow Policies using the queue URL. For example:

```terraform
import {
  to = aws_sqs_queue_redrive_allow_policy.test
  id = "https://queue.amazonaws.com/0123456789012/myqueue"
}
```

Using `terraform import`, import SQS Queue Redrive Allow Policies using the queue URL. For example:

```console
% terraform import aws_sqs_queue_redrive_allow_policy.test https://queue.amazonaws.com/0123456789012/myqueue
```
