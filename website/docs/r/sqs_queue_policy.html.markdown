---
subcategory: "SQS (Simple Queue)"
layout: "aws"
page_title: "AWS: aws_sqs_queue_policy"
description: |-
  Provides a SQS Queue Policy resource.
---

# Resource: aws_sqs_queue_policy

Allows you to set a policy of an SQS Queue while referencing the ARN of the queue within the policy.

!> AWS will hang indefinitely when creating or updating an [`aws_sqs_queue`](/docs/providers/aws/r/sqs_queue.html) with an associated policy if `Version = "2012-10-17"` is not explicitly set in the policy. [See below](#timeout-problems-creatingupdating) for an example of how to avoid this issue.

## Example Usage

### Basic Usage

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

### Timeout Problems Creating/Updating

If `Version = "2012-10-17"` is not explicitly set in the policy, AWS may hang, causing the AWS provider to time out. To avoid this, make sure to include `Version` as shown in the example below.

```terraform
resource "aws_s3_bucket" "example" {
  bucket = "brodobaggins"
}

resource "aws_sqs_queue" "example" {
  name = "be-giant"
}

resource "aws_sqs_queue_policy" "example" {
  queue_url = aws_sqs_queue.example.id

  policy = jsonencode({
    Version = "2012-10-17" # !! Important !!
    Statement = [{
      Sid    = "Cejuwdam"
      Effect = "Allow"
      Principal = {
        Service = "s3.amazonaws.com"
      }
      Action   = "SQS:SendMessage"
      Resource = aws_sqs_queue.example.arn
      Condition = {
        ArnLike = {
          "aws:SourceArn" = aws_s3_bucket.example.arn
        }
      }
    }]
  })
}
```

## Argument Reference

This resource supports the following arguments:

* `policy` - (Required) JSON policy for the SQS queue. For more information about building AWS IAM policy documents with Terraform, see the [AWS IAM Policy Document Guide](https://learn.hashicorp.com/terraform/aws/iam-policy). Ensure that `Version = "2012-10-17"` is set in the policy or AWS may hang in creating the queue.
* `queue_url` - (Required) URL of the SQS Queue to which to attach the policy.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import SQS Queue Policies using the queue URL. For example:

```terraform
import {
  to = aws_sqs_queue_policy.test
  id = "https://queue.amazonaws.com/123456789012/myqueue"
}
```

Using `terraform import`, import SQS Queue Policies using the queue URL. For example:

```console
% terraform import aws_sqs_queue_policy.test https://queue.amazonaws.com/123456789012/myqueue
```
