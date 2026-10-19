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

* `channel_class` - (Required) Class of the channel. Valid values: `STANDARD`, `SINGLE_PIPELINE`.
* `destinations` - (Required) Destinations for the channel. See [`destinations` Block](#destinations-block) for details.
* `encoder_settings` - (Required) Encoder settings. See [`encoder_settings` Block](#encoder_settings-block) for details.
* `input_specification` - (Required) Specification of network and file inputs for the channel. See [`input_specification` Block](#input_specification-block) for details.
* `name` - (Required) Name of the channel.

The following arguments are optional:

* `cdi_input_specification` - (Optional) Specification of CDI inputs for this channel. See [`cdi_input_specification` Block](#cdi_input_specification-block) for details.
* `input_attachments` - (Optional) Input attachments for the channel. See [`input_attachments` Block](#input_attachments-block) for details.
* `log_level` - (Optional) Level of verbosity for the logs written to CloudWatch Logs.
* `maintenance` - (Optional) Maintenance settings for this channel. See [`maintenance` Block](#maintenance-block) for details.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `role_arn` - (Optional) ARN of the IAM role for the channel.
* `start_channel` - (Optional) Whether to start/stop channel. Default: `false`.
* `tags` - (Optional) Map of tags to assign to the channel. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `vpc` - (Optional) Settings for the VPC outputs. See [`vpc` Block](#vpc-block) for details.

### `input_specification` Block

* `codec` - (Required) Codec to use for the input.
* `input_resolution` - (Required) Maximum input resolution.
* `maximum_bitrate` - (Required) Maximum input bitrate for an input.

### `cdi_input_specification` Block

* `resolution` - (Required) Maximum CDI input resolution.

### `destinations` Block

* `id` - (Required) User-specified ID used in an output group or an output.
* `media_package_settings` - (Optional) Destination settings for a MediaPackage output; one destination for both encoders. See [`media_package_settings` Block](#media_package_settings-block) for details.
* `multiplex_settings` - (Optional) Destination settings for a Multiplex output; one destination for both encoders. See [`multiplex_settings` Block](#multiplex_settings-block) for details.
* `settings` - (Optional) Destination settings for a standard output; one destination for each redundant encoder. See [`settings` Block](#settings-block) for details.

### `media_package_settings` Block

* `channel_id` - (Required) ID of the channel in MediaPackage that is the destination for this output group.

### `multiplex_settings` Block

* `multiplex_id` - (Required) ID of the Multiplex that the encoder is providing output to.
* `program_name` - (Required) Program name of the Multiplex program that the encoder is providing output to.

### `settings` Block

* `password_param` - (Optional) Key used to extract the password from EC2 Parameter store.
* `stream_name` - (Optional) Stream name for RTMP destinations (URLs of type `rtmp://`).
* `url` - (Optional) URL for the destination.
* `username` - (Optional) Username for the destination.

### `input_attachments` Block

* `automatic_input_failover_settings` - (Optional) Settings that define the conditions for declaring the input unhealthy and failing over to a different input. See [`automatic_input_failover_settings` Block](#automatic_input_failover_settings-block) for details.
* `input_attachment_name` - (Required) User-specified name for the attachment.
* `input_id` - (Required) ID of the input.
* `input_settings` - (Optional) Settings of an input. See [`input_settings` Block](#input_settings-block) for details.

### `automatic_input_failover_settings` Block

* `error_clear_time_msec` - (Optional) Amount of time, in milliseconds, that a recovered input must have no failover conditions before it is considered healthy. This value is particularly important when `input_preference` is set to `PRIMARY_INPUT_PREFERRED`, because after this time, MediaLive switches back to the primary input.
* `failover_condition` - (Optional) Failover conditions. If any of these conditions occur, MediaLive fails over to the other input. See [`failover_condition` Block](#failover_condition-block) for details.
* `input_preference` - (Optional) Input preference when deciding which input to make active when a previously failed input has recovered.
* `secondary_input_id` - (Required) Input ID of the secondary input in the automatic input failover pair.

### `failover_condition` Block

* `failover_condition_settings` - (Optional) Failover condition type-specific settings. See [`failover_condition_settings` Block](#failover_condition_settings-block) for details.

### `failover_condition_settings` Block

* `audio_silence_settings` - (Optional) MediaLive fails over if the specified audio selector is silent for the specified period. See [`audio_silence_settings` Block](#audio_silence_settings-block) for details.
* `input_loss_settings` - (Optional) MediaLive fails over if content is not detected in this input for the specified period. See [`input_loss_settings` Block](#input_loss_settings-block) for details.
* `video_black_settings` - (Optional) MediaLive fails over if content is considered black for the specified period. See [`video_black_settings` Block](#video_black_settings-block) for details.

### `audio_silence_settings` Block

* `audio_selector_name` - (Required) Name of the audio selector in the input that MediaLive monitors to detect silence. Select your most important rendition. Leave blank if you did not create an audio selector in this input.
* `audio_silence_threshold_msec` - (Optional) Amount of time, in milliseconds, that the active input must be silent before automatic input failover occurs. Silence is defined as audio loss or audio quieter than -50 dBFS.

### `input_loss_settings` Block

* `input_loss_threshold_msec` - (Optional) Amount of time, in milliseconds, that no input is detected before an input failover occurs.

### `video_black_settings` Block

* `black_detect_threshold` - (Optional) Threshold below which MediaLive considers a pixel to be black, expressed as a decimal fraction of white. For example, `0.1` means 10% white (90% black). For the input to be considered black, every pixel in a frame must be below this threshold. The range is 0.0 to 1.0, with any number of decimal places.
* `video_black_threshold_msec` - (Optional) Amount of time, in milliseconds, that the active input must be black before automatic input failover occurs.

### `input_settings` Block

* `audio_selector` - (Optional) Select the audio stream to decode for inputs that have multiple. See [`audio_selector` Block](#audio_selector-block) for details.
* `caption_selector` - (Optional) Select the caption input to use for inputs that have multiple available. See [`caption_selector` Block](#caption_selector-block) for details.
* `deblock_filter` - (Optional) Whether to enable the deblock filter when filtering.
* `denoise_filter` - (Optional) Whether to enable the denoise filter when filtering.
* `filter_strength` - (Optional) Magnitude of filtering from 1 (minimal) to 5 (strongest).
* `input_filter` - (Optional) Turns on the filter for the input.
* `network_input_settings` - (Optional) Input settings. See [`network_input_settings` Block](#network_input_settings-block) for details.
* `scte35_pid` - (Optional) PID from which to read SCTE-35 messages.
* `smpte2038_data_preference` - (Optional) Whether to extract applicable ancillary data from a SMPTE-2038 source in the input.
* `source_end_behavior` - (Optional) Loop input if it is a file.
* `video_selector` - (Optional) Information about one video to extract from the input. See [`video_selector` Block](#video_selector-block) for details.

### `audio_selector` Block

* `name` - (Required) Name of the audio selector.
* `selector_settings` - (Optional) Audio selector settings. See [`audio_selector.selector_settings` Block](#audio_selectorselector_settings-block) for details.

### `audio_selector.selector_settings` Block

* `audio_hls_rendition_selection` - (Optional) HLS rendition selection. See [`audio_hls_rendition_selection` Block](#audio_hls_rendition_selection-block) for details.
* `audio_language_selection` - (Optional) Language selection. See [`audio_language_selection` Block](#audio_language_selection-block) for details.
* `audio_pid_selection` - (Optional) PID selection. See [`audio_pid_selection` Block](#audio_pid_selection-block) for details.
* `audio_track_selection` - (Optional) Track selection. See [`audio_track_selection` Block](#audio_track_selection-block) for details.

### `audio_hls_rendition_selection` Block

* `group_id` - (Required) GROUP-ID in the `#EXT-X-MEDIA` tag of the target HLS audio rendition.
* `name` - (Required) NAME in the `#EXT-X-MEDIA` tag of the target HLS audio rendition.

### `audio_language_selection` Block

* `language_code` - (Required) Three-letter language code from within an audio source.
* `language_selection_policy` - (Optional) How to handle a change in the language of the audio stream after it is selected. When set to `strict`, the transport stream demux strictly identifies audio streams by their language descriptor. If a PMT update occurs such that an audio stream matching the initially selected language is no longer present, then mute is encoded until the language returns. When set to `loose`, the demux chooses another audio stream in the program with the same stream type on a PMT update if it cannot find one with the same language.

### `audio_pid_selection` Block

* `pid` - (Required) PID from within a source.

### `audio_track_selection` Block

* `dolby_e_decode` - (Optional) Decoding options for Dolby E streams. These should be Dolby E frames carried in PCM streams tagged with SMPTE-337. See [`dolby_e_decode` Block](#dolby_e_decode-block) for details.
* `tracks` - (Required) One or more unique audio tracks from within a source. See [`tracks` Block](#tracks-block) for details.

### `dolby_e_decode` Block

* `program_selection` - (Required) Program ID of the Dolby E program to extract from the specified track. Applies only to Dolby E. One program is extracted per audio selector. To select multiple programs, create multiple selectors with the same track and different program numbers. `All channels` ignores the program IDs and includes all the channels in this selector; useful if metadata is known to be incorrect.

### `tracks` Block

* `track` - (Required) 1-based integer value that maps to a specific audio track.

### `caption_selector` Block

* `language_code` - (Optional) When specified, indicates the three-letter language code of the caption track to extract from the source.
* `name` - (Required) Name of the caption selector.
* `selector_settings` - (Optional) Caption selector settings. See [`caption_selector.selector_settings` Block](#caption_selectorselector_settings-block) for details.

### `caption_selector.selector_settings` Block

* `ancillary_source_settings` - (Optional) Ancillary source settings. See [`ancillary_source_settings` Block](#ancillary_source_settings-block) for details.
* `arib_source_settings` - (Optional) ARIB source settings. Can be set as an empty block.
* `dvb_sub_source_settings` - (Optional) DVB Sub source settings. See [`dvb_sub_source_settings` Block](#dvb_sub_source_settings-block) for details.
* `embedded_source_settings` - (Optional) Embedded source settings. See [`embedded_source_settings` Block](#embedded_source_settings-block) for details.
* `scte20_source_settings` - (Optional) SCTE-20 source settings. See [`scte20_source_settings` Block](#scte20_source_settings-block) for details.
* `scte27_source_settings` - (Optional) SCTE-27 source settings. See [`scte27_source_settings` Block](#scte27_source_settings-block) for details.
* `teletext_source_settings` - (Optional) Teletext source settings. See [`teletext_source_settings` Block](#teletext_source_settings-block) for details.

### `ancillary_source_settings` Block

* `source_ancillary_channel_number` - (Optional) Captions channel, from 1 to 4, to extract from the ancillary captions. Complete this field if you plan to convert the ancillary captions to another format. Leave blank if you plan to choose Embedded as the captions destination in the output, because MediaLive ignores the field.

### `dvb_sub_source_settings` Block

* `ocr_language` - (Optional) Language to consider when translating the image-based source to text, for a referencing WebVTT caption description.
* `pid` - (Optional) PID for the source content when using DVB-Sub with Burn-In or SMPTE-TT. Unused for DVB-Sub passthrough. All DVB-Sub content is passed through, regardless of selectors.

### `embedded_source_settings` Block

* `convert_608_to_708` - (Optional) When set to `upconvert`, 608 data is both passed through via the "608 compatibility bytes" fields of the 708 wrapper and translated into 708. 708 data present in the source content is discarded.
* `scte20_detection` - (Optional) Set to `auto` to handle streams with intermittent or non-aligned SCTE-20 and Embedded captions.
* `source_608_channel_number` - (Optional) 608/708 channel number within the video track from which to extract captions. Unused for passthrough.

### `scte20_source_settings` Block

* `convert_608_to_708` - (Optional) When set to `upconvert`, 608 data is both passed through via the "608 compatibility bytes" fields of the 708 wrapper and translated into 708. 708 data present in the source content is discarded.
* `source_608_channel_number` - (Optional) 608/708 channel number within the video track from which to extract captions. Unused for passthrough.

### `scte27_source_settings` Block

* `ocr_language` - (Optional) Language to consider when translating the image-based source to text, for a referencing WebVTT caption description.
* `pid` - (Optional) PID field used with the caption selector `language_code` field as follows. Specify PID and language: extracts captions from that PID; the language is informational. Specify PID and omit language: extracts the specified PID. Omit PID and specify language: extracts the specified language, whichever PID that happens to be. Omit PID and omit language: valid only if the source is DVB-Sub that is being passed through; all languages are passed through.

### `teletext_source_settings` Block

* `output_rectangle` - (Optional) Region where TTML style captions are displayed. See [`output_rectangle` Block](#output_rectangle-block) for details.
* `page_number` - (Optional) Teletext page number within the data stream from which to extract captions. Range of 0x100 (256) to 0x8FF (2303), specified as a hexadecimal string with no `0x` prefix. Unused for passthrough.

### `output_rectangle` Block

* `height` - (Required) Height of the rectangle as a percentage of the underlying frame height. The `top_offset` and rectangle height must add up to 100% or less. This field corresponds to `tts:extent` - Y in the TTML standard.
* `left_offset` - (Required) Position of the left edge of the rectangle as a percentage of the underlying frame width, relative to the left edge of the frame. Applies only if you plan to convert these source captions to EBU-TT-D or TTML in an output. You can define a display rectangle smaller than the underlying video frame; if you specify a value for one of the offset or dimension fields, you must specify a value for all of them. This field corresponds to `tts:origin` - X in the TTML standard.
* `top_offset` - (Required) Position of the top edge of the rectangle as a percentage of the underlying frame height, relative to the top edge of the frame. This field corresponds to `tts:origin` - Y in the TTML standard.
* `width` - (Required) Width of the rectangle as a percentage of the underlying frame width. The `left_offset` and rectangle width must add up to 100% or less. This field corresponds to `tts:extent` - X in the TTML standard.

### `network_input_settings` Block

* `hls_input_settings` - (Optional) HLS input settings when the URI is for an HLS manifest. See [`hls_input_settings` Block](#hls_input_settings-block) for details.
* `server_validation` - (Optional) Check HTTPS server certificates.

### `hls_input_settings` Block

* `bandwidth` - (Optional) Bitrate, in bits per second, as in an HLS manifest.
* `buffer_segments` - (Optional) Buffer segments.
* `retries` - (Optional) Number of consecutive times that attempts to read a manifest or segment must fail before the input is considered unavailable.
* `retry_interval` - (Optional) Number of seconds between retries when an attempt to read a manifest or segment fails.
* `scte35_source` - (Optional) Source for the SCTE-35 messages that MediaLive ingests.

### `video_selector` Block

* `color_space` - (Optional) Color space of the input.
* `color_space_usage` - (Optional) How MediaLive applies the color space metadata from the input.

### `maintenance` Block

* `maintenance_day` - (Required) Day of the week to use for maintenance.
* `maintenance_start_time` - (Required) Hour maintenance will start.

### `vpc` Block

* `public_address_allocation_ids` - (Required) List of public address allocation IDs to associate with ENIs created in the output VPC. Specify one for `SINGLE_PIPELINE`, two for `STANDARD` channels.
* `security_group_ids` - (Optional) List of up to 5 EC2 VPC security group IDs to attach to the output VPC network interfaces. The VPC default security group is used if none are specified.
* `subnet_ids` - (Required) List of VPC subnet IDs from the same VPC. For a `STANDARD` channel, subnet IDs must map to two unique availability zones.

### `encoder_settings` Block

* `audio_descriptions` - (Optional) Audio descriptions for the channel. See [`audio_descriptions` Block](#audio_descriptions-block) for details.
* `avail_blanking` - (Optional) Settings for ad avail blanking. See [`avail_blanking` Block](#avail_blanking-block) for details.
* `caption_descriptions` - (Optional) Caption descriptions for the channel. See [`caption_descriptions` Block](#caption_descriptions-block) for details.
* `global_configuration` - (Optional) Configuration settings that apply to the event as a whole. See [`global_configuration` Block](#global_configuration-block) for details.
* `motion_graphics_configuration` - (Optional) Settings for motion graphics. See [`motion_graphics_configuration` Block](#motion_graphics_configuration-block) for details.
* `nielsen_configuration` - (Optional) Nielsen configuration settings. See [`nielsen_configuration` Block](#nielsen_configuration-block) for details.
* `output_groups` - (Required) Output groups for the channel. See [`output_groups` Block](#output_groups-block) for details.
* `timecode_config` - (Required) Settings used to acquire and adjust timecode information from inputs. See [`timecode_config` Block](#timecode_config-block) for details.
* `video_descriptions` - (Optional) Video descriptions for the channel. See [`video_descriptions` Block](#video_descriptions-block) for details.

### `audio_descriptions` Block

* `audio_normalization_settings` - (Optional) Advanced audio normalization settings. See [`audio_normalization_settings` Block](#audio_normalization_settings-block) for details.
* `audio_selector_name` - (Required) Name of the audio selector used as the source for this audio description.
* `audio_type` - (Optional) Applies only if `audio_type_control` is `useConfigured`. Values for `audio_type` are defined in ISO-IEC 13818-1.
* `audio_type_control` - (Optional) How audio type is determined.
* `audio_watermark_settings` - (Optional) Settings to configure one or more solutions that insert audio watermarks in the audio encode. See [`audio_watermark_settings` Block](#audio_watermark_settings-block) for details.
* `codec_settings` - (Optional) Audio codec settings. See [`audio_descriptions.codec_settings` Block](#audio_descriptionscodec_settings-block) for details.
* `language_code` - (Optional) ISO 639-2 three-digit code for the language of this audio description.
* `language_code_control` - (Optional) How to handle the audio language code.
* `name` - (Required) Name of this audio description.
* `remix_settings` - (Optional) Settings to remix the audio. See [`remix_settings` Block](#remix_settings-block) for details.
* `stream_name` - (Optional) Stream name used to identify the audio description.

### `audio_normalization_settings` Block

* `algorithm` - (Optional) Audio normalization algorithm to use. `itu17701` conforms to the CALM Act specification; `itu17702` conforms to the EBU R-128 specification.
* `algorithm_control` - (Optional) Algorithm control for the audio description.
* `target_lkfs` - (Optional) Target LKFS (loudness) to adjust volume to.

### `audio_watermark_settings` Block

* `nielsen_watermarks_settings` - (Optional) Settings to configure Nielsen watermarks in the audio encode. See [`nielsen_watermarks_settings` Block](#nielsen_watermarks_settings-block) for details.

### `nielsen_watermarks_settings` Block

* `nielsen_cbet_settings` - (Optional) Settings to insert watermarks of type Nielsen CBET. See [`nielsen_cbet_settings` Block](#nielsen_cbet_settings-block) for details.
* `nielsen_distribution_type` - (Optional) Distribution types to assign to the watermarks. Options are `PROGRAM_CONTENT` and `FINAL_DISTRIBUTOR`.
* `nielsen_naes_ii_nw_settings` - (Optional) Settings to insert watermarks of type Nielsen NAES, II (N2) and Nielsen NAES VI (NW). See [`nielsen_naes_ii_nw_settings` Block](#nielsen_naes_ii_nw_settings-block) for details.

### `nielsen_cbet_settings` Block

* `cbet_check_digit_string` - (Required) CBET check digits to use for the watermark.
* `cbet_stepaside` - (Required) Method of CBET insertion mode when prior encoding is detected on the same layer.
* `csid` - (Required) CBET source ID to use in the watermark.

### `nielsen_naes_ii_nw_settings` Block

* `check_digit_string` - (Required) Check digit string for the watermark.
* `sid` - (Required) Nielsen source ID to include in the watermark.

### `audio_descriptions.codec_settings` Block

* `aac_settings` - (Optional) AAC settings. See [`aac_settings` Block](#aac_settings-block) for details.
* `ac3_settings` - (Optional) AC3 settings. See [`ac3_settings` Block](#ac3_settings-block) for details.
* `eac3_atmos_settings` - (Optional) EAC3 Atmos settings. See [`eac3_atmos_settings` Block](#eac3_atmos_settings-block) for details.
* `eac3_settings` - (Optional) EAC3 settings. See [`eac3_settings` Block](#eac3_settings-block) for details.
* `mp2_settings` - (Optional) MP2 settings. See [`mp2_settings` Block](#mp2_settings-block) for details.
* `pass_through_settings` - (Optional) Pass-through settings. Can be set as an empty block.
* `wav_settings` - (Optional) WAV settings. See [`wav_settings` Block](#wav_settings-block) for details.

### `aac_settings` Block

* `bitrate` - (Optional) Average bitrate in bits/second.
* `coding_mode` - (Optional) Mono, stereo, or 5.1 channel layout.
* `input_type` - (Optional) Set to `broadcasterMixedAd` when input contains pre-mixed main audio and AD (narration) as a stereo pair.
* `profile` - (Optional) AAC profile.
* `rate_control_mode` - (Optional) Rate control mode.
* `raw_format` - (Optional) Sets LATM/LOAS AAC output for raw containers.
* `sample_rate` - (Optional) Sample rate in Hz.
* `spec` - (Optional) Whether to use MPEG-2 AAC audio instead of MPEG-4 AAC audio for raw or MPEG-2 Transport Stream containers.
* `vbr_quality` - (Optional) VBR quality level. Only used if `rate_control_mode` is `VBR`.

### `ac3_settings` Block

* `bitrate` - (Optional) Average bitrate in bits/second.
* `bitstream_mode` - (Optional) Bitstream mode (bsmod) for the emitted AC-3 stream.
* `coding_mode` - (Optional) Dolby Digital coding mode.
* `dialnorm` - (Optional) Sets the dialnorm of the output.
* `drc_profile` - (Optional) When set to `filmStandard`, adds dynamic range compression signaling to the output bitstream as defined in the Dolby Digital specification.
* `lfe_filter` - (Optional) When set to `enabled`, applies a 120Hz lowpass filter to the LFE channel prior to encoding.
* `metadata_control` - (Optional) Metadata control.

### `eac3_atmos_settings` Block

* `bitrate` - (Optional) Average bitrate in bits/second.
* `coding_mode` - (Optional) Dolby Digital Plus with Dolby Atmos coding mode.
* `dialnorm` - (Optional) Sets the dialnorm for the output.
* `drc_line` - (Optional) Dolby dynamic range compression profile.
* `drc_rf` - (Optional) Profile for heavy Dolby dynamic range compression.
* `height_trim` - (Optional) Height dimensional trim.
* `surround_trim` - (Optional) Surround dimensional trim.

### `eac3_settings` Block

* `attenuation_control` - (Optional) Attenuation control.
* `bitrate` - (Optional) Average bitrate in bits/second.
* `bitstream_mode` - (Optional) Bitstream mode (bsmod) for the emitted AC-3 stream.
* `coding_mode` - (Optional) Dolby Digital Plus coding mode.
* `dc_filter` - (Optional) Whether to activate a DC filter for the channels.
* `dialnorm` - (Optional) Sets the dialnorm for the output.
* `drc_line` - (Optional) Sets the Dolby dynamic range compression profile.
* `drc_rf` - (Optional) Profile for heavy Dolby dynamic range compression.
* `lfe_control` - (Optional) When set to `lfe`, the EAC3 output contains an LFE channel.
* `lfe_filter` - (Optional) When set to `enabled`, applies a 120Hz lowpass filter to the LFE channel prior to encoding.
* `lo_ro_center_mix_level` - (Optional) Left total/right total center mix level.
* `lo_ro_surround_mix_level` - (Optional) Left total/right total surround mix level.
* `lt_rt_center_mix_level` - (Optional) Left only/right only center mix level.
* `lt_rt_surround_mix_level` - (Optional) Left only/right only surround mix level.
* `metadata_control` - (Optional) Metadata control.
* `passthrough_control` - (Optional) When set to `whenPossible`, input DD+ audio is passed through if present.
* `phase_control` - (Optional) Whether to invert the phase of the surround channels.
* `stereo_downmix` - (Optional) Stereo downmix preference.
* `surround_ex_mode` - (Optional) Whether to encode the stream as Dolby Digital EX.
* `surround_mode` - (Optional) Whether to encode the stream as a surround sound stream.

### `mp2_settings` Block

* `bitrate` - (Optional) Average bitrate in bits/second.
* `coding_mode` - (Optional) MPEG-2 coding mode.
* `sample_rate` - (Optional) Sample rate in Hz.

### `wav_settings` Block

* `bit_depth` - (Optional) Bits per sample.
* `coding_mode` - (Optional) WAV coding mode.
* `sample_rate` - (Optional) Sample rate in Hz.

### `remix_settings` Block

* `channel_mappings` - (Required) Mapping of input channels to output channels. See [`channel_mappings` Block](#channel_mappings-block) for details.
* `channels_in` - (Optional) Number of input channels to be used.
* `channels_out` - (Optional) Number of output channels to be produced.

### `channel_mappings` Block

* `input_channel_levels` - (Required) Levels for each of the input channels. See [`input_channel_levels` Block](#input_channel_levels-block) for details.
* `output_channel` - (Required) Index of the output channel being produced.

### `input_channel_levels` Block

* `gain` - (Required) Remixing value, from -60 to 6 decibels.
* `input_channel` - (Required) Index of the input channel used as a source.

### `output_groups` Block

* `name` - (Optional) Custom output group name defined by the user.
* `output_group_settings` - (Required) Settings associated with the output group. See [`output_group_settings` Block](#output_group_settings-block) for details.
* `outputs` - (Required) List of outputs. See [`outputs` Block](#outputs-block) for details.

### `output_group_settings` Block

* `archive_group_settings` - (Optional) Archive group settings. See [`archive_group_settings` Block](#archive_group_settings-block) for details.
* `frame_capture_group_settings` - (Optional) Frame capture group settings. See [`frame_capture_group_settings` Block](#frame_capture_group_settings-block) for details.
* `hls_group_settings` - (Optional) HLS group settings. See [`hls_group_settings` Block](#hls_group_settings-block) for details.
* `media_package_group_settings` - (Optional) Media package group settings. See [`media_package_group_settings` Block](#media_package_group_settings-block) for details.
* `ms_smooth_group_settings` - (Optional) Microsoft Smooth group settings. See [`ms_smooth_group_settings` Block](#ms_smooth_group_settings-block) for details.
* `multiplex_group_settings` - (Optional) Multiplex group settings. Can be set as an empty block.
* `rtmp_group_settings` - (Optional) RTMP group settings. See [`rtmp_group_settings` Block](#rtmp_group_settings-block) for details.
* `udp_group_settings` - (Optional) UDP group settings. See [`udp_group_settings` Block](#udp_group_settings-block) for details.

### `archive_group_settings` Block

* `archive_cdn_settings` - (Optional) Parameters that control the interactions with the CDN. See [`archive_cdn_settings` Block](#archive_cdn_settings-block) for details.
* `destination` - (Required) Directory and base filename where archive files are written. See [`destination` Block](#destination-block) for details.
* `rollover_interval` - (Optional) Number of seconds to write to an archive file before closing and starting a new one.

### `archive_cdn_settings` Block

* `archive_s3_settings` - (Optional) Archive S3 settings. See [`archive_s3_settings` Block](#archive_s3_settings-block) for details.

### `archive_s3_settings` Block

* `canned_acl` - (Optional) Canned ACL to apply to each S3 request.

### `frame_capture_group_settings` Block

* `destination` - (Required) Parameters that identify where to write the frame captures. See [`destination` Block](#destination-block) for details.
* `frame_capture_cdn_settings` - (Optional) Parameters that control the interactions with the CDN. See [`frame_capture_cdn_settings` Block](#frame_capture_cdn_settings-block) for details.

### `frame_capture_cdn_settings` Block

* `frame_capture_s3_settings` - (Optional) Frame capture S3 settings. See [`frame_capture_s3_settings` Block](#frame_capture_s3_settings-block) for details.

### `frame_capture_s3_settings` Block

* `canned_acl` - (Optional) Canned ACL to apply to each S3 request.

### `hls_group_settings` Block

* `ad_markers` - (Optional) List of markers used to insert ad avail markers in the HLS output.
* `base_url_content` - (Optional) Prepends a value to the URI of each media (segment) file in the manifest.
* `base_url_content1` - (Optional) Alternate base URL for content, used for redundant manifest configurations.
* `base_url_manifest` - (Optional) Prepends a value to the URI of each variant manifest file in the master manifest.
* `base_url_manifest1` - (Optional) Alternate base URL for manifests, used for redundant manifest configurations.
* `caption_language_mappings` - (Optional) Mapping of up to 4 caption channels to caption languages. See [`caption_language_mappings` Block](#caption_language_mappings-block) for details.
* `caption_language_setting` - (Optional) How audio and caption language setting tags are written in the manifest.
* `client_cache` - (Optional) When set to `disabled`, sets the `#EXT-X-ALLOW-CACHE:no` tag in the manifest.
* `codec_specification` - (Optional) Specification to use for the codec in the manifest.
* `constant_iv` - (Optional) 128-bit, 16-byte hex value represented by a 32-character text string, used with the key for encrypting blocks.
* `destination` - (Required) Place where the segments and manifests are written. See [`destination` Block](#destination-block) for details.
* `directory_structure` - (Optional) Place for the media and manifest files.
* `discontinuity_tags` - (Optional) How the `EXT-X-DISCONTINUITY` tags are inserted.
* `encryption_type` - (Optional) Encryption type.
* `hls_cdn_settings` - (Optional) Parameters that control the interactions with the CDN. See [`hls_cdn_settings` Block](#hls_cdn_settings-block) for details.
* `hls_id3_segment_tagging` - (Optional) State of HLS ID3 segment tagging.
* `iframe_only_playlists` - (Optional) When set to `standard`, an I-Frame only playlist is written out for each video output in the output group.
* `incomplete_segment_behavior` - (Optional) Behavior of the segment that is in progress at a channel restart or input switch.
* `index_n_segments` - (Optional) Number of segments to keep in the manifest.
* `input_loss_action` - (Optional) Parameter that controls output group behavior on an input loss.
* `iv_in_manifest` - (Optional) Whether the initialization vector is included in the key method on the segment.
* `iv_source` - (Optional) Whether the initialization vector follows the segment number or is from the group.
* `keep_segments` - (Optional) Number of segments to retain in the destination directory.
* `key_format` - (Optional) Value for the key URI format in the key method.
* `key_format_versions` - (Optional) Versions of the key format in the key method.
* `key_provider_settings` - (Optional) Key provider settings. See [`key_provider_settings` Block](#key_provider_settings-block) for details.
* `manifest_compression` - (Optional) Whether the manifest is compressed with gzip.
* `manifest_duration_format` - (Optional) Format of the segment duration in the manifest.
* `min_segment_length` - (Optional) Minimum length, in seconds, for each segment.
* `mode` - (Optional) HLS mode.
* `output_selection` - (Optional) Generation of the manifest and segments or only variant manifests.
* `program_date_time` - (Optional) Whether to include an `EXT-X-PROGRAM-DATE-TIME` tag.
* `program_date_time_clock` - (Optional) Clock source for the `EXT-X-PROGRAM-DATE-TIME` tag.
* `program_date_time_period` - (Optional) Period, in seconds, between `EXT-X-PROGRAM-DATE-TIME` tags.
* `redundant_manifest` - (Optional) When set to `enabled`, writes a secondary manifest for each output group.
* `segment_length` - (Optional) Length, in seconds, of each segment.
* `segments_per_subdirectory` - (Optional) Number of segments to write to a subdirectory before starting a new one.
* `stream_inf_resolution` - (Optional) Whether to include the resolution in the `EXT-X-STREAM-INF` tag.
* `timed_metadata_id3_frame` - (Optional) ID3 frame used for timed metadata.
* `timed_metadata_id3_period` - (Optional) Period, in seconds, between ID3 frames for timed metadata.
* `timestamp_delta_milliseconds` - (Optional) Value, in milliseconds, to offset the program date time from the input timestamp.
* `ts_file_mode` - (Optional) Whether segments are placed in subdirectories.

### `caption_language_mappings` Block

* `caption_channel` - (Required) Caption channel to install in the output.
* `language_code` - (Required) Three-character ISO 639-2 language code.
* `language_description` - (Required) Textual description of the language.

### `hls_cdn_settings` Block

* `hls_akamai_settings` - (Optional) Akamai settings. See [`hls_akamai_settings` Block](#hls_akamai_settings-block) for details.
* `hls_basic_put_settings` - (Optional) Basic put settings. See [`hls_basic_put_settings` Block](#hls_basic_put_settings-block) for details.
* `hls_media_store_settings` - (Optional) Media store settings. See [`hls_media_store_settings` Block](#hls_media_store_settings-block) for details.
* `hls_s3_settings` - (Optional) S3 settings. See [`hls_s3_settings` Block](#hls_s3_settings-block) for details.
* `hls_webdav_settings` - (Optional) WebDAV settings. See [`hls_webdav_settings` Block](#hls_webdav_settings-block) for details.

### `hls_akamai_settings` Block

* `connection_retry_interval` - (Optional) Number of seconds to wait before retrying a connection to the CDN if the connection is lost.
* `filecache_duration` - (Optional) Size, in seconds, of the file cache for streaming outputs.
* `http_transfer_mode` - (Optional) HTTP transfer mode.
* `num_retries` - (Optional) Number of retry attempts that are made before the channel is put into an error state.
* `restart_delay` - (Optional) Number of seconds to wait until a restart is initiated.
* `salt` - (Optional) Salt for authenticated Akamai.
* `token` - (Optional) Token parameter for authenticated Akamai.

### `hls_basic_put_settings` Block

* `connection_retry_interval` - (Optional) Number of seconds to wait before retrying a connection to the CDN if the connection is lost.
* `filecache_duration` - (Optional) Size, in seconds, of the file cache for streaming outputs.
* `num_retries` - (Optional) Number of retry attempts that are made before the channel is put into an error state.
* `restart_delay` - (Optional) Number of seconds to wait until a restart is initiated.

### `hls_media_store_settings` Block

* `connection_retry_interval` - (Optional) Number of seconds to wait before retrying a connection to the CDN if the connection is lost.
* `filecache_duration` - (Optional) Size, in seconds, of the file cache for streaming outputs.
* `media_store_storage_class` - (Optional) MediaStore storage class to use.
* `num_retries` - (Optional) Number of retry attempts that are made before the channel is put into an error state.
* `restart_delay` - (Optional) Number of seconds to wait until a restart is initiated.

### `hls_s3_settings` Block

* `canned_acl` - (Optional) Canned ACL to apply to each S3 request.

### `hls_webdav_settings` Block

* `connection_retry_interval` - (Optional) Number of seconds to wait before retrying a connection to the CDN if the connection is lost.
* `filecache_duration` - (Optional) Size, in seconds, of the file cache for streaming outputs.
* `http_transfer_mode` - (Optional) HTTP transfer mode.
* `num_retries` - (Optional) Number of retry attempts that are made before the channel is put into an error state.
* `restart_delay` - (Optional) Number of seconds to wait until a restart is initiated.

### `key_provider_settings` Block

* `static_key_settings` - (Optional) Static key settings. See [`static_key_settings` Block](#static_key_settings-block) for details.

### `static_key_settings` Block

* `key_provider_server` - (Optional) Location of the license server used for protecting content. See [`key_provider_server` Block](#key_provider_server-block) for details.
* `static_key_value` - (Required) Static key value as a 32-character hexadecimal string.

### `key_provider_server` Block

* `password_param` - (Optional) Key used to extract the password from EC2 Parameter store.
* `uri` - (Required) Path to a file accessible to the live stream.
* `username` - (Optional) Username for the server.

### `media_package_group_settings` Block

* `destination` - (Required) MediaPackage channel destination. See [`destination` Block](#destination-block) for details.

### `ms_smooth_group_settings` Block

* `acquisition_point_id` - (Optional) Identifier used to associate the segments between the output and the ingest server.
* `audio_only_timecode_control` - (Optional) Whether to put a timecode in the single video track of an audio-only output.
* `certificate_mode` - (Optional) When set to `selfSigned`, accepts self-signed certificates from the destination server.
* `connection_retry_interval` - (Optional) Number of seconds to wait before retrying a connection to the server if the connection is lost.
* `destination` - (Required) Smooth Streaming publish point on an IIS server. See [`destination` Block](#destination-block) for details.
* `event_id` - (Optional) Microsoft Smooth channel ID that is sent to the IIS server.
* `event_id_mode` - (Optional) How the event ID is generated.
* `event_stop_behavior` - (Optional) When set to `sendEos`, sends an end-of-stream to the IIS server when a stream stops.
* `filecache_duration` - (Optional) Size, in seconds, of the file cache for streaming outputs.
* `fragment_length` - (Optional) Length, in seconds, of mp4 fragments to generate.
* `input_loss_action` - (Optional) Parameter that controls output group behavior on an input loss.
* `num_retries` - (Optional) Number of retry attempts.
* `restart_delay` - (Optional) Number of seconds to wait until a restart is initiated.
* `segmentation_mode` - (Optional) When set to `useInputSegmentation`, segmentation markers from the input are used.
* `send_delay_ms` - (Optional) Number of milliseconds to delay the output from the second pipeline.
* `sparse_track_type` - (Optional) Type of the sparse track that carries timed metadata.
* `stream_manifest_behavior` - (Optional) When set to `send`, sends a stream manifest so the publishing point does not start until the first segment is sent.
* `timestamp_offset` - (Optional) Timestamp offset for the event.
* `timestamp_offset_mode` - (Optional) Type of timestamp date offset to use.

### `rtmp_group_settings` Block

* `ad_markers` - (Optional) List of ad markers for this output group.
* `authentication_scheme` - (Optional) Authentication scheme to use when connecting with the CDN.
* `cache_full_behavior` - (Optional) Behavior when the content cache fills up.
* `cache_length` - (Optional) Cache length, in seconds, used to calculate buffer size.
* `caption_data` - (Optional) Controls the types of data that pass to `onCaptionInfo` outputs.
* `input_loss_action` - (Optional) Behavior of the RTMP group if the input becomes unavailable.
* `restart_delay` - (Optional) Number of seconds to wait until a restart is initiated.

### `udp_group_settings` Block

* `input_loss_action` - (Optional) Behavior of last resort when input video is lost.
* `timed_metadata_id3_frame` - (Optional) ID3 frame that has the timecode.
* `timed_metadata_id3_period` - (Optional) Timed metadata interval, in seconds.

### `outputs` Block

* `audio_description_names` - (Optional) Names of the audio descriptions used as audio sources for the output.
* `caption_description_names` - (Optional) Names of the caption descriptions used as caption sources for the output.
* `output_name` - (Optional) Name used to identify an output.
* `output_settings` - (Required) Settings for the output. See [`output_settings` Block](#output_settings-block) for details.
* `video_description_name` - (Optional) Name of the video description used as the video source for the output.

### `output_settings` Block

* `archive_output_settings` - (Optional) Archive output settings. See [`archive_output_settings` Block](#archive_output_settings-block) for details.
* `frame_capture_output_settings` - (Optional) Frame capture output settings. See [`frame_capture_output_settings` Block](#frame_capture_output_settings-block) for details.
* `hls_output_settings` - (Optional) HLS output settings. See [`hls_output_settings` Block](#hls_output_settings-block) for details.
* `media_package_output_settings` - (Optional) Media package output settings. Can be set as an empty block.
* `ms_smooth_output_settings` - (Optional) Microsoft Smooth output settings. See [`ms_smooth_output_settings` Block](#ms_smooth_output_settings-block) for details.
* `multiplex_output_settings` - (Optional) Multiplex output settings. See [`multiplex_output_settings` Block](#multiplex_output_settings-block) for details.
* `rtmp_output_settings` - (Optional) RTMP output settings. See [`rtmp_output_settings` Block](#rtmp_output_settings-block) for details.
* `udp_output_settings` - (Optional) UDP output settings. See [`udp_output_settings` Block](#udp_output_settings-block) for details.

### `archive_output_settings` Block

* `container_settings` - (Required) Settings specific to the container type of the file. See [`archive_output_settings.container_settings` Block](#archive_output_settingscontainer_settings-block) for details.
* `extension` - (Optional) Output file extension.
* `name_modifier` - (Optional) String concatenated to the end of the destination filename. Required for multiple outputs of the same type.

### `archive_output_settings.container_settings` Block

* `m2ts_settings` - (Optional) M2TS settings. See [`m2ts_settings` Block](#m2ts_settings-block) for details.
* `raw_settings` - (Optional) Raw settings. Can be set as an empty block.

### `m2ts_settings` Block

* `absent_input_audio_behavior` - (Optional) When set to `drop`, output audio streams are removed from the program if the selected input audio stream is removed from the input.
* `arib` - (Optional) When set to `enabled`, uses ARIB-compliant captioning.
* `arib_captions_pid` - (Optional) Packet identifier (PID) for ARIB captions in the transport stream.
* `arib_captions_pid_control` - (Optional) How to pass along the ARIB captions.
* `audio_buffer_model` - (Optional) When set to `dvb`, uses DVB buffer model for Dolby Digital audio.
* `audio_frames_per_pes` - (Optional) Number of audio frames to insert for each PES packet.
* `audio_pids` - (Optional) Packet identifier (PID) of the elementary audio streams in the transport stream.
* `audio_stream_type` - (Optional) When set to `atsc`, uses stream type `0x81` for AC3 and stream type `0x87` for EAC3.
* `bitrate` - (Optional) Output bitrate, in bits per second, of the transport stream.
* `buffer_model` - (Optional) Controls the timing accuracy for output network traffic.
* `cc_descriptor` - (Optional) When set to `enabled`, generates a captions service descriptor in the PMT.
* `dvb_nit_settings` - (Optional) DVB Network Information Table (NIT) settings. See [`dvb_nit_settings` Block](#dvb_nit_settings-block) for details.
* `dvb_sdt_settings` - (Optional) DVB Service Description Table (SDT) settings. See [`dvb_sdt_settings` Block](#dvb_sdt_settings-block) for details.
* `dvb_sub_pids` - (Optional) Packet identifier (PID) for input source DVB Subtitle data to this output.
* `dvb_tdt_settings` - (Optional) DVB Time and Date Table (TDT) settings. See [`dvb_tdt_settings` Block](#dvb_tdt_settings-block) for details.
* `dvb_teletext_pid` - (Optional) Packet identifier (PID) for input source DVB Teletext data to this output.
* `ebif` - (Optional) When set to `passthrough`, passes through EBIF messages from the input source to this output.
* `ebp_audio_interval` - (Optional) When set to `videoAndFixedIntervals`, adds Encoder Boundary Points for each elementary audio stream.
* `ebp_lookahead_ms` - (Optional) Amount of time, in milliseconds, that the encoder looks ahead for Encoder Boundary Points.
* `ebp_placement` - (Optional) Controls placement of Encoder Boundary Points on the video PID and audio PIDs.
* `ecm_pid` - (Optional) Packet identifier (PID) for ECM in the transport stream.
* `es_rate_in_pes` - (Optional) When set to `include`, writes the stream to the PES header.
* `etv_platform_pid` - (Optional) Packet identifier (PID) for input source ETV Platform data to this output.
* `etv_signal_pid` - (Optional) Packet identifier (PID) for input source ETV Signal data to this output.
* `fragment_time` - (Optional) Length, in seconds, of each fragment, used only with EBP markers.
* `klv` - (Optional) When set to `passthrough`, passes through KLV data from the SCTE-35 source.
* `klv_data_pids` - (Optional) Packet identifier (PID) for input source KLV data to this output.
* `nielsen_id3_behavior` - (Optional) When set to `passthrough`, passes through Nielsen ID3.
* `null_packet_bitrate` - (Optional) Value, in bits per second, to insert null packets into the transport stream.
* `pat_interval` - (Optional) Number of milliseconds between instances of this table in the output transport stream.
* `pcr_control` - (Optional) When set to `pcrEveryPesPacket`, a Program Clock Reference value is inserted for every Packetized Elementary Stream.
* `pcr_period` - (Optional) Maximum time, in milliseconds, between Program Clock References.
* `pcr_pid` - (Optional) Packet identifier (PID) of the Program Clock Reference in the transport stream.
* `pmt_interval` - (Optional) Number of milliseconds between instances of this table in the output transport stream.
* `pmt_pid` - (Optional) Packet identifier (PID) for the Program Map Table in the transport stream.
* `program_num` - (Optional) Value of the program number field in the Program Map Table.
* `rate_mode` - (Optional) When set to `vbr`, does not insert null packets into the transport stream.
* `scte27_pids` - (Optional) Packet identifier (PID) for input source SCTE-27 data to this output.
* `scte35_control` - (Optional) When set to `passthrough`, passes any SCTE-35 signals from the input source to this output.
* `scte35_pid` - (Optional) Packet identifier (PID) of the SCTE-35 stream in the transport stream.
* `segmentation_markers` - (Optional) Type of segmentation markers to include.
* `segmentation_style` - (Optional) Whether segmentation markers are included at the start of a stream.
* `segmentation_time` - (Optional) Length, in seconds, of each segment.
* `timed_metadata_behavior` - (Optional) When set to `passthrough`, passes through timed metadata.
* `timed_metadata_pid` - (Optional) Packet identifier (PID) of the timed metadata stream in the transport stream.
* `transport_stream_id` - (Optional) Value of the transport stream ID field in the Program Map Table.
* `video_pid` - (Optional) Packet identifier (PID) of the elementary video stream in the transport stream.

### `dvb_nit_settings` Block

* `network_id` - (Required) Numeric value placed in the Network Information Table.
* `network_name` - (Required) Network name text placed in the Network Information Table.
* `rep_interval` - (Optional) Number of milliseconds between instances of this table in the output transport stream.

### `dvb_sdt_settings` Block

* `output_sdt` - (Optional) Service Description Table output behavior.
* `rep_interval` - (Optional) Number of milliseconds between instances of this table in the output transport stream.
* `service_name` - (Optional) Service name placed in the `service_descriptor` in the Service Description Table.
* `service_provider_name` - (Optional) Service provider name placed in the `service_descriptor` in the Service Description Table.

### `dvb_tdt_settings` Block

* `rep_interval` - (Optional) Number of milliseconds between instances of this table in the output transport stream.

### `frame_capture_output_settings` Block

* `name_modifier` - (Optional) String concatenated to the end of the destination filename.

### `hls_output_settings` Block

* `h265_packaging_type` - (Optional) Only applicable when this output is referencing an H.265 video description. Specifies whether MP4 segments should be packaged as HEV1 or HVC1.
* `hls_settings` - (Required) Settings regarding the underlying stream. See [`hls_settings` Block](#hls_settings-block) for details.
* `name_modifier` - (Optional) String concatenated to the end of the destination filename. Required for multiple outputs of the same type.
* `segment_modifier` - (Optional) String concatenated to the end of the destination filename.

### `hls_settings` Block

* `audio_only_hls_settings` - (Optional) Audio-only HLS settings. See [`audio_only_hls_settings` Block](#audio_only_hls_settings-block) for details.
* `fmp4_hls_settings` - (Optional) fMP4 HLS settings. See [`fmp4_hls_settings` Block](#fmp4_hls_settings-block) for details.
* `frame_capture_hls_settings` - (Optional) Frame capture HLS settings. Can be set as an empty block.
* `standard_hls_settings` - (Optional) Standard HLS settings. See [`standard_hls_settings` Block](#standard_hls_settings-block) for details.

### `audio_only_hls_settings` Block

* `audio_group_id` - (Optional) Group to which the audio rendition belongs.
* `audio_only_image` - (Optional) Image for an audio-only output. See [`audio_only_image` Block](#audio_only_image-block) for details.
* `audio_track_type` - (Optional) Four types of audio-only tracks.
* `segment_type` - (Optional) Whether the stream is an audio-only stream.

### `audio_only_image` Block

* `password_param` - (Optional) Key used to extract the password from EC2 Parameter store.
* `uri` - (Required) Path to a file accessible to the live stream.
* `username` - (Optional) Username for the server.

### `fmp4_hls_settings` Block

* `audio_rendition_sets` - (Optional) List of audio rendition sets, separated by commas.
* `nielsen_id3_behavior` - (Optional) When set to `passthrough`, passes through Nielsen ID3.
* `timed_metadata_behavior` - (Optional) When set to `passthrough`, passes through timed metadata.

### `standard_hls_settings` Block

* `audio_rendition_sets` - (Optional) List of audio rendition sets, separated by commas.
* `m3u8_settings` - (Required) Settings for the M3U8 container. See [`m3u8_settings` Block](#m3u8_settings-block) for details.

### `m3u8_settings` Block

* `audio_frames_per_pes` - (Optional) Number of audio frames to insert for each PES packet.
* `audio_pids` - (Optional) Packet identifier (PID) of the elementary audio streams in the transport stream.
* `ecm_pid` - (Optional) Packet identifier (PID) for ECM in the transport stream.
* `nielsen_id3_behavior` - (Optional) When set to `passthrough`, passes through Nielsen ID3.
* `pat_interval` - (Optional) Number of milliseconds between instances of this table in the output transport stream.
* `pcr_control` - (Optional) When set to `pcrEveryPesPacket`, a Program Clock Reference value is inserted for every Packetized Elementary Stream.
* `pcr_period` - (Optional) Maximum time, in milliseconds, between Program Clock References.
* `pcr_pid` - (Optional) Packet identifier (PID) of the Program Clock Reference in the transport stream.
* `pmt_interval` - (Optional) Number of milliseconds between instances of this table in the output transport stream.
* `pmt_pid` - (Optional) Packet identifier (PID) for the Program Map Table in the transport stream.
* `program_num` - (Optional) Value of the program number field in the Program Map Table.
* `scte35_behavior` - (Optional) When set to `passthrough`, passes through SCTE-35.
* `scte35_pid` - (Optional) Packet identifier (PID) of the SCTE-35 stream in the transport stream.
* `timed_metadata_behavior` - (Optional) When set to `passthrough`, passes through timed metadata.
* `timed_metadata_pid` - (Optional) Packet identifier (PID) of the timed metadata stream in the transport stream.
* `transport_stream_id` - (Optional) Value of the transport stream ID field in the Program Map Table.
* `video_pid` - (Optional) Packet identifier (PID) of the elementary video stream in the transport stream.

### `ms_smooth_output_settings` Block

* `h265_packaging_type` - (Optional) Only applicable when this output is referencing an H.265 video description. Specifies whether MP4 segments should be packaged as HEV1 or HVC1.
* `name_modifier` - (Optional) String concatenated to the end of the destination filename.

### `multiplex_output_settings` Block

* `destination` - (Required) Multiplex destination. See [`destination` Block](#destination-block) for details.

### `rtmp_output_settings` Block

* `certificate_mode` - (Optional) Whether to allow self-signed or verified RTMP certificates.
* `connection_retry_interval` - (Optional) Number of seconds to wait before retrying a connection to the Flash Media Server if the connection is lost.
* `destination` - (Required) RTMP endpoint excluding the stream name. See [`destination` Block](#destination-block) for details.
* `num_retries` - (Optional) Number of retry attempts.

### `udp_output_settings` Block

* `buffer_msec` - (Optional) UDP output buffering, in milliseconds.
* `container_settings` - (Required) UDP container settings. See [`udp_output_settings.container_settings` Block](#udp_output_settingscontainer_settings-block) for details.
* `destination` - (Required) Destination address and port number for RTP or UDP packets. See [`destination` Block](#destination-block) for details.
* `fec_output_settings` - (Optional) Settings for enabling and adjusting Forward Error Correction on UDP outputs. See [`fec_output_settings` Block](#fec_output_settings-block) for details.

### `udp_output_settings.container_settings` Block

* `m2ts_settings` - (Optional) M2TS settings. See [`m2ts_settings` Block](#m2ts_settings-block) for details.

### `fec_output_settings` Block

* `column_depth` - (Optional) Height of the FEC protection matrix.
* `include_fec` - (Optional) Enables column only or column and row based FEC.
* `row_length` - (Optional) Width of the FEC protection matrix.

### `destination` Block

* `destination_ref_id` - (Required) Reference ID for the destination.

### `timecode_config` Block

* `source` - (Required) Source for the timecode associated with the event outputs.
* `sync_threshold` - (Optional) Threshold, in frames, beyond which output timecode is resynchronized to the input timecode.

### `video_descriptions` Block

* `codec_settings` - (Optional) Video codec settings. See [`video_descriptions.codec_settings` Block](#video_descriptionscodec_settings-block) for details.
* `height` - (Optional) Output video height, in pixels.
* `name` - (Required) Name of the video description.
* `respond_to_afd` - (Optional) How to respond to the AFD values that might be in the input video.
* `scaling_behavior` - (Optional) Behavior on how to scale.
* `sharpness` - (Optional) Strength of the anti-alias filter used for scaling.
* `width` - (Optional) Output video width, in pixels.

### `video_descriptions.codec_settings` Block

* `frame_capture_settings` - (Optional) Frame capture settings. See [`frame_capture_settings` Block](#frame_capture_settings-block) for details.
* `h264_settings` - (Optional) H264 settings. See [`h264_settings` Block](#h264_settings-block) for details.
* `h265_settings` - (Optional) H265 settings. See [`h265_settings` Block](#h265_settings-block) for details.

### `frame_capture_settings` Block

* `capture_interval` - (Optional) Frequency at which to capture frames for inclusion in the output.
* `capture_interval_units` - (Optional) Unit for the frame capture interval.

### `h264_settings` Block

* `adaptive_quantization` - (Optional) Enables or disables adaptive quantization.
* `afd_signaling` - (Optional) Whether AFD values are written into the output stream.
* `bitrate` - (Optional) Average bitrate in bits/second.
* `buf_fill_pct` - (Optional) Percentage of the buffer that should initially be filled.
* `buf_size` - (Optional) Size of the buffer in bits.
* `color_metadata` - (Optional) Whether color space metadata is included in the output.
* `entropy_encoding` - (Optional) Entropy encoding mode.
* `filter_settings` - (Optional) Filters to apply to an encode. See [`filter_settings` Block](#filter_settings-block) for details.
* `fixed_afd` - (Optional) Four-bit AFD value to write on all frames of video in the output stream.
* `flicker_aq` - (Optional) Makes adjustments within each frame to reduce flicker on the I-frames.
* `force_field_pictures` - (Optional) Whether coding is performed on a field basis or on a frame basis.
* `framerate_control` - (Optional) How the output video frame rate is specified.
* `framerate_denominator` - (Optional) Framerate denominator.
* `framerate_numerator` - (Optional) Framerate numerator.
* `gop_b_reference` - (Optional) GOP-B reference.
* `gop_closed_cadence` - (Optional) Frequency of closed GOPs.
* `gop_num_b_frames` - (Optional) Number of B-frames between reference frames.
* `gop_size` - (Optional) GOP size in units of either frames or seconds per `gop_size_units`.
* `gop_size_units` - (Optional) Whether `gop_size` is specified in frames or seconds.
* `level` - (Optional) H264 level.
* `look_ahead_rate_control` - (Optional) Amount of lookahead.
* `max_bitrate` - (Optional) Maximum bitrate, used to accommodate expected spikes in the complexity of the video.
* `min_i_interval` - (Optional) Minimum interval, in frames, between two I-frames.
* `num_ref_frames` - (Optional) Number of reference frames to use.
* `par_control` - (Optional) How the output pixel aspect ratio is specified.
* `par_denominator` - (Optional) Pixel aspect ratio denominator.
* `par_numerator` - (Optional) Pixel aspect ratio numerator.
* `profile` - (Optional) H264 profile.
* `quality_level` - (Optional) Quality level.
* `qvbr_quality_level` - (Optional) Target quality for the video encode.
* `rate_control_mode` - (Optional) Rate control mode.
* `scan_type` - (Optional) Scan type of the output.
* `scene_change_detect` - (Optional) Scene change detection.
* `slices` - (Optional) Number of slices per picture.
* `softness` - (Optional) Softness.
* `spatial_aq` - (Optional) Makes adjustments within each frame based on spatial variation of content complexity.
* `subgop_length` - (Optional) Subgop length.
* `syntax` - (Optional) Produces a bitstream compliant with SMPTE RP-2027.
* `temporal_aq` - (Optional) Makes adjustments within each frame based on temporal variation of content complexity.
* `timecode_insertion` - (Optional) How timecodes are inserted into the video elementary stream.

### `filter_settings` Block

* `temporal_filter_settings` - (Optional) Temporal filter settings. See [`temporal_filter_settings` Block](#temporal_filter_settings-block) for details.

### `temporal_filter_settings` Block

* `post_filter_sharpening` - (Optional) Post filter sharpening.
* `strength` - (Optional) Filter strength.

### `h265_settings` Block

* `adaptive_quantization` - (Optional) Enables or disables adaptive quantization.
* `afd_signaling` - (Optional) Whether AFD values are written into the output stream.
* `alternative_transfer_function` - (Optional) Whether EML inserts an Alternative Transfer Function SEI message.
* `bitrate` - (Required) Average bitrate in bits/second.
* `buf_size` - (Optional) Size of the buffer in bits.
* `color_metadata` - (Optional) Whether color space metadata is included in the output.
* `color_space_settings` - (Optional) Color metadata for the output. See [`color_space_settings` Block](#color_space_settings-block) for details.
* `filter_settings` - (Optional) Filters to apply to an encode. See [`filter_settings` Block](#filter_settings-block) for details.
* `fixed_afd` - (Optional) Four-bit AFD value to write on all frames of video in the output stream.
* `flicker_aq` - (Optional) Makes adjustments within each frame to reduce flicker on the I-frames.
* `framerate_denominator` - (Required) Framerate denominator.
* `framerate_numerator` - (Required) Framerate numerator.
* `gop_closed_cadence` - (Optional) Frequency of closed GOPs.
* `gop_size` - (Optional) GOP size in units of either frames or seconds per `gop_size_units`.
* `gop_size_units` - (Optional) Whether `gop_size` is specified in frames or seconds.
* `level` - (Optional) H265 level.
* `look_ahead_rate_control` - (Optional) Amount of lookahead.
* `max_bitrate` - (Optional) Maximum bitrate, used to accommodate expected spikes in the complexity of the video.
* `min_i_interval` - (Optional) Minimum interval, in frames, between two I-frames.
* `min_qp` - (Optional) Sets the minimum QP.
* `mv_over_picture_boundaries` - (Optional) Enables or disables motion vectors over picture boundaries.
* `mv_temporal_predictor` - (Optional) Enables or disables the motion vector temporal predictor.
* `par_denominator` - (Optional) Pixel aspect ratio denominator.
* `par_numerator` - (Optional) Pixel aspect ratio numerator.
* `profile` - (Optional) H265 profile.
* `qvbr_quality_level` - (Optional) Target quality for the video encode.
* `rate_control_mode` - (Optional) Rate control mode.
* `scan_type` - (Optional) Scan type of the output.
* `scene_change_detect` - (Optional) Scene change detection.
* `slices` - (Optional) Number of slices per picture.
* `tier` - (Optional) H265 tier in the output.
* `tile_height` - (Optional) Height of tiles.
* `tile_padding` - (Optional) Enables or disables padding of tiles.
* `tile_width` - (Optional) Width of tiles.
* `timecode_burnin_settings` - (Optional) Apply a burned-in timecode. See [`timecode_burnin_settings` Block](#timecode_burnin_settings-block) for details.
* `timecode_insertion` - (Optional) How timecodes are inserted into the video elementary stream.
* `treeblock_size` - (Optional) Size of the treeblock.

### `color_space_settings` Block

* `color_space_passthrough_settings` - (Optional) Sets the colorspace metadata to be passed through. Can be set as an empty block.
* `dolby_vision81_settings` - (Optional) Sets the colorspace to Dolby Vision81. Can be set as an empty block.
* `hdr10_settings` - (Optional) Sets the colorspace to be HDR10. See [`hdr10_settings` Block](#hdr10_settings-block) for details.
* `rec601_settings` - (Optional) Sets the colorspace to Rec. 601. Can be set as an empty block.
* `rec709_settings` - (Optional) Sets the colorspace to Rec. 709. Can be set as an empty block.

### `hdr10_settings` Block

* `max_cll` - (Optional) Sets the MaxCLL value for HDR10.
* `max_fall` - (Optional) Sets the MaxFALL value for HDR10.

### `timecode_burnin_settings` Block

* `prefix` - (Optional) Prefix on the burned-in timecode.
* `timecode_burnin_font_size` - (Optional) Size of the burned-in timecode.
* `timecode_burnin_position` - (Optional) Position of the burned-in timecode.

### `avail_blanking` Block

* `avail_blanking_image` - (Optional) Blanking image to be used. See [`avail_blanking_image` Block](#avail_blanking_image-block) for details.
* `state` - (Optional) When set to `enabled`, causes video, audio, and captions to be blanked when insertion metadata is added.

### `avail_blanking_image` Block

* `password_param` - (Optional) Key used to extract the password from EC2 Parameter store.
* `uri` - (Required) Path to a file accessible to the live stream.
* `username` - (Optional) Username for the server.

### `caption_descriptions` Block

* `accessibility` - (Optional) Whether the caption track implements accessibility features such as written descriptions of spoken dialog, music, and sounds.
* `caption_selector_name` - (Required) Input caption selector to use as a caption source when generating output captions. Must match a caption selector name.
* `destination_settings` - (Optional) Additional settings for the captions destination that depend on the destination type. See [`destination_settings` Block](#destination_settings-block) for details.
* `language_code` - (Optional) ISO 639-2 three-digit code.
* `language_description` - (Optional) Human-readable information to indicate the captions available for players. For example, English or Spanish.
* `name` - (Required) Name of the caption description. Used to associate a caption description with an output. Must be unique within an event.

### `destination_settings` Block

* `arib_destination_settings` - (Optional) ARIB destination settings. Can be set as an empty block.
* `burn_in_destination_settings` - (Optional) Burn-in destination settings. See [`burn_in_destination_settings` Block](#burn_in_destination_settings-block) for details.
* `dvb_sub_destination_settings` - (Optional) DVB Sub destination settings. See [`dvb_sub_destination_settings` Block](#dvb_sub_destination_settings-block) for details.
* `ebu_tt_d_destination_settings` - (Optional) EBU TT D destination settings. See [`ebu_tt_d_destination_settings` Block](#ebu_tt_d_destination_settings-block) for details.
* `embedded_destination_settings` - (Optional) Embedded destination settings. Can be set as an empty block.
* `embedded_plus_scte20_destination_settings` - (Optional) Embedded Plus SCTE-20 destination settings. Can be set as an empty block.
* `rtmp_caption_info_destination_settings` - (Optional) RTMP caption info destination settings. Can be set as an empty block.
* `scte20_plus_embedded_destination_settings` - (Optional) SCTE-20 Plus Embedded destination settings. Can be set as an empty block.
* `scte27_destination_settings` - (Optional) SCTE-27 destination settings. Can be set as an empty block.
* `smpte_tt_destination_settings` - (Optional) SMPTE TT destination settings. Can be set as an empty block.
* `teletext_destination_settings` - (Optional) Teletext destination settings. Can be set as an empty block.
* `ttml_destination_settings` - (Optional) TTML destination settings. See [`ttml_destination_settings` Block](#ttml_destination_settings-block) for details.
* `webvtt_destination_settings` - (Optional) WebVTT destination settings. See [`webvtt_destination_settings` Block](#webvtt_destination_settings-block) for details.

### `burn_in_destination_settings` Block

~> **Note:** All burn-in and DVB-Sub font settings must match.

* `alignment` - (Optional) Justification of the caption text. When no explicit `x_position` or `y_position` is provided, `centered` places the captions at the bottom center of the output and `left` aligns captions to the bottom left. If `x_position` and `y_position` are given with the alignment parameter, the font is justified (left or centered) relative to those coordinates. Selecting `smart` justification left-justifies live subtitles and center-justifies pre-recorded subtitles.
* `background_color` - (Optional) Color of the rectangle behind the captions.
* `background_opacity` - (Optional) Opacity of the background rectangle. 255 is opaque; 0 is transparent. Leaving this parameter out is equivalent to setting it to 0 (transparent).
* `font` - (Optional) External font file used for caption burn-in. File extension must be `ttf` or `tte`. Although you can select output fonts for many different types of input captions, embedded, STL, and teletext sources use a strict grid system; using external fonts with these caption sources could cause unexpected display of proportional fonts. See [`font` Block](#font-block) for details.
* `font_color` - (Optional) Color of the burned-in captions. This option is not valid for source captions that are STL, 608/embedded, or teletext. These source settings are already pre-defined by the caption stream.
* `font_opacity` - (Optional) Opacity of the burned-in captions. 255 is opaque; 0 is transparent.
* `font_resolution` - (Optional) Font resolution in DPI (dots per inch). Default is 96 dpi.
* `font_size` - (Optional) When set to `auto`, font size scales depending on the size of the output. A positive integer sets the exact font size in points.
* `outline_color` - (Required) Font outline color. This option is not valid for source captions that are either 608/embedded or teletext. These source settings are already pre-defined by the caption stream.
* `outline_size` - (Optional) Font outline size in pixels. This option is not valid for source captions that are either 608/embedded or teletext. These source settings are already pre-defined by the caption stream.
* `shadow_color` - (Optional) Color of the shadow cast by the captions.
* `shadow_opacity` - (Optional) Opacity of the shadow. 255 is opaque; 0 is transparent. Leaving this parameter out is equivalent to setting it to 0 (transparent).
* `shadow_x_offset` - (Optional) Horizontal offset, in pixels, of the shadow relative to the captions. A value of -2 results in a shadow offset 2 pixels to the left.
* `shadow_y_offset` - (Optional) Vertical offset, in pixels, of the shadow relative to the captions. A value of -2 results in a shadow offset 2 pixels above the text.
* `teletext_grid_control` - (Required) Whether a fixed grid size is used to generate the output subtitles bitmap. Only applicable for Teletext inputs and DVB-Sub/Burn-in outputs.
* `x_position` - (Optional) Horizontal position, in pixels, of the caption relative to the left side of the output. A value of 10 results in the captions starting 10 pixels from the left of the output. If no explicit `x_position` is provided, the horizontal caption position is determined by the alignment parameter.
* `y_position` - (Optional) Vertical position, in pixels, of the caption relative to the top of the output. A value of 10 results in the captions starting 10 pixels from the top of the output. If no explicit `y_position` is provided, the caption is positioned towards the bottom of the output.

### `font` Block

* `password_param` - (Optional) Key used to extract the password from EC2 Parameter store.
* `uri` - (Required) Path to a file accessible to the live stream.
* `username` - (Optional) Username for the server.

### `dvb_sub_destination_settings` Block

~> **Note:** All burn-in and DVB-Sub font settings must match.

* `alignment` - (Optional) Justification of the caption text. When no explicit `x_position` or `y_position` is provided, `centered` places the captions at the bottom center of the output and `left` aligns captions to the bottom left. If `x_position` and `y_position` are given with the alignment parameter, the font is justified (left or centered) relative to those coordinates. Selecting `smart` justification left-justifies live subtitles and center-justifies pre-recorded subtitles. This option is not valid for source captions that are STL or 608/embedded. These source settings are already pre-defined by the caption stream.
* `background_color` - (Optional) Color of the rectangle behind the captions.
* `background_opacity` - (Optional) Opacity of the background rectangle. 255 is opaque; 0 is transparent. Leaving this parameter blank is equivalent to setting it to 0 (transparent).
* `font` - (Optional) External font file used for caption burn-in. File extension must be `ttf` or `tte`. Although you can select output fonts for many different types of input captions, embedded, STL, and teletext sources use a strict grid system; using external fonts with these caption sources could cause unexpected display of proportional fonts. See [`font` Block](#font-block) for details.
* `font_color` - (Optional) Color of the burned-in captions. This option is not valid for source captions that are STL, 608/embedded, or teletext. These source settings are already pre-defined by the caption stream.
* `font_opacity` - (Optional) Opacity of the burned-in captions. 255 is opaque; 0 is transparent.
* `font_resolution` - (Optional) Font resolution in DPI (dots per inch). Default is 96 dpi.
* `font_size` - (Optional) When set to `auto`, font size scales depending on the size of the output. A positive integer sets the exact font size in points.
* `outline_color` - (Optional) Font outline color. This option is not valid for source captions that are either 608/embedded or teletext. These source settings are already pre-defined by the caption stream.
* `outline_size` - (Optional) Font outline size in pixels. This option is not valid for source captions that are either 608/embedded or teletext. These source settings are already pre-defined by the caption stream.
* `shadow_color` - (Optional) Color of the shadow cast by the captions.
* `shadow_opacity` - (Optional) Opacity of the shadow. 255 is opaque; 0 is transparent. Leaving this parameter blank is equivalent to setting it to 0 (transparent).
* `shadow_x_offset` - (Optional) Horizontal offset, in pixels, of the shadow relative to the captions. A value of -2 results in a shadow offset 2 pixels to the left.
* `shadow_y_offset` - (Optional) Vertical offset, in pixels, of the shadow relative to the captions. A value of -2 results in a shadow offset 2 pixels above the text.
* `teletext_grid_control` - (Optional) Whether a fixed grid size is used to generate the output subtitles bitmap. Only applicable for Teletext inputs and DVB-Sub/Burn-in outputs.
* `x_position` - (Optional) Horizontal position, in pixels, of the caption relative to the left side of the output. A value of 10 results in the captions starting 10 pixels from the left of the output. If no explicit `x_position` is provided, the horizontal caption position is determined by the alignment parameter. This option is not valid for source captions that are STL, 608/embedded, or teletext. These source settings are already pre-defined by the caption stream.
* `y_position` - (Optional) Vertical position, in pixels, of the caption relative to the top of the output. A value of 10 results in the captions starting 10 pixels from the top of the output. If no explicit `y_position` is provided, the caption is positioned towards the bottom of the output. This option is not valid for source captions that are STL, 608/embedded, or teletext. These source settings are already pre-defined by the caption stream.

### `ebu_tt_d_destination_settings` Block

* `copyright_holder` - (Optional) Name of the copyright holder to include in the copyright tag in the captions metadata.
* `fill_line_gap` - (Optional) How to handle the gap between the lines in multi-line captions. `enabled` fills the gap with the captions background color as specified in the input captions; `disabled` leaves the gap unfilled.
* `font_family` - (Optional) Font family to include in the font data attached to the EBU-TT captions. Valid only if `style_control` is set to `include`. If you leave this field empty, the font family is set to monospaced. You specify only the font family; all other style information (color, bold, position, and so on) is copied from the input captions, and the size is always set to 100% to allow the downstream player to choose the size. Enter a comma-separated list of font names in order of preference, a generic font family such as serif, or default to let the downstream player choose the font.
* `style_control` - (Optional) Style information to include in the font data that is attached to the EBU-TT captions. `include` takes the style information from the source captions and includes it in the font data, valid only if the source captions are Embedded or Teletext. `exclude` sets the font family to monospaced and does not include any other style information.

### `ttml_destination_settings` Block

* `style_control` - (Required) Whether to include style information when writing the TTML captions. This field is not currently supported and does not affect the output styling. Leave the default value.

### `webvtt_destination_settings` Block

* `style_control` - (Required) Whether the color and position of the source captions is passed through to the WebVTT output captions. `PASSTHROUGH` is valid only if the source captions are EMBEDDED or TELETEXT. `NO_STYLE_DATA` does not pass through the style, and the output captions do not contain any font styling information.

### `global_configuration` Block

* `initial_audio_gain` - (Optional) Value to set the initial audio gain for the live event.
* `input_end_action` - (Optional) Action to take when the current input completes. When `switchAndLoopInputs` is configured, the encoder restarts at the beginning of the first input. When `none` is configured, the encoder transcodes black, a solid color, or a user-specified slate image per the input loss behavior configuration until the next input switch occurs, which is controlled through the Channel Schedule API.
* `input_loss_behavior` - (Optional) Settings for system actions when input is lost. See [`input_loss_behavior` Block](#input_loss_behavior-block) for details.
* `output_locking_mode` - (Optional) How MediaLive pipelines are synchronized. `PIPELINE_LOCKING` attempts to synchronize the output of each pipeline to the other. `EPOCH_LOCKING` attempts to synchronize the output of each pipeline to the Unix epoch.
* `output_timing_source` - (Optional) Whether the rate of frames emitted by the live encoder is paced by its system clock, which optionally may be locked to another source via NTP, or locked to the clock of the input source.
* `support_low_framerate_inputs` - (Optional) Adjusts the video input buffer for streams with very low video framerates. This is commonly set to enabled for music channels with less than one video frame per second.

### `input_loss_behavior` Block

* `black_frame_msec` - (Optional) Amount of time, in milliseconds, to display black video when the input is lost.
* `input_loss_image_color` - (Optional) Color to display, as a hexadecimal RGBA value, when `input_loss_image_type` is `color`.
* `input_loss_image_slate` - (Optional) Image to display when `input_loss_image_type` is `slate`. See [`input_loss_image_slate` Block](#input_loss_image_slate-block) for details.
* `input_loss_image_type` - (Optional) Whether to display a solid color or a slate image when input is lost.
* `repeat_frame_msec` - (Optional) Amount of time, in milliseconds, to repeat the previous picture when the input is lost.

### `input_loss_image_slate` Block

* `password_param` - (Optional) Key used to extract the password from EC2 Parameter store.
* `uri` - (Required) Path to a file accessible to the live stream.
* `username` - (Optional) Username for the server.

### `motion_graphics_configuration` Block

* `motion_graphics_insertion` - (Optional) Whether to activate motion graphics insertion.
* `motion_graphics_settings` - (Required) Motion graphics settings. See [`motion_graphics_settings` Block](#motion_graphics_settings-block) for details.

### `motion_graphics_settings` Block

* `html_motion_graphics_settings` - (Optional) HTML motion graphics settings. Can be set as an empty block.

### `nielsen_configuration` Block

* `distributor_id` - (Optional) Distributor ID assigned to your organization by Nielsen.
* `nielsen_pcm_to_id3_tagging` - (Optional) Enables Nielsen PCM to ID3 tagging.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Channel.
* `channel_id` - ID of the Channel.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).
* `vpc[0].availability_zones` - Availability zones where the network interfaces are created.
* `vpc[0].network_interface_ids` - Network interface IDs created in the output VPC.

## Timeouts

[Configuration options](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts):

* `create` - (Default `15m`)
* `update` - (Default `15m`)
* `delete` - (Default `15m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_medialive_channel.example
  identity = {
    channel_id = "1234567"
  }
}

resource "aws_medialive_channel" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `channel_id` (String) ID of the MediaLive Channel.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

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
