---
subcategory: "Elemental MediaPackage"
layout: "aws"
page_title: "AWS: aws_media_package_channel"
description: |-
  Provides an AWS Elemental MediaPackage Channel.
---

# Resource: aws_media_package_channel

Provides an AWS Elemental MediaPackage Channel.

## Example Usage

```terraform
resource "aws_media_package_channel" "kittens" {
  channel_id  = "kitten-channel"
  description = "A channel dedicated to amusing videos of kittens."
}
```

## Argument Reference

This resource supports the following arguments:

* `channel_id` - (Required) A unique identifier describing the channel
* `description` - (Optional) A description of the channel
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The same as `channel_id`
* `arn` - The ARN of the channel
* `hls_ingest` - A single item list of HLS ingest information
    * `ingest_endpoints` - A list of the ingest endpoints
        * `password` - The password
        * `url` - The URL
        * `username` - The username
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Media Package Channels using the channel ID. For example:

```terraform
import {
  to = aws_media_package_channel.kittens
  id = "kittens-channel"
}
```

Using `terraform import`, import Media Package Channels using the channel ID. For example:

```console
% terraform import aws_media_package_channel.kittens kittens-channel
```
