---
subcategory: "SNS (Simple Notification)"
layout: "aws"
page_title: "AWS: aws_sns_topic"
description: |-
  Provides an SNS topic resource.
---

# Resource: aws_sns_topic

Provides an SNS topic resource

## Example Usage

```terraform
resource "aws_sns_topic" "user_updates" {
  name = "user-updates-topic"
}
```

## Example with Delivery Policy

```hcl
resource "aws_sns_topic" "user_updates" {
  name            = "user-updates-topic"
  delivery_policy = <<EOF
{
  "http": {
    "defaultHealthyRetryPolicy": {
      "minDelayTarget": 20,
      "maxDelayTarget": 20,
      "numRetries": 3,
      "numMaxDelayRetries": 0,
      "numNoDelayRetries": 0,
      "numMinDelayRetries": 0,
      "backoffFunction": "linear"
    },
    "disableSubscriptionOverrides": false,
    "defaultThrottlePolicy": {
      "maxReceivesPerSecond": 1
    }
  }
}
EOF
}
```

## Example with Server-side encryption (SSE)

```terraform
resource "aws_sns_topic" "user_updates" {
  name              = "user-updates-topic"
  kms_master_key_id = "alias/aws/sns"
}
```

## Example with First-In-First-Out (FIFO)

```hcl
resource "aws_sns_topic" "user_updates" {
  name                        = "user-updates-topic.fifo"
  fifo_topic                  = true
  content_based_deduplication = true
}
```

## Message Delivery Status Arguments

The `<endpoint>_success_feedback_role_arn` and `<endpoint>_failure_feedback_role_arn` arguments are used to give Amazon SNS write access to use CloudWatch Logs on your behalf. The `<endpoint>_success_feedback_sample_rate` argument is for specifying the sample rate percentage (0-100) of successfully delivered messages. After you configure the  `<endpoint>_failure_feedback_role_arn` argument, then all failed message deliveries generate CloudWatch Logs.

## Argument Reference

This resource supports the following arguments:

* `name` - (Optional) The name of the topic. Topic names must be made up of only uppercase and lowercase ASCII letters, numbers, underscores, and hyphens, and must be between 1 and 256 characters long. For a FIFO (first-in-first-out) topic, the name must end with the `.fifo` suffix. If omitted, Terraform will assign a random, unique name. Conflicts with `name_prefix`
* `name_prefix` - (Optional) Creates a unique name beginning with the specified prefix. Conflicts with `name`
* `display_name` - (Optional) The display name for the topic
* `policy` - (Optional) The fully-formed AWS policy as JSON. For more information about building AWS IAM policy documents with Terraform, see the [AWS IAM Policy Document Guide](https://learn.hashicorp.com/terraform/aws/iam-policy).
* `delivery_policy` - (Optional) The SNS delivery policy. More details in the [AWS documentation](https://docs.aws.amazon.com/sns/latest/dg/DeliveryPolicies.html).
* `application_success_feedback_role_arn` - (Optional) The IAM role permitted to receive success feedback for this topic
* `application_success_feedback_sample_rate` - (Optional) Percentage of success to sample
* `application_failure_feedback_role_arn` - (Optional) IAM role for failure feedback
* `http_success_feedback_role_arn` - (Optional) The IAM role permitted to receive success feedback for this topic
* `http_success_feedback_sample_rate` - (Optional) Percentage of success to sample
* `http_failure_feedback_role_arn` - (Optional) IAM role for failure feedback
* `kms_master_key_id` - (Optional) The ID of an AWS-managed customer master key (CMK) for Amazon SNS or a custom CMK. For more information, see [Key Terms](https://docs.aws.amazon.com/sns/latest/dg/sns-server-side-encryption.html#sse-key-terms)
* `signature_version` - (Optional) If `SignatureVersion` should be [1 (SHA1) or 2 (SHA256)](https://docs.aws.amazon.com/sns/latest/dg/sns-verify-signature-of-message.html). The signature version corresponds to the hashing algorithm used while creating the signature of the notifications, subscription confirmations, or unsubscribe confirmation messages sent by Amazon SNS.
* `tracing_config` - (Optional) Tracing mode of an Amazon SNS topic. Valid values: `"PassThrough"`, `"Active"`.
* `fifo_topic` - (Optional) Boolean indicating whether or not to create a FIFO (first-in-first-out) topic. FIFO topics can't deliver messages to customer managed endpoints, such as email addresses, mobile apps, SMS, or HTTP(S) endpoints. These endpoint types aren't guaranteed to preserve strict message ordering. Default is `false`.
* `archive_policy` - (Optional) The message archive policy for FIFO topics. More details in the [AWS documentation](https://docs.aws.amazon.com/sns/latest/dg/message-archiving-and-replay-topic-owner.html).
* `content_based_deduplication` - (Optional) Enables content-based deduplication for FIFO topics. For more information, see the [related documentation](https://docs.aws.amazon.com/sns/latest/dg/fifo-message-dedup.html)
* `lambda_success_feedback_role_arn` - (Optional) The IAM role permitted to receive success feedback for this topic
* `lambda_success_feedback_sample_rate` - (Optional) Percentage of success to sample
* `lambda_failure_feedback_role_arn` - (Optional) IAM role for failure feedback
* `sqs_success_feedback_role_arn` - (Optional) The IAM role permitted to receive success feedback for this topic
* `sqs_success_feedback_sample_rate` - (Optional) Percentage of success to sample
* `sqs_failure_feedback_role_arn` - (Optional) IAM role for failure feedback
* `firehose_success_feedback_role_arn` - (Optional) The IAM role permitted to receive success feedback for this topic
* `firehose_success_feedback_sample_rate` - (Optional) Percentage of success to sample
* `firehose_failure_feedback_role_arn` - (Optional) IAM role for failure feedback
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The ARN of the SNS topic
* `arn` - The ARN of the SNS topic, as a more obvious property (clone of id)
* `beginning_archive_time` - The oldest timestamp at which a FIFO topic subscriber can start a replay.
* `owner` - The AWS Account ID of the SNS topic owner
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import SNS Topics using the topic `arn`. For example:

```terraform
import {
  to = aws_sns_topic.user_updates
  id = "arn:aws:sns:us-west-2:0123456789012:my-topic"
}
```

Using `terraform import`, import SNS Topics using the topic `arn`. For example:

```console
% terraform import aws_sns_topic.user_updates arn:aws:sns:us-west-2:0123456789012:my-topic
```
