---
subcategory: "End User Messaging"
layout: "aws"
page_title: "AWS: aws_pinpoint_event_stream"
description: |-
  Provides an End User Messaging Event Stream resource.
---

# Resource: aws_pinpoint_event_stream

~> **NOTE:** This resource is deprecated. AWS End User Messaging event streams are being discontinued on October 30, 2026. After that date, this resource will no longer be available. For SMS/Voice event delivery, use [`aws_pinpointsmsvoicev2_configuration_set`](pinpointsmsvoicev2_configuration_set.html) with an event destination.

Provides an End User Messaging Event Stream resource.

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

* `application_id` - (Required, **Deprecated**) Application ID.
* `destination_stream_arn` - (Required, **Deprecated**) Amazon Resource Name (ARN) of the Amazon Kinesis stream or Firehose delivery stream to which you want to publish events.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `role_arn` - (Required, **Deprecated**) IAM role that authorizes AWS End User Messaging to publish events to the stream in your account.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import End User Messaging Event Stream using the `application-id`. For example:

```terraform
import {
  to = aws_pinpoint_event_stream.stream
  id = "application-id"
}
```

Using `terraform import`, import End User Messaging Event Stream using the `application-id`. For example:

```console
% terraform import aws_pinpoint_event_stream.stream application-id
```
