---
subcategory: "Elemental MediaPackage Version 2"
layout: "aws"
page_title: "AWS: aws_media_packagev2_channel"
description: |-
  Creates an AWS Elemental MediaPackage Version 2 Channel.
---

# Resource: aws_media_packagev2_channel

Creates an AWS Elemental MediaPackage Version 2 Channel.

## Example Usage

```terraform
resource "aws_media_packagev2_channel_group" "exampleGroup" {
  name        = "example"
  description = "channel group for example channels"
}

resource "aws_media_packagev2_channel" "channel" {
  channel_group_name = aws_media_packagev2_channel_group.exampleGroup.name
  name               = "channel"
  description        = "channel in the example channel group"
}
```

## Argument Reference

This resource supports the following arguments:

* `channel_group_name` - (Required) The name of the channel group to create this channel in.
* `name` - (Required) A unique identifier naming the channel.
* `description` - (Optional) A description of the channel.
* `input_type` - (Optional) The input type of the stream being pushed to the channel E.g. `HLS` or `CMAF`. See [Supported input types](https://docs.aws.amazon.com/mediapackage/latest/userguide/supported-inputs.html#suported-inputs-codecs-live). The default value is `HLS`
* `input_switch_configuration` - (Optional) Defines the input switching configuration for the channel.
* `output_header_configuration` - (Optional) Defines the output header configuration for the channel.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `input_switch_configuration` Configuration Block

* `mqcs_input_switching` - (Optional) A Boolean value set determines if Media Quality Confidence Score should be enabled. This can only be set to `true` if the `input_type` is `CMAF`. See [Media Quality Scores](https://docs.aws.amazon.com/mediapackage/latest/userguide/mqcs.html)

### `output_header_configuration` Configuration Block

* `publish_mqcs` - (Optional) A Boolean value set determines if MQCS will show up as CMSD keys in the CMSD-Static output headers. This can only be set to `true` if the `input_type` is `CMAF`. See [CMSD headers from AWS Elemental MediaPackage](https://docs.aws.amazon.com/mediapackage/latest/userguide/mqcs.html)

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The ARN of the channel.
* `description` - The same as `description`.
* `ingest_endpoints` - List of ingest endpoints for this channel.
    * `id` - Id of the ingest endpoint. `1` or `2`.
    * `url` - Url of the ingest endpoint.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import an Elemental MediaPackage Version 2 Channel using the channel group's `name` and channel `name` separated by a `/`. For example:

```terraform
import {
  to = aws_media_packagev2_channel.example
  id = "ChannelGroupName/ChannelName"
}
```

Using `terraform import`, import Elemental MediaPackage Version 2 Channel using the channel group's `name` and channel `name` separated by a `/`. For example:

```console
% terraform import aws_media_packagev2_channel.example "ChannelGroupName/ChannelName"
```
