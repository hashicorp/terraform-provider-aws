---
subcategory: "ECR (Elastic Container Registry)"
layout: "aws"
page_title: "AWS: aws_ecr_images"
description: |-
    Provides a list of images for a specified ECR Repository
---

# Data Source: aws_ecr_images

The ECR Images data source allows the list of images in a specified repository to be retrieved.

## Example Usage

```terraform
data "aws_ecr_images" "example" {
  repository_name = "my-repository"
}

output "image_digests" {
  value = [for img in data.aws_ecr_images.example.image_ids : img.image_digest if img.image_digest != null]
}

output "image_tags" {
  value = [for img in data.aws_ecr_images.example.image_ids : img.image_tag if img.image_tag != null]
}
```

## Argument Reference

This data source supports the following arguments:

* `repository_name` - (Required) Name of the ECR Repository.
* `registry_id` - (Optional) ID of the Registry where the repository resides.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - The name of the repository.
* `image_ids` - List of image objects containing image digest and tags. Each object has the following attributes:
  * `image_digest` - The sha256 digest of the image manifest.
  * `image_tag` - The tag associated with the image.
