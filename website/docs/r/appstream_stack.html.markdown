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
resource "aws_appstream_stack" "appstream_stack" {
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

The following arguments are supported:

* `name` - (Optional) A unique name for the AppStream stack.
* `name_prefix` - (Optional) Creates a unique name beginning with the specified prefix. Conflicts with `name`.
* `description` - (Optional) Description for the AppStream stack.
* `display_name` - (Optional) The stack name to display.
* `embed_host_domains` - (Optional) The domains where AppStream 2.0 streaming sessions can be embedded in an iframe. You must approve the domains that you want to host embedded AppStream 2.0 streaming sessions.
* `redirect_url` - (Optional) The URL that users are redirected to after their streaming session ends.
* `feedback_url` - (Optional) The URL that users are redirected to after they click the Send Feedback link. If no URL is specified, no Send Feedback link is displayed. .
* `storage_connectors` - (Optional) The storage connectors to enable. (documented below)
* `user_settings` - (Optional) The actions that are enabled or disabled for users during their streaming sessions. By default, these actions are enabled. (documented below)
* `application_settings` - (Optional) settings for application settings persistence.


The `storage_connectors` object supports the following:

* `connector_type` - (Required) The type of storage connector. Valid values are: `HOMEFOLDERS`, `GOOGLE_DRIVE`, `ONE_DRIVE`
* `domains` - (Optional) The names of the domains for the account.
* `resource_identifier` - (Optional) The ARN of the storage connector.

The `user_settings` object supports the following:

* `action` - (Required) The action that is enabled or disabled. Valid values are: `CLIPBOARD_COPY_FROM_LOCAL_DEVICE`, `CLIPBOARD_COPY_TO_LOCAL_DEVICE`, `FILE_UPLOAD`,`FILE_DOWNLOAD`,`PRINTING_TO_LOCAL_DEVICE`,`DOMAIN_PASSWORD_SIGNIN`,`DOMAIN_SMART_CARD_SIGNIN`,
* `permission` - (Required) Indicates whether the action is enabled or disabled. Valid values are: `ENABLED` , `DISABLED`


## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Unique identifier (ID) of the appstream stack.
* `arn` - Amazon Resource Name (ARN) of the appstream stack.
* `created_time` - The date and time, in UTC and extended RFC 3339 format, when the stack was created.
