---
subcategory: "AppConfig"
layout: "aws"
page_title: "AWS: aws_appconfig_environment"
description: |-
  Provides an AppConfig Environment resource.
---

# Resource: aws_appconfig_environment

Provides an AppConfig Environment resource.

## Example Usage

### AppConfig Environment

```hcl
resource "aws_appconfig_environment" "production" {
  name           = "production"
  description    = "Production"
  application_id = aws_appconfig_application.test.id
  tags = {
    Type = "Production"
  }
  monitors {
    alarm_arn      = "arn:aws:cloudwatch:us-east-1:111111111111:alarm:test-appconfig"
    alarm_role_arn = "arn:aws:iam::111111111111:role/service-role/test-appconfig"
  }
}

resource "aws_appconfig_application" "test" {
  name        = "test"
  description = "Test"
  tags = {
    Type = "Test"
  }
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) The environment name. Must be between 1 and 64 characters in length.
- `application_id` - (Required) The application id. Must be between 4 and 7 characters in length.
- `description` - (Optional) The description of the environment. Can be at most 1024 characters.
- `monitors` - (Optional) Amazon CloudWatch alarms to monitor during the deployment process. Detailed below.
- `tags` - (Optional) A map of tags to assign to the resource.

### monitor

- `alarm_arn` - (Optional) ARN of the Amazon CloudWatch alarm. Must be between 20 and 2048 characters in length.
- `alarm_role_arn` - (Optional) ARN of an IAM role for AWS AppConfig to monitor AlarmArn. Must be between 20 and 2048 characters in length.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `arn` - The Amazon Resource Name (ARN) of the AppConfig Environment.
- `id` - The AppConfig Environment ID

## Import

`aws_appconfig_environment` can be imported by the Application ID and Environment ID, e.g.

```
$ terraform import aws_appconfig_environment.production 71abcde/11xxxxx
```
