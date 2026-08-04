---
subcategory: "WorkSpaces"
layout: "aws"
page_title: "AWS: aws_workspaces_pool"
description: |-
  Manages an AWS WorkSpaces Pool.
---
# Resource: aws_workspaces_pool

Manages a WorkSpaces Pool in the AWS WorkSpaces service.

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
  pool_name    = "example-pool"
  description  = "Example WorkSpaces Pool"
  directory_id = aws_workspaces_directory.example.directory_id
  running_mode = "AUTO_STOP"

  capacity {
    desired_user_sessions = 10
  }
}
```

### With Application Settings

```terraform
resource "aws_workspaces_pool" "example" {
  bundle_id    = data.aws_workspaces_bundle.example.id
  pool_name    = "example-pool"
  description  = "Example WorkSpaces Pool with Application Settings"
  directory_id = aws_workspaces_directory.example.directory_id
  running_mode = "AUTO_STOP"

  capacity {
    desired_user_sessions = 10
  }

  application_settings = [{
    status         = "ENABLED"
    settings_group = "my-settings-group"
  }]
}
```

### With Timeout Settings

```terraform
resource "aws_workspaces_pool" "example" {
  bundle_id    = data.aws_workspaces_bundle.example.id
  pool_name    = "example-pool"
  description  = "Example WorkSpaces Pool with Timeout Settings"
  directory_id = aws_workspaces_directory.example.directory_id
  running_mode = "AUTO_STOP"

  capacity {
    desired_user_sessions = 10
  }

  timeout_settings = [{
    disconnect_timeout_in_seconds      = 900
    idle_disconnect_timeout_in_seconds = 900
    max_user_duration_in_seconds       = 14400
  }]
}
```

## Argument Reference

The following arguments are required:

* `bundle_id` - (Required) ID of the bundle for the WorkSpaces Pool.
* `capacity` - (Required) Capacity configuration for the WorkSpaces Pool. See [`capacity`](#capacity) below.
* `description` - (Required) Description of the WorkSpaces Pool.
* `directory_id` - (Required) ID of the directory for the WorkSpaces Pool.
* `pool_name` - (Required) Name of the WorkSpaces Pool. This cannot be changed after creation.
* `running_mode` - (Required) Running mode of the WorkSpaces Pool. Valid values are `AUTO_STOP` and `ALWAYS_ON`.

The following arguments are optional:

* `application_settings` - (Optional) Application settings configuration for the WorkSpaces Pool. See [`application_settings`](#application_settings) below.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `timeout_settings` - (Optional) Timeout settings configuration for the WorkSpaces Pool. See [`timeout_settings`](#timeout_settings) below.

### `capacity` Block

* `desired_user_sessions` - (Required) Desired number of user sessions for the WorkSpaces Pool.

### `application_settings`

* `settings_group` - (Optional) Name of the settings group for the application settings.
* `status` - (Required) Status of the application settings. Valid values are `ENABLED` and `DISABLED`.

### `timeout_settings`

* `disconnect_timeout_in_seconds` - (Optional) Time after disconnection when a user is logged out of their WorkSpace. Must be between 1 and 36000.
* `idle_disconnect_timeout_in_seconds` - (Optional) Time after inactivity when a user is disconnected from their WorkSpace. Must be between 1 and 36000.
* `max_user_duration_in_seconds` - (Optional) Maximum time that a user can be connected to their WorkSpace. Must be between 1 and 432000.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `capacity_status` - Capacity status of the WorkSpaces Pool. See [`capacity_status`](#capacity_status) below.
* `created_at` - Date and time the WorkSpaces Pool was created.
* `pool_arn` - ARN of the WorkSpaces Pool.
* `pool_id` - ID of the WorkSpaces Pool.
* `s3_bucket_name` - S3 bucket where application settings are stored when `application_settings` is enabled.
* `state` - Current state of the WorkSpaces Pool.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

### `capacity_status`

* `active_user_sessions` - Number of user sessions that are currently being used for WorkSpaces in the pool.
* `actual_user_sessions` - Number of user sessions currently being used for WorkSpaces in the pool.
* `available_user_sessions` - Number of user sessions available for WorkSpaces in the pool.
* `desired_user_sessions` - Number of user sessions required for WorkSpaces in the pool.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `5 minutes`)
* `update` - (Default `5 minutes`)
* `delete` - (Default `5 minutes`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_workspaces_pool.example
  identity = {
    "pool_id" = "wspool-12345678"
  }
}

resource "aws_workspaces_pool" "example" {
  # Configuration omitted for brevity
}
```

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

### Identity Schema

#### Required

* `pool_id` (String) WorkSpaces Pool identifier.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.
