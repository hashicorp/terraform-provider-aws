---
subcategory: "WorkSpaces"
layout: "aws"
page_title: "AWS: aws_workspaces_workspace"
description: |-
  Provides a workspaces in AWS Workspaces Service.
---

# Resource: aws_workspaces_workspace

Provides a workspace in [AWS WorkSpaces](https://docs.aws.amazon.com/workspaces/latest/adminguide/amazon-workspaces.html) Service

~> **NOTE:** AWS WorkSpaces service requires [`workspaces_DefaultRole`](https://docs.aws.amazon.com/workspaces/latest/adminguide/workspaces-access-control.html#create-default-role) IAM role to operate normally.

## Example Usage

```terraform
data "aws_workspaces_bundle" "value_windows_10" {
  bundle_id = "wsb-bh8rsxt14" # Value with Windows 10 (English)
}

data "aws_kms_key" "workspaces" {
  key_id = "alias/aws/workspaces"
}

resource "aws_workspaces_workspace" "example" {
  directory_id = aws_workspaces_directory.example.id
  bundle_id    = data.aws_workspaces_bundle.value_windows_10.id
  user_name    = "john.doe"

  root_volume_encryption_enabled = true
  user_volume_encryption_enabled = true
  volume_encryption_key          = data.aws_kms_key.workspaces.arn

  workspace_properties {
    compute_type_name                         = "VALUE"
    user_volume_size_gib                      = 10
    root_volume_size_gib                      = 80
    running_mode                              = "AUTO_STOP"
    running_mode_auto_stop_timeout_in_minutes = 60
  }

  tags = {
    Department = "IT"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `account` - (Optional) AWS account ID where this resource will be managed. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `bundle_id` - (Required) ID of the bundle for the WorkSpace.
* `directory_id` - (Required) ID of the directory for the WorkSpace.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `root_volume_encryption_enabled` - (Optional) Whether the data stored on the root volume is encrypted.
* `tags` - (Optional) Tags for the WorkSpace. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `user_name` - (Required) User name of the user for the WorkSpace. This user name must exist in the directory for the WorkSpace.
* `user_volume_encryption_enabled` - (Optional) Whether the data stored on the user volume is encrypted.
* `volume_encryption_key` - (Optional) ARN of a symmetric AWS KMS customer master key (CMK) used to encrypt data stored on your WorkSpace. Amazon WorkSpaces does not support asymmetric CMKs.
* `workspace_properties` - (Optional) WorkSpace properties. Defined below.

### `workspace_properties` Block

* `compute_type_name` - (Optional) Compute type. For more information, see [Amazon WorkSpaces Bundles](http://aws.amazon.com/workspaces/details/#Amazon_WorkSpaces_Bundles). Valid values are `VALUE`, `STANDARD`, `PERFORMANCE`, `POWER`, `GRAPHICS`, `POWERPRO`, `GENERALPURPOSE_4XLARGE`, `GENERALPURPOSE_8XLARGE`, `GRAPHICSPRO`, `GRAPHICS_G4DN`, `GRAPHICSPRO_G4DN`, `GRAPHICS_G6_XLARGE`, `GRAPHICS_G6_2XLARGE`, `GRAPHICS_G6_4XLARGE`, `GRAPHICS_G6_8XLARGE`, `GRAPHICS_G6_16XLARGE`, `GRAPHICS_GR6_4XLARGE`, `GRAPHICS_GR6_8XLARGE`, `GRAPHICS_G6F_LARGE`, `GRAPHICS_G6F_XLARGE`, `GRAPHICS_G6F_2XLARGE`, `GRAPHICS_G6F_4XLARGE`, and `GRAPHICS_GR6F_4XLARGE`.
* `root_volume_size_gib` - (Optional) Size of the root volume.
* `running_mode` - (Optional) Running mode. For more information, see [Manage the WorkSpace Running Mode](https://docs.aws.amazon.com/workspaces/latest/adminguide/running-mode.html). Valid values are `AUTO_STOP` and `ALWAYS_ON`.
* `running_mode_auto_stop_timeout_in_minutes` - (Optional) Time after a user logs off when WorkSpaces are automatically stopped. Configured in 60-minute intervals.
* `user_volume_size_gib` - (Optional) Size of the user storage.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `computer_name` - Name of the WorkSpace, as seen by the operating system.
* `id` - WorkSpaces ID.
* `ip_address` - IP address of the WorkSpace.
* `state` - Operational state of the WorkSpace.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `30m`)
- `update` - (Default `10m`)
- `delete` - (Default `10m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import WorkSpaces using their ID. For example:

```terraform
import {
  to = aws_workspaces_workspace.example
  id = "ws-9z9zmbkhv"
}
```

Using `terraform import`, import WorkSpaces using their ID. For example:

```console
% terraform import aws_workspaces_workspace.example ws-9z9zmbkhv
```
