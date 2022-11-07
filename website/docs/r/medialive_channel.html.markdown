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

* `resolution` - (Required) - Maximum CDI input resolution.

### Destinations

* `id` - (Required) User-specified id. Ths is used in an output group or an output.
* `media_package_settings` - (Optional) Destination settings for a MediaPackage output; one destination for both encoders. See [Media Package Settings](#media-package-settings) for more details.
* `multiplex_settings` - (Optional) Destination settings for a Multiplex output; one destination for both encoders. See [Multiplex Settings](#multiplex-settings) for more details.
* `settings` - (Optional) Destination settings for a standard output; one destination for each redundant encoder. See [Settings](#settings) for more details.

### Encoder Settings

* `audio_descriptions` - (Required) Audio descriptions for the channel. See [Audio Descriptions](#audio-descriptions) for more details.
* `output_groups` - (Required) Output groups for the channel. See [Output Groups](#output-groups) for more details.

### Input Attachments

* `input_attachment_name` - (Optional) User-specified name for the attachment.
* `input_id` - (Required) The ID of the input.
* `input_settings` - (Optional) Settings of an input. See [Input Settings](#input-settings) for more details

### Input Settings

* `audio_selectors` - (Optional) Used to select the audio stream to decode for inputs that have multiple. See [Audio Selectors](#audio-selectors) for more details.

### Audio Selectors

* `name` - (Required) The name of the audio selector.

### Maintenance

* `maintenance_day` - (Optional) The day of the week to use for maintenance.
* `maintenance_start_time` - (Optional) The hour maintenance will start.

### Media Package Settings

* `channel_id` - (Required) ID of the channel in MediaPackage that is the destination for this output group.

### Multiplex Settings

* `multiplex_id` - (Required) The ID of the Multiplex that the encoder is providing output to.
* `program_name` - (Optional) The program name of the Multiplex program that the encoder is providing output to.

### Settings

* `password_param` - (Optional) Key used to extract the password from EC2 Parameter store.
* `stream_name` - (Optional) Stream name RTMP destinations (URLs of type rtmp://)
* `url` - (Optional) A URL specifying a destination.
* `username` - (Optional) Username for destination.

### Audio Descriptions

* `audio_selector_name` - (Required) The name of the audio selector used as the source for this AudioDescription.
* `name` - (Required) The name of this audio description.
* `audio_normalization_settings` - (Optional) Advanced audio normalization settings. See [Audio Normalization Settings](#audio-normalization-settings) for more details.
* `audio_type` - (Optional) Applies only if audioTypeControl is useConfigured. The values for audioType are defined in ISO-IEC 13818-1.
* `audio_type_control` - (Optional) Determined how audio type is determined.
* `audio_watermark_settings` - (Optional) Settings to configure one or more solutions that insert audio watermarks in the audio encode. See [Audio Watermark Settings](#audio-watermark-settings) for more details.

### Audio Normalization Settings

* `algorithm` - (Optional) Audio normalization algorithm to use. itu17701 conforms to the CALM Act specification, itu17702 to the EBU R-128 specification.
* `algorithm_control` - (Optional) Algorithm control for the audio description.
* `target_lkfs` - (Optional) Target LKFS (loudness) to adjust volume to.

### Audio Watermark Settings

* `nielsen_watermark_settings` - (Optional) Settings to configure Nielsen Watermarks in the audio encode. See [Nielsen Watermark Settings](#nielsen-watermark-settings) for more details.

### Nielsen Watermark Settings

* `nielsen_cbet_settings` - (Optional) Used to insert watermarks of type Nielsen CBET. See [Nielsen CBET Settings](#nielsen-cbet-settings) for more details.
* `nielsen_distribution_type` - (Optional) Distribution types to assign to the watermarks. Options are `PROGRAM_CONTENT` and `FINAL_DISTRIBUTOR`.
* `nielsen_naes_ii_nw_settings` - (Optional) Used to insert watermarks of type Nielsen NAES, II (N2) and Nielsen NAES VI (NW). See [Nielsen NAES II NW Settings](#nielsen-naes-ii-nw-settings) for more details.

### Nielsen CBET Settings

* `cbet_check_digit` - (Required) CBET check digits to use for the watermark.
* `cbet_stepaside` - (Required) Determines the method of CBET insertion mode when prior encoding is detected on the same layer.
* `csid` - (Required) CBET source ID to use in the watermark.

### Nielsen NAES II NW Settings

* `check_digit` - (Required) Check digit string for the watermark.
* `sid` - (Required) The Nielsen Source ID to include in the watermark.

### Output Groups
* `output_group_settings` - (Required) Settings associated with the output group.
* `outputs` - (Required) List of outputs.
* `name` - (Optional) Custom output group name defined by the user.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of the Channel.
* `channel_id` - ID of the Channel.

## Timeouts

[Configuration options](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts):

* `create` - (Default `15m`)
* `update` - (Default `15m`)
* `delete` - (Default `15m`)

## Import

MediaLive Channel can be imported using the `channel_id`, e.g.,

```
$ terraform import aws_medialive_channel.example 1234567
```
