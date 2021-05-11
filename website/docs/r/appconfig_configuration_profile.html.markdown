---
subcategory: "AppConfig"
layout: "aws"
page_title: "AWS: aws_appconfig_configuration_profile"
description: |-
  Provides an AppConfig Configuration Profile resource.
---

# Resource: aws_appconfig_configuration_profile

Provides an AppConfig Configuration Profile resource.

## Example Usage

### AppConfig Configuration Profile

```hcl
resource "aws_appconfig_configuration_profile" "production" {
  name           = "test"
  description    = "test"
  application_id = aws_appconfig_application.test.id
  validators {
    content = "arn:aws:lambda:us-east-1:111111111111:function:test"
    type    = "LAMBDA"
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
- `location_uri` - (Optional) A URI to locate the configuration.
- `validators` - (Optional) A list of methods for validating the configuration. Detailed below.
- `retrieval_role_arn` - (Optional) The ARN of an IAM role with permission to access the configuration at the specified LocationUri.
- `tags` - (Optional) A map of tags to assign to the resource.

### validator

- `content` - (Optional) Either the JSON Schema content or the Amazon Resource Name (ARN) of an AWS Lambda function.
- `type` - (Optional) AWS AppConfig supports validators of type `JSON_SCHEMA` and `LAMBDA`

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `arn` - The Amazon Resource Name (ARN) of the AppConfig Configuration Profile.
- `id` - The AppConfig Configuration Profile ID

## Import

`aws_appconfig_configuration_profile` can be imported by the Application ID and Configuration Profile ID, e.g.

```
$ terraform import aws_appconfig_configuration_profile.test 71abcde/11xxxxx
```
