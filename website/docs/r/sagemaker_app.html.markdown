---
subcategory: "SageMaker"
layout: "aws"
page_title: "AWS: aws_sagemaker_app"
description: |-
  Provides a SageMaker App resource.
---

# Resource: aws_sagemaker_app

Provides a SageMaker App resource.

## Example Usage

### Basic usage

```terraform
resource "aws_sagemaker_app" "example" {
  domain_id         = aws_sagemaker_domain.example.id
  user_profile_name = aws_sagemaker_user_profile.example.user_profile_name
  app_name          = "example"
  app_type          = "JupyterServer"
}
```

## Argument Reference

This resource supports the following arguments:

* `app_name` - (Required) The name of the app.
* `app_type` - (Required) The type of app. Valid values are `JupyterServer`, `KernelGateway`, `RStudioServerPro`, `RSessionGateway` and `TensorBoard`.
* `domain_id` - (Required) The domain ID.
* `resource_spec` - (Optional) The instance type and the Amazon Resource Name (ARN) of the SageMaker image created on the instance.See [Resource Spec](#resource-spec) below.
* `space_name` - (Optional) The name of the space. At least one of `user_profile_name` or `space_name` required.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `user_profile_name` - (Optional) The user profile name. At least one of `user_profile_name` or `space_name` required.

### Resource Spec

* `instance_type` - (Optional) The instance type that the image version runs on. For valid values see [SageMaker Instance Types](https://docs.aws.amazon.com/sagemaker/latest/dg/notebooks-available-instance-types.html).
* `lifecycle_config_arn` - (Optional) The Amazon Resource Name (ARN) of the Lifecycle Configuration attached to the Resource.
* `sagemaker_image_arn` - (Optional) The ARN of the SageMaker image that the image version belongs to.
* `sagemaker_image_version_alias` - (Optional) The SageMaker Image Version Alias.
* `sagemaker_image_version_arn` - (Optional) The ARN of the image version created on the instance.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The Amazon Resource Name (ARN) of the app.
* `arn` - The Amazon Resource Name (ARN) of the app.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import SageMaker Apps using the `id`. For example:

```terraform
import {
  to = aws_sagemaker_app.example
  id = "arn:aws:sagemaker:us-west-2:012345678912:app/domain-id/user-profile-name/app-type/app-name"
}
```

Using `terraform import`, import SageMaker Apps using the `id`. For example:

```console
% terraform import aws_sagemaker_app.example arn:aws:sagemaker:us-west-2:012345678912:app/domain-id/user-profile-name/app-type/app-name
```
