---
subcategory: "EventBridge"
layout: "aws"
page_title: "AWS: aws_cloudwatch_event_rule"
description: |-
  Get information on an EventBridge (Cloudwatch) rule.
---

# Data Source: aws_cloudwatch_event_rule

Use this data source to get information about an EventBridge rule. 

~> **Note:** EventBridge was formerly known as CloudWatch Events. The functionality is identical.

## Example Usage

```terraform
data "aws_cloudwatch_event_rule" "examplerule" {
  name = "examplerule"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the rule
* `event_bus_name` - (Optional) The name or ARN of the event bus associated with the rule. If you omit this, the default event bus is used

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The Amazon Resource Name (ARN) of the rule
* `event_pattern` - The event pattern
* `schedule_expression` - The scheduling expression
* `state` - Specifies whether the rule is enabled or disabled
* `description` - The description of the rule
* `created_by` - The account ID of the user that created the rule. If you use PutRule to put a rule on an event bus in another account, the other account is the owner of the rule, and the rule ARN includes the account ID for that account. However, the value for CreatedBy is the account ID as the account that created the rule in the other account
* `managed_by` - If this is a managed rule, created by an Amazon Web Services service on your behalf, this field displays the principal name of the Amazon Web Services service that created the rule
* `role_arn` - The Amazon Resource Name (ARN) of the IAM role associated with the rule
