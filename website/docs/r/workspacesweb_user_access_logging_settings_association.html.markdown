---
subcategory: "WorkSpaces Web"
layout: "aws"
page_title: "AWS: aws_workspacesweb_user_access_logging_settings_association"
description: |-
  Terraform resource for managing an AWS WorkSpaces Web User Access Logging Settings Association.
---

# Resource: aws_workspacesweb_user_access_logging_settings_association

Terraform resource for managing an AWS WorkSpaces Web User Access Logging Settings Association.

## Example Usage

### Basic Usage

```terraform
resource "aws_workspacesweb_portal" "example" {
  display_name = "example"
}

resource "aws_kinesis_stream" "example" {
  name        = "amazon-workspaces-web-example"
  shard_count = 1
}

resource "aws_workspacesweb_user_access_logging_settings" "example" {
  kinesis_stream_arn = aws_kinesis_stream.example.arn
}

resource "aws_workspacesweb_user_access_logging_settings_association" "example" {
  user_access_logging_settings_arn = aws_workspacesweb_user_access_logging_settings.example.user_access_logging_settings_arn
  portal_arn                       = aws_workspacesweb_portal.example.portal_arn
}
```

## Argument Reference

The following arguments are required:

* `user_access_logging_settings_arn` - (Required) ARN of the user access logging settings to associate with the portal. Forces replacement if changed.
* `portal_arn` - (Required) ARN of the portal to associate with the user access logging settings. Forces replacement if changed.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import WorkSpaces Web User Access Logging Settings Association using the `user_access_logging_settings_arn,portal_arn`. For example:

```terraform
import {
  to = aws_workspacesweb_user_access_logging_settings_association.example
  id = "arn:aws:workspaces-web:us-west-2:123456789012:userAccessLoggingSettings/user_access_logging_settings-id-12345678,arn:aws:workspaces-web:us-west-2:123456789012:portal/portal-id-12345678"
}
```
