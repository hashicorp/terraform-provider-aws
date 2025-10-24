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
# List all images in a public repository
data "aws_ecrpublic_images" "example" {
  repository_name = "public-repo-name"
}

# Filter images by tag
data "aws_ecrpublic_images" "by_tag" {
  repository_name = "public-repo-name"

  image_ids {
    image_tag = "latest"
  }
}

# Filter images by multiple tags
data "aws_ecrpublic_images" "multi_tag" {
  repository_name = "public-repo-name"

  image_ids {
    image_tag = "latest"
  }

  image_ids {
    image_tag = "v1.0.0"
  }
}

# Filter by image digest
data "aws_ecrpublic_images" "by_digest" {
  repository_name = "public-repo-name"

  image_ids {
    image_digest = "sha256:example1234567890abcdef"
  }
}

# Using registry_id for cross-account access
data "aws_ecrpublic_images" "cross_account" {
  repository_name = "public-repo-name"
  registry_id     = "012345678901"
}
```

## Argument Reference

* `repository_name` - (Required) Name of the public repository.
* `registry_id` - (Optional) The AWS account ID associated with the public registry that contains the repository. If not specified, the default public registry is assumed.
* `image_ids` - (Optional) One or more image ID filters. Each image ID can use either a tag or digest (or both). See [Image IDs](#image-ids) below for more details.

### Image IDs

The `image_ids` configuration block supports the following:

* `image_tag` - (Optional) The tag used for the image.
* `image_digest` - (Optional) The digest of the image manifest.

## Attribute Reference

This data source exports the following attributes:

* `images` - List of images returned. `Each image contains:
    * `digest` - Image digest.
    * `tags` - List of image tags.
    * `size_in_bytes` - Image size in bytes.
    * `pushed_at` - Timestamp when image was pushed.
    * `artifact_media_type` - Media type of the artifact.
    * `image_manifest_media_type` - Media type of the image manifest.
    * `registry_id` - AWS account ID associated with the public registry.
    * `repository_name` - Name of the repository.
