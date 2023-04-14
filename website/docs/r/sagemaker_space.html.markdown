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

The following arguments are supported:

* `space_name` - (Required) The name of the space.
* `domain_id` - (Required) The ID of the associated Domain.
* `space_settings` - (Required) A collection of space settings. See [Space Settings](#space-settings) below.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### Space Settings

* `jupyter_server_app_settings` - (Optional) The Jupyter server's app settings. See [Jupyter Server App Settings](#jupyter-server-app-settings) below.
* `kernel_gateway_app_settings` - (Optional) The kernel gateway app settings. See [Kernel Gateway App Settings](#kernel-gateway-app-settings) below.

#### Kernel Gateway App Settings

* `default_resource_spec` - (Optional) The default instance type and the Amazon Resource Name (ARN) of the SageMaker image created on the instance. see [Default Resource Spec](#default-resource-spec) below.
* `custom_image` - (Optional) A list of custom SageMaker images that are configured to run as a KernelGateway app. see [Custom Image](#custom-image) below.
* `lifecycle_config_arns` - (Optional) The Amazon Resource Name (ARN) of the Lifecycle Configurations.

#### Jupyter Server App Settings

* `code_repository` - (Optional) A list of Git repositories that SageMaker automatically displays to users for cloning in the JupyterServer application. see [Code Repository](#code-repository) below.
* `default_resource_spec` - (Optional) The default instance type and the Amazon Resource Name (ARN) of the SageMaker image created on the instance. see [Default Resource Spec](#default-resource-spec) below.
* `lifecycle_config_arns` - (Optional) The Amazon Resource Name (ARN) of the Lifecycle Configurations.

##### Code Repository

* `repository_url` - (Optional) The URL of the Git repository.

##### Default Resource Spec

* `instance_type` - (Optional) The instance type.
* `lifecycle_config_arn` - (Optional) The Amazon Resource Name (ARN) of the Lifecycle Configuration attached to the Resource.
* `sagemaker_image_arn` - (Optional) The Amazon Resource Name (ARN) of the SageMaker image created on the instance.
* `sagemaker_image_version_arn` - (Optional) The ARN of the image version created on the instance.

##### Custom Image

* `app_image_config_name` - (Required) The name of the App Image Config.
* `image_name` - (Required) The name of the Custom Image.
* `image_version_number` - (Optional) The version number of the Custom Image.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The space's Amazon Resource Name (ARN).
* `arn` - The space's Amazon Resource Name (ARN).
* `home_efs_file_system_uid` - The ID of the space's profile in the Amazon Elastic File System volume.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

SageMaker Spaces can be imported using the `id`, e.g.,

```
$ terraform import aws_sagemaker_space.test_space arn:aws:sagemaker:us-west-2:123456789012:space/domain-id/space-name
```
