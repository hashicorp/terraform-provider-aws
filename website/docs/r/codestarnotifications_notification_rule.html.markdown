---
subcategory: "CodeStar Notifications"
layout: "aws"
page_title: "AWS: aws_codestarnotifications_notification_rule"
description: |-
  Provides a CodeStar Notifications Rule
---

# Resource: aws_codestarnotifications_notification_rule

Provides a CodeStar Notifications Rule.

## Example Usage

```terraform
resource "aws_codecommit_repository" "code" {
  repository_name = "example-code-repo"
}

resource "aws_sns_topic" "notif" {
  name = "notification"
}

data "aws_iam_policy_document" "notif_access" {
  statement {
    actions = ["sns:Publish"]

    principals {
      type        = "Service"
      identifiers = ["codestar-notifications.amazonaws.com"]
    }

    resources = [aws_sns_topic.notif.arn]
  }
}

resource "aws_sns_topic_policy" "default" {
  arn    = aws_sns_topic.notif.arn
  policy = data.aws_iam_policy_document.notif_access.json
}

resource "aws_codestarnotifications_notification_rule" "commits" {
  detail_type    = "BASIC"
  event_type_ids = ["codecommit-repository-comments-on-commits"]

  name     = "example-code-repo-commits"
  resource = aws_codecommit_repository.code.arn

  target {
    address = aws_sns_topic.notif.arn
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `detail_type` - (Required) The level of detail to include in the notifications for this resource. Possible values are `BASIC` and `FULL`.
* `event_type_ids` - (Required) A list of event types associated with this notification rule.
  For list of allowed events see [here](https://docs.aws.amazon.com/codestar-notifications/latest/userguide/concepts.html#concepts-api).
* `name` - (Required) The name of notification rule.
* `resource` - (Required) The ARN of the resource to associate with the notification rule.
* `status` - (Optional) The status of the notification rule. Possible values are `ENABLED` and `DISABLED`, default is `ENABLED`.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `target` - (Optional) Configuration blocks containing notification target information. Can be specified multiple times. At least one target must be specified on creation.

An `target` block supports the following arguments:

* `address` - (Required) The ARN of notification rule target. For example, a SNS Topic ARN.
* `type` - (Optional) The type of the notification target. Default value is `SNS`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The codestar notification rule ARN.
* `arn` - The codestar notification rule ARN.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import CodeStar notification rule using the ARN. For example:

```terraform
import {
  to = aws_codestarnotifications_notification_rule.foo
  id = "arn:aws:codestar-notifications:us-west-1:0123456789:notificationrule/2cdc68a3-8f7c-4893-b6a5-45b362bd4f2b"
}
```

Using `terraform import`, import CodeStar notification rule using the ARN. For example:

```console
% terraform import aws_codestarnotifications_notification_rule.foo arn:aws:codestar-notifications:us-west-1:0123456789:notificationrule/2cdc68a3-8f7c-4893-b6a5-45b362bd4f2b
```
