---
subcategory: "Amazon Q Business"
layout: "aws"
page_title: "AWS: aws_qbusiness_plugin"
description: |-
  Provides a Q Business Plugin resource.
---

# Resource: aws_qbusiness_plugin

Provides a Q Business Plugin resource.

## Example Usage

```terraform
resource "aws_qbusiness_plugin" "example" {
  application_id = aws_qbusiness_app.test.id

  basic_auth_configuration {
    role_arn   = aws_iam_role.test.arn
    secret_arn = aws_secretsmanager_secret.test.arn
  }

  display_name = "Plugin"
  server_url   = "https://yourinstance.service-now.com"
  state        = "ENABLED"
  type         = "SERVICE_NOW"
}
```

## Argument Reference

This resource supports the following arguments:

* `application_id` - (Required) Identifier of the Amazon Q application associated with the plugin.
* `basic_auth_configuration` - (Optional) TInformation about the basic authentication credentials used to configure a plugin. Conflicts with `oauth2_client_credential_configuration`
* `display_name` - (Required) The name of the Amazon Q plugin.
* `oauth2_client_credential_configuration` - (Optional) Information about the OAuth 2.0 authentication credential/token used to configure a plugin. Conflicts with `basic_auth_configuration`
* `server_url` - (Required) Source URL used for plugin configuration.
* `state` - (Required) State of plugin. Valid value are `ENABLED` and `DISABLED`
* `type` - (Required) Type of plugin. Valid value are `SERVICE_NOW`, `SALESFORCE`, `JIRA`, `ZENDESK` and `CUSTOM`
* `custom_plugin_configuration` - (Required if `type` is `CUSTOM`) Custom plugin configuration.

`basic_auth_configuration` supports the following:

* `role_arn` - (Required) ARN of an IAM role used by Amazon Q to access the basic authentication credentials stored in a Secrets Manager secret.
* `secret_arn` - (Required) ARN of the Secrets Manager secret that stores the basic authentication credentials used for plugin configuration.

`oauth2_client_credential_configuration` supports the following:

* `role_arn` - (Required) ARN of an IAM role used by Amazon Q to access the basic authentication credentials stored in a Secrets Manager secret.
* `secret_arn` - (Required) ARN of the Secrets Manager secret that stores the basic authentication credentials used for plugin configuration.

`custom_plugin_configuration` supports the following

* `description` - (Required) A description for your custom plugin configuration.
* `api_schema_type` - (Required) The type of OpenAPI schema to use. Valid value is `OPEN_API_V3`.
* `payload` - (Optional) The JSON or YAML-formatted payload defining the OpenAPI schema for a custom plugin.
* `s3` - (Optional) S3 location for the schema

`s3` supports the following

* `bucket` - (Required) The name of the S3 bucket where the OpenAPI schema is stored.
* `key` - (Required) The key of the OpenAPI schema object in the S3 bucket.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `plugin_id` - ID of the Q Business Plugin.
* `arn` - ARN of the Q Business Plugin.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
