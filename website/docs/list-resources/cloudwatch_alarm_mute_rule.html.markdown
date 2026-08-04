---
subcategory: "CloudWatch"
layout: "aws"
page_title: "AWS: aws_cloudwatch_alarm_mute_rule"
description: |-
  Lists CloudWatch Alarm Mute Rule resources.
---

# List Resource: aws_cloudwatch_alarm_mute_rule

Lists CloudWatch Alarm Mute Rule resources.

## Example Usage

```terraform
list "aws_cloudwatch_alarm_mute_rule" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
