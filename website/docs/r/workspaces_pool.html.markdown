---
subcategory: "WorkSpaces"
layout: "aws"
page_title: "AWS: aws_workspaces_pool"
description: |-
  Terraform resource for managing an AWS WorkSpaces Pool.
---
# Resource: aws_workspaces_pool

Provides a WorkSpaces Pool in AWS WorkSpaces Service.

## Example Usage

### Basic Usage

```terraform
data "aws_workspaces_bundle" "example" {
  owner = "AMAZON"
  name  = "Standard with Windows 10 (Server 2022 based) (WSP)"
}

resource "aws_workspaces_directory" "example" {
  subnet_ids = [
    aws_subnet.example_c.id,
    aws_subnet.example_d.id
  ]
  workspace_type                  = "POOLS"
  workspace_directory_name        = "example-directory"
  workspace_directory_description = "Example WorkSpaces Directory for Pools"
  user_identity_type              = "CUSTOMER_MANAGED"
}

resource "aws_workspaces_pool" "example" {
  bundle_id    = data.aws_workspaces_bundle.example.id
  name         = "example-pool"
  description  = "Example WorkSpaces Pool"
  directory_id = aws_workspaces_directory.example.directory_id

  capacity {
    desired_user_sessions = 10
  }
}
```

### With Application Settings

```terraform
resource "aws_workspaces_pool" "example" {
  bundle_id    = data.aws_workspaces_bundle.example.id
  name         = "example-pool"
  description  = "Example WorkSpaces Pool with Application Settings"
  directory_id = aws_workspaces_directory.example.directory_id

  capacity {
    desired_user_sessions = 10
  }

  application_settings {
    status         = "ENABLED"
    settings_group = "my-settings-group"
  }
}
```

### With Timeout Settings

```terraform
resource "aws_workspaces_pool" "example" {
  bundle_id    = data.aws_workspaces_bundle.example.id
  name         = "example-pool"
  description  = "Example WorkSpaces Pool with Timeout Settings"
  directory_id = aws_workspaces_directory.example.directory_id

  capacity {
    desired_user_sessions = 10
  }

  timeout_settings {
    disconnect_timeout_in_seconds      = 900
    idle_disconnect_timeout_in_seconds = 900
    max_user_duration_in_seconds       = 14400
  }
}
```

## Argument Reference

The following arguments are required:

* `bundle_id` - (Required) ID of the bundle for the WorkSpaces Pool.
* `capacity` - (Required) Information about the capacity of the WorkSpaces Pool. Defined below.
* `description` - (Required) Description of the WorkSpaces Pool.
* `directory_id` - (Required) ID of the directory for the WorkSpaces Pool.
* `name` - (Required) Name of the WorkSpaces Pool. This cannot be changed after creation.

The following arguments are optional:

* `application_settings` - (Optional) Information about the application settings for the WorkSpaces Pool. Defined below.
* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `timeout_settings` - (Optional) Information about the timeout settings for the WorkSpaces Pool. Defined below.

### capacity

* `desired_user_sessions` - (Required) The desired number of user sessions for the WorkSpaces Pool.

### application_settings

* `settings_group` - (Optional) The name of the settings group for the application settings.
* `status` - (Required) The status of the application settings. Valid values are `ENABLED` and `DISABLED`.

### timeout_settings

* `disconnect_timeout_in_seconds` - (Optional) The time after disconnection when a user is logged out of their WorkSpace. Must be between 1 and 36000.
* `idle_disconnect_timeout_in_seconds` - (Optional) The time after inactivity when a user is disconnected from their WorkSpace. Must be between 1 and 36000.
* `max_user_duration_in_seconds` - (Optional) The maximum time that a user can be connected to their WorkSpace. Must be between 1 and 432000.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the WorkSpaces Pool.
* `id` - ID of the WorkSpaces Pool.
* `state` - Current state of the WorkSpaces Pool.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `5 minutes`)
* `update` - (Default `5 minutes`)
* `delete` - (Default `5 minutes`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import WorkSpaces Pool using the pool ID. For example:

```terraform
import {
  to = aws_workspaces_pool.example
  id = "wspool-12345678"
}
```

Using `terraform import`, import WorkSpaces Pool using the pool ID. For example:

```console
% terraform import aws_workspaces_pool.example wspool-12345678
```
