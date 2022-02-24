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

```terraform
resource "aws_appconfig_configuration_profile" "example" {
  application_id = aws_appconfig_application.example.id
  description    = "Example Configuration Profile"
  name           = "example-configuration-profile-tf"
  location_uri   = "hosted"

  validator {
    content = aws_lambda_function.example.arn
    type    = "LAMBDA"
  }

  tags = {
    Type = "AppConfig Configuration Profile"
  }
}
```

## Argument Reference

The following arguments are supported:

* `application_id` - (Required, Forces new resource) The application ID. Must be between 4 and 7 characters in length.
* `location_uri` - (Required, Forces new resource) A URI to locate the configuration. You can specify the AWS AppConfig hosted configuration store, Systems Manager (SSM) document, an SSM Parameter Store parameter, or an Amazon S3 object. For the hosted configuration store, specify `hosted`. For an SSM document, specify either the document name in the format `ssm-document://<Document_name>` or the Amazon Resource Name (ARN). For a parameter, specify either the parameter name in the format `ssm-parameter://<Parameter_name>` or the ARN. For an Amazon S3 object, specify the URI in the following format: `s3://<bucket>/<objectKey>`.
* `name` - (Required) The name for the configuration profile. Must be between 1 and 64 characters in length.
* `description` - (Optional) The description of the configuration profile. Can be at most 1024 characters.
* `retrieval_role_arn` - (Optional) The ARN of an IAM role with permission to access the configuration at the specified `location_uri`. A retrieval role ARN is not required for configurations stored in the AWS AppConfig `hosted` configuration store. It is required for all other sources that store your configuration.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `validator` - (Optional) A set of methods for validating the configuration. Maximum of 2. See [Validator](#validator) below for more details.

### Validator

The `validator` block supports the following:

* `content` - (Optional, Required when `type` is `LAMBDA`) Either the JSON Schema content or the Amazon Resource Name (ARN) of an AWS Lambda function.
* `type` - (Optional) The type of validator. Valid values: `JSON_SCHEMA` and `LAMBDA`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The Amazon Resource Name (ARN) of the AppConfig Configuration Profile.
* `configuration_profile_id` - The configuration profile ID.
* `id` - The AppConfig configuration profile ID and application ID separated by a colon (`:`).
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

AppConfig Configuration Profiles can be imported by using the configuration profile ID and application ID separated by a colon (`:`), e.g.,

```
$ terraform import aws_appconfig_configuration_profile.example 71abcde:11xxxxx
```
