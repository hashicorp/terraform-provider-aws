---
subcategory: "ECR Public"
layout: "aws"
page_title: "AWS: aws_ecrpublic_images"
description: |-
  Provides a list of images for a specified AWS ECR Public Repository.
---

# Data Source: aws_ecrpublic_images

The ECR Public Images data source allows the list of images in a specified public repository to be retrieved.

## Example Usage

```terraform
data "aws_ecrpublic_images" "example" {
  repository_name = "my-public-repository"
}

output "image_digests" {
  value = [for img in data.aws_ecrpublic_images.example.images : img.digest if img.digest != null]
}

output "image_tags" {
  value = distinct(flatten([for img in data.aws_ecrpublic_images.example.images : img.tags]))
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `repository_name` - (Required) Name of the public repository.
* `registry_id` - (Optional) AWS account ID associated with the public registry that contains the repository. If not specified, the default public registry is assumed.
* `image_ids` - (Optional) One or more image ID filters. Each image ID can use either a tag or digest (or both). Each object has the following attributes:
    * `image_tag` - (Optional) Tag used for the image.
    * `image_digest` - (Optional) Digest of the image manifest.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `images` - List of images returned. Each image contains:
    * `digest` - Image digest.
    * `tags` - List of image tags.
    * `size_in_bytes` - Image size in bytes.
    * `pushed_at` - Timestamp when image was pushed.
    * `artifact_media_type` - Media type of the artifact.
    * `image_manifest_media_type` - Media type of the image manifest.
    * `registry_id` - AWS account ID associated with the public registry.
    * `repository_name` - Name of the repository.
