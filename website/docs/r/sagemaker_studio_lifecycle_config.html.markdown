---
subcategory: "SageMaker"
layout: "aws"
page_title: "AWS: aws_sagemaker_studio_lifecycle_config"
description: |-
  Provides a SageMaker Studio Lifecycle Config resource.
---

# Resource: aws_sagemaker_studio_lifecycle_config

Provides a SageMaker Studio Lifecycle Config resource.

## Example Usage

### Basic usage

```terraform
resource "aws_sagemaker_studio_lifecycle_config" "example" {
  studio_lifecycle_config_name     = "example"
  studio_lifecycle_config_app_type = "JupyterServer"
  studio_lifecycle_config_content  = base64encode("echo Hello")
}
```

## Argument Reference

This resource supports the following arguments:

- `studio_lifecycle_config_name` - (Required) The name of the Studio Lifecycle Configuration to create.
- `studio_lifecycle_config_app_type` - (Required) The App type that the Lifecycle Configuration is attached to. Valid values are `JupyterServer`, `JupyterLab`, `CodeEditor` and `KernelGateway`.
- `studio_lifecycle_config_content` - (Required) The content of your Studio Lifecycle Configuration script. This content must be base64 encoded.
- `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

- `id` - The name of the Studio Lifecycle Config.
- `arn` - The Amazon Resource Name (ARN) assigned by AWS to this Studio Lifecycle Config.
- `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import SageMaker Studio Lifecycle Configs using the `studio_lifecycle_config_name`. For example:

```terraform
import {
  to = aws_sagemaker_studio_lifecycle_config.example
  id = "example"
}
```

Using `terraform import`, import SageMaker Studio Lifecycle Configs using the `studio_lifecycle_config_name`. For example:

```console
% terraform import aws_sagemaker_studio_lifecycle_config.example example
```
