---
subcategory: "SES (Simple Email)"
layout: "aws"
page_title: "AWS: aws_ses_identity_notification_topic"
description: |-
  Setting AWS SES Identity Notification Topic
---

# Resource: aws_ses_identity_notification_topic

Resource for managing SES Identity Notification Topics

## Example Usage

```terraform
resource "aws_ses_identity_notification_topic" "test" {
  topic_arn                = aws_sns_topic.example.arn
  notification_type        = "Bounce"
  identity                 = aws_ses_domain_identity.example.domain
  include_original_headers = true
}
```

## Argument Reference

This resource supports the following arguments:

* `topic_arn` - (Optional) The Amazon Resource Name (ARN) of the Amazon SNS topic. Can be set to `""` (an empty string) to disable publishing.
* `notification_type` - (Required) The type of notifications that will be published to the specified Amazon SNS topic. Valid Values: `Bounce`, `Complaint` or `Delivery`.
* `identity` - (Required) The identity for which the Amazon SNS topic will be set. You can specify an identity by using its name or by using its Amazon Resource Name (ARN).
* `include_original_headers` - (Optional) Whether SES should include original email headers in SNS notifications of this type. `false` by default.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Identity Notification Topics using the ID of the record. The ID is made up as `IDENTITY|TYPE` where `IDENTITY` is the SES Identity and `TYPE` is the Notification Type. For example:

```terraform
import {
  to = aws_ses_identity_notification_topic.test
  id = "example.com|Bounce"
}
```

Using `terraform import`, import Identity Notification Topics using the ID of the record. The ID is made up as `IDENTITY|TYPE` where `IDENTITY` is the SES Identity and `TYPE` is the Notification Type. For example:

```console
% terraform import aws_ses_identity_notification_topic.test 'example.com|Bounce'
```
