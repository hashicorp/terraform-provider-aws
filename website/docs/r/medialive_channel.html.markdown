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
      name = "example-video"
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
        video_description_name  = "example-video"
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
* `start_channel` - (Optional) Whether to start/stop channel. Default: `false`
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
* `timecode_config` - (Required) Contains settings used to acquire and adjust timecode information from inputs. See [Timecode Config](#timecode-config) for more details.
* `video_descriptions` - (Required) Video Descriptions. See [Video Descriptions](#video-descriptions) for more details.
* `caption_descriptions` - (Optional) Caption Descriptions. See [Caption Descriptions](#caption-descriptions) for more details.
* `global_configuration` - (Optional) Configuration settings that apply to the event as a whole. See [Global Configuration](#global-configuration) for more details.
* `motion_graphics_configuration` - (Optional) Settings for motion graphics. See [Motion Graphics Configuration](#motion-graphics-configuration) for more details.
* `nielsen_configuration` - (Optional) Nielsen configuration settings. See [Nielsen Configuration](#nielsen-configuration) for more details.
* `avail_blanking` - (Optional) Settings for ad avail blanking. See [Avail Blanking](#avail-blanking) for more details.

### Input Attachments

* `input_attachment_name` - (Optional) User-specified name for the attachment.
* `input_id` - (Required) The ID of the input.
* `input_settings` - (Optional) Settings of an input. See [Input Settings](#input-settings) for more details

### Input Settings

* `audio_selectors` - (Optional) Used to select the audio stream to decode for inputs that have multiple. See [Audio Selectors](#audio-selectors) for more details.
* `caption_selectors` - (Optional) Used to select the caption input to use for inputs that have multiple available. See [Caption Selectors](#caption-selectors) for more details.
* `deblock_filter` - (Optional) Enable or disable the deblock filter when filtering.
* `denoise_filter` - (Optional) Enable or disable the denoise filter when filtering.
* `filter_strength` - (Optional) Adjusts the magnitude of filtering from 1 (minimal) to 5 (strongest).
* `input_filter` - (Optional) Turns on the filter for the input.
* `network_input_settings` - (Optional) Input settings. See [Network Input Settings](#network-input-settings) for more details.
* `scte35_pid` - (Optional) PID from which to read SCTE-35 messages.
* `smpte2038_data_preference` - (Optional) Specifies whether to extract applicable ancillary data from a SMPTE-2038 source in the input.
* `source_end_behavior` - (Optional) Loop input if it is a file.

### Audio Selectors

* `name` - (Required) The name of the audio selector.

### Caption Selectors

* `name` - (Optional) The name of the caption selector.
* `language_code` - (Optional) When specified this field indicates the three letter language code of the caption track to extract from the source.

### Network Input Settings

* `hls_input_settings` - (Optional) Specifies HLS input settings when the uri is for a HLS manifest. See [HLS Input Settings](#hls-input-settings) for more details.
* `server_validation` - (Optional) Check HTTPS server certificates.

### HLS Input Settings

* `bandwidth` - (Optional) The bitrate is specified in bits per second, as in an HLS manifest.
* `buffer_segments` - (Optional) Buffer segments.
* `retries` - (Optional) The number of consecutive times that attempts to read a manifest or segment must fail before the input is considered unavailable.
* `retry_interval` - (Optional) The number of seconds between retries when an attempt to read a manifest or segment fails.
* `scte35_source_type` - (Optional) Identifies the source for the SCTE-35 messages that MediaLive will ingest.

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
* `codec_settings` - (Optional) Audio codec settings. See [Audio Codec Settings](#audio-codec-settings) for more details.

### Audio Normalization Settings

* `algorithm` - (Optional) Audio normalization algorithm to use. itu17701 conforms to the CALM Act specification, itu17702 to the EBU R-128 specification.
* `algorithm_control` - (Optional) Algorithm control for the audio description.
* `target_lkfs` - (Optional) Target LKFS (loudness) to adjust volume to.

### Audio Watermark Settings

* `nielsen_watermark_settings` - (Optional) Settings to configure Nielsen Watermarks in the audio encode. See [Nielsen Watermark Settings](#nielsen-watermark-settings) for more details.

### Audio Codec Settings

* `aac_settings` - (Optional) Aac Settings. See [AAC Settings](#aac-settings) for more details.
* `ac3_settings` - (Optional) Ac3 Settings. See [AC3 Settings](#ac3-settings) for more details.
* `eac3_atmos_settings` - (Optional) - Eac3 Atmos Settings. See [EAC3 Atmos Settings](#eac3-atmos-settings)
* `eac3_settings` - (Optional) - Eac3 Settings. See [EAC3 Settings](#eac3-settings)

### AAC Settings

* `bitrate` - (Optional) Average bitrate in bits/second.
* `coding_mode` - (Optional) Mono, Stereo, or 5.1 channel layout.
* `input_type` - (Optional) Set to "broadcasterMixedAd" when input contains pre-mixed main audio + AD (narration) as a stereo pair.
* `profile` - (Optional) AAC profile.
* `rate_control_mode` - (Optional) The rate control mode.
* `raw_format` - (Optional) Sets LATM/LOAS AAC output for raw containers.
* `sample_rate` - (Optional) Sample rate in Hz.
* `spec` - (Optional) Use MPEG-2 AAC audio instead of MPEG-4 AAC audio for raw or MPEG-2 Transport Stream containers.
* `vbr_quality` - (Optional) VBR Quality Level - Only used if rateControlMode is VBR.

### AC3 Settings

* `bitrate` - (Optional) Average bitrate in bits/second.
* `bitstream_mode` - (Optional) Specifies the bitstream mode (bsmod) for the emitted AC-3 stream.
* `coding_mode` - (Optional) Dolby Digital coding mode.
* `dialnorm` - (Optional) Sets the dialnorm of the output.
* `drc_profile` - (Optional) If set to filmStandard, adds dynamic range compression signaling to the output bitstream as defined in the Dolby Digital specification.
* `lfe_filter` - (Optional) When set to enabled, applies a 120Hz lowpass filter to the LFE channel prior to encoding.
* `metadata_control` - (Optional) Metadata control.

### EAC3 Atmos Settings

* `bitrate` - (Optional) Average bitrate in bits/second.
* `coding_mode` - (Optional) Dolby Digital Plus with dolby Atmos coding mode.
* `dialnorm` - (Optional) Sets the dialnorm for the output.
* `drc_line` - (Optional) Sets the Dolby dynamic range compression profile.
* `drc_rf` - (Optional) Sets the profile for heavy Dolby dynamic range compression.
* `height_trim` - (Optional) Height dimensional trim.
* `surround_trim` - (Optional) Surround dimensional trim.

### EAC3 Settings

* `attenuation_control` - (Optional) Sets the attenuation control.
* `bitrate` - (Optional) Average bitrate in bits/second.
* `bitstream_mode` - (Optional) Specifies the bitstream mode (bsmod) for the emitted AC-3 stream.
* `coding_mode` - (Optional) Dolby Digital Plus coding mode.

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

* `output_group_settings` - (Required) Settings associated with the output group. See [Output Group Settings](#output-group-settings) for more details.
* `outputs` - (Required) List of outputs. See [Outputs](#outputs) for more details.
* `name` - (Optional) Custom output group name defined by the user.

### Output Group Settings

* `archive_group_settings` - (Optional) Archive group settings. See [Archive Group Settings](#archive-group-settings) for more details.
* `media_package_group_settings` - (Optional) Media package group settings. See [Media Package Group Settings](#media-package-group-settings) for more details.
* `multiplex_group_sttings` - (Optional) Multiplex group settings. Attribute can be passed as an empty block.
* `rtmp_group_settings` - (Optional) RTMP group settings. See [RTMP Group Settings](#rtmp-group-settings) for more details.
* `udp_group_sttings` - (Optional) UDP group settings. See [UDP Group Settings](#udp-group-settings) for more details.

### Outputs

* `output_settings` - (Required) Settings for output. See [Output Settings](#output-settings) for more details.
* `audio_description_names` - (Optional) The names of the audio descriptions used as audio sources for the output.
* `caption_description_names` - (Optional) The names of the caption descriptions used as caption sources for the output.
* `output_name` - (Required) The name used to identify an output.
* `video_description_name` - (Optional) The name of the video description used as video source for the output.

### Timecode Config

* `source` - (Optional) The source for the timecode that will be associated with the events outputs.
* `sync_threshold` - (Optional) Threshold in frames beyond which output timecode is resynchronized to the input timecode.

### Video Descriptions

* `name` - (Required) The name of the video description.
* `codec_settings` - (Optional) The video codec settings. See [Video Codec Settings](#video-codec-settings) for more details.
* `height` - Output video height in pixels.
* `respond_to_afd` - (Optional) Indicate how to respond to the AFD values that might be in the input video.
* `scaling_behavior` - (Optional) Behavior on how to scale.
* `sharpness` - (Optional) Changes the strength of the anti-alias filter used for scaling.
* `width` - (Optional) Output video width in pixels.

### Video Codec Settings

* `frame_capture_settings` - (Optional) Frame capture settings. See [Frame Capture Settings](#frame-capture-settings) for more details.
* `h264_settings` - (Optional) H264 settings. See [H264 Settings](#h264-settings) for more details.

### Frame Capture Settings

* `capture_interval` - (Optional) The frequency at which to capture frames for inclusion in the output.
* `capture_interval_units` - (Optional) Unit for the frame capture interval.

### H264 Settings

* `adaptive_quantization` - (Optional) Enables or disables adaptive quantization.
* `afd_signaling` - (Optional) Indicates that AFD values will be written into the output stream.
* `bitrate` - (Optional) Average bitrate in bits/second.
* `buf_fil_pct` - (Optional) Percentage of the buffer that should initially be filled.
* `buf_size` - (Optional) Size of buffer in bits.
* `color_metadata` - (Optional) Includes color space metadata in the output.
* `entropy_encoding` - (Optional) Entropy encoding mode.
* `filter_settings` - (Optional) Filters to apply to an encode. See [H264 Filter Settings](#h264-filter-settings) for more details.
* `fixed_afd` - (Optional) Four bit AFD value to write on all frames of video in the output stream.
* `flicer_aq` - (Optional) Makes adjustments within each frame to reduce flicker on the I-frames.
* `force_field_pictures` - (Optional) Controls whether coding is performed on a field basis or on a frame basis.
* `framerate_control` - (Optional) Indicates how the output video frame rate is specified.
* `framerate_denominator` - (Optional) Framerate denominator.
* `framerate_numerator` - (Optional) Framerate numerator.
* `gop_b_reference` - (Optional) GOP-B reference.
* `gop_closed_cadence` - (Optional) Frequency of closed GOPs.
* `gop_num_b_frames` - (Optional) Number of B-frames between reference frames.
* `gop_size` - (Optional) GOP size in units of either frames of seconds per `gop_size_units`.
* `gop_size_units` - (Optional) Indicates if the `gop_size` is specified in frames or seconds.
* `level` - (Optional) H264 level.
* `look_ahead_rate_control` - (Optional) Amount of lookahead.
* `max_bitrate` - (Optional) Set the maximum bitrate in order to accommodate expected spikes in the complexity of the video.
* `min_interval` - (Optional) Min interval.
* `num_ref_frames` - (Optional) Number of reference frames to use.
* `par_control` - (Optional) Indicates how the output pixel aspect ratio is specified.
* `par_denominator` - (Optional) Pixel Aspect Ratio denominator.
* `par_numerator` - (Optional) Pixel Aspect Ratio numerator.
* `profile` - (Optional) H264 profile.
* `quality_level` - (Optional) Quality level.
* `qvbr_quality_level` - (Optional) Controls the target quality for the video encode.
* `rate_control_mode` - (Optional) Rate control mode.
* `scan_type` - (Optional) Sets the scan type of the output.
* `scene_change_detect` - (Optional) Scene change detection.
* `slices` - (Optional) Number of slices per picture.
* `softness` - (Optional) Softness.
* `spatial_aq` - (Optional) Makes adjustments within each frame based on spatial variation of content complexity.
* `subgop_length` - (Optional) Subgop length.
* `syntax` - (Optional) Produces a bitstream compliant with SMPTE RP-2027.
* `temporal_aq` - (Optional) Makes adjustments within each frame based on temporal variation of content complexity.
* `timecode_insertion` - (Optional) Determines how timecodes should be inserted into the video elementary stream.

### H264 Filter Settings

* `temporal_filter_settings` - (Optional) Temporal filter settings. See [Temporal Filter Settings](#temporal-filter-settings)

### H265 Settings

* `adaptive_quantization` - (Optional) Enables or disables adaptive quantization.
* `afd_signaling` - (Optional) Indicates that AFD values will be written into the output stream.
* `alternative_transfer_function` - (Optional) Whether or not EML should insert an Alternative Transfer Function SEI message.
* `bitrate` - (Required) Average bitrate in bits/second.
* `buf_size` - (Optional) Size of buffer in bits.
* `color_metadata` - (Optional) Includes color space metadata in the output.
* `color_space_settings` (Optional) Define the color metadata for the output. [H265 Color Space Settings](#h265-color-space-settings) for more details.
* `filter_settings` - (Optional) Filters to apply to an encode. See [H265 Filter Settings](#h265-filter-settings) for more details.
* `fixed_afd` - (Optional) Four bit AFD value to write on all frames of video in the output stream.
* `flicer_aq` - (Optional) Makes adjustments within each frame to reduce flicker on the I-frames.
* `framerate_denominator` - (Required) Framerate denominator.
* `framerate_numerator` - (Required) Framerate numerator.
* `gop_closed_cadence` - (Optional) Frequency of closed GOPs.
* `gop_size` - (Optional) GOP size in units of either frames of seconds per `gop_size_units`.
* `gop_size_units` - (Optional) Indicates if the `gop_size` is specified in frames or seconds.
* `level` - (Optional) H265 level.
* `look_ahead_rate_control` - (Optional) Amount of lookahead.
* `max_bitrate` - (Optional) Set the maximum bitrate in order to accommodate expected spikes in the complexity of the video.
* `min_interval` - (Optional) Min interval.
* `par_denominator` - (Optional) Pixel Aspect Ratio denominator.
* `par_numerator` - (Optional) Pixel Aspect Ratio numerator.
* `profile` - (Optional) H265 profile.
* `qvbr_quality_level` - (Optional) Controls the target quality for the video encode.
* `rate_control_mode` - (Optional) Rate control mode.
* `scan_type` - (Optional) Sets the scan type of the output.
* `scene_change_detect` - (Optional) Scene change detection.
* `slices` - (Optional) Number of slices per picture.
* `tier` - (Optional) Set the H265 tier in the output.
* `timecode_burnin_settings` - (Optional) Apply a burned in timecode. See [H265 Timecode Burnin Settings](#h265-timecode-burnin-settings) for more details.
* `timecode_insertion` = (Optional) Determines how timecodes should be inserted into the video elementary stream.

### H265 Color Space Settings

* `color_space_passthrough_settings` - (Optional) Sets the colorspace metadata to be passed through.
* `dolby_vision81_settings` - (Optional) Set the colorspace to Dolby Vision81.
* `hdr10_settings` - (Optional) Set the colorspace to be HDR10. See [H265 HDR10 Settings](#h265-hdr10-settings) for more details.
* `rec601_settings` - (Optional) Set the colorspace to Rec. 601.
* `rec709_settings` - (Optional) Set the colorspace to Rec. 709.

### H265 HDR10 Settings

* `max_cll` - (Optional) Sets the MaxCLL value for HDR10.
* `max_fall` - (Optional) Sets the MaxFALL value for HDR10.

### H265 Filter Settings

* `temporal_filter_settings` - (Optional) Temporal filter settings. See [Temporal Filter Settings](#temporal-filter-settings)

### H265 Timecode Burnin Settings

* `timecode_burnin_font_size` - (Optional) Sets the size of the burned in timecode.
* `timecode_burnin_position` - (Optional) Sets the position of the burned in timecode.
* `prefix` - (Optional) Set a prefix on the burned in timecode.

### Temporal Filter Settings

* `post_filter_sharpening` - (Optional) Post filter sharpening.
* `strength` - (Optional) Filter strength.

### Caption Descriptions

* `accessibility` - (Optional) Indicates whether the caption track implements accessibility features such as written descriptions of spoken dialog, music, and sounds.
* `caption_selector_name` - (Required) Specifies which input caption selector to use as a caption source when generating output captions. This field should match a captionSelector name.
* `destination_settings` - (Optional) Additional settings for captions destination that depend on the destination type. See [Destination Settings](#destination-settings) for more details.
* `language_code` - (Optional) ISO 639-2 three-digit code.
* `language_description` - (Optional) Human readable information to indicate captions available for players (eg. English, or Spanish).
* `name` - (Required) Name of the caption description. Used to associate a caption description with an output. Names must be unique within an event.

### Destination Settings

* `arib_destination_settings` - (Optional) Arib Destination Settings.
* `burn_in_destination_settings` - (Optional) Burn In Destination Settings. See [Burn In Destination Settings](#burn-in-destination-settings) for more details.
* `dvb_sub_destination_settings` - (Optional) Dvb Sub Destination Settings. See [Dvb Sub Destination Settings](#dvb-sub-destination-settings) for more details.
* `ebu_tt_d_destination_settings` - (Optional) Ebu Tt D Destination Settings. See [Ebu Tt D Destination Settings](#ebu-tt-d-destination-settings) for more details.
* `embedded_destination_settings` - (Optional) Embedded Destination Settings.
* `embedded_plus_scte20_destination_settings` - (Optional) Embedded Plus Scte20 Destination Settings.
* `rtmp_caption_info_destination_settings` - (Optional) Rtmp Caption Info Destination Settings.
* `scte20_plus_embedded_destination_settings` - (Optional) Scte20 Plus Embedded Destination Settings.
* `scte27_destination_settings` – (Optional) Scte27 Destination Settings.
* `smpteTt_destination_settings` – (Optional) Smpte Tt Destination Settings.
* `teletext_destination_settings` – (Optional) Teletext Destination Settings.
* `ttml_destination_settings` – (Optional) Ttml Destination Settings. See [Ttml Destination Settings](#ttml-destination-settings) for more details.
* `webvtt_destination_settings` - (Optional) Webvtt Destination Settings. See [Webvtt Destination Settings](#webvtt-destination-settings) for more details.

### Burn In Destination Settings

* `alignment` – (Optional) If no explicit xPosition or yPosition is provided, setting alignment to centered will place the captions at the bottom center of the output. Similarly, setting a left alignment will align captions to the bottom left of the output. If x and y positions are given in conjunction with the alignment parameter, the font will be justified (either left or centered) relative to those coordinates. Selecting “smart” justification will left-justify live subtitles and center-justify pre-recorded subtitles. All burn-in and DVB-Sub font settings must match.
* `background_color` – (Optional) Specifies the color of the rectangle behind the captions. All burn-in and DVB-Sub font settings must match.
* `background_opacity` – (Optional) Specifies the opacity of the background rectangle. 255 is opaque; 0 is transparent. Leaving this parameter out is equivalent to setting it to 0 (transparent). All burn-in and DVB-Sub font settings must match.
* `font` – (Optional) External font file used for caption burn-in. File extension must be ‘ttf’ or ‘tte’. Although the user can select output fonts for many different types of input captions, embedded, STL and teletext sources use a strict grid system. Using external fonts with these caption sources could cause unexpected display of proportional fonts. All burn-in and DVB-Sub font settings must match. See [Font](#font) for more details.
* `font_color` – (Optional) Specifies the color of the burned-in captions. This option is not valid for source captions that are STL, 608/embedded or teletext. These source settings are already pre-defined by the caption stream. All burn-in and DVB-Sub font settings must match.
* `font_opacity` – (Optional) Specifies the opacity of the burned-in captions. 255 is opaque; 0 is transparent. All burn-in and DVB-Sub font settings must match.
* `font_resolution` – (Optional) Font resolution in DPI (dots per inch); default is 96 dpi. All burn-in and DVB-Sub font settings must match.
* `font_size` – (Optional) When set to ‘auto’ fontSize will scale depending on the size of the output. Giving a positive integer will specify the exact font size in points. All burn-in and DVB-Sub font settings must match.
* `outline_color` – (Optional) Specifies font outline color. This option is not valid for source captions that are either 608/embedded or teletext. These source settings are already pre-defined by the caption stream. All burn-in and DVB-Sub font settings must match.
* `outline_size` – (Optional) Specifies font outline size in pixels. This option is not valid for source captions that are either 608/embedded or teletext. These source settings are already pre-defined by the caption stream. All burn-in and DVB-Sub font settings must match.
* `shadow_color` – (Optional) Specifies the color of the shadow cast by the captions. All burn-in and DVB-Sub font settings must match.
* `shadow_opacity` – (Optional) Specifies the opacity of the shadow. 255 is opaque; 0 is transparent. Leaving this parameter out is equivalent to setting it to 0 (transparent). All burn-in and DVB-Sub font settings must match.
* `shadow_x_offset` – (Optional) Specifies the horizontal offset of the shadow relative to the captions in pixels. A value of -2 would result in a shadow offset 2 pixels to the left. All burn-in and DVB-Sub font settings must match.
* `shadow_y_offset` – (Optional) Specifies the vertical offset of the shadow relative to the captions in pixels. A value of -2 would result in a shadow offset 2 pixels above the text. All burn-in and DVB-Sub font settings must match.
* `teletext_grid_control` – (Optional) Controls whether a fixed grid size will be used to generate the output subtitles bitmap. Only applicable for Teletext inputs and DVB-Sub/Burn-in outputs.
* `x_position` – (Optional) Specifies the horizontal position of the caption relative to the left side of the output in pixels. A value of 10 would result in the captions starting 10 pixels from the left of the output. If no explicit xPosition is provided, the horizontal caption position will be determined by the alignment parameter. All burn-in and DVB-Sub font settings must match.
* `y_position` – (Optional) Specifies the vertical position of the caption relative to the top of the output in pixels. A value of 10 would result in the captions starting 10 pixels from the top of the output. If no explicit yPosition is provided, the caption will be positioned towards the bottom of the output. All burn-in and DVB-Sub font settings must match.

### Dvb Sub Destination Settings

* `alignment` – (Optional) If no explicit xPosition or yPosition is provided, setting alignment to centered will place the captions at the bottom center of the output. Similarly, setting a left alignment will align captions to the bottom left of the output. If x and y positions are given in conjunction with the alignment parameter, the font will be justified (either left or centered) relative to those coordinates. Selecting “smart” justification will left-justify live subtitles and center-justify pre-recorded subtitles. This option is not valid for source captions that are STL or 608/embedded. These source settings are already pre-defined by the caption stream. All burn-in and DVB-Sub font settings must match.
* `background_color` – (Optional) Specifies the color of the rectangle behind the captions. All burn-in and DVB-Sub font settings must match.
* `background_opacity` – (Optional) Specifies the opacity of the background rectangle. 255 is opaque; 0 is transparent. Leaving this parameter blank is equivalent to setting it to 0 (transparent). All burn-in and DVB-Sub font settings must match.
* `font` – (Optional) External font file used for caption burn-in. File extension must be ‘ttf’ or ‘tte’. Although the user can select output fonts for many different types of input captions, embedded, STL and teletext sources use a strict grid system. Using external fonts with these caption sources could cause unexpected display of proportional fonts. All burn-in and DVB-Sub font settings must match. See [Font](#font) for more details.
* `font_color` – (Optional) Specifies the color of the burned-in captions. This option is not valid for source captions that are STL, 608/embedded or teletext. These source settings are already pre-defined by the caption stream. All burn-in and DVB-Sub font settings must match.
* `font_opacity` – (Optional) Specifies the opacity of the burned-in captions. 255 is opaque; 0 is transparent. All burn-in and DVB-Sub font settings must match.
* `font_resolution` – (Optional) Font resolution in DPI (dots per inch); default is 96 dpi. All burn-in and DVB-Sub font settings must match.
* `font_size` – (Optional) When set to auto fontSize will scale depending on the size of the output. Giving a positive integer will specify the exact font size in points. All burn-in and DVB-Sub font settings must match.
* `outline_color` – (Optional) Specifies font outline color. This option is not valid for source captions that are either 608/embedded or teletext. These source settings are already pre-defined by the caption stream. All burn-in and DVB-Sub font settings must match.
* `outline_size` – (Optional) Specifies font outline size in pixels. This option is not valid for source captions that are either 608/embedded or teletext. These source settings are already pre-defined by the caption stream. All burn-in and DVB-Sub font settings must match.
* `shadow_color` – (Optional) Specifies the color of the shadow cast by the captions. All burn-in and DVB-Sub font settings must match.
* `shadow_opacity` – (Optional) Specifies the opacity of the shadow. 255 is opaque; 0 is transparent. Leaving this parameter blank is equivalent to setting it to 0 (transparent). All burn-in and DVB-Sub font settings must match.
* `shadow_x_offset` – (Optional) Specifies the horizontal offset of the shadow relative to the captions in pixels. A value of -2 would result in a shadow offset 2 pixels to the left. All burn-in and DVB-Sub font settings must match.
* `shadow_y_offset` – (Optional) Specifies the vertical offset of the shadow relative to the captions in pixels. A value of -2 would result in a shadow offset 2 pixels above the text. All burn-in and DVB-Sub font settings must match.
* `teletext_grid_control` – (Optional) Controls whether a fixed grid size will be used to generate the output subtitles bitmap. Only applicable for Teletext inputs and DVB-Sub/Burn-in outputs.
* `x_position` – (Optional) Specifies the horizontal position of the caption relative to the left side of the output in pixels. A value of 10 would result in the captions starting 10 pixels from the left of the output. If no explicit xPosition is provided, the horizontal caption position will be determined by the alignment parameter. This option is not valid for source captions that are STL, 608/embedded or teletext. These source settings are already pre-defined by the caption stream. All burn-in and DVB-Sub font settings must match.
* `y_position` – (Optional) Specifies the vertical position of the caption relative to the top of the output in pixels. A value of 10 would result in the captions starting 10 pixels from the top of the output. If no explicit yPosition is provided, the caption will be positioned towards the bottom of the output. This option is not valid for source captions that are STL, 608/embedded or teletext. These source settings are already pre-defined by the caption stream. All burn-in and DVB-Sub font settings must match.

### Ebu Tt D Destination Settings

* `copyright_holder` – (Optional) Complete this field if you want to include the name of the copyright holder in the copyright tag in the captions metadata.
* `fill_line_gap` – (Optional) Specifies how to handle the gap between the lines (in multi-line captions). - enabled: Fill with the captions background color (as specified in the input captions). - disabled: Leave the gap unfilled.
* `font_family` – (Optional) Specifies the font family to include in the font data attached to the EBU-TT captions. Valid only if styleControl is set to include. If you leave this field empty, the font family is set to “monospaced”. (If styleControl is set to exclude, the font family is always set to “monospaced”.) You specify only the font family. All other style information (color, bold, position and so on) is copied from the input captions. The size is always set to 100% to allow the downstream player to choose the size. - Enter a list of font families, as a comma-separated list of font names, in order of preference. The name can be a font family (such as “Arial”), or a generic font family (such as “serif”), or “default” (to let the downstream player choose the font). - Leave blank to set the family to “monospace”.
* `style_control` – (Optional) Specifies the style information (font color, font position, and so on) to include in the font data that is attached to the EBU-TT captions. - include: Take the style information (font color, font position, and so on) from the source captions and include that information in the font data attached to the EBU-TT captions. This option is valid only if the source captions are Embedded or Teletext. - exclude: In the font data attached to the EBU-TT captions, set the font family to “monospaced”. Do not include any other style information.

### Ttml Destination Settings

* `style_control` – (Optional) This field is not currently supported and will not affect the output styling. Leave the default value.

### Webvtt Destination Settings

* `style_control` - (Optional) Controls whether the color and position of the source captions is passed through to the WebVTT output captions. PASSTHROUGH - Valid only if the source captions are EMBEDDED or TELETEXT. NO\_STYLE\_DATA - Don’t pass through the style. The output captions will not contain any font styling information.

### Font

* `password_param` – (Optional) Key used to extract the password from EC2 Parameter store.
* `uri` – (Required) Path to a file accessible to the live stream.
* `username` – (Optional) Username to be used.

### Global Configuration

* `initial_audio_gain` – (Optional) Value to set the initial audio gain for the Live Event.
* `input_end_action` – (Optional) Indicates the action to take when the current input completes (e.g. end-of-file). When switchAndLoopInputs is configured the encoder will restart at the beginning of the first input. When “none” is configured the encoder will transcode either black, a solid color, or a user specified slate images per the “Input Loss Behavior” configuration until the next input switch occurs (which is controlled through the Channel Schedule API).
* `input_loss_behavior` - (Optional) Settings for system actions when input is lost. See [Input Loss Behavior](#input-loss-behavior) for more details.
* `output_locking_mode` – (Optional) Indicates how MediaLive pipelines are synchronized. PIPELINE\_LOCKING - MediaLive will attempt to synchronize the output of each pipeline to the other. EPOCH\_LOCKING - MediaLive will attempt to synchronize the output of each pipeline to the Unix epoch.
* `output_timing_source` – (Optional) Indicates whether the rate of frames emitted by the Live encoder should be paced by its system clock (which optionally may be locked to another source via NTP) or should be locked to the clock of the source that is providing the input stream.
* `support_low_framerate_inputs` – (Optional) Adjusts video input buffer for streams with very low video framerates. This is commonly set to enabled for music channels with less than one video frame per second.

### Input Loss Behavior

* `password_param` – (Optional) Key used to extract the password from EC2 Parameter store.
* `uri` – (Required) Path to a file accessible to the live stream.
* `username` – (Optional) Username to be used.

### Motion Graphics Configuration

* `motion_graphics_insertion` – (Optional) Motion Graphics Insertion.
* `motion_graphics_settings`– (Required) Motion Graphics Settings. See [Motion Graphics Settings](#motion-graphics-settings) for more details.

### Motion Graphics Settings

* `html_motion_graphics_settings` – (Optional) Html Motion Graphics Settings.

### Nielsen Configuration

* `distributor_id` – (Optional) Enter the Distributor ID assigned to your organization by Nielsen.
* `nielsen_pcm_to_id3_tagging` – (Optional) Enables Nielsen PCM to ID3 tagging.

### Avail Blanking

* `avail_blanking_image` - (Optional) Blanking image to be used. See [Avail Blanking Image](#avail-blanking-image) for more details.
* `state` - (Optional) When set to enabled, causes video, audio and captions to be blanked when insertion metadata is added.

### Avail Blanking Image

* `uri` - (Required) Path to a file accessible to the live stream.
* `password_param` - (Optional) Key used to extract the password from EC2 Parameter store.
* `username` - (Optional). Username to be used.

### Archive Group Settings

* `destination` - (Required) A director and base filename where archive files should be written. See [Destination](#destination) for more details.
* `archive_cdn_settings` - (Optional) Parameters that control the interactions with the CDN. See [Archive CDN Settings](#archive-cdn-settings) for more details.
* `rollover_interval` - (Optional) Number of seconds to write to archive file before closing and starting a new one.

### Media Package Group Settings

* `destination` - (Required) A director and base filename where archive files should be written. See [Destination](#destination) for more details.

### RTMP Group Settings

* `ad_markers` - (Optional) The ad marker type for this output group.
* `authentication_scheme` - (Optional) Authentication scheme to use when connecting with CDN.
* `cache_full_behavior` - (Optional) Controls behavior when content cache fills up.
* `cache_length` - (Optional) Cache length in seconds, is used to calculate buffer size.
* `caption_data` - (Optional) Controls the types of data that passes to onCaptionInfo outputs.
* `input_loss_action` - (Optional) Controls the behavior of the RTMP group if input becomes unavailable.
* `restart_delay` - (Optional) Number of seconds to wait until a restart is initiated.

### UDP Group Settings

* `input_loss_action` - (Optional) Specifies behavior of last resort when input video os lost.
* `timed_metadata_id3_frame` - (Optional) Indicates ID3 frame that has the timecode.
* `timed_metadta_id3_period`- (Optional) Timed metadata interval in seconds.

### Destination

* `destination_ref_id` - (Required) Reference ID for the destination.

### Archive CDN Settings

* `archive_s3_settings` - (Optional) Archive S3 Settings. See [Archive S3 Settings](#archive-s3-settings) for more details.

### Archive S3 Settings

* `canned_acl` - (Optional) Specify the canned ACL to apply to each S3 request.

### Output Settings

* `archive_output_settings` - (Optional) Archive output settings. See [Archive Output Settings](#archive-output-settings) for more details.
* `media_package_output_settings` - (Optional) Media package output settings. This can be set as an empty block.
* `multiplex_output_settings` - (Optional) Multiplex output settings. See [Multiplex Output Settings](#multiplex-output-settings) for more details.
* `rtmp_output_settings` - (Optional) RTMP output settings. See [RTMP Output Settings](#rtmp-output-settings) for more details.
* `udp_output_settings` - (Optional) UDP output settings. See [UDP Output Settings](#udp-output-settings) for more details

### Archive Output Settings

* `container_settings` - (Required) Settings specific to the container type of the file. See [Container Settings](#container-settings) for more details.
* `extension` - (Optional) Output file extension.
* `name_modifier` - (Optional) String concatenated to the end of the destination filename. Required for multiple outputs of the same type.

### Multiplex Output Settings

* `destination` - (Required) Destination is a multiplex. See [Destination](#destination) for more details.

### RTMP Output Settings

- `destination` - (Required) The RTMP endpoint excluding the stream name. See [Destination](#destination) for more details.
- `certificate_mode` - (Optional) Setting to allow self signed or verified RTMP certificates.
- `connection_retry_interval` - (Optional) Number of seconds to wait before retrying connection to the flash media server if the connection is lost.
- `num_retries` - (Optional) Number of retry attempts.

### Container Settings

* `m2ts_settings` - (Optional) M2ts Settings. See [M2ts Settings](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-medialive-channel-m2tssettings.html) for more details.
* `raw_settings`- (Optional) Raw Settings. This can be set as an empty block.

### UDP Output Settings

* `container_settings` - (Required) UDP container settings. See [Container Settings](#container-settings) for more details.
* `destination` - (Required) Destination address and port number for RTP or UDP packets. See [Destination](#destination) for more details.
* `buffer_msec` - (Optional) UDP output buffering in milliseconds.
* `fec_output_setting` - (Optional) Settings for enabling and adjusting Forward Error Correction on UDP outputs. See [FEC Output Settings](#fec-output-settings) for more details.

### FEC Output Settings

* `column_depth` - (Optional) The height of the FEC protection matrix.
* `include_fec` - (Optional) Enables column only or column and row based FEC.
* `row_length` - (Optional) The width of the FEC protection matrix.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Channel.
* `channel_id` - ID of the Channel.

## Timeouts

[Configuration options](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts):

* `create` - (Default `15m`)
* `update` - (Default `15m`)
* `delete` - (Default `15m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import MediaLive Channel using the `channel_id`. For example:

```terraform
import {
  to = aws_medialive_channel.example
  id = "1234567"
}
```

Using `terraform import`, import MediaLive Channel using the `channel_id`. For example:

```console
% terraform import aws_medialive_channel.example 1234567
```
