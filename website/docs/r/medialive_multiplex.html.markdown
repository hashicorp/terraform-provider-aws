---
subcategory: "Elemental MediaLive"
layout: "aws"
page_title: "AWS: aws_medialive_multiplex"
description: |-
  Terraform resource for managing an AWS MediaLive Multiplex.
---

# Resource: aws_medialive_multiplex

Terraform resource for managing an AWS MediaLive Multiplex.

## Example Usage

### Basic Usage

```terraform
data "aws_availability_zones" "available" {
  state = "available"
}

resource "aws_medialive_multiplex" "example" {
  name               = "example-multiplex-changed"
  availability_zones = [data.aws_availability_zones.available.names[0], data.aws_availability_zones.available.names[1]]

  multiplex_settings {
    transport_stream_bitrate                = 1000000
    transport_stream_id                     = 1
    transport_stream_reserved_bitrate       = 1
    maximum_video_buffer_delay_milliseconds = 1000
  }

  start_multiplex = true

  tags = {
    tag1 = "value1"
  }
}
```

## Argument Reference

The following arguments are required:

* `availability_zones` - (Required) A list of availability zones. You must specify exactly two.
* `multiplex_settings`- (Required) Multiplex settings. See [Multiplex Settings](#multiplex-settings) for more details.
* `name` - (Required) name of Multiplex.

The following arguments are optional:

* `start_multiplex` - (Optional) Whether to start the Multiplex. Defaults to `false`.
* `tags` - (Optional) A map of tags to assign to the Multiplex. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### Multiplex Settings

* `transport_stream_bitrate` - (Required) Transport stream bit rate.
* `transport_stream_id` - (Required) Unique ID for each multiplex.
* `transport_stream_reserved_bitrate` - (Optional) Transport stream reserved bit rate.
* `maximum_video_buffer_delay_milliseconds` - (Optional) Maximum video buffer delay.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Multiplex.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import MediaLive Multiplex using the `id`. For example:

```terraform
import {
  to = aws_medialive_multiplex.example
  id = "12345678"
}
```

Using `terraform import`, import MediaLive Multiplex using the `id`. For example:

```console
% terraform import aws_medialive_multiplex.example 12345678
```
