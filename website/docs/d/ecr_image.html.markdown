---
subcategory: "ECR (Elastic Container Registry)"
layout: "aws"
page_title: "AWS: aws_ecr_image"
description: |-
    Provides details about an ECR Image
---

# Data Source: aws_ecr_image

The ECR Image data source allows the details of an image with a particular tag or digest to be retrieved.

## Example Usage

```terraform
data "aws_ecr_image" "service_image" {
  repository_name = "my/service"
  image_tag       = "latest"
}
```

## Argument Reference

This data source supports the following arguments:

* `registry_id` - (Optional) ID of the Registry where the repository resides.
* `repository_name` - (Required) Name of the ECR Repository.
* `image_digest` - (Optional) Sha256 digest of the image manifest. At least one of `image_digest`, `image_tag`, or `most_recent` must be specified.
* `image_tag` - (Optional) Tag associated with this image. At least one of `image_digest`, `image_tag`, or `most_recent` must be specified.
* `most_recent` - (Optional) Return the most recently pushed image. At least one of `image_digest`, `image_tag`, or `most_recent` must be specified.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - SHA256 digest of the image manifest.
* `image_pushed_at` - Date and time, expressed as a unix timestamp, at which the current image was pushed to the repository.
* `image_size_in_bytes` - Size, in bytes, of the image in the repository.
* `image_tags` - List of tags associated with this image.
