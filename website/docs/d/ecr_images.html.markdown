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

### Filter by Tag Status

```terraform
data "aws_ecr_images" "tagged_only" {
  repository_name = "my-repository"
  tag_status      = "TAGGED"
}
```

### Limit Results

```terraform
data "aws_ecr_images" "limited" {
  repository_name = "my-repository"
  max_results     = 10
}
```

### Get Detailed Image Information

```terraform
data "aws_ecr_images" "detailed" {
  repository_name = "my-repository"
  describe_images = true
}

output "image_details" {
  value = data.aws_ecr_images.detailed.image_details
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `registry_id` - (Optional) ID of the Registry where the repository resides.
* `repository_name` - (Required) Name of the ECR Repository.
* `tag_status` - (Optional) Filter images by tag status. Valid values: `TAGGED`, `UNTAGGED`, `ANY`.
* `max_results` - (Optional) Maximum number of images to return.
* `describe_images` - (Optional) Whether to call DescribeImages API to get detailed image information. Defaults to `false`.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `image_ids` - List of image objects containing image digest and tags. Each object has the following attributes:
    * `image_digest` - The sha256 digest of the image manifest.
    * `image_tag` - The tag associated with the image.
* `image_details` - List of detailed image information (only populated when `describe_images` is `true`). Each object has the following attributes:
    * `image_digest` - The sha256 digest of the image manifest.
    * `image_pushed_at` - The date and time when the image was pushed to the repository.
    * `image_size_in_bytes` - The size of the image in bytes.
    * `image_tags` - List of tags associated with the image.
    * `registry_id` - The AWS account ID associated with the registry.
    * `repository_name` - The name of the repository.
