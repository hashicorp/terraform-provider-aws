---
subcategory: "AppStream"
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
    action     = "FILE_UPLOAD"
    permission = "ENABLED"
  }
  user_settings {
    action     = "FILE_DOWNLOAD"
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

* `application_settings` - (Optional) Settings for application settings persistence.
* `description` - (Optional) Description for the AppStream stack.
* `display_name` - (Optional) Stack name to display.
* `embed_host_domains` - (Optional) Domains where AppStream 2.0 streaming sessions can be embedded in an iframe. You must approve the domains that you want to host embedded AppStream 2.0 streaming sessions.
* `feedback_url` - (Optional) URL that users are redirected to after they click the Send Feedback link. If no URL is specified, no Send Feedback link is displayed. .
* `redirect_url` - (Optional) URL that users are redirected to after their streaming session ends.
* `storage_connectors` - (Optional) Configuration block for the storage connectors to enable. See below.
* `user_settings` - (Optional) Configuration block for the actions that are enabled or disabled for users during their streaming sessions. By default, these actions are enabled. See below.

### `storage_connectors`

* `connector_type` - (Required) Type of storage connector. Valid values are: `HOMEFOLDERS`, `GOOGLE_DRIVE`, `ONE_DRIVE`.
* `domains` - (Optional) Names of the domains for the account.
* `resource_identifier` - (Optional) ARN of the storage connector.

### `user_settings`

* `action` - (Required) Action that is enabled or disabled. Valid values are: `CLIPBOARD_COPY_FROM_LOCAL_DEVICE`,  `CLIPBOARD_COPY_TO_LOCAL_DEVICE`, `FILE_UPLOAD`, `FILE_DOWNLOAD`, `PRINTING_TO_LOCAL_DEVICE`, `DOMAIN_PASSWORD_SIGNIN`, `DOMAIN_SMART_CARD_SIGNIN`.
* `permission` - (Required) Indicates whether the action is enabled or disabled. Valid values are: `ENABLED`, `DISABLED`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of the appstream stack.
* `created_time` - Date and time, in UTC and extended RFC 3339 format, when the stack was created.
* `id` - Unique ID of the appstream stack.


## Import

`aws_appstream_stack` can be imported using the id, e.g.,

```
$ terraform import aws_appstream_stack.example stackID
```
