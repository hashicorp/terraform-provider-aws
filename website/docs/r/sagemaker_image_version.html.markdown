---
subcategory: "SageMaker"
layout: "aws"
page_title: "AWS: aws_sagemaker_image_version"
description: |-
  Provides a SageMaker Image Version resource.
---

# Resource: aws_sagemaker_image_version

Provides a SageMaker Image Version resource.

## Example Usage

### Basic usage

```terraform
resource "aws_sagemaker_image_version" "test" {
  image_name = aws_sagemaker_image.test.id
  base_image = "012345678912.dkr.ecr.us-west-2.amazonaws.com/image:latest"
}
```

## Argument Reference

This resource supports the following arguments:

* `image_name` - (Required) The name of the image. Must be unique to your account.
* `base_image` - (Required) The registry path of the container image on which this image version is based.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The name of the Image.
* `arn` - The Amazon Resource Name (ARN) assigned by AWS to this Image Version.
* `image_arn`- The Amazon Resource Name (ARN) of the image the version is based on.
* `container_image` - The registry path of the container image that contains this image version.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import SageMaker Image Versions using the `name`. For example:

```terraform
import {
  to = aws_sagemaker_image_version.test_image
  id = "my-code-repo"
}
```

Using `terraform import`, import SageMaker Image Versions using the `name`. For example:

```console
% terraform import aws_sagemaker_image_version.test_image my-code-repo
```
