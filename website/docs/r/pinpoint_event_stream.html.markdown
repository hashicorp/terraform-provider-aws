---
subcategory: "Pinpoint"
layout: "aws"
page_title: "AWS: aws_pinpoint_event_stream"
description: |-
  Provides a Pinpoint Event Stream resource.
---

# Resource: aws_pinpoint_event_stream

Provides a Pinpoint Event Stream resource.

## Example Usage

```terraform
resource "aws_pinpoint_event_stream" "stream" {
  application_id         = aws_pinpoint_app.app.application_id
  destination_stream_arn = aws_kinesis_stream.test_stream.arn
  role_arn               = aws_iam_role.test_role.arn
}

resource "aws_pinpoint_app" "app" {}

resource "aws_kinesis_stream" "test_stream" {
  name        = "pinpoint-kinesis-test"
  shard_count = 1
}

data "aws_iam_policy_document" "assume_role" {
  statement {
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["pinpoint.us-east-1.amazonaws.com"]
    }

    actions = ["sts:AssumeRole"]
  }
}

resource "aws_iam_role" "test_role" {
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

data "aws_iam_policy_document" "test_role_policy" {
  statement {
    effect = "Allow"

    actions = [
      "kinesis:PutRecords",
      "kinesis:DescribeStream",
    ]

    resources = ["arn:aws:kinesis:us-east-1:*:*/*"]
  }
}
resource "aws_iam_role_policy" "test_role_policy" {
  name   = "test_policy"
  role   = aws_iam_role.test_role.id
  policy = data.aws_iam_policy_document.test_role_policy.json
}
```

## Argument Reference

This resource supports the following arguments:

* `application_id` - (Required) The application ID.
* `destination_stream_arn` - (Required) The Amazon Resource Name (ARN) of the Amazon Kinesis stream or Firehose delivery stream to which you want to publish events.
* `role_arn` - (Required) The IAM role that authorizes Amazon Pinpoint to publish events to the stream in your account.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Pinpoint Event Stream using the `application-id`. For example:

```terraform
import {
  to = aws_pinpoint_event_stream.stream
  id = "application-id"
}
```

Using `terraform import`, import Pinpoint Event Stream using the `application-id`. For example:

```console
% terraform import aws_pinpoint_event_stream.stream application-id
```
