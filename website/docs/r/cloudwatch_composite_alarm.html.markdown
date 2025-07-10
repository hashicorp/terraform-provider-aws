---
subcategory: "CloudWatch"
layout: "aws"
page_title: "AWS: aws_cloudwatch_composite_alarm"
description: |-
  Provides a CloudWatch Composite Alarm resource.
---

# Resource: aws_cloudwatch_composite_alarm

Provides a CloudWatch Composite Alarm resource.

~> **NOTE:** An alarm (composite or metric) cannot be destroyed when there are other composite alarms depending on it. This can lead to a cyclical dependency on update, as Terraform will unsuccessfully attempt to destroy alarms before updating the rule. Consider using `depends_on`, references to alarm names, and two-stage updates.

## Example Usage

```terraform
resource "aws_cloudwatch_composite_alarm" "example" {
  alarm_description = "This is a composite alarm!"
  alarm_name        = "example-composite-alarm"

  alarm_actions = aws_sns_topic.example.arn
  ok_actions    = aws_sns_topic.example.arn

  alarm_rule = <<EOF
ALARM(${aws_cloudwatch_metric_alarm.alpha.alarm_name}) OR
ALARM(${aws_cloudwatch_metric_alarm.bravo.alarm_name})
EOF

  actions_suppressor {
    alarm            = "suppressor-alarm"
    extension_period = 10
    wait_period      = 20
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `actions_enabled` - (Optional, Forces new resource) Indicates whether actions should be executed during any changes to the alarm state of the composite alarm. Defaults to `true`.
* `actions_suppressor` - (Optional) Actions will be suppressed if the suppressor alarm is in the ALARM state.
    * `alarm` - (Required) Can be an AlarmName or an Amazon Resource Name (ARN) from an existing alarm.
    * `extension_period` - (Required) The maximum time in seconds that the composite alarm waits after suppressor alarm goes out of the `ALARM` state. After this time, the composite alarm performs its actions.
    * `wait_period` - (Required) The maximum time in seconds that the composite alarm waits for the suppressor alarm to go into the `ALARM` state. After this time, the composite alarm performs its actions.
* `alarm_actions` - (Optional) The set of actions to execute when this alarm transitions to the `ALARM` state from any other state. Each action is specified as an ARN. Up to 5 actions are allowed.
* `alarm_description` - (Optional) The description for the composite alarm.
* `alarm_name` - (Required) The name for the composite alarm. This name must be unique within the region.
* `alarm_rule` - (Required) An expression that specifies which other alarms are to be evaluated to determine this composite alarm's state. For syntax, see [Creating a Composite Alarm](https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/Create_Composite_Alarm.html). The maximum length is 10240 characters.
* `insufficient_data_actions` - (Optional) The set of actions to execute when this alarm transitions to the `INSUFFICIENT_DATA` state from any other state. Each action is specified as an ARN. Up to 5 actions are allowed.
* `ok_actions` - (Optional) The set of actions to execute when this alarm transitions to an `OK` state from any other state. Each action is specified as an ARN. Up to 5 actions are allowed.
* `tags` - (Optional) A map of tags to associate with the alarm. Up to 50 tags are allowed. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The ARN of the composite alarm.
* `id` - The ID of the composite alarm resource, which is equivalent to its `alarm_name`.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import a CloudWatch Composite Alarm using the `alarm_name`. For example:

```terraform
import {
  to = aws_cloudwatch_composite_alarm.test
  id = "my-alarm"
}
```

Using `terraform import`, import a CloudWatch Composite Alarm using the `alarm_name`. For example:

```console
% terraform import aws_cloudwatch_composite_alarm.test my-alarm
```
