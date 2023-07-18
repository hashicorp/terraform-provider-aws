---
subcategory: "AppConfig"
layout: "aws"
page_title: "AWS: aws_appconfig_environment"
description: |-
  Terraform data source for managing an AWS AppConfig Environment.
---

# Data Source: aws_appconfig_environment

Provides access to an AppConfig Environment.

## Example Usage

### Basic Usage

```terraform
data "aws_appconfig_environment" "example" {
  application_id = "b5d5gpj"
  environment_id = "qrbb1c1"
}
```

## Argument Reference

The following arguments are required:

* `application_id` - (Required) ID of the AppConfig Application to which this Environment belongs.
* `environment_id` - (Required) ID of the AppConfig Environment.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the environment.
* `name` - Name of the environment.
* `description` - Name of the environment.
* `monitor` - Set of Amazon CloudWatch alarms to monitor during the deployment process.
    * `alarm_arn` - ARN of the Amazon CloudWatch alarm.
    * `alarm_role_arn` - ARN of an IAM role for AWS AppConfig to monitor.
* `state` - State of the environment. Possible values are `READY_FOR_DEPLOYMENT`, `DEPLOYING`, `ROLLING_BACK`
  or `ROLLED_BACK`.
* `tags` - Map of tags for the resource.
