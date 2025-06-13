---
subcategory: "Elemental MediaPackage"
layout: "aws"
page_title: "AWS: aws_media_package_origin_endpoint"
description: |-
  Provides an AWS Elemental MediaPackage Origin Endpoint.
---

# Resource: aws_media_package_origin_endpoint

Provides an AWS Elemental MediaPackage Origin Endpoint.

## Example Usage

```terraform
resource "aws_media_package_origin_endpoint" "test" {
  channel_id = aws_media_package_channel.test.id
  id         = "test-origin-endpoint"

  authorization {
    cdn_identifier_secret = "example-cdn-identifier-secret"
    secrets_role_arn      = "arn:aws:iam::123456789012:role/example-role"
  }

  cmaf_package {
    encryption {
      speke_key_provider {
        resource_id = "example-resource-id"
        role_arn    = "arn:aws:iam::123456789012:role/example-role"
        system_ids  = ["example-system-id"]
        url         = "https://example.com"
      }
    }
    hls_manifests {
      id = "example-hls-manifest-id"
    }
    segment_duration_seconds = 6
    segment_prefix           = "example-segment-prefix"
    stream_selection {
      max_video_bits_per_second = 1000000
      min_video_bits_per_second = 500000
      stream_order              = "ORIGINAL"
    }
  }
  
  hls_package {
    ad_markers = "PASSTHROUGH"
    encryption {
      speke_key_provider {
        resource_id = "example-resource-id"
        role_arn    = "arn:aws:iam::123456789012:role/example-role"
        system_ids  = ["example-system-id"]
        url         = "https://example.com"
      }
    }
    include_iframe_only_stream = true
    playlist_type              = "EVENT"
    playlist_window_seconds    = 60
    program_date_time_interval_seconds = 600
    segment_duration_seconds   = 6
    stream_selection {
      max_video_bits_per_second = 1000000
      min_video_bits_per_second = 500000
      stream_order              = "ORIGINAL"
    }
    use_audio_rendition_group = true
  }

  mss_package {
    encryption {
      speke_key_provider {
        resource_id = "example-resource-id"
        role_arn    = "arn:aws:iam::123456789012:role/example-role"
        system_ids  = ["example-system-id"]
        url         = "https://example.com"
      }
    }
    manifest_window_seconds = 60
    segment_duration_seconds = 6
    stream_selection {
      max_video_bits_per_second = 1000000
      min_video_bits_per_second = 500000
      stream_order              = "ORIGINAL"
    }
  }

  dash_package {
    ad_triggers = ["SPLICE_INSERT"]
    ads_on_delivery_restrictions = "NONE"
    encryption {
      speke_key_provider {
        resource_id = "example-resource-id"
        role_arn    = "arn:aws:iam::123456789012:role/example-role"
        system_ids  = ["edef8ba9-79d6-4ace-a3c8-27dcd51d21ed"]
        url         = "https://example.com/hoge"
        encryption_contract_configuration {
          preset_speke20_audio = "PRESET-AUDIO-1"
          preset_speke20_video = "PRESET-VIDEO-1"
        }
      }
      key_rotation_interval_seconds = 0
    }
    include_iframe_only_stream = true
    manifest_layout            = "FULL"
    manifest_window_seconds    = 60
    min_buffer_time_seconds    = 60
    min_update_period_seconds  = 60
    period_triggers = ["ADS"]
    profile = "NONE"
    segment_duration_seconds = 6
    stream_selection {
      max_video_bits_per_second = 1000000
      min_video_bits_per_second = 500000
      stream_order              = "ORIGINAL"
    }
    suggested_presentation_delay_seconds = 60
    utc_timing = "HTTP-HEAD"
    utc_timing_uri = "https://example.com/hoge"
  }

  description   = "example description"
  manifest_name = "example-manifest-name"
  origination   = "ALLOW"
  startover_window_seconds = 3600
  tags = {
    Name = "example"
  }
  time_delay_seconds = 60
  whitelist = ["192.168.0.1/24"]
}
```

## Argument Reference

This resource supports the following arguments:
> Only one of cmaf_package, hls_package, mss_package, dash_package needs to be defined

* `channel_id` - (Required) The ID of the channel that the origin endpoint is associated with.
* `id` - (Required) A unique identifier describing the origin endpoint.
* `authorization` - (Optional) Parameters for CDN authorization.
    * `cdn_identifier_secret` - (Required) The secret to authenticate the CDN.
    * `secrets_role_arn` - (Required) The ARN of the role that allows MediaPackage to access the secret.
* `cmaf_package` - (Optional) Parameters for the CMAF package.
    * `encryption` - (Optional) Parameters for encryption.
        * `speke_key_provider` - (Required) Parameters for the SPEKE key provider.
            * `resource_id` - (Required) The resource ID.
            * `role_arn` - (Required) The ARN of the role that allows MediaPackage to access the key provider.
            * `system_ids` - (Required) A list of system IDs.
            * `url` - (Required) The URL of the key provider.
            * `certificate_arn` - (Optional) The ARN of the certificate.
            * `encryption_contract_configuration` - (Optional) Parameters for the encryption contract configuration.
                * `preset_speke20_audio` - (Optional) The preset for SPEKE 2.0 audio.
                * `preset_speke20_video` - (Optional) The preset for SPEKE 2.0 video.
    * `hls_manifests` - (Optional) A list of HLS manifest parameters.
        * `id` - (Required) The ID of the HLS manifest.
    * `segment_duration_seconds` - (Optional) The duration (in seconds) of each segment.
    * `segment_prefix` - (Optional) The prefix for each segment.
    * `stream_selection` - (Optional) Parameters for stream selection.
        * `max_video_bits_per_second` - (Optional) The maximum video bitrate.
        * `min_video_bits_per_second` - (Optional) The minimum video bitrate.
        * `stream_order` - (Optional) The order of the streams.
* `hls_package` - (Optional) Parameters for the HLS package.
    * `ad_markers` - (Optional) The ad markers.
    * `encryption` - (Optional) Parameters for encryption.
        * `speke_key_provider` - (Required) Parameters for the SPEKE key provider.
            * `resource_id` - (Required) The resource ID.
            * `role_arn` - (Required) The ARN of the role that allows MediaPackage to access the key provider.
            * `system_ids` - (Required) A list of system IDs.
            * `url` - (Required) The URL of the key provider.
            * `certificate_arn` - (Optional) The ARN of the certificate.
            * `encryption_contract_configuration` - (Optional) Parameters for the encryption contract configuration.
                * `preset_speke20_audio` - (Optional) The preset for SPEKE 2.0 audio.
                * `preset_speke20_video` - (Optional) The preset for SPEKE 2.0 video.
    * `include_iframe_only_stream` - (Optional) Whether to include an I-frame only stream.
    * `playlist_type` - (Optional) The type of playlist.
    * `playlist_window_seconds` - (Optional) The duration (in seconds) of the playlist window.
    * `program_date_time_interval_seconds` - (Optional) The interval (in seconds) for program date time.
    * `segment_duration_seconds` - (Optional) The duration (in seconds) of each segment.
    * `stream_selection` - (Optional) Parameters for stream selection.
        * `max_video_bits_per_second` - (Optional) The maximum video bitrate.
        * `min_video_bits_per_second` - (Optional) The minimum video bitrate.
        * `stream_order` - (Optional) The order of the streams.
    * `use_audio_rendition_group` - (Optional) Whether to use an audio rendition group.
* `mss_package` - (Optional) Parameters for the MSS package.
    * `encryption` - (Optional) Parameters for encryption.
        * `speke_key_provider` - (Required) Parameters for the SPEKE key provider.
            * `resource_id` - (Required) The resource ID.
            * `role_arn` - (Required) The ARN of the role that allows MediaPackage to access the key provider.
            * `system_ids` - (Required) A list of system IDs.
            * `url` - (Required) The URL of the key provider.
            * `certificate_arn` - (Optional) The ARN of the certificate.
            * `encryption_contract_configuration` - (Optional) Parameters for the encryption contract configuration.
                * `preset_speke20_audio` - (Optional) The preset for SPEKE 2.0 audio.
                * `preset_speke20_video` - (Optional) The preset for SPEKE 2.0 video.
    * `manifest_window_seconds` - (Optional) The duration (in seconds) of the manifest window.
    * `segment_duration_seconds` - (Optional) The duration (in seconds) of each segment.
    * `stream_selection` - (Optional) Parameters for stream selection.
        * `max_video_bits_per_second` - (Optional) The maximum video bitrate.
        * `min_video_bits_per_second` - (Optional) The minimum video bitrate.
        * `stream_order` - (Optional) The order of the streams.
* `dash_package` - (Optional) Parameters for the DASH package.
    * `ad_triggers` - (Optional) A list of ad triggers.
    * `ads_on_delivery_restrictions` - (Optional) The ads on delivery restrictions.
    * `encryption` - (Optional) Parameters for encryption.
        * `speke_key_provider` - (Required) Parameters for the SPEKE key provider.
            * `resource_id` - (Required) The resource ID.
            * `role_arn` - (Required) The ARN of the role that allows MediaPackage to access the key provider.
            * `system_ids` - (Required) A list of system IDs.
            * `url` - (Required) The URL of the key provider.
            * `certificate_arn` - (Optional) The ARN of the certificate.
            * `encryption_contract_configuration` - (Optional) Parameters for the encryption contract configuration.
                * `preset_speke20_audio` - (Optional) The preset for SPEKE 2.0 audio.
                * `preset_speke20_video` - (Optional) The preset for SPEKE 2.0 video.
        * `key_rotation_interval_seconds` - (Optional) The interval (in seconds) for key rotation.
    * `include_iframe_only_stream` - (Optional) Whether to include an I-frame only stream.
    * `manifest_layout` - (Optional) The layout of the manifest.
    * `manifest_window_seconds` - (Optional) The duration (in seconds) of the manifest window.
    * `min_buffer_time_seconds` - (Optional) The duration (in seconds) of the minimum buffer time.
    * `min_update_period_seconds` - (Optional) The duration (in seconds) of the minimum update period.
    * `period_triggers` - (Optional) A list of period triggers.
    * `profile` - (Optional) The profile.
    * `segment_duration_seconds` - (Optional) The duration (in seconds) of each segment.
    * `segment_template_format` - (Optional) The format of the segment template.
    * `stream_selection` - (Optional) Parameters for stream selection.
        * `max_video_bits_per_second` - (Optional) The maximum video bitrate.
        * `min_video_bits_per_second` - (Optional) The minimum video bitrate.
        * `stream_order` - (Optional) The order of the streams.
    * `suggested_presentation_delay_seconds` - (Optional) The duration (in seconds) of the suggested presentation delay.
    * `utc_timing` - (Optional) The UTC timing.
    * `utc_timing_uri` - (Optional) The URI of the UTC timing.
* `description` - (Optional) A description of the origin endpoint.
* `manifest_name` - (Optional) The name of the manifest.
* `origination` - (Optional) The origination setting.
* `startover_window_seconds` - (Optional) The duration (in seconds) of the startover window.
* `tags` - (Optional) A map of tags to assign to the resource.
* `time_delay_seconds` - (Optional) The duration (in seconds) of the time delay.
* `whitelist` - (Optional) A list of IP ranges to whitelist.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The same as `id`
* `arn` - The ARN of the origin endpoint
* `url` - The URL of the origin endpoint

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Media Package Origin Endpoints using the origin endpoint ID. For example:

```terraform
import {
  to = aws_media_package_origin_endpoint.test
  id = "test-origin-endpoint"
}
```

Using `terraform import`, import Media Package Origin Endpoints using the origin endpoint ID. For example:

```console
% terraform import aws_media_package_origin_endpoint.test test-origin-endpoint
```
