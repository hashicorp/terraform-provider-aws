---
subcategory: "WorkSpaces Web"
layout: "aws"
page_title: "AWS: aws_workspacesweb_user_access_logging_settings"
description: |-
  Terraform resource for managing an AWS WorkSpaces Web User Access Logging Settings.
---

# Resource: aws_workspacesweb_user_access_logging_settings

Terraform resource for managing an AWS WorkSpaces Web User Access Logging Settings resource. Once associated with a web portal, user access logging settings control how user access events are logged to Amazon Kinesis.

## Example Usage

### Basic Usage

```terraform
resource "aws_kinesis_stream" "example" {
  name        = "amazon-workspaces-web-example-stream"
  shard_count = 1
}

resource "aws_workspacesweb_user_access_logging_settings" "example" {
  kinesis_stream_arn = aws_kinesis_stream.example.arn
}
```

### With Tags

```terraform
resource "aws_kinesis_stream" "example" {
  name        = "example-stream"
  shard_count = 1
}

resource "aws_workspacesweb_user_access_logging_settings" "example" {
  kinesis_stream_arn = aws_kinesis_stream.example.arn
  tags = {
    Name        = "example-user-access-logging-settings"
    Environment = "Production"
  }
}
```

## Argument Reference

The following arguments are required:

* `kinesis_stream_arn` - (Required) ARN of the Kinesis stream.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `associated_portal_arns` - List of web portal ARNs that this user access logging settings resource is associated with.
* `user_access_logging_settings_arn` - ARN of the user access logging settings resource.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import WorkSpaces Web User Access Logging Settings using the `user_access_logging_settings_arn`. For example:

```terraform
import {
  to = aws_workspacesweb_user_access_logging_settings.example
  id = "arn:aws:workspaces-web:us-west-2:123456789012:userAccessLoggingSettings/abcdef12345"
}
```

Using `terraform import`, import WorkSpaces Web User Access Logging Settings using the `user_access_logging_settings_arn`. For example:

```console
% terraform import aws_workspacesweb_user_access_logging_settings.example arn:aws:workspaces-web:us-west-2:123456789012:userAccessLoggingSettings/abcdef12345
```
