---
subcategory: "SQS (Simple Queue)"
layout: "aws"
page_title: "AWS: aws_sqs_queue_policy"
description: |-
  Provides a SQS Queue Policy resource.
---

# Resource: aws_sqs_queue_policy

Allows you to set a policy of an SQS Queue
while referencing ARN of the queue within the policy.

## Example Usage

```terraform
resource "aws_sqs_queue" "q" {
  name = "examplequeue"
}

data "aws_iam_policy_document" "test" {
  statement {
    sid    = "First"
    effect = "Allow"

    principals {
      type        = "*"
      identifiers = ["*"]
    }

    actions   = ["sqs:SendMessage"]
    resources = [aws_sqs_queue.q.arn]

    condition {
      test     = "ArnEquals"
      variable = "aws:SourceArn"
      values   = [aws_sns_topic.example.arn]
    }
  }
}

resource "aws_sqs_queue_policy" "test" {
  queue_url = aws_sqs_queue.q.id
  policy    = data.aws_iam_policy_document.test.json
}
```

## Argument Reference

The following arguments are supported:

* `queue_url` - (Required) The URL of the SQS Queue to which to attach the policy
* `policy` - (Required) The JSON policy for the SQS queue. For more information about building AWS IAM policy documents with Terraform, see the [AWS IAM Policy Document Guide](https://learn.hashicorp.com/terraform/aws/iam-policy).

## Attributes Reference

No additional attributes are exported.

## Import

SQS Queue Policies can be imported using the queue URL, e.g.,

```
$ terraform import aws_sqs_queue_policy.test https://queue.amazonaws.com/0123456789012/myqueue
```
