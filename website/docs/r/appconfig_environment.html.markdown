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

The following arguments are supported:

* `application_id` - (Required, Forces new resource) The AppConfig application ID. Must be between 4 and 7 characters in length.
* `name` - (Required) The name for the environment. Must be between 1 and 64 characters in length.
* `description` - (Optional) The description of the environment. Can be at most 1024 characters.
* `monitor` - (Optional) Set of Amazon CloudWatch alarms to monitor during the deployment process. Maximum of 5. See [Monitor](#monitor) below for more details.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### Monitor

The `monitor` block supports the following:

* `alarm_arn` - (Required) ARN of the Amazon CloudWatch alarm.
* `alarm_role_arn` - (Optional) ARN of an IAM role for AWS AppConfig to monitor `alarm_arn`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The Amazon Resource Name (ARN) of the AppConfig Environment.
* `id` - The AppConfig environment ID and application ID separated by a colon (`:`).
* `environment_id` - The AppConfig environment ID.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

AppConfig Environments can be imported by using the environment ID and application ID separated by a colon (`:`), e.g.,

```
$ terraform import aws_appconfig_environment.example 71abcde:11xxxxx
```
