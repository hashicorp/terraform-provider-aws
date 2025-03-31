---
subcategory: "SQS (Simple Queue)"
layout: "aws"
page_title: "AWS: aws_sqs_queue"
description: |-
  Provides a SQS resource.
---

# Resource: aws_sqs_queue

Amazon SQS (Simple Queue Service) is a fully managed message queuing service that enables decoupling and scaling of microservices, distributed systems, and serverless applications. This resource allows you to create, configure, and manage an SQS queue, which acts as a reliable message buffer between producers and consumers. With support for standard and FIFO queues, SQS ensures secure, scalable, and asynchronous message processing. Use this resource to define queue attributes, configure access policies, and integrate seamlessly with AWS services like Lambda, SNS, and EC2.

!> AWS will hang indefinitely, leading to a `timeout while waiting` error, when creating or updating an `aws_sqs_queue` with an associated [`aws_sqs_queue_policy`](/docs/providers/aws/r/sqs_queue_policy.html) if `Version = "2012-10-17"` is not explicitly set in the policy.

!> AWS will hang indefinitely and trigger a `timeout while waiting` error when creating or updating an `aws_sqs_queue` if `kms_data_key_reuse_period_seconds` is set to a non-default value, `sqs_managed_sse_enabled` is `false` (explicitly or by default), and `kms_master_key_id` is not set.

## Example Usage

```terraform
resource "aws_sqs_queue" "terraform_queue" {
  name                      = "terraform-example-queue"
  delay_seconds             = 90
  max_message_size          = 2048
  message_retention_seconds = 86400
  receive_wait_time_seconds = 10
  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.terraform_queue_deadletter.arn
    maxReceiveCount     = 4
  })

  tags = {
    Environment = "production"
  }
}
```

## FIFO queue

```terraform
resource "aws_sqs_queue" "terraform_queue" {
  name                        = "terraform-example-queue.fifo"
  fifo_queue                  = true
  content_based_deduplication = true
}
```

## High-throughput FIFO queue

```terraform
resource "aws_sqs_queue" "terraform_queue" {
  name                  = "terraform-example-queue.fifo"
  fifo_queue            = true
  deduplication_scope   = "messageGroup"
  fifo_throughput_limit = "perMessageGroupId"
}
```

## Dead-letter queue

```terraform
resource "aws_sqs_queue" "terraform_queue" {
  name = "terraform-example-queue"

  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.terraform_queue_deadletter.arn
    maxReceiveCount     = 4
  })
}

resource "aws_sqs_queue" "terraform_queue_deadletter" {
  name = "terraform-example-deadletter-queue"
}

resource "aws_sqs_queue_redrive_allow_policy" "terraform_queue_redrive_allow_policy" {
  queue_url = aws_sqs_queue.terraform_queue_deadletter.id

  redrive_allow_policy = jsonencode({
    redrivePermission = "byQueue",
    sourceQueueArns   = [aws_sqs_queue.terraform_queue.arn]
  })
}
```

## Server-side encryption (SSE)

Using [SSE-SQS](https://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/sqs-configure-sqs-sse-queue.html):

```terraform
resource "aws_sqs_queue" "terraform_queue" {
  name                    = "terraform-example-queue"
  sqs_managed_sse_enabled = true
}
```

Using [SSE-KMS](https://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/sqs-configure-sse-existing-queue.html):

```terraform
resource "aws_sqs_queue" "terraform_queue" {
  name                              = "terraform-example-queue"
  kms_master_key_id                 = "alias/aws/sqs"
  kms_data_key_reuse_period_seconds = 300
}
```

## Argument Reference

This resource supports the following arguments:

* `content_based_deduplication` - (Optional) Enables content-based deduplication for FIFO queues. For more information, see the [related documentation](http://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/FIFO-queues.html#FIFO-queues-exactly-once-processing).
* `deduplication_scope` - (Optional) Specifies whether message deduplication occurs at the message group or queue level. Valid values are `messageGroup` and `queue` (default).
* `delay_seconds` - (Optional) Time in seconds that the delivery of all messages in the queue will be delayed. An integer from 0 to 900 (15 minutes). The default for this attribute is 0 seconds.
* `fifo_queue` - (Optional) Boolean designating a FIFO queue. If not set, it defaults to `false` making it standard.
* `fifo_throughput_limit` - (Optional) Specifies whether the FIFO queue throughput quota applies to the entire queue or per message group. Valid values are `perQueue` (default) and `perMessageGroupId`.
* `kms_data_key_reuse_period_seconds` - (Optional) Length of time, in seconds, for which Amazon SQS can reuse a data key to encrypt or decrypt messages before calling AWS KMS again. An integer representing seconds, between 60 seconds (1 minute) and 86,400 seconds (24 hours). The default is 300 (5 minutes).
* `kms_master_key_id` - (Optional) ID of an AWS-managed customer master key (CMK) for Amazon SQS or a custom CMK. For more information, see [Key Terms](http://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/sqs-server-side-encryption.html#sqs-sse-key-terms).
* `max_message_size` - (Optional) Limit of how many bytes a message can contain before Amazon SQS rejects it. An integer from 1024 bytes (1 KiB) up to 262144 bytes (256 KiB). The default for this attribute is 262144 (256 KiB).
* `message_retention_seconds` - (Optional) Number of seconds Amazon SQS retains a message. Integer representing seconds, from 60 (1 minute) to 1209600 (14 days). The default for this attribute is 345600 (4 days).
* `name` - (Optional) Name of the queue. Queue names must be made up of only uppercase and lowercase ASCII letters, numbers, underscores, and hyphens, and must be between 1 and 80 characters long. For a FIFO (first-in-first-out) queue, the name must end with the `.fifo` suffix. If omitted, Terraform will assign a random, unique name. Conflicts with `name_prefix`.
* `name_prefix` - (Optional) Creates a unique name beginning with the specified prefix. Conflicts with `name`.
* `policy` - (Optional) JSON policy for the SQS queue. For more information about building AWS IAM policy documents with Terraform, see the [AWS IAM Policy Document Guide](https://learn.hashicorp.com/terraform/aws/iam-policy). Terraform will only perform drift detection of its value when present in a configuration. It is preferred to use the `aws_sqs_queue_policy` resource instead.
* `receive_wait_time_seconds` - (Optional) Time for which a ReceiveMessage call will wait for a message to arrive (long polling) before returning. An integer from 0 to 20 (seconds). The default for this attribute is 0, meaning that the call will return immediately.
* `redrive_allow_policy` - (Optional) JSON policy to set up the Dead Letter Queue redrive permission, see [AWS docs](https://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/SQSDeadLetterQueue.html). Terraform will only perform drift detection of its value when present in a configuration. It is preferred to use the `aws_sqs_queue_redrive_allow_policy` resource instead.
* `redrive_policy` - (Optional) JSON policy to set up the Dead Letter Queue, see [AWS docs](https://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/SQSDeadLetterQueue.html). Terraform will only perform drift detection of its value when present in a configuration. It is preferred to use the `aws_sqs_queue_redrive_policy` resource instead. **Note:** when specifying `maxReceiveCount`, you must specify it as an integer (`5`), and not a string (`"5"`).
* `sqs_managed_sse_enabled` - (Optional) Boolean to enable server-side encryption (SSE) of message content with SQS-owned encryption keys. See [Encryption at rest](https://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/sqs-server-side-encryption.html). Terraform will only perform drift detection of its value when present in a configuration.
* `tags` - (Optional) Map of tags to assign to the queue. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `visibility_timeout_seconds` - (Optional) Visibility timeout for the queue. An integer from 0 to 43200 (12 hours). The default for this attribute is 30. For more information about visibility timeout, see [AWS docs](https://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/AboutVT.html).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the SQS queue.
* `id` - URL for the created Amazon SQS queue.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `url` - Same as `id`: The URL for the created Amazon SQS queue.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `3m`)
- `update` - (Default `3m`)
- `delete` - (Default `3m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import SQS Queues using the queue `url`. For example:

```terraform
import {
  to = aws_sqs_queue.public_queue
  id = "https://queue.amazonaws.com/80398EXAMPLE/MyQueue"
}
```

Using `terraform import`, import SQS Queues using the queue `url`. For example:

```console
% terraform import aws_sqs_queue.public_queue https://queue.amazonaws.com/80398EXAMPLE/MyQueue
```
