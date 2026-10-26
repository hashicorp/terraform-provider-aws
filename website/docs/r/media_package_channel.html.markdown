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

* `channel_id` - (Required) Unique identifier describing the channel
* `description` - (Optional) Description of the channel
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the channel
* `hls_ingest` - Single item list of HLS ingest information
    * `ingest_endpoints` - List of the ingest endpoints
        * `password` - Password
        * `url` - URL
        * `username` - Username
* `id` - Same as `channel_id`
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

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
