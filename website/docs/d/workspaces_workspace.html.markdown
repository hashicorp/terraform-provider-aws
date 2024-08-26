---
subcategory: "WorkSpaces"
layout: "aws"
page_title: "AWS: aws_workspaces_workspace"
description: |-
  Get information about a WorkSpace in AWS Workspaces Service.
---

# Resource: aws_workspaces_workspace

Use this data source to get information about a workspace in [AWS Workspaces](https://docs.aws.amazon.com/workspaces/latest/adminguide/amazon-workspaces.html) Service.

## Example Usage

### Filter By Workspace ID

```terraform
data "aws_workspaces_workspace" "example" {
  workspace_id = "ws-cj5xcxsz5"
}
```

### Filter By Directory ID & User Name

```terraform
data "aws_workspaces_workspace" "example" {
  directory_id = "d-9967252f57"
  user_name    = "Example"
}
```

## Argument Reference

This data source supports the following arguments:

* `bundle_id` - (Optional) ID of the bundle for the WorkSpace.
* `directory_id` - (Optional) ID of the directory for the WorkSpace. You have to specify `user_name` along with `directory_id`. You cannot combine this parameter with `workspace_id`.
* `root_volume_encryption_enabled` - (Optional) Indicates whether the data stored on the root volume is encrypted.
* `tags` - (Optional) Tags for the WorkSpace.
* `user_name` – (Optional) User name of the user for the WorkSpace. This user name must exist in the directory for the WorkSpace. You cannot combine this parameter with `workspace_id`.
* `user_volume_encryption_enabled` – (Optional) Indicates whether the data stored on the user volume is encrypted.
* `volume_encryption_key` – (Optional) Symmetric AWS KMS customer master key (CMK) used to encrypt data stored on your WorkSpace. Amazon WorkSpaces does not support asymmetric CMKs.
* `workspace_id` - (Optional) ID of the WorkSpace. You cannot combine this parameter with `directory_id`.
* `workspace_properties` – (Optional) WorkSpace properties.

`workspace_properties` supports the following:

* `compute_type_name` – (Optional) Compute type. For more information, see [Amazon WorkSpaces Bundles](http://aws.amazon.com/workspaces/details/#Amazon_WorkSpaces_Bundles). Valid values are `VALUE`, `STANDARD`, `PERFORMANCE`, `POWER`, `GRAPHICS`, `POWERPRO` and `GRAPHICSPRO`.
* `root_volume_size_gib` – (Optional) Size of the root volume.
* `running_mode` – (Optional) Running mode. For more information, see [Manage the WorkSpace Running Mode](https://docs.aws.amazon.com/workspaces/latest/adminguide/running-mode.html). Valid values are `AUTO_STOP` and `ALWAYS_ON`.
* `running_mode_auto_stop_timeout_in_minutes` – (Optional) Time after a user logs off when WorkSpaces are automatically stopped. Configured in 60-minute intervals.
* `user_volume_size_gib` – (Optional) Size of the user storage.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - Workspaces ID.
* `ip_address` - IP address of the WorkSpace.
* `computer_name` - Name of the WorkSpace, as seen by the operating system.
* `state` - Operational state of the WorkSpace.
