---
subcategory: "ECS (Elastic Container)"
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

This resource supports the following arguments:

* `name` - (Required) Name of the account setting to set.
* `value` - (Required) State of the setting.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - ARN that identifies the account setting.
* `prinicpal_arn` - ARN that identifies the account setting.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import ECS Account Setting defaults using the `name`. For example:

```terraform
import {
  to = aws_ecs_account_setting_default.example
  id = "taskLongArnFormat"
}
```

Using `terraform import`, import ECS Account Setting defaults using the `name`. For example:

```console
% terraform import aws_ecs_account_setting_default.example taskLongArnFormat
```
