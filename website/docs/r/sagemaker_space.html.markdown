---
subcategory: "SageMaker"
layout: "aws"
page_title: "AWS: aws_sagemaker_space"
description: |-
  Provides a SageMaker Space resource.
---

# Resource: aws_sagemaker_space

Provides a SageMaker Space resource.

## Example Usage

### Basic usage

```terraform
resource "aws_sagemaker_space" "example" {
  domain_id  = aws_sagemaker_domain.test.id
  space_name = "example"
}
```

## Argument Reference

This resource supports the following arguments:

* `domain_id` - (Required) The ID of the associated Domain.
* `ownership_settings` - (Optional) A collection of ownership settings. Required if `space_sharing_settings` is set. See [`ownership_settings` Block](#ownership_settings-block) below.
* `space_display_name` - (Optional) The name of the space that appears in the SageMaker Studio UI.
* `space_name` - (Required) The name of the space.
* `space_settings` - (Required) A collection of space settings. See [`space_settings` Block](#space_settings-block) below.
* `space_sharing_settings` - (Optional) A collection of space sharing settings. Required if `ownership_settings` is set. See [`space_sharing_settings` Block](#space_sharing_settings-block) below.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `ownership_settings` Block

The `ownership_settings` block supports the following arguments:

* `owner_user_profile_name` - (Required) The user profile who is the owner of the private space.

### `space_settings` Block

The `space_settings` block supports the following arguments:

* `app_type` - (Optional) The type of app created within the space.
* `code_editor_app_settings` - (Optional) The Code Editor application settings. See [`code_editor_app_settings` Block](#code_editor_app_settings-block) below.
* `custom_file_system` - (Optional) A file system, created by you, that you assign to a space for an Amazon SageMaker Domain. See [`custom_file_system` Block](#custom_file_system-block) below.
* `jupyter_lab_app_settings` - (Optional) The settings for the JupyterLab application. See [`jupyter_lab_app_settings` Block](#jupyter_lab_app_settings-block) below.
* `jupyter_server_app_settings` - (Optional) The Jupyter server's app settings. See [`jupyter_server_app_settings` Block](#jupyter_server_app_settings-block) below.
* `kernel_gateway_app_settings` - (Optional) The kernel gateway app settings. See [`kernel_gateway_app_settings` Block](#kernel_gateway_app_settings-block) below.
* `space_storage_settings` - (Optional) The storage settings. See [`space_storage_settings` Block](#space_storage_settings-block) below.

### `space_sharing_settings` Block

The `space_sharing_settings` block supports the following argument:

* `sharing_type` - (Required) Specifies the sharing type of the space. Valid values are `Private` and `Shared`.

### `code_editor_app_settings` Block

The `code_editor_app_settings` block supports the following argument:

* `default_resource_spec` - (Required) The default instance type and the Amazon Resource Name (ARN) of the SageMaker image created on the instance. See [`default_resource_spec` Block](#default_resource_spec-block) below.

### `custom_file_system` Block

The `custom_file_system` block supports the following argument:

* `efs_file_system` - (Required) A custom file system in Amazon EFS. See [`efs_file_system` Block](#efs_file_system-block) below.

### `jupyter_lab_app_settings` Block

The `jupyter_lab_app_settings` block supports the following arguments:

* `code_repository` - (Optional) A list of Git repositories that SageMaker automatically displays to users for cloning in the JupyterServer application. See [`code_repository` Block](#code_repository-block) below.
* `default_resource_spec` - (Required) The default instance type and the Amazon Resource Name (ARN) of the SageMaker image created on the instance. See [`default_resource_spec` Block](#default_resource_spec-block) below.

### `jupyter_server_app_settings` Block

The `jupyter_server_app_settings` block supports the following arguments:

* `code_repository` - (Optional) A list of Git repositories that SageMaker automatically displays to users for cloning in the JupyterServer application. See [`code_repository` Block](#code_repository-block) below.
* `default_resource_spec` - (Optional) The default instance type and the Amazon Resource Name (ARN) of the SageMaker image created on the instance. See [`default_resource_spec` Block](#default_resource_spec-block) below.
* `lifecycle_config_arns` - (Optional) The Amazon Resource Name (ARN) of the Lifecycle Configurations.

### `kernel_gateway_app_settings` Block

The `kernel_gateway_app_settings` block supports the following arguments:

* `default_resource_spec` - (Required) The default instance type and the Amazon Resource Name (ARN) of the SageMaker image created on the instance. See [`default_resource_spec` Block](#default_resource_spec-block) below.
* `custom_image` - (Optional) A list of custom SageMaker images that are configured to run as a KernelGateway app. See [`custom_image` Block](#custom_image-block) below.
* `lifecycle_config_arns` - (Optional) The Amazon Resource Name (ARN) of the Lifecycle Configurations.

### `space_storage_settings` Block

The `space_storage_settings` block supports the following argument:

* `ebs_storage_settings` - (Required) A collection of EBS storage settings for a space. See [`ebs_storage_settings` Block](#ebs_storage_settings-block) below.

### `code_repository` Block

The `code_repository` block supports the following argument:

* `repository_url` - (Required) The URL of the Git repository.

### `default_resource_spec` Block

The `default_resource_spec` block supports the following arguments:

* `instance_type` - (Optional) The instance type.
* `lifecycle_config_arn` - (Optional) The Amazon Resource Name (ARN) of the Lifecycle Configuration attached to the Resource.
* `sagemaker_image_arn` - (Optional) The Amazon Resource Name (ARN) of the SageMaker image created on the instance.
* `sagemaker_image_version_alias` - (Optional) The SageMaker Image Version Alias.
* `sagemaker_image_version_arn` - (Optional) The ARN of the image version created on the instance.

### `efs_file_system` Block

The `efs_file_system` block supports the following argument:

* `file_system_id` - (Required) The ID of your Amazon EFS file system.

### `custom_image` Block

The `custom_image` block supports the following arguments:

* `app_image_config_name` - (Required) The name of the App Image Config.
* `image_name` - (Required) The name of the Custom Image.
* `image_version_number` - (Optional) The version number of the Custom Image.

### `ebs_storage_settings` Block

The `ebs_storage_settings` block supports the following argument:

* `ebs_volume_size_in_gb` - (Required) The size of an EBS storage volume for a space.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The space's Amazon Resource Name (ARN).
* `home_efs_file_system_uid` - The ID of the space's profile in the Amazon Elastic File System volume.
* `id` - The space's Amazon Resource Name (ARN).
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `url` - Returns the URL of the space. If the space is created with Amazon Web Services IAM Identity Center (Successor to Amazon Web Services Single Sign-On) authentication, users can navigate to the URL after appending the respective redirect parameter for the application type to be federated through Amazon Web Services IAM Identity Center.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import SageMaker Spaces using the `id`. For example:

```terraform
import {
  to = aws_sagemaker_space.test_space
  id = "arn:aws:sagemaker:us-west-2:123456789012:space/domain-id/space-name"
}
```

Using `terraform import`, import SageMaker Spaces using the `id`. For example:

```console
% terraform import aws_sagemaker_space.test_space arn:aws:sagemaker:us-west-2:123456789012:space/domain-id/space-name
```
