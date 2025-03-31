---
subcategory: "SageMaker AI"
layout: "aws"
page_title: "AWS: aws_sagemaker_image"
description: |-
  Provides a SageMaker AI Image resource.
---

# Resource: aws_sagemaker_image

Provides a SageMaker AI Image resource.

## Example Usage

### Basic usage

```terraform
resource "aws_sagemaker_image" "example" {
  image_name = "example"
  role_arn   = aws_iam_role.test.arn
}
```

## Argument Reference

This resource supports the following arguments:

* `image_name` - (Required) The name of the image. Must be unique to your account.
* `role_arn` - (Required) The Amazon Resource Name (ARN) of an IAM role that enables Amazon SageMaker AI to perform tasks on your behalf.
* `display_name` - (Optional) The display name of the image. When the image is added to a domain (must be unique to the domain).
* `description` - (Optional) The description of the image.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The name of the Image.
* `arn` - The Amazon Resource Name (ARN) assigned by AWS to this Image.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import SageMaker AI Code Images using the `name`. For example:

```terraform
import {
  to = aws_sagemaker_image.test_image
  id = "my-code-repo"
}
```

Using `terraform import`, import SageMaker AI Code Images using the `name`. For example:

```console
% terraform import aws_sagemaker_image.test_image my-code-repo
```
