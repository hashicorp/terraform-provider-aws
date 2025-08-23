---
subcategory: "WorkSpaces Web"
layout: "aws"
page_title: "AWS: aws_workspacesweb_user_settings"
description: |-
  Terraform resource for managing an AWS WorkSpaces Web User Settings.
---

# Resource: aws_workspacesweb_user_settings

Terraform resource for managing an AWS WorkSpaces Web User Settings resource. Once associated with a web portal, user settings control how users can transfer data between a streaming session and their local devices.

## Example Usage

### Basic Usage

```terraform
resource "aws_workspacesweb_user_settings" "example" {
  copy_allowed     = "Enabled"
  download_allowed = "Enabled"
  paste_allowed    = "Enabled"
  print_allowed    = "Enabled"
  upload_allowed   = "Enabled"
}
```

### With Toolbar Configuration

```terraform
resource "aws_workspacesweb_user_settings" "example" {
  copy_allowed     = "Enabled"
  download_allowed = "Enabled"
  paste_allowed    = "Enabled"
  print_allowed    = "Enabled"
  upload_allowed   = "Enabled"

  toolbar_configuration {
    toolbar_type         = "Docked"
    visual_mode          = "Dark"
    hidden_toolbar_items = ["Webcam", "Microphone"]
  }
}
```

### Complete Example

```terraform
resource "aws_kms_key" "example" {
  description             = "KMS key for WorkSpaces Web User Settings"
  deletion_window_in_days = 7
}

resource "aws_workspacesweb_user_settings" "example" {
  copy_allowed                       = "Enabled"
  download_allowed                   = "Enabled"
  paste_allowed                      = "Enabled"
  print_allowed                      = "Enabled"
  upload_allowed                     = "Enabled"
  deep_link_allowed                  = "Enabled"
  disconnect_timeout_in_minutes      = 30
  idle_disconnect_timeout_in_minutes = 15
  customer_managed_key               = aws_kms_key.example.arn

  additional_encryption_context = {
    Environment = "Production"
  }

  toolbar_configuration {
    toolbar_type           = "Docked"
    visual_mode            = "Dark"
    hidden_toolbar_items   = ["Webcam", "Microphone"]
    max_display_resolution = "size1920X1080"
  }

  cookie_synchronization_configuration {
    allowlist {
      domain = "example.com"
      path   = "/path"
    }
    blocklist {
      domain = "blocked.com"
    }
  }

  tags = {
    Name = "example-user-settings"
  }
}
```

## Argument Reference

The following arguments are required:

* `copy_allowed` - (Required) Specifies whether the user can copy text from the streaming session to the local device. Valid values are `Enabled` or `Disabled`.
* `download_allowed` - (Required) Specifies whether the user can download files from the streaming session to the local device. Valid values are `Enabled` or `Disabled`.
* `paste_allowed` - (Required) Specifies whether the user can paste text from the local device to the streaming session. Valid values are `Enabled` or `Disabled`.
* `print_allowed` - (Required) Specifies whether the user can print to the local device. Valid values are `Enabled` or `Disabled`.
* `upload_allowed` - (Required) Specifies whether the user can upload files from the local device to the streaming session. Valid values are `Enabled` or `Disabled`.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `additional_encryption_context` - (Optional) Additional encryption context for the user settings.
* `associated_portal_arns` - (Optional) List of web portal ARNs to associate with the user settings.
* `cookie_synchronization_configuration` - (Optional) Configuration that specifies which cookies should be synchronized from the end user's local browser to the remote browser. Detailed below.
* `customer_managed_key` - (Optional) ARN of the customer managed KMS key.
* `deep_link_allowed` - (Optional) Specifies whether the user can use deep links that open automatically when connecting to a session. Valid values are `Enabled` or `Disabled`.
* `disconnect_timeout_in_minutes` - (Optional) Amount of time that a streaming session remains active after users disconnect. Value must be between 1 and 600 minutes.
* `idle_disconnect_timeout_in_minutes` - (Optional) Amount of time that users can be idle before they are disconnected from their streaming session. Value must be between 0 and 60 minutes.
* `toolbar_configuration` - (Optional) Configuration of the toolbar. Detailed below.
* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### cookie_synchronization_configuration

* `allowlist` - (Required) List of cookie specifications that are allowed to be synchronized to the remote browser.
    * `domain` - (Required) Domain of the cookie.
    * `name` - (Optional) Name of the cookie.
    * `path` - (Optional) Path of the cookie.
* `blocklist` - (Optional) List of cookie specifications that are blocked from being synchronized to the remote browser.
    * `domain` - (Required) Domain of the cookie.
    * `name` - (Optional) Name of the cookie.
    * `path` - (Optional) Path of the cookie.

### toolbar_configuration

* `hidden_toolbar_items` - (Optional) List of toolbar items to be hidden.
* `max_display_resolution` - (Optional) Maximum display resolution that is allowed for the session.
* `toolbar_type` - (Optional) Type of toolbar displayed during the session.
* `visual_mode` - (Optional) Visual mode of the toolbar.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `user_settings_arn` - ARN of the user settings resource.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import WorkSpaces Web User Settings using the `user_settings_arn`. For example:

```terraform
import {
  to = aws_workspacesweb_user_settings.example
  id = "arn:aws:workspaces-web:us-west-2:123456789012:usersettings/abcdef12345"
}
```

Using `terraform import`, import WorkSpaces Web User Settings using the `user_settings_arn`. For example:

```console
% terraform import aws_workspacesweb_user_settings.example arn:aws:workspacesweb:us-west-2:123456789012:usersettings/abcdef12345
```
