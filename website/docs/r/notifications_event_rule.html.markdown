---
subcategory: "User Notifications"
layout: "aws"
page_title: "AWS: aws_notifications_event_rule"
description: |-
  Terraform resource for managing an AWS User Notifications Event Rule.
---
# Resource: aws_notifications_event_rule

Terraform resource for managing an AWS User Notifications Event Rule.

## Example Usage

### Basic Usage

```terraform
resource "aws_notifications_notification_configuration" "example" {
  name        = "example"
  description = "example configuration"
}

resource "aws_notifications_event_rule" "example" {
  event_pattern = jsonencode({
    detail = {
      state = {
        value = ["ALARM"]
      }
    }
  })
  event_type                     = "CloudWatch Alarm State Change"
  notification_configuration_arn = aws_notifications_notification_configuration.example.arn
  regions                        = ["us-east-1", "us-west-2"]
  source                         = "aws.cloudwatch"
}
```

## Argument Reference

The following arguments are required:

* `event_type` - (Required) Type of event to match. Must be between 1 and 128 characters, and match the pattern `([a-zA-Z0-9 \-\(\)])+`.
* `notification_configuration_arn` - (Required) ARN of the notification configuration to associate with this event rule. Must match the pattern `arn:aws:notifications::[0-9]{12}:configuration/[a-z0-9]{27}`.
* `regions` - (Required) Set of AWS regions where the event rule will be applied. Each region must be between 2 and 25 characters, and match the pattern `([a-z]{1,2})-([a-z]{1,15}-)+([0-9])`.
* `source` - (Required) Source of the event. Must be between 1 and 36 characters, and match the pattern `aws.([a-z0-9\-])+`.

The following arguments are optional:

* `event_pattern` - (Optional) JSON string defining the event pattern to match. Maximum length is 4096 characters.
* `timeouts` - (Optional) Configuration block specifying how long to wait for the event rule to be created, updated, or deleted.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Event Rule.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import User Notifications Event Rule using the `arn`. For example:

```terraform
import {
  to = aws_notifications_event_rule.example
  id = "arn:aws:notifications::123456789012:configuration/abc123def456ghi789jkl012mno345/rule/abc123def456ghi789jkl012mno345"
}
```

Using `terraform import`, import User Notifications Event Rule using the `arn`. For example:

```console
% terraform import aws_notifications_event_rule.example arn:aws:notifications::123456789012:configuration/abc123def456ghi789jkl012mno345/rule/abc123def456ghi789jkl012mno345
```
