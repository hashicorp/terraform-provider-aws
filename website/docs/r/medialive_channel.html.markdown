---
subcategory: "Elemental MediaLive"
layout: "aws"
page_title: "AWS: aws_medialive_channel"
description: |-
  Terraform resource for managing an AWS MediaLive Channel.
---

# Resource: aws_medialive_channel

Terraform resource for managing an AWS MediaLive Channel.

## Example Usage

### Basic Usage

```terraform
resource "aws_medialive_channel" "example" {
  name          = "example-channel"
  channel_class = "STANDARD"
  role_arn      = aws_iam_role.example.arn

  input_specification {
    codec            = "AVC"
    input_resolution = "HD"
    maximum_bitrate  = "MAX_20_MBPS"
  }

  input_attachments {
    input_attachment_name = "example-input"
    input_id              = aws_medialive_input.example.id

  }

  destinations {
    id = "destination"

    settings {
      url = "s3://${aws_s3_bucket.main.id}/test1"
    }

    settings {
      url = "s3://${aws_s3_bucket.main2.id}/test2"
    }
  }

  encoder_settings {
    timecode_config {
      source = "EMBEDDED"
    }

    audio_descriptions {
      audio_selector_name = "example audio selector"
      name                = "audio-selector"
    }

    video_descriptions {
      name = "example-vidoe"
    }

    output_groups {
      output_group_settings {
        archive_group_settings {
          destination {
            destination_ref_id = "destination"
          }
        }
      }

      outputs {
        output_name             = "example-name"
        video_description_name  = "example-vidoe"
        audio_description_names = ["audio-selector"]
        output_settings {
          archive_output_settings {
            name_modifier = "_1"
            extension     = "m2ts"
            container_settings {
              m2ts_settings {
                audio_buffer_model = "ATSC"
                buffer_model       = "MULTIPLEX"
                rate_mode          = "CBR"
              }
            }
          }
        }
      }
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `channel_class` - (Required) Concise argument description.
* `destinations` - (Required) Destinations for channel. See [Destinations](#destinations) for more details.
* `encoder_settings` - (Required) Encoder settings. See [Encoder Settings](#encoder-settings) for more details.
* `input_specification` - (Required) Specification of network and file inputs for the channel. 
* `name` - (Required) Name of the Channel.


The following arguments are optional:

* `cdi_input_specification` - (Optional) Specification of CDI inputs for this channel. See [CDI Input Specification](#cdi-input-specification) for more details.
* `input_attachments` - (Optional) Input attachments for the channel. See [Input Attachments](#input-attachments) for more details.
* `log_level` - (Optional) The log level to write to Cloudwatch logs.
* `maintenance` - (Optional) Maintenance settings for this channel. See [Maintenance](#maintenance) for more details.
* `role_arn` - (Optional) Concise argument description.
* `tags` - (Optional) A map of tags to assign to the channel. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `vpc` - (Optional) Settings for the VPC outputs.

### CDI Input Specification


### Destinations


### Encoder Settings


### Input Attachments


### Maintenance


## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the Channel.
* `arn` - ARN of the Channel.

## Timeouts

[Configuration options](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts):

* `create` - (Default `15m`)
* `update` - (Default `15m`)
* `delete` - (Default `15m`)

## Import

MediaLive Channel can be imported using the `id`, e.g.,

```
$ terraform import aws_medialive_channel.example 1234567
```
