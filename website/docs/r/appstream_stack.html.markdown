---
subcategory: "AppStream 2.0"
layout: "aws"
page_title: "AWS: aws_appstream_stack"
description: |-
  Provides an AppStream stack
---

# Resource: aws_appstream_stack

Provides an AppStream stack.

## Example Usage

```terraform
resource "aws_appstream_stack" "example" {
  name         = "stack name"
  description  = "stack description"
  display_name = "stack display name"
  feedback_url = "http://your-domain/feedback"
  redirect_url = "http://your-domain/redirect"

  storage_connectors {
    connector_type = "HOMEFOLDERS"
  }

  user_settings {
    action     = "CLIPBOARD_COPY_FROM_LOCAL_DEVICE"
    permission = "ENABLED"
  }
  user_settings {
    action     = "CLIPBOARD_COPY_TO_LOCAL_DEVICE"
    permission = "ENABLED"
  }
  user_settings {
    action     = "DOMAIN_PASSWORD_SIGNIN"
    permission = "ENABLED"
  }
  user_settings {
    action     = "DOMAIN_SMART_CARD_SIGNIN"
    permission = "DISABLED"
  }
  user_settings {
    action     = "FILE_DOWNLOAD"
    permission = "ENABLED"
  }
  user_settings {
    action     = "FILE_UPLOAD"
    permission = "ENABLED"
  }
  user_settings {
    action     = "PRINTING_TO_LOCAL_DEVICE"
    permission = "ENABLED"
  }

  application_settings {
    enabled        = true
    settings_group = "SettingsGroup"
  }

  tags = {
    TagName = "TagValue"
  }
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Unique name for the AppStream stack.

The following arguments are optional:

* `access_endpoints` - (Optional) Set of configuration blocks defining the interface VPC endpoints. Users of the stack can connect to AppStream 2.0 only through the specified endpoints.
  See [`access_endpoints`](#access_endpoints) below.
* `application_settings` - (Optional) Settings for application settings persistence.
  See [`application_settings`](#application_settings) below.
* `description` - (Optional) Description for the AppStream stack.
* `display_name` - (Optional) Stack name to display.
* `embed_host_domains` - (Optional) Domains where AppStream 2.0 streaming sessions can be embedded in an iframe. You must approve the domains that you want to host embedded AppStream 2.0 streaming sessions.
* `feedback_url` - (Optional) URL that users are redirected to after they click the Send Feedback link. If no URL is specified, no Send Feedback link is displayed. .
* `redirect_url` - (Optional) URL that users are redirected to after their streaming session ends.
* `storage_connectors` - (Optional) Configuration block for the storage connectors to enable.
  See [`storage_connectors`](#storage_connectors) below.
* `user_settings` - (Optional) Configuration block for the actions that are enabled or disabled for users during their streaming sessions. If not provided, these settings are configured automatically by AWS. If provided, the terraform configuration should include a block for each configurable action.
  See [`user_settings`](#user_settings) below.
* `streaming_experience_settings` - (Optional) The streaming protocol you want your stack to prefer. This can be UDP or TCP. Currently, UDP is only supported in the Windows native client.
  See [`streaming_experience_settings`](#streaming_experience_settings) below.
* `tags` - (Optional) Key-value mapping of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `access_endpoints`

* `endpoint_type` - (Required) Type of the interface endpoint.
  See the [`AccessEndpoint` AWS API documentation](https://docs.aws.amazon.com/appstream2/latest/APIReference/API_AccessEndpoint.html) for valid values.
* `vpce_id` - (Optional) ID of the VPC in which the interface endpoint is used.

### `application_settings`

* `enabled` - (Required) Whether application settings should be persisted.
* `settings_group` - (Optional) Name of the settings group.
  Required when `enabled` is `true`.
  Can be up to 100 characters.

### `storage_connectors`

* `connector_type` - (Required) Type of storage connector.
  Valid values are `HOMEFOLDERS`, `GOOGLE_DRIVE`, or `ONE_DRIVE`.
* `domains` - (Optional) Names of the domains for the account.
* `resource_identifier` - (Optional) ARN of the storage connector.

### `user_settings`

* `action` - (Required) Action that is enabled or disabled.
  Valid values are `CLIPBOARD_COPY_FROM_LOCAL_DEVICE`,  `CLIPBOARD_COPY_TO_LOCAL_DEVICE`, `FILE_UPLOAD`, `FILE_DOWNLOAD`, `PRINTING_TO_LOCAL_DEVICE`, `DOMAIN_PASSWORD_SIGNIN`, or `DOMAIN_SMART_CARD_SIGNIN`.
* `permission` - (Required) Whether the action is enabled or disabled.
  Valid values are `ENABLED` or `DISABLED`.

### `streaming_experience_settings`

* `preferred_protocol` - (Optional) The preferred protocol that you want to use while streaming your application.
  Valid values are `TCP` and `UDP`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the appstream stack.
* `created_time` - Date and time, in UTC and extended RFC 3339 format, when the stack was created.
* `id` - Unique ID of the appstream stack.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_appstream_stack` using the id. For example:

```terraform
import {
  to = aws_appstream_stack.example
  id = "stackID"
}
```

Using `terraform import`, import `aws_appstream_stack` using the id. For example:

```console
% terraform import aws_appstream_stack.example stackID
```
