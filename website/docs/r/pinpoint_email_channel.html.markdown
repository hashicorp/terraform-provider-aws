---
subcategory: "End User Messaging"
layout: "aws"
page_title: "AWS: aws_pinpoint_email_channel"
description: |-
  Provides an End User Messaging Email Channel resource.
---

# Resource: aws_pinpoint_email_channel

~> **NOTE:** This resource is deprecated. AWS End User Messaging email features are being discontinued on October 30, 2026. Migrate to Amazon SES using [`aws_ses_domain_identity`](ses_domain_identity.html), [`aws_sesv2_email_identity`](sesv2_email_identity.html), and related SES/SESv2 resources. See the [AWS End User Messaging migration guide](https://docs.aws.amazon.com/pinpoint/latest/userguide/migrate.html) for details.

Provides an End User Messaging Email Channel resource.

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

* `application_id` - (Required, **Deprecated**) Application ID.
* `configuration_set` - (Optional, **Deprecated**) ARN of the Amazon SES configuration set that you want to apply to messages that you send through the channel.
* `enabled` - (Optional, **Deprecated**) Whether the channel is enabled or disabled. Defaults to `true`.
* `from_address` - (Required, **Deprecated**) Email address used to send emails from. You can use email only (`user@example.com`) or friendly address (`User <user@example.com>`). This field comply with [RFC 5322](https://www.ietf.org/rfc/rfc5322.txt).
* `identity` - (Required, **Deprecated**) ARN of an identity verified with SES.
* `orchestration_sending_role_arn` - (Optional, **Deprecated**) ARN of an IAM role for AWS End User Messaging to use to send email from your campaigns or journeys through Amazon SES.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `role_arn` - (Optional, **Deprecated**) ARN of an IAM Role used to submit events to Mobile Analytics' event ingestion service.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `messages_per_second` - (**Deprecated**) Messages per second that can be sent.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import End User Messaging Email Channel using the `application-id`. For example:

```terraform
import {
  to = aws_pinpoint_email_channel.email
  id = "application-id"
}
```

Using `terraform import`, import End User Messaging Email Channel using the `application-id`. For example:

```console
% terraform import aws_pinpoint_email_channel.email application-id
```
