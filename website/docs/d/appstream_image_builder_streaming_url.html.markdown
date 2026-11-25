---
subcategory: "AppStream 2.0"
layout: "aws"
page_title: "AWS: aws_appstream_image_builder_streaming_url"
description: |-
  Provides details about an AWS AppStream 2.0 Image Builder Streaming URL.
---

# Data Source: aws_appstream_image_builder_streaming_url

Provides details about an AWS AppStream 2.0 Image Builder Streaming URL. Creates a URL to start an image builder streaming session.

## Example Usage

### Basic Usage

```terraform
resource "aws_appstream_image_builder" "example" {
  name          = "example"
  instance_type = "stream.standard.small"
  image_name    = "AppStream-WinServer2019-10-05-2022"
}

data "aws_appstream_image_builder_streaming_url" "example" {
  name = aws_appstream_image_builder.example.name
}
```

### With Custom Validity

```terraform
data "aws_appstream_image_builder_streaming_url" "example" {
  name     = "existing-image-builder"
  validity = 7200  # 2 hours
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the image builder.

The following arguments are optional:

* `validity` - (Optional) Duration (in seconds) for which the streaming URL will be valid. Must be a value from 1 to 604800 (1 week). Defaults to 3600 (1 hour).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - Name of the image builder.
* `streaming_url` - Streaming URL for the image builder.
* `expires` - Time when the streaming URL expires (RFC3339 format).

