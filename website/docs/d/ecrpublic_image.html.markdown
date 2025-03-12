---
subcategory: "ECR Public"
layout: "aws"
page_title: "AWS: aws_ecrpublic_image"
description: |-
  The ECR public Image data source allows the details of an image with a particular tag or digest to be retrieved.
---

# Data Source: aws_ecrpublic_image

Terraform data source for managing an AWS ECR Public Image.

## Example Usage

### Basic Usage

```terraform
data "aws_ecrpublic_image" "example" {
   repository_name = "my/service"
   image_tag       = "latest"
}
```

## Argument Reference

The following arguments are required:

* `repository_name` - (Required) Name of the ECR Public Repository.

The following arguments are optional:

* `registry_id` - (Optional) AWS Account ID of the Registry where the repository resides. Default is yours.
* `image_digest` - (Optional) Sha256 digest of the image manifest. At least one of `image_digest`, `image_tag`, or `most_recent` must be specified.
* `image_tag` - (Optional) Tag associated with this image. At least one of `image_digest`, `image_tag`, or `most_recent` must be specified.
* `most_recent` - (Optional) Return the most recently pushed image. At least one of `image_digest`, `image_tag`, or `most_recent` must be specified.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - SHA256 digest of the image manifest.
* `image_pushed_at` - Date and time, expressed as a unix timestamp, at which the current image was pushed to the repository.
* `image_size_in_bytes` - Size, in bytes, of the image in the repository.
* `image_tags` - List of tags associated with this image.
* `image_uri` - The URI for the specific image version specified by `image_tag` or `image_digest`.