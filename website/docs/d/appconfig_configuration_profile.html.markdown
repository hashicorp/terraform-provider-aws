---
subcategory: "AppConfig"
layout: "aws"
page_title: "AWS: aws_appconfig_configuration_profile"
description: |-
  Terraform data source for managing an AWS AppConfig Configuration Profile.
---

# Data Source: aws_appconfig_configuration_profile

Provides access to an AppConfig Configuration Profile.

## Example Usage

### Basic Usage

```terraform
data "aws_appconfig_configuration_profile" "example" {
  application_id           = "b5d5gpj"
  configuration_profile_id = "qrbb1c1"
}
```

## Argument Reference

The following arguments are required:

* `application_id` - (Required) ID of the AppConfig application to which this configuration profile belongs.
* `configuration_profile_id` - (Required) ID of the Configuration Profile.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of the Configuration Profile.
* `description` - Description of the Configuration Profile.
* `id` - AppConfig Configuration Profile ID and Application ID separated by a colon `(:)`.
* `location_uri` - Location URI of the Configuration Profile.
* `name` - Name of the Configuration Profile.
* `retrieval_role_arn` - ARN of an IAM role with permission to access the configuration at the specified location_uri.
* `tags` - Map of tags for the resource.
* `validator` - Nested list of methods for validating the configuration.
    * `content` - Either the JSON Schema content or the ARN of an AWS Lambda function.
    * `type` - Type of validator. Valid values: JSON_SCHEMA and LAMBDA.
