---
subcategory: "ECS"
layout: "aws"
page_title: "AWS: aws_ecs_account_setting_default"
description: |-
  Provides an ECS Default account setting.
---

# Resource: aws_ecs_account_setting_default

Provides an ECS default account setting for a specific ECS Resource name within a specific region. More information can be found on the [ECS Developer Guide](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/ecs-account-settings.html).

~> **NOTE:** The AWS API does not delete this resource. When you run `destroy`, the provider will attempt to disable the setting.

~> **NOTE:** Your AWS account may not support disabling `containerInstanceLongArnFormat`, `serviceLongArnFormat`, and `taskLongArnFormat`. If your account does not support disabling these, "destroying" this resource will not disable the setting nor cause a Terraform error. However, the AWS Provider will log an AWS error: `InvalidParameterException: You can no longer disable Long Arn settings`.

## Example Usage

```terraform
resource "aws_ecs_account_setting_default" "test" {
  name  = "taskLongArnFormat"
  value = "enabled"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the account setting to set. Valid values are `serviceLongArnFormat`, `taskLongArnFormat`, `containerInstanceLongArnFormat`, `awsvpcTrunking` and `containerInsights`.
* `value` - (Required) State of the setting. Valid values are `enabled` and `disabled`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ARN that identifies the account setting.
* `prinicpal_arn` - ARN that identifies the account setting.

## Import

ECS Account Setting defaults can be imported using the `name`, e.g.,

```
$ terraform import aws_ecs_account_setting_default.example taskLongArnFormat
```
