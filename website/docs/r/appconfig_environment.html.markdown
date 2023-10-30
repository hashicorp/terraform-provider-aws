---
subcategory: "AppConfig"
layout: "aws"
page_title: "AWS: aws_appconfig_environment"
description: |-
  Provides an AppConfig Environment resource.
---

# Resource: aws_appconfig_environment

Provides an AppConfig Environment resource for an [`aws_appconfig_application` resource](appconfig_application.html.markdown). One or more environments can be defined for an application.

## Example Usage

```terraform
resource "aws_appconfig_environment" "example" {
  name           = "example-environment-tf"
  description    = "Example AppConfig Environment"
  application_id = aws_appconfig_application.example.id

  monitor {
    alarm_arn      = aws_cloudwatch_metric_alarm.example.arn
    alarm_role_arn = aws_iam_role.example.arn
  }

  tags = {
    Type = "AppConfig Environment"
  }
}

resource "aws_appconfig_application" "example" {
  name        = "example-application-tf"
  description = "Example AppConfig Application"

  tags = {
    Type = "AppConfig Application"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `application_id` - (Required, Forces new resource) AppConfig application ID. Must be between 4 and 7 characters in length.
* `name` - (Required) Name for the environment. Must be between 1 and 64 characters in length.
* `description` - (Optional) Description of the environment. Can be at most 1024 characters.
* `monitor` - (Optional) Set of Amazon CloudWatch alarms to monitor during the deployment process. Maximum of 5. See [Monitor](#monitor) below for more details.
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### Monitor

The `monitor` block supports the following:

* `alarm_arn` - (Required) ARN of the Amazon CloudWatch alarm.
* `alarm_role_arn` - (Optional) ARN of an IAM role for AWS AppConfig to monitor `alarm_arn`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the AppConfig Environment.
* `id` - (**Deprecated**) AppConfig environment ID and application ID separated by a colon (`:`).
* `environment_id` - AppConfig environment ID.
* `state` - State of the environment. Possible values are `READY_FOR_DEPLOYMENT`, `DEPLOYING`, `ROLLING_BACK`
  or `ROLLED_BACK`.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import AppConfig Environments using the environment ID and application ID separated by a colon (`:`). For example:

```terraform
import {
  to = aws_appconfig_environment.example
  id = "71abcde:11xxxxx"
}
```

Using `terraform import`, import AppConfig Environments using the environment ID and application ID separated by a colon (`:`). For example:

```console
% terraform import aws_appconfig_environment.example 71abcde:11xxxxx
```
