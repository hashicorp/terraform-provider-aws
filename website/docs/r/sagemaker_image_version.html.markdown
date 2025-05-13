---
subcategory: "SageMaker AI"
layout: "aws"
page_title: "AWS: aws_sagemaker_image_version"
description: |-
  Provides a SageMaker AI Image Version resource.
---

# Resource: aws_sagemaker_image_version

Provides a SageMaker AI Image Version resource.

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
* `horovod` - (Optional) Indicates Horovod compatibility.
* `job_type` - (Optional) Indicates SageMaker AI job type compatibility. Valid values are: `TRAINING`, `INFERENCE`, and `NOTEBOOK_KERNEL`.
* `ml_framework` - (Optional) The machine learning framework vended in the image version.
* `processor` - (Optional) Indicates CPU or GPU compatibility. Valid values are: `CPU` and `GPU`.
* `programming_lang` - (Optional) The supported programming language and its version.
* `release_notes` - (Optional) The maintainer description of the image version.
* `vendor_guidance` - (Optional) The stability of the image version, specified by the maintainer. Valid values are: `NOT_PROVIDED`, `STABLE`, `TO_BE_ARCHIVED`, and `ARCHIVED`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The image name and version in the format `name:version`.
* `arn` - The Amazon Resource Name (ARN) assigned by AWS to this Image Version.
* `version`- The version of the image. If not specified, the latest version is described.
* `container_image` - The registry path of the container image that contains this image version.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import SageMaker AI Image Versions using the `name:version` format. For example:

```terraform
import {
  to = aws_sagemaker_image_version.test_image
  id = "my-image:1"
}
```

Using `terraform import`, import SageMaker AI Image Versions using the `name:version` format. For example:

```console
% terraform import aws_sagemaker_image_version.test_image my-image:1
```

For backward compatibility, importing using just the image name is still supported, but the resource ID will be automatically updated to the `name:version` format after import.
