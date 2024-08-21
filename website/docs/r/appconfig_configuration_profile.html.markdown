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

This resource supports the following arguments:

* `application_id` - (Required, Forces new resource) Application ID. Must be between 4 and 7 characters in length.
* `location_uri` - (Required, Forces new resource) URI to locate the configuration. You can specify the AWS AppConfig hosted configuration store, Systems Manager (SSM) document, an SSM Parameter Store parameter, or an Amazon S3 object. For the hosted configuration store, specify `hosted`. For an SSM document, specify either the document name in the format `ssm-document://<Document_name>` or the ARN. For a parameter, specify either the parameter name in the format `ssm-parameter://<Parameter_name>` or the ARN. For an Amazon S3 object, specify the URI in the following format: `s3://<bucket>/<objectKey>`.
* `name` - (Required) Name for the configuration profile. Must be between 1 and 128 characters in length.
* `description` - (Optional) Description of the configuration profile. Can be at most 1024 characters.
* `kms_key_identifier` - (Optional) The identifier for an Key Management Service key to encrypt new configuration data versions in the AppConfig hosted configuration store. This attribute is only used for hosted configuration types. The identifier can be an KMS key ID, alias, or the Amazon Resource Name (ARN) of the key ID or alias.
* `retrieval_role_arn` - (Optional) ARN of an IAM role with permission to access the configuration at the specified `location_uri`. A retrieval role ARN is not required for configurations stored in the AWS AppConfig `hosted` configuration store. It is required for all other sources that store your configuration.
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `type` - (Optional) Type of configurations contained in the profile. Valid values: `AWS.AppConfig.FeatureFlags` and `AWS.Freeform`.  Default: `AWS.Freeform`.
* `validator` - (Optional) Set of methods for validating the configuration. Maximum of 2. See [Validator](#validator) below for more details.

### Validator

The `validator` block supports the following:

* `content` - (Optional, Required when `type` is `LAMBDA`) Either the JSON Schema content or the ARN of an AWS Lambda function.
* `type` - (Optional) Type of validator. Valid values: `JSON_SCHEMA` and `LAMBDA`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the AppConfig Configuration Profile.
* `configuration_profile_id` - The configuration profile ID.
* `id` - AppConfig configuration profile ID and application ID separated by a colon (`:`).
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import AppConfig Configuration Profiles using the configuration profile ID and application ID separated by a colon (`:`). For example:

```terraform
import {
  to = aws_appconfig_configuration_profile.example
  id = "71abcde:11xxxxx"
}
```

Using `terraform import`, import AppConfig Configuration Profiles using the configuration profile ID and application ID separated by a colon (`:`). For example:

```console
% terraform import aws_appconfig_configuration_profile.example 71abcde:11xxxxx
```
