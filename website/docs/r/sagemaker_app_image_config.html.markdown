---
subcategory: "SageMaker"
layout: "aws"
page_title: "AWS: aws_sagemaker_app_image_config"
description: |-
  Provides a SageMaker App Image Config resource.
---

# Resource: aws_sagemaker_app_image_config

Provides a SageMaker App Image Config resource.

## Example Usage

### Basic usage

```terraform
resource "aws_sagemaker_app_image_config" "test" {
  app_image_config_name = "example"

  kernel_gateway_image_config {
    kernel_spec {
      name = "example"
    }
  }
}
```

### Default File System Config

```terraform
resource "aws_sagemaker_app_image_config" "test" {
  app_image_config_name = "example"

  kernel_gateway_image_config {
    kernel_spec {
      name = "example"
    }

    file_system_config {}
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `app_image_config_name` - (Required) The name of the App Image Config.
* `code_editor_app_image_config` - (Optional) The CodeEditorAppImageConfig. You can only specify one image kernel in the AppImageConfig API. This kernel is shown to users before the image starts. After the image runs, all kernels are visible in Code Editor. See [Code Editor App Image Config](#code-editor-app-image-config) details below.
* `jupyter_lab_image_config` - (Optional) The JupyterLabAppImageConfig. You can only specify one image kernel in the AppImageConfig API. This kernel is shown to users before the image starts. After the image runs, all kernels are visible in JupyterLab. See [Jupyter Lab Image Config](#jupyter-lab-image-config) details below.
* `kernel_gateway_image_config` - (Optional) The configuration for the file system and kernels in a SageMaker image running as a KernelGateway app. See [Kernel Gateway Image Config](#kernel-gateway-image-config) details below.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### Code Editor App Image Config

* `container_config` - (Optional) The configuration used to run the application image container. See [Container Config](#container-config) details below.
* `file_system_config` - (Optional) The URL where the Git repository is located. See [File System Config](#file-system-config) details below.

### Jupyter Lab Image Config

* `container_config` - (Optional) The configuration used to run the application image container. See [Container Config](#container-config) details below.
* `file_system_config` - (Optional) The URL where the Git repository is located. See [File System Config](#file-system-config) details below.

#### Container Config

* `container_arguments` - (Optional) The arguments for the container when you're running the application.
* `container_entrypoint` - (Optional) The entrypoint used to run the application in the container.
* `container_environment_variables` - (Optional) The environment variables to set in the container.

### Kernel Gateway Image Config

* `file_system_config` - (Optional) The URL where the Git repository is located. See [File System Config](#file-system-config) details below.
* `kernel_spec` - (Required) The default branch for the Git repository. See [Kernel Spec](#kernel-spec) details below.

#### File System Config

* `default_gid` - (Optional) The default POSIX group ID (GID). If not specified, defaults to `100`. Valid values are `0` and `100`.
* `default_uid` - (Optional) The default POSIX user ID (UID). If not specified, defaults to `1000`. Valid values are `0` and `1000`.
* `mount_path` - (Optional) The path within the image to mount the user's EFS home directory. The directory should be empty. If not specified, defaults to `/home/sagemaker-user`.

~> **Note:** When specifying `default_gid` and `default_uid`, Valid value pairs are [`0`, `0`] and [`100`, `1000`].

#### Kernel Spec

* `name` - (Required) The name of the kernel.
* `display_name` - (Optional) The display name of the kernel.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The name of the App Image Config.
* `arn` - The Amazon Resource Name (ARN) assigned by AWS to this App Image Config.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import SageMaker App Image Configs using the `name`. For example:

```terraform
import {
  to = aws_sagemaker_app_image_config.example
  id = "example"
}
```

Using `terraform import`, import SageMaker App Image Configs using the `name`. For example:

```console
% terraform import aws_sagemaker_app_image_config.example example
```
