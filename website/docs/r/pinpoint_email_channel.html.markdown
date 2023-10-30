---
subcategory: "Pinpoint"
layout: "aws"
page_title: "AWS: aws_pinpoint_email_channel"
description: |-
  Provides a Pinpoint Email Channel resource.
---

# Resource: aws_pinpoint_email_channel

Provides a Pinpoint Email Channel resource.

## Example Usage

```terraform
resource "aws_pinpoint_email_channel" "email" {
  application_id = aws_pinpoint_app.app.application_id
  from_address   = "user@example.com"
  role_arn       = aws_iam_role.role.arn
}

resource "aws_pinpoint_app" "app" {}

resource "aws_ses_domain_identity" "identity" {
  domain = "example.com"
}

data "aws_iam_policy_document" "assume_role" {
  statement {
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["pinpoint.amazonaws.com"]
    }

    actions = ["sts:AssumeRole"]
  }
}

resource "aws_iam_role" "role" {
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

data "aws_iam_policy_document" "role_policy" {
  statement {
    effect = "Allow"

    actions = [
      "mobileanalytics:PutEvents",
      "mobileanalytics:PutItems",
    ]

    resources = ["*"]
  }
}

resource "aws_iam_role_policy" "role_policy" {
  name   = "role_policy"
  role   = aws_iam_role.role.id
  policy = data.aws_iam_policy_document.role_policy.json
}
```

## Argument Reference

This resource supports the following arguments:

* `application_id` - (Required) The application ID.
* `enabled` - (Optional) Whether the channel is enabled or disabled. Defaults to `true`.
* `configuration_set` - (Optional) The ARN of the Amazon SES configuration set that you want to apply to messages that you send through the channel.
* `from_address` - (Required) The email address used to send emails from. You can use email only (`user@example.com`) or friendly address (`User <user@example.com>`). This field comply with [RFC 5322](https://www.ietf.org/rfc/rfc5322.txt).
* `identity` - (Required) The ARN of an identity verified with SES.
* `role_arn` - (Optional) The ARN of an IAM Role used to submit events to Mobile Analytics' event ingestion service.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `messages_per_second` - Messages per second that can be sent.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Pinpoint Email Channel using the `application-id`. For example:

```terraform
import {
  to = aws_pinpoint_email_channel.email
  id = "application-id"
}
```

Using `terraform import`, import Pinpoint Email Channel using the `application-id`. For example:

```console
% terraform import aws_pinpoint_email_channel.email application-id
```
