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

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `registry_id` - (Optional) ID of the Registry where the repository resides.
* `repository_name` - (Required) Name of the ECR Repository.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `image_ids` - List of image objects containing image digest and tags. Each object has the following attributes:
    * `image_digest` - The sha256 digest of the image manifest.
    * `image_tag` - The tag associated with the image.
