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
* `vpc` - (Optional) Settings for the VPC outputs. See [VPC](#vpc) for more details.

### CDI Input Specification

* `resolution` - (Required) - Maximum CDI input resolution.

### Destinations

* `id` - (Required) User-specified id. Ths is used in an output group or an output.
* `media_package_settings` - (Optional) Destination settings for a MediaPackage output; one destination for both encoders. See [Media Package Settings](#media-package-settings) for more details.
* `multiplex_settings` - (Optional) Destination settings for a Multiplex output; one destination for both encoders. See [Multiplex Settings](#multiplex-settings) for more details.
* `settings` - (Optional) Destination settings for a standard output; one destination for each redundant encoder. See [Settings](#settings) for more details.

### Encoder Settings

* `output_groups` - (Required) Output groups for the channel. See [Output Groups](#output-groups) for more details.
* `timecode_config` - (Required) Contains settings used to acquire and adjust timecode information from inputs. See [Timecode Config](#timecode-config) for more details.
* `video_descriptions` - (Required) Video Descriptions. See [Video Descriptions](#video-descriptions) for more details.
* `audio_descriptions` - (Optional) Audio descriptions for the channel. See [Audio Descriptions](#audio-descriptions) for more details.
* `avail_blanking` - (Optional) Settings for ad avail blanking. See [Avail Blanking](#avail-blanking) for more details.
* `caption_descriptions` - (Optional) Caption Descriptions. See [Caption Descriptions](#caption-descriptions) for more details.
* `global_configuration` - (Optional) Configuration settings that apply to the event as a whole. See [Global Configuration](#global-configuration) for more details.
* `motion_graphics_configuration` - (Optional) Settings for motion graphics. See [Motion Graphics Configuration](#motion-graphics-configuration) for more details.
* `nielsen_configuration` - (Optional) Nielsen configuration settings. See [Nielsen Configuration](#nielsen-configuration) for more details.

### Input Attachments

* `input_attachment_name` - (Optional) User-specified name for the attachment.
* `input_id` - (Required) The ID of the input.
* `input_settings` - (Optional) Settings of an input. See [Input Settings](#input-settings) for more details.
* `automatic_input_failover_settings` - (Optional) User-specified settings for defining what the conditions are for declaring the input unhealthy and failing over to a different input. See [Automatic Input Failover Settings](#automatic-input-failover-settings) for more details.

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
* `selector_settings` - (Optional) The audio selector settings. See [Audio Selector Settings](#audio-selector-settings) for more details.

### Audio Selector Settings

* `audio_hls_rendition_selection` - (Optional) Audio HLS Rendition Selection. See [Audio HLS Rendition Selection](#audio-hls-rendition-selection) for more details.
* `audio_language_selection` - (Optional) Audio Language Selection. See [Audio Language Selection](#audio-language-selection) for more details.
* `audio_pid_selection` - (Optional) Audio Pid Selection. See [Audio PID Selection](#audio-pid-selection) for more details.
* `audio_track_selection` - (Optional) Audio Track Selection. See [Audio Track Selection](#audio-track-selection) for more details.

### Audio HLS Rendition Selection

* `group_id` - (Required) Specifies the GROUP-ID in the #EXT-X-MEDIA tag of the target HLS audio rendition.
* `name` - (Required) Specifies the NAME in the #EXT-X-MEDIA tag of the target HLS audio rendition.

### Audio Language Selection

* `language_code` - (Required) Selects a specific three-letter language code from within an audio source.
* `language_selection_policy` - (Optional) When set to “strict”, the transport stream demux strictly identifies audio streams by their language descriptor. If a PMT update occurs such that an audio stream matching the initially selected language is no longer present then mute will be encoded until the language returns. If “loose”, then on a PMT update the demux will choose another audio stream in the program with the same stream type if it can’t find one with the same language.

### Audio PID Selection

* `pid` - (Required) Selects a specific PID from within a source.

### Audio Track Selection

* `tracks` - (Required) Selects one or more unique audio tracks from within a source. See [Audio Tracks](#audio-tracks) for more details.
* `dolby_e_decode` - (Optional) Configure decoding options for Dolby E streams - these should be Dolby E frames carried in PCM streams tagged with SMPTE-337. See [Dolby E Decode](#dolby-e-decode) for more details.

### Audio Tracks

* `track` - (Required) 1-based integer value that maps to a specific audio track.

### Dolby E Decode

* `program_selection` - (Required) Applies only to Dolby E. Enter the program ID (according to the metadata in the audio) of the Dolby E program to extract from the specified track. One program extracted per audio selector. To select multiple programs, create multiple selectors with the same Track and different Program numbers. “All channels” means to ignore the program IDs and include all the channels in this selector; useful if metadata is known to be incorrect.

### Caption Selectors

* `name` - (Optional) The name of the caption selector.
* `language_code` - (Optional) When specified this field indicates the three letter language code of the caption track to extract from the source.
* `selector_settings` - (Optional) Caption selector settings. See [Caption Selector Settings](#caption-selector-settings) for more details.

### Caption Selector Settings

* `ancillary_source_settings` - (Optional) Ancillary Source Settings. See [Ancillary Source Settings](#ancillary-source-settings) for more details.
* `arib_source_settings` - (Optional) ARIB Source Settings.
* `dvb_sub_source_settings` - (Optional) DVB Sub Source Settings. See [DVB Sub Source Settings](#dvb-sub-source-settings) for more details.
* `embedded_source_settings` - (Optional) Embedded Source Settings. See [Embedded Source Settings](#embedded-source-settings) for more details.
* `scte20_source_settings` - (Optional) SCTE20 Source Settings. See [SCTE 20 Source Settings](#scte-20-source-settings) for more details.
* `scte27_source_settings` - (Optional) SCTE27 Source Settings. See [SCTE 27 Source Settings](#scte-27-source-settings) for more details.
* `teletext_source_settings` - (Optional) Teletext Source Settings. See [Teletext Source Settings](#teletext-source-settings) for more details.

### Ancillary Source Settings

* `source_ancillary_channel_number` - (Optional) Specifies the number (1 to 4) of the captions channel you want to extract from the ancillary captions. If you plan to convert the ancillary captions to another format, complete this field. If you plan to choose Embedded as the captions destination in the output (to pass through all the channels in the ancillary captions), leave this field blank because MediaLive ignores the field.

### DVB Sub Source Settings

* `ocr_language` - (Optional) If you will configure a WebVTT caption description that references this caption selector, use this field to provide the language to consider when translating the image-based source to text.
* `pid` - (Optional) When using DVB-Sub with Burn-In or SMPTE-TT, use this PID for the source content. Unused for DVB-Sub passthrough. All DVB-Sub content is passed through, regardless of selectors.

### Embedded Source Settings

* `convert_608_to_708` - (Optional) If upconvert, 608 data is both passed through via the “608 compatibility bytes” fields of the 708 wrapper as well as translated into 708. 708 data present in the source content will be discarded.
* `scte20_detection` - (Optional) Set to “auto” to handle streams with intermittent and/or non-aligned SCTE-20 and Embedded captions.
* `source_608_channel_number` - (Optional) Specifies the 608/708 channel number within the video track from which to extract captions. Unused for passthrough.

### SCTE 20 Source Settings

* `convert_608_to_708` – (Optional) If upconvert, 608 data is both passed through via the “608 compatibility bytes” fields of the 708 wrapper as well as translated into 708. 708 data present in the source content will be discarded.
* `source_608_channel_number` - (Optional) Specifies the 608/708 channel number within the video track from which to extract captions. Unused for passthrough.

### SCTE 27 Source Settings

* `ocr_language` - (Optional) If you will configure a WebVTT caption description that references this caption selector, use this field to provide the language to consider when translating the image-based source to text.
* `pid` - (Optional) The pid field is used in conjunction with the caption selector languageCode field as follows: - Specify PID and Language: Extracts captions from that PID; the language is "informational". - Specify PID and omit Language: Extracts the specified PID. - Omit PID and specify Language: Extracts the specified language, whichever PID that happens to be. - Omit PID and omit Language: Valid only if source is DVB-Sub that is being passed through; all languages will be passed through.

### Teletext Source Settings

* `output_rectangle` - (Optional) Optionally defines a region where TTML style captions will be displayed. See [Caption Rectangle](#caption-rectangle) for more details.
* `page_number` - (Optional) Specifies the teletext page number within the data stream from which to extract captions. Range of 0x100 (256) to 0x8FF (2303). Unused for passthrough. Should be specified as a hexadecimal string with no “0x” prefix.

### Caption Rectangle

* `height` - (Required) See the description in left\_offset. For height, specify the entire height of the rectangle as a percentage of the underlying frame height. For example, "80" means the rectangle height is 80% of the underlying frame height. The top\_offset and rectangle\_height must add up to 100% or less. This field corresponds to tts:extent - Y in the TTML standard.
* `left_offset` - (Required) Applies only if you plan to convert these source captions to EBU-TT-D or TTML in an output. (Make sure to leave the default if you don’t have either of these formats in the output.) You can define a display rectangle for the captions that is smaller than the underlying video frame. You define the rectangle by specifying the position of the left edge, top edge, bottom edge, and right edge of the rectangle, all within the underlying video frame. The units for the measurements are percentages. If you specify a value for one of these fields, you must specify a value for all of them. For leftOffset, specify the position of the left edge of the rectangle, as a percentage of the underlying frame width, and relative to the left edge of the frame. For example, "10" means the measurement is 10% of the underlying frame width. The rectangle left edge starts at that position from the left edge of the frame. This field corresponds to tts:origin - X in the TTML standard.
* `top_offset` - (Required) See the description in left\_offset. For top\_offset, specify the position of the top edge of the rectangle, as a percentage of the underlying frame height, and relative to the top edge of the frame. For example, "10" means the measurement is 10% of the underlying frame height. The rectangle top edge starts at that position from the top edge of the frame. This field corresponds to tts:origin - Y in the TTML standard.
* `width` - (Required) See the description in left\_offset. For width, specify the entire width of the rectangle as a percentage of the underlying frame width. For example, "80" means the rectangle width is 80% of the underlying frame width. The left\_offset and rectangle\_width must add up to 100% or less. This field corresponds to tts:extent - X in the TTML standard.

### Network Input Settings

* `hls_input_settings` - (Optional) Specifies HLS input settings when the URI is for a HLS manifest. See [HLS Input Settings](#hls-input-settings) for more details.
* `server_validation` - (Optional) Check HTTPS server certificates.

### HLS Input Settings

* `bandwidth` - (Optional) The bitrate is specified in bits per second, as in an HLS manifest.
* `buffer_segments` - (Optional) Buffer segments.
* `retries` - (Optional) The number of consecutive times that attempts to read a manifest or segment must fail before the input is considered unavailable.
* `retry_interval` - (Optional) The number of seconds between retries when an attempt to read a manifest or segment fails.
* `scte35_source_type` - (Optional) Identifies the source for the SCTE-35 messages that MediaLive will ingest.

### Automatic Input Failover Settings

* `secondary_input_id` - (Required) The input ID of the secondary input in the automatic input failover pair.
* `error_clear_time_msec` - (Optional) This clear time defines the requirement a recovered input must meet to be considered healthy. The input must have no failover conditions for this length of time. Enter a time in milliseconds. This value is particularly important if the input\_preference for the failover pair is set to PRIMARY\_INPUT\_PREFERRED, because after this time, MediaLive will switch back to the primary input.
* `failover_condition` - (Optional) A list of failover conditions. If any of these conditions occur, MediaLive will perform a failover to the other input. See [Failover Condition Block](#failover-condition-block) for more details.
* `input_preference` - (Optional) Input preference when deciding which input to make active when a previously failed input has recovered.

### Failover Condition Block

* `failover_condition_settings` - (Optional) Failover condition type-specific settings. See [Failover Condition Settings](#failover-condition-settings) for more details.

### Failover Condition Settings

* `audio_silence_settings` - (Optional) MediaLive will perform a failover if the specified audio selector is silent for the specified period. See [Audio Silence Failover Settings](#audio-silence-failover-settings) for more details.
* `input_loss_settings` - (Optional) MediaLive will perform a failover if content is not detected in this input for the specified period. See [Input Loss Failover Settings](#input-loss-failover-settings) for more details.
* `video_black_settings` - (Optional) MediaLive will perform a failover if content is considered black for the specified period. See [Video Black Failover Settings](#video-black-failover-settings) for more details.

### Audio Silence Failover Settings

* `audio_selector_name` - (Required) The name of the audio selector in the input that MediaLive should monitor to detect silence. Select your most important rendition. If you didn't create an audio selector in this input, leave blank.
* `audio_silence_threshold_msec` - (Optional) The amount of time (in milliseconds) that the active input must be silent before automatic input failover occurs. Silence is defined as audio loss or audio quieter than -50 dBFS.

### Input Loss Failover Settings

* `input_loss_threshold_msec` - (Optional) The amount of time (in milliseconds) that no input is detected. After that time, an input failover will occur.

### Video Black Failover Settings

* `black_detect_threshold` - (Optional) A value used in calculating the threshold below which MediaLive considers a pixel to be 'black'. For the input to be considered black, every pixel in a frame must be below this threshold. The threshold is calculated as a percentage (expressed as a decimal) of white. Therefore .1 means 10% white (or 90% black). Note how the formula works for any color depth. For example, if you set this field to 0.1 in 10-bit color depth: (10230.1=102.3), which means a pixel value of 102 or less is 'black'. If you set this field to .1 in an 8-bit color depth: (2550.1=25.5), which means a pixel value of 25 or less is 'black'. The range is 0.0 to 1.0, with any number of decimal places.
* `video_black_threshold_msec` - (Optional) The amount of time (in milliseconds) that the active input must be black before automatic input failover occurs.

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

* `attenuation_control` - (Optional) Sets the attenuation control.
* `bitrate` - (Optional) Average bitrate in bits/second.
* `bitstream_mode` - (Optional) Specifies the bitstream mode (bsmod) for the emitted AC-3 stream.
* `coding_mode` - (Optional) Dolby Digital coding mode.
* `dialnorm` - (Optional) Sets the dialnorm of the output.
* `drc_profile` - (Optional) If set to filmStandard, adds dynamic range compression signaling to the output bitstream as defined in the Dolby Digital specification.
* `lfe_filter` - (Optional) When set to enabled, applies a 120Hz lowpass filter to the LFE channel prior to encoding.
* `metadata_control` - (Optional) Metadata control.

### EAC3 Atmos Settings

* `bitrate` - (Optional) Average bitrate in bits/second.
* `coding_mode` - (Optional) Dolby Digital Plus with Dolby Atmos coding mode.
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
* `timezone` - (Optional) Choose the timezone for the time stamps in the watermark. If not provided, the timestamps will be in Coordinated Universal Time (UTC).

### Output Groups

* `output_group_settings` - (Required) Settings associated with the output group. See [Output Group Settings](#output-group-settings) for more details.
* `outputs` - (Required) List of outputs. See [Outputs](#outputs) for more details.
* `name` - (Optional) Custom output group name defined by the user.

### Output Group Settings

* `archive_group_settings` - (Optional) Archive group settings. See [Archive Group Settings](#archive-group-settings) for more details.
* `frame_capture_group_settings` - (Optional) Frame Capture Group Settings. See [Frame Capture Group Settings](#frame-capture-group-settings) for more details.
* `hls_group_settings` - (Optional) HLS Group Settings. See [HLS Group Settings](#hls-group-settings) for more details.
* `ms_smooth_group_settings` - (Optional) MS Smooth Group Settings. See [MS Smooth Group Settings](#ms-smooth-group-settings) for more details.
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
* `timecode_burnin_settings` - (Optional) Apply a burned in timecode. See [Timecode Burnin Settings](#timecode-burnin-settings) for more details.

### H264 Settings

* `adaptive_quantization` - (Optional) Enables or disables adaptive quantization.
* `afd_signaling` - (Optional) Indicates that AFD values will be written into the output stream.
* `bitrate` - (Optional) Average bitrate in bits/second.
* `buf_fil_pct` - (Optional) Percentage of the buffer that should initially be filled.
* `buf_size` - (Optional) Size of buffer in bits.
* `color_metadata` - (Optional) Includes color space metadata in the output.
* `color_space_settings` (Optional) Define the color metadata for the output. [H264 Color Space Settings](#h264-color-space-settings) for more details.
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
* `timecode_burnin_settings` - (Optional) Apply a burned in timecode. See [Timecode Burnin Settings](#timecode-burnin-settings) for more details.
* `timecode_insertion` - (Optional) Determines how timecodes should be inserted into the video elementary stream.

### H264 Color Space Settings

* `color_space_passthrough_settings` - (Optional) Sets the colorspace metadata to be passed through.
* `rec601_settings` - (Optional) Set the colorspace to Rec. 601.
* `rec709_settings` - (Optional) Set the colorspace to Rec. 709.

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
* `timecode_burnin_settings` - (Optional) Apply a burned in timecode. See [Timecode Burnin Settings](#timecode-burnin-settings) for more details.
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

### Timecode Burnin Settings

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

* `arib_destination_settings` - (Optional) ARIB Destination Settings.
* `burn_in_destination_settings` - (Optional) Burn In Destination Settings. See [Burn In Destination Settings](#burn-in-destination-settings) for more details.
* `dvb_sub_destination_settings` - (Optional) DVB Sub Destination Settings. See [DVB Sub Destination Settings](#dvb-sub-destination-settings) for more details.
* `ebu_tt_d_destination_settings` - (Optional) EBU TT D Destination Settings. See [EBU TT D Destination Settings](#ebu-tt-d-destination-settings) for more details.
* `embedded_destination_settings` - (Optional) Embedded Destination Settings.
* `embedded_plus_scte20_destination_settings` - (Optional) Embedded Plus SCTE20 Destination Settings.
* `rtmp_caption_info_destination_settings` - (Optional) RTMP Caption Info Destination Settings.
* `scte20_plus_embedded_destination_settings` - (Optional) SCTE20 Plus Embedded Destination Settings.
* `scte27_destination_settings` – (Optional) SCTE27 Destination Settings.
* `smpte_tt_destination_settings` – (Optional) SMPTE TT Destination Settings.
* `teletext_destination_settings` – (Optional) Teletext Destination Settings.
* `ttml_destination_settings` – (Optional) TTML Destination Settings. See [TTML Destination Settings](#ttml-destination-settings) for more details.
* `webvtt_destination_settings` - (Optional) WebVTT Destination Settings. See [WebVTT Destination Settings](#webvtt-destination-settings) for more details.

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

### DVB Sub Destination Settings

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

### EBU TT D Destination Settings

* `copyright_holder` – (Optional) Complete this field if you want to include the name of the copyright holder in the copyright tag in the captions metadata.
* `fill_line_gap` – (Optional) Specifies how to handle the gap between the lines (in multi-line captions). - enabled: Fill with the captions background color (as specified in the input captions). - disabled: Leave the gap unfilled.
* `font_family` – (Optional) Specifies the font family to include in the font data attached to the EBU-TT captions. Valid only if styleControl is set to include. If you leave this field empty, the font family is set to “monospaced”. (If styleControl is set to exclude, the font family is always set to “monospaced”.) You specify only the font family. All other style information (color, bold, position and so on) is copied from the input captions. The size is always set to 100% to allow the downstream player to choose the size. - Enter a list of font families, as a comma-separated list of font names, in order of preference. The name can be a font family (such as “Arial”), or a generic font family (such as “serif”), or “default” (to let the downstream player choose the font). - Leave blank to set the family to “monospace”.
* `style_control` – (Optional) Specifies the style information (font color, font position, and so on) to include in the font data that is attached to the EBU-TT captions. - include: Take the style information (font color, font position, and so on) from the source captions and include that information in the font data attached to the EBU-TT captions. This option is valid only if the source captions are Embedded or Teletext. - exclude: In the font data attached to the EBU-TT captions, set the font family to “monospaced”. Do not include any other style information.

### TTML Destination Settings

* `style_control` – (Optional) This field is not currently supported and will not affect the output styling. Leave the default value.

### WebVTT Destination Settings

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
* `output_locking_settings` - (Optional) Advanced output locking settings. See [Output Locking Settings](#output-locking-settings) for more details.
* `output_timing_source` – (Optional) Indicates whether the rate of frames emitted by the Live encoder should be paced by its system clock (which optionally may be locked to another source via NTP) or should be locked to the clock of the source that is providing the input stream.
* `support_low_framerate_inputs` – (Optional) Adjusts video input buffer for streams with very low video framerates. This is commonly set to enabled for music channels with less than one video frame per second.

### Input Loss Behavior

* `password_param` – (Optional) Key used to extract the password from EC2 Parameter store.
* `uri` – (Required) Path to a file accessible to the live stream.
* `username` – (Optional) Username to be used.

### Output Locking Settings

* `epoch_locking_settings` - (Optional) Epoch Locking Settings. See [Epoch Locking Settings](#epoch-locking-settings) for more details.
* `pipeline_locking_settings` - (Optional) Pipeline Locking Settings.

### Epoch Locking Settings

* `custom_epoch` - (Optional) Enter a value here to use a custom epoch, instead of the standard epoch (which started at 1970-01-01T00:00:00 UTC). Specify the start time of the custom epoch, in YYYY-MM-DDTHH:MM:SS in UTC. The time must be 2000-01-01T00:00:00 or later. Always set the MM:SS portion to 00:00.
* `jam_sync_time` - (Optional) Enter a time for the jam sync. The default is midnight UTC. When epoch locking is enabled, MediaLive performs a daily jam sync on every output encode to ensure timecodes don’t diverge from the wall clock. The jam sync applies only to encodes with frame rate of 29.97 or 59.94 FPS. To override, enter a time in HH:MM:SS in UTC. Always set the MM:SS portion to 00:00.

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

### Frame Capture Group Settings

* `destination` - (Required) A director and base filename where archive files should be written. See [Destination](#destination) for more details.
* `frame_capture_cdn_settings` - (Optional) Parameters that control interactions with the CDN. See [Frame Capture CDN Settings](#frame-capture-cdn-settings) for more details.

### Frame Capture CDN Settings

* `frame_capture_s3_settings` - (Optional) Frame Capture S3 Settings. See [Frame Capture S3 Settings](#frame-capture-s3-settings) for more details.

### Frame Capture S3 Settings

* `canned_acl` - (Optional) Specify the canned ACL to apply to each S3 request. Defaults to none.

### HLS Group Settings

* `destination` - (Required) A director and base filename where archive files should be written. See [Destination](#destination) for more details.
* `ad_markers` - (Optional) Choose one or more ad marker types to pass SCTE35 signals through to this group of Apple HLS outputs.
* `base_url_content` - (Optional) A partial URI prefix that will be prepended to each output in the media .m3u8 file. Can be used if base manifest is delivered from a different URL than the main .m3u8 file.
* `base_url_content1` - (Optional) One value per output group. This field is required only if you are completing Base URL content A, and the downstream system has notified you that the media files for pipeline 1 of all outputs are in a location different from the media files for pipeline 0.
* `base_url_manifest` - (Optional) A partial URI prefix that will be prepended to each output in the media .m3u8 file. Can be used if base manifest is delivered from a different URL than the main .m3u8 file.
* `base_url_manifest1` - (Optional) One value per output group. Complete this field only if you are completing Base URL manifest A, and the downstream system has notified you that the child manifest files for pipeline 1 of all outputs are in a location different from the child manifest files for pipeline 0.
* `caption_language_mappings` - (Optional) Mapping of up to 4 caption channels to caption languages. Is only meaningful if captionLanguageSetting is set to "insert". See [Caption Language Mappings](#caption-language-mappings) for more details.
* `caption_language_setting` - (Optional) Applies only to 608 Embedded output captions. insert: Include CLOSED-CAPTIONS lines in the manifest. Specify at least one language in the CC1 Language Code field. One CLOSED-CAPTION line is added for each Language Code you specify. Make sure to specify the languages in the order in which they appear in the original source (if the source is embedded format) or the order of the caption selectors (if the source is other than embedded). Otherwise, languages in the manifest will not match up properly with the output captions. none: Include CLOSED-CAPTIONS=NONE line in the manifest. omit: Omit any CLOSED-CAPTIONS line from the manifest.
* `client_cache` - (Optional) When set to "disabled", sets the #EXT-X-ALLOW-CACHE:no tag in the manifest, which prevents clients from saving media segments for later replay.
* `codec_specification - (Optional) Specification to use (RFC-6381 or the default RFC-4281) during m3u8 playlist generation.
* `constant_iv` - (Optional) For use with encryptionType. This is a 128-bit, 16-byte hex value represented by a 32-character text string. If ivSource is set to "explicit" then this parameter is required and is used as the IV for encryption.
* `directory_structure` - (Optional) Place segments in subdirectories.
* `discontinuity_tags` - (Optional) Specifies whether to insert EXT-X-DISCONTINUITY tags in the HLS child manifests for this output group. Typically, choose Insert because these tags are required in the manifest (according to the HLS specification) and serve an important purpose. Choose Never Insert only if the downstream system is doing real-time failover (without using the MediaLive automatic failover feature) and only if that downstream system has advised you to exclude the tags.
* `encryption_type` - (Optional) Encrypts the segments with the given encryption scheme. Exclude this parameter if no encryption is desired.
* `hls_cdn_settings` - (Optional) Parameters that control interactions with the CDN. See [HLS CDN Settings](#hls-cdn-settings) for more details.
* `hls_id3_segment_tagging` - (Optional) State of HLS ID3 Segment Tagging.
* `iframe_only_playlists` - (Optional) DISABLED: Do not create an I-frame-only manifest, but do create the master and media manifests (according to the Output Selection field). STANDARD: Create an I-frame-only manifest for each output that contains video, as well as the other manifests (according to the Output Selection field). The I-frame manifest contains a #EXT-X-I-FRAMES-ONLY tag to indicate it is I-frame only, and one or more #EXT-X-BYTERANGE entries identifying the I-frame position. For example, #EXT-X-BYTERANGE:160364@1461888".
* `incomplete_segment_behavior` - (Optional) Specifies whether to include the final (incomplete) segment in the media output when the pipeline stops producing output because of a channel stop, a channel pause or a loss of input to the pipeline. Auto means that MediaLive decides whether to include the final segment, depending on the channel class and the types of output groups. Suppress means to never include the incomplete segment. We recommend you choose Auto and let MediaLive control the behavior.
* `index_n_segments` - (Optional) Applies only if Mode field is LIVE. Specifies the maximum number of segments in the media manifest file. After this maximum, older segments are removed from the media manifest. This number must be smaller than the number in the Keep Segments field.
* `input_loss_action` - (Optional) Parameter that control output group behavior on input loss.
* `iv_in_manifest` - (Optional) For use with encryptionType. The IV (Initialization Vector) is a 128-bit number used in conjunction with the key for encrypting blocks. If set to "include", IV is listed in the manifest, otherwise the IV is not in the manifest.
* `iv_source` - (Optional) For use with encryptionType. The IV (Initialization Vector) is a 128-bit number used in conjunction with the key for encrypting blocks. If this setting is "followsSegmentNumber", it will cause the IV to change every segment (to match the segment number). If this is set to "explicit", you must enter a constantIv value.
* `keep_segments` - (Optional) Applies only if Mode field is LIVE. Specifies the number of media segments to retain in the destination directory. This number should be bigger than indexNSegments (Num segments). We recommend (value = (2 x indexNsegments) + 1). If this "keep segments" number is too low, the following might happen: the player is still reading a media manifest file that lists this segment, but that segment has been removed from the destination directory (as directed by indexNSegments). This situation would result in a 404 HTTP error on the player.
* `key_format` - (Optional) The value specifies how the key is represented in the resource identified by the URI. If parameter is absent, an implicit value of "identity" is used. A reverse DNS string can also be given.
* `key_format_versions` - (Optional) Either a single positive integer version value or a slash delimited list of version values (1/2/3).
* `key_provider_settings` - (Optional) The key provider settings. See [Key Provider Settings](#key-provider-settings) for more details.
* `manifest_compression` - (Optional) When set to gzip, compresses HLS playlist.
* `manifest_duration_format` - (Optional) Indicates whether the output manifest should use floating point or integer values for segment duration.
* `min_segment_length` - (Optional) Minimum length of MPEG-2 Transport Stream segments in seconds. When set, minimum segment length is enforced by looking ahead and back within the specified range for a nearby avail and extending the segment size if needed.
* `mode` - (Optional) If "vod", all segments are indexed and kept permanently in the destination and manifest. If "live", only the number segments specified in keepSegments and indexNSegments are kept; newer segments replace older segments, which may prevent players from rewinding all the way to the beginning of the event. VOD mode uses HLS EXT-X-PLAYLIST-TYPE of EVENT while the channel is running, converting it to a "VOD" type manifest on completion of the stream.
* `output_selection` - (Optional) MANIFESTS\_AND\_SEGMENTS: Generates manifests (master manifest, if applicable, and media manifests) for this output group. VARIANT\_MANIFESTS\_AND\_SEGMENTS: Generates media manifests for this output group, but not a master manifest. SEGMENTS\_ONLY: Does not generate any manifests for this output group.
* `program_date_time` - (Optional) Includes or excludes EXT-X-PROGRAM-DATE-TIME tag in .m3u8 manifest files. The value is calculated using the program date time clock.
* `program_date_time_clock` - (Optional) Specifies the algorithm used to drive the HLS EXT-X-PROGRAM-DATE-TIME clock. Options include: INITIALIZE\_FROM\_OUTPUT\_TIMECODE: The PDT clock is initialized as a function of the first output timecode, then incremented by the EXTINF duration of each encoded segment. SYSTEM\_CLOCK: The PDT clock is initialized as a function of the UTC wall clock, then incremented by the EXTINF duration of each encoded segment. If the PDT clock diverges from the wall clock by more than 500ms, it is resynchronized to the wall clock.
* `program_date_time_period` - (Optional) Period of insertion of EXT-X-PROGRAM-DATE-TIME entry, in seconds.
* `redundant_manifest` - (Optional) ENABLED: The master manifest (.m3u8 file) for each pipeline includes information about both pipelines: first its own media files, then the media files of the other pipeline. This feature allows playout device that support stale manifest detection to switch from one manifest to the other, when the current manifest seems to be stale. There are still two destinations and two master manifests, but both master manifests reference the media files from both pipelines. DISABLED: The master manifest (.m3u8 file) for each pipeline includes information about its own pipeline only. For an HLS output group with MediaPackage as the destination, the DISABLED behavior is always followed. MediaPackage regenerates the manifests it serves to players so a redundant manifest from MediaLive is irrelevant.
* `segment_length` - (Optional) Length of MPEG-2 Transport Stream segments to create in seconds. Note that segments will end on the next keyframe after this duration, so actual segment length may be longer.
* `segments_per_subdirectory` - (Optional) Number of segments to write to a subdirectory before starting a new one. directoryStructure must be subdirectoryPerStream for this setting to have an effect.
* `stream_inf_resolution` - (Optional) Include or exclude RESOLUTION attribute for video in EXT-X-STREAM-INF tag of variant manifest.
* `timed_metadata_id3_frame` - (Optional) Indicates ID3 frame that has the timecode.
* `timed_metadata_id3_period` - (Optional) Timed Metadata interval in seconds.
* `timestamp_delta_milliseconds` - (Optional) Provides an extra millisecond delta offset to fine tune the timestamps.
* `ts_file_mode` - (Optional) SEGMENTED\_FILES: Emit the program as segments - multiple .ts media files. SINGLE\_FILE: Applies only if Mode field is VOD. Emit the program as a single .ts media file. The media manifest includes #EXT-X-BYTERANGE tags to index segments for playback. A typical use for this value is when sending the output to AWS Elemental MediaConvert, which can accept only a single media file. Playback while the channel is running is not guaranteed due to HTTP server caching.

### HLS CDN Settings

* `hls_akamai_settings` - (Optional) HLS Akamai Settings. See [HLS Akamai Settings](#hls-akamai-settings) for more details.
* `hls_basic_put_settings` - (Optional) HLS Basic Put Settings. See [HLS Basic Put Settings](#hls-basic-put-settings) for more details.
* `hls_media_store_settings` - (Optional) HLS Media Store Settings. See [HLS Media Store Settings](#hls-media-store-settings) for more details.
* `hls_s3_settings` - (Optional) HLS S3 Settings. See [HLS S3 Settings](#hls-s3-settings) for more details.
* `hls_webdav_settings` - (Optional) HLS WebDAV Settings. See [HLS WebDAV Settings](#hls-webdav-settings) for more details.

### HLS Akamai Settings

* `connection_retry_interval` - (Optional) Number of seconds to wait before retrying connection to the CDN if the connection is lost.
* `filecache_duration` - (Optional) Size in seconds of file cache for streaming outputs.
* `http_transfer_mode` - (Optional) Specify whether or not to use chunked transfer encoding to Akamai. User should contact Akamai to enable this feature.
* `num_retries` - (Optional) Number of retry attempts that will be made before the Live Event is put into an error state. Applies only if the CDN destination URI begins with "s3" or "mediastore". For other URIs, the value is always 3.
* `restart_delay` - (Optional) If a streaming output fails, number of seconds to wait until a restart is initiated. A value of 0 means never restart.
* `salt` - (Optional) Salt for authenticated Akamai.
* `token` - (Optional) Token parameter for authenticated akamai. If not specified, gda is used.

### HLS Basic Put Settings

* `connection_retry_interval` - (Optional) Number of seconds to wait before retrying connection to the CDN if the connection is lost.
* `filecache_duration` - (Optional) Size in seconds of file cache for streaming outputs.
* `num_retries` - (Optional) Number of retry attempts that will be made before the Live Event is put into an error state. Applies only if the CDN destination URI begins with "s3" or "mediastore". For other URIs, the value is always 3.
* `restart_delay` - (Optional) If a streaming output fails, number of seconds to wait until a restart is initiated. A value of 0 means never restart.

### HLS Media Store Settings

* `connection_retry_interval` - (Optional) Number of seconds to wait before retrying connection to the CDN if the connection is lost.
* `filecache_duration` - (Optional) Size in seconds of file cache for streaming outputs.
* `media_store_storage_class` - (Optional) When set to temporal, output files are stored in non-persistent memory for faster reading and writing.
* `num_retries` - (Optional) Number of retry attempts that will be made before the Live Event is put into an error state. Applies only if the CDN destination URI begins with "s3" or "mediastore". For other URIs, the value is always 3.
* `restart_delay` - (Optional) If a streaming output fails, number of seconds to wait until a restart is initiated. A value of 0 means never restart.

### HLS S3 Settings

* `canned_acl` - (Optional) Specify the canned ACL to apply to each S3 request. Defaults to none.

### HLS WebDAV Settings

* `connection_retry_interval` - (Optional) Number of seconds to wait before retrying connection to the CDN if the connection is lost.
* `filecache_duration` - (Optional) Size in seconds of file cache for streaming outputs.
* `http_transfer_mode` - (Optional) Specify whether or not to use chunked transfer encoding to WebDAV.
* `num_retries` - (Optional) Number of retry attempts that will be made before the Live Event is put into an error state. Applies only if the CDN destination URI begins with "s3" or "mediastore". For other URIs, the value is always 3.
* `restart_delay` - (Optional) If a streaming output fails, number of seconds to wait until a restart is initiated. A value of 0 means never restart.

### Key Provider Settings

* `static_key_settings` - (Optional) Static Key Settings. See [Static Key Settings](#static-key-settings) for more details.

### Static Key Settings

* `static_key_value` - (Required) Static key value as a 32 character hexadecimal string.
* `key_provider_server` - (Optional) The URL of the license server used for protecting content. See [Key Provider Server](#key-provider-server) for more details.

### Key Provider Server

* `password_param` – (Optional) Key used to extract the password from EC2 Parameter store.
* `uri` – (Required) Path to a file accessible to the live stream.
* `username` – (Optional) Username to be used.

### MS Smooth Group Settings

* `destination` - (Required) Smooth Streaming publish point on an IIS server. Elemental Live acts as a "Push" encoder to IIS. See [Destination](#destination) for more details.
* `acquisition_point_id` - (Optional) The ID to include in each message in the sparse track. Ignored if sparseTrackType is NONE.
* `audio_only_timecode_control` - (Optional) If set to passthrough for an audio-only MS Smooth output, the fragment absolute time will be set to the current timecode. This option does not write timecodes to the audio elementary stream.
* `certificate_mode` - (Optional) If set to verifyAuthenticity, verify the https certificate chain to a trusted Certificate Authority (CA). This will cause https outputs to self-signed certificates to fail.
* `connection_retry_interval` - (Optional) Number of seconds to wait before retrying connection to the IIS server if the connection is lost. Content will be cached during this time and the cache will be be delivered to the IIS server once the connection is re-established.
* `event_id` - (Optional) MS Smooth event ID to be sent to the IIS server. Should only be specified if eventIdMode is set to useConfigured.
* `event_id_mode` - (Optional) Specifies whether or not to send an event ID to the IIS server. If no event ID is sent and the same Live Event is used without changing the publishing point, clients might see cached video from the previous run. Options: - "useConfigured" - use the value provided in eventId. - "useTimestamp" - generate and send an event ID based on the current timestamp. - "noEventId" - do not send an event ID to the IIS server.
* `event_stop_behavior` - (Optional) When set to sendEos, send EOS signal to IIS server when stopping the event.
* `filecache_duration` - (Optional) Size in seconds of file cache for streaming outputs.
* `fragment_length` - (Optional) Length of mp4 fragments to generate (in seconds). Fragment length must be compatible with GOP size and framerate.
* `input_loss_action` - (Optional) Parameter that control output group behavior on input loss.
* `num_retries` - (Optional) Number of retry attempts.
* `restart_delay` - (Optional) Number of seconds before initiating a restart due to output failure, due to exhausting the numRetries on one segment, or exceeding filecacheDuration.
* `send_delay_ms` - (Optional) Number of milliseconds to delay the output from the second pipeline.
* `sparse_track_type` - (Optional) Identifies the type of data to place in the sparse track: - SCTE35: Insert SCTE-35 messages from the source content. With each message, insert an IDR frame to start a new segment. - SCTE35\_WITHOUT\_SEGMENTATION: Insert SCTE-35 messages from the source content. With each message, insert an IDR frame but don't start a new segment. - NONE: Don't generate a sparse track for any outputs in this output group.
* `stream_manifest_behavior` - (Optional) When set to send, send stream manifest so publishing point doesn't start until all streams start.
* `timestamp_offset` - (Optional) Timestamp offset for the event. Only used if timestampOffsetMode is set to useConfiguredOffset.
* `timestamp_offset_mode` - (Optional) Type of timestamp date offset to use. - useEventStartDate: Use the date the event was started as the offset. - useConfiguredOffset: Use an explicitly configured date as the offset.

### Caption Language Mappings

* `caption_channel` - (Required) The closed caption channel being described by this CaptionLanguageMapping. Each channel mapping must have a unique channel number (maximum of 4).
* `language_code` - (Required) Three character ISO 639-2 language code (see http://www.loc.gov/standards/iso639-2).
* `language_description` - (Required) Textual description of language.

### Media Package Group Settings

* `destination` - (Required) A director and base filename where archive files should be written. See [Destination](#destination) for more details.

### RTMP Group Settings

* `ad_markers` - (Optional) The ad marker type for this output group.
* `authentication_scheme` - (Optional) Authentication scheme to use when connecting with CDN.
* `cache_full_behavior` - (Optional) Controls behavior when content cache fills up.
* `cache_length` - (Optional) Cache length in seconds, is used to calculate buffer size.
* `caption_data` - (Optional) Controls the types of data that passes to onCaptionInfo outputs.
* `include_filler_nal_units` - (Optional) Applies only when the rate control mode (in the codec settings) is CBR (constant bit rate). Controls whether the RTMP output stream is padded (with FILL NAL units) in order to achieve a constant bit rate that is truly constant. When there is no padding, the bandwidth varies (up to the bitrate value in the codec settings). We recommend that you choose Auto.
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
* `frame_capture_output_settings` - (Optional) Frame capture output settings. See [Frame Capture Output Settings](#frame-capture-output-settings) for more details.
* `hls_output_settings` - (Optional) HLS output settings. See [HLS Output Settings](#hls-output-settings) for more details.
* `media_package_output_settings` - (Optional) Media package output settings. This can be set as an empty block.
* `multiplex_output_settings` - (Optional) Multiplex output settings. See [Multiplex Output Settings](#multiplex-output-settings) for more details.
* `rtmp_output_settings` - (Optional) RTMP output settings. See [RTMP Output Settings](#rtmp-output-settings) for more details.
* `udp_output_settings` - (Optional) UDP output settings. See [UDP Output Settings](#udp-output-settings) for more details.

### Archive Output Settings

* `container_settings` - (Required) Settings specific to the container type of the file. See [Container Settings](#container-settings) for more details.
* `extension` - (Optional) Output file extension.
* `name_modifier` - (Optional) String concatenated to the end of the destination filename. Required for multiple outputs of the same type.

### Frame Capture Output Settings

* `name_modifier` - (Optional) Required if the output group contains more than one output. This modifier forms part of the output file name.

### HLS Output Settings

* `hls_settings` - (Required) Settings regarding the underlying stream. These settings are different for audio-only outputs. See [HLS Settings](#hls-settings) for more details.
* `h265_packaging_type` - (Optional) Only applicable when this output is referencing an H.265 video description. Specifies whether MP4 segments should be packaged as HEV1 or HVC1.
* `name_modifier` - (Optional) String concatenated to the end of the destination filename. Accepts "Format Identifiers":#formatIdentifierParameters.
* `segment_modifier` - (Optional) String concatenated to end of segment filenames.

### HLS Settings

* `audio_only_hls_settings` - (Optional) Audio Only HLS Settings. See [Audio Only HLS Settings](#audio-only-hls-settings) for more details.
* `fmp4_hls_settings` - (Optional) FMP4 HLS Settings. See [FMP4 HLS Settings](#fmp4-hls-settings) for more details.
* `frame_capture_hls_settings` - (Optional) Frame Capture HLS Settings.
* `standard_hls_settings` - (Optional) Standard HLS Settings. See [Standard HLS Settings](#standard-hls-settings) for more details.

### Audio Only HLS Settings

* `audio_group_id` - (Optional) Specifies the group to which the audio Rendition belongs.
* `audio_only_image` - (Optional) Specifies the .jpg or .png image to use as the cover art for an audio-only output. We recommend a low bit-size file because the image increases the output audio bandwidth. The image is attached to the audio as an ID3 tag, frame type APIC, picture type 0x10, as per the "ID3 tag version 2.4.0 - Native Frames" standard.
* `audio_track_type` - (Optional) Four types of audio-only tracks are supported: Audio-Only Variant Stream The client can play back this audio-only stream instead of video in low-bandwidth scenarios. Represented as an EXT-X-STREAM-INF in the HLS manifest. Alternate Audio, Auto Select, Default Alternate rendition that the client should try to play back by default. Represented as an EXT-X-MEDIA in the HLS manifest with DEFAULT=YES, AUTOSELECT=YES Alternate Audio, Auto Select, Not Default Alternate rendition that the client may try to play back by default. Represented as an EXT-X-MEDIA in the HLS manifest with DEFAULT=NO, AUTOSELECT=YES Alternate Audio, not Auto Select Alternate rendition that the client will not try to play back by default. Represented as an EXT-X-MEDIA in the HLS manifest with DEFAULT=NO, AUTOSELECT=NO
* `segment_type` - (Optional) Specifies the segment type.

### FMP4 HLS Settings

* `audio_rendition_sets` - (Optional) List all the audio groups that are used with the video output stream. Input all the audio GROUP-IDs that are associated to the video, separate by ','.
* `nielsen_id3_behavior` - (Optional) If set to passthrough, Nielsen inaudible tones for media tracking will be detected in the input audio and an equivalent ID3 tag will be inserted in the output.
* `timed_metadata_behavior` - (Optional) When set to passthrough, timed metadata is passed through from input to output.

### Standard HLS Settings

* `m3u8_settings` - (Required) Settings information for the .m3u8 container. See [M3U8 Settings](#m3u8-settings) for more details.
* `audio_rendition_sets` - (Optional) List all the audio groups that are used with the video output stream. Input all the audio GROUP-IDs that are associated to the video, separate by ','.

### M3U8 Settings

* `audio_frames_per_pes` - (Optional) The number of audio frames to insert for each PES packet.
* `audio_pids` - (Optional) Packet Identifier (PID) of the elementary audio stream(s) in the transport stream. Multiple values are accepted, and can be entered in ranges and/or by comma separation. Can be entered as decimal or hexadecimal values.
* `klv_behavior` - (Optional) If set to passthrough, passes any KLV data from the input source to this output.
* `klv_data_pids` - (Optional) Packet Identifier (PID) for input source KLV data to this output. Multiple values are accepted, and can be entered in ranges and/or by comma separation. Can be entered as decimal or hexadecimal values. Each PID specified must be in the range of 32 (or 0x20)..8182 (or 0x1ff6).
* `nielsen_id3_behavior` - (Optional) If set to passthrough, Nielsen inaudible tones for media tracking will be detected in the input audio and an equivalent ID3 tag will be inserted in the output.
* `pat_interval` - (Optional) The number of milliseconds between instances of this table in the output transport stream. A value of "0" writes out the PMT once per segment file.
* `pcr_control` - (Optional) When set to pcrEveryPesPacket, a Program Clock Reference value is inserted for every Packetized Elementary Stream (PES) header. This parameter is effective only when the PCR PID is the same as the video or audio elementary stream.
* `pcr_period` - (Optional) Maximum time in milliseconds between Program Clock References (PCRs) inserted into the transport stream.
* `pcr_pid` - (Optional) Packet Identifier (PID) of the Program Clock Reference (PCR) in the transport stream. When no value is given, the encoder will assign the same value as the Video PID. Can be entered as a decimal or hexadecimal value.
* `pmt_interval` - (Optional) The number of milliseconds between instances of this table in the output transport stream. A value of "0" writes out the PMT once per segment file.
* `pmt_pid` - (Optional) Packet Identifier (PID) for the Program Map Table (PMT) in the transport stream. Can be entered as a decimal or hexadecimal value.
* `program_num` - (Optional) The value of the program number field in the Program Map Table.
* `scte35_behavior` - (Optional) If set to passthrough, passes any SCTE-35 signals from the input source to this output.
* `scte35_pid` - (Optional) Packet Identifier (PID) of the SCTE-35 stream in the transport stream. Can be entered as a decimal or hexadecimal value.
* `timed_metadata_behavior` - (Optional) When set to passthrough, timed metadata is passed through from input to output.
* `timed_metadata_pid` - (Optional) Packet Identifier (PID) of the timed metadata stream in the transport stream. Can be entered as a decimal or hexadecimal value. Valid values are 32 (or 0x20)..8182 (or 0x1ff6).
* `transport_stream_id` - (Optional) The value of the transport stream ID field in the Program Map Table.
* `video_pid` - (Optional) Packet Identifier (PID) of the elementary video stream in the transport stream. Can be entered as a decimal or hexadecimal value.

### Multiplex Output Settings

* `destination` - (Required) Destination is a multiplex. See [Destination](#destination) for more details.

### RTMP Output Settings

- `destination` - (Required) The RTMP endpoint excluding the stream name. See [Destination](#destination) for more details.
- `certificate_mode` - (Optional) Setting to allow self signed or verified RTMP certificates.
- `connection_retry_interval` - (Optional) Number of seconds to wait before retrying connection to the flash media server if the connection is lost.
- `num_retries` - (Optional) Number of retry attempts.

### Container Settings

* `m2ts_settings` - (Optional) M2TS Settings. See [M2TS Settings](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-medialive-channel-m2tssettings.html) for more details.
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

### VPC

* `subnet_ids` - (Required) A list of VPC subnet IDs from the same VPC. If STANDARD channel, subnet IDs must be mapped to two unique availability zones (AZ).
* `public_address_allocation_ids` - (Required) List of public address allocation ids to associate with ENIs that will be created in Output VPC. Must specify one for SINGLE_PIPELINE, two for STANDARD channels.
* `security_group_ids` - (Optional) A list of up to 5 EC2 VPC security group IDs to attach to the Output VPC network interfaces. If none are specified then the VPC default security group will be used.

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
