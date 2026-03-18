---
subcategory: "Elemental MediaConnect"
layout: "aws"
page_title: "AWS: aws_mediaconnect_flow"
description: |-
  Provides an AWS Elemental MediaConnect Flow resource.
---

# Resource: aws_mediaconnect_flow

Provides an AWS Elemental MediaConnect Flow resource.

## Example Usage

### Basic SRT Listener Flow

```terraform
resource "aws_mediaconnect_flow" "example" {
  name = "example-flow"

  source {
    name        = "example-source"
    description = "Example SRT listener source"
    protocol    = "srt-listener"
    ingest_port = 5000
  }
}
```

### Flow with RTP Source and Whitelist

```terraform
resource "aws_mediaconnect_flow" "example" {
  name = "example-flow"

  source {
    name           = "example-source"
    description    = "Example RTP source"
    protocol       = "rtp"
    whitelist_cidr = "10.0.0.0/16"
  }
}
```

### Flow with Output

```terraform
resource "aws_mediaconnect_flow" "example" {
  name = "example-flow"

  source {
    name        = "example-source"
    protocol    = "srt-listener"
    ingest_port = 5000
  }

  output {
    name        = "example-output"
    description = "Example output"
    protocol    = "srt-listener"
    port        = 5001
  }
}
```

### Flow with Entitlement

```terraform
resource "aws_mediaconnect_flow" "example" {
  name = "example-flow"

  source {
    name        = "example-source"
    protocol    = "srt-listener"
    ingest_port = 5000
  }

  entitlement {
    name        = "example-entitlement"
    description = "Example entitlement"
    subscriber  = ["123456789012"]
  }
}
```

### Flow with Source Failover (Multiple Sources)

```terraform
resource "aws_mediaconnect_flow" "example" {
  name = "example-flow"

  source {
    name        = "primary-source"
    description = "Primary source"
    protocol    = "srt-listener"
    ingest_port = 5000
  }

  source {
    name        = "backup-source"
    description = "Backup source"
    protocol    = "srt-listener"
    ingest_port = 5001
  }

  source_failover_config {
    failover_mode   = "FAILOVER"
    recovery_window = 200
    state           = "ENABLED"

    source_priority {
      primary_source = "primary-source"
    }
  }
}
```

### Flow Started on Creation

```terraform
resource "aws_mediaconnect_flow" "example" {
  name       = "example-flow"
  start_flow = true

  source {
    name        = "example-source"
    protocol    = "srt-listener"
    ingest_port = 5000
  }
}
```

### Flow with Source Monitoring

```terraform
resource "aws_mediaconnect_flow" "example" {
  name = "example-flow"

  source {
    name        = "example-source"
    protocol    = "srt-listener"
    ingest_port = 5000
  }

  source_monitoring_config {
    thumbnail_state                = "ENABLED"
    content_quality_analysis_state = "ENABLED"

    video_monitoring_setting {
      black_frames {
        state             = "ENABLED"
        threshold_seconds = 5
      }

      frozen_frames {
        state             = "ENABLED"
        threshold_seconds = 5
      }
    }
  }
}
```

### Flow with Maintenance Window

```terraform
resource "aws_mediaconnect_flow" "example" {
  name = "example-flow"

  source {
    name        = "example-source"
    protocol    = "srt-listener"
    ingest_port = 5000
  }

  maintenance {
    day        = "Monday"
    start_hour = "02:00"
  }
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required, Forces new resource) The name of the flow.
* `source` - (Required) One or more source blocks for the flow. At least one is required. For multi-source failover, specify multiple `source` blocks along with `source_failover_config`. See [Source](#source) below.

The following arguments are optional:

* `availability_zone` - (Optional, Forces new resource) The Availability Zone that you want to create the flow in.
* `entitlement` - (Optional) The entitlements for the flow. See [Entitlement](#entitlement) below.
* `flow_size` - (Optional) The processing capacity of the flow. Valid values: `MEDIUM`, `LARGE`, `LARGE_4X`.
* `maintenance` - (Optional) The maintenance settings for the flow. See [Maintenance](#maintenance) below.
* `media_stream` - (Optional) The media streams associated with the flow. See [Media Stream](#media-stream) below.
* `output` - (Optional) The outputs for the flow. See [Output](#output) below.
* `source_failover_config` - (Optional) The source failover configuration. See [Source Failover Config](#source-failover-config) below.
* `source_monitoring_config` - (Optional) The source monitoring configuration. See [Source Monitoring Config](#source-monitoring-config) below.
* `start_flow` - (Optional) Whether to start the flow after creation. Defaults to `false`. When set to `true`, the flow will be started (status `ACTIVE`). When set to `false`, the flow remains in `STANDBY`.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `vpc_interface` - (Optional) The VPC interfaces for this flow. See [VPC Interface](#vpc-interface) below.

### Source

* `name` - (Required, Forces new resource) The name of the source.
* `description` - (Optional) A description for the source.
* `protocol` - (Optional) The protocol that is used by the source. Valid values: `zixi-push`, `rtp-fec`, `rtp`, `zixi-pull`, `rist`, `st2110-jpegxs`, `cdi`, `srt-listener`, `srt-caller`, `udp`, `ndi-speed-hq`.
* `whitelist_cidr` - (Optional) The range of IP addresses that should be allowed to contribute content to your source, in CIDR notation (e.g., `10.0.0.0/16`).
* `ingest_port` - (Optional) The port that the flow will be listening on for incoming content.
* `max_bitrate` - (Optional) The smoothing max bitrate (in bps) for RIST, RTP, and RTP-FEC streams.
* `max_latency` - (Optional) The maximum latency in milliseconds for RIST-based and Zixi-based streams.
* `min_latency` - (Optional) The minimum latency in milliseconds for SRT-based streams.
* `max_sync_buffer` - (Optional) The size of the buffer (in milliseconds) to use to sync incoming source data.
* `sender_control_port` - (Optional) The port that the flow uses to send outbound requests to initiate connection with the sender.
* `sender_ip_address` - (Optional) The IP address that the flow communicates with to initiate connection with the sender.
* `stream_id` - (Optional) The stream ID for Zixi and SRT caller-based streams.
* `listener_address` - (Optional) Source IP or domain name for SRT-caller protocol.
* `listener_port` - (Optional) Source port for SRT-caller protocol.
* `entitlement_arn` - (Optional) The ARN of the entitlement that allows you to subscribe to content from another AWS account.
* `vpc_interface_name` - (Optional) The name of the VPC interface to use for this source.
* `decryption` - (Optional) The decryption settings for the source. See [Encryption](#encryption) below.
* `gateway_bridge_source` - (Optional) The source configuration for a flow receiving from a bridge. See [Gateway Bridge Source](#gateway-bridge-source) below.
* `media_stream_source_configuration` - (Optional) The media stream source configurations. See [Media Stream Source Configuration](#media-stream-source-configuration) below.

### Encryption

The `encryption` and `decryption` blocks share the same structure:

* `role_arn` - (Required) The ARN of the role that you created during setup for encryption.
* `algorithm` - (Required) The type of algorithm for encryption. Valid values: `aes128`, `aes192`, `aes256`.
* `key_type` - (Optional) The type of key used for encryption. Valid values: `speke`, `static-key`, `srt-password`.
* `secret_arn` - (Optional) The ARN of the secret in Secrets Manager that stores the encryption key (for static-key encryption).
* `constant_initialization_vector` - (Optional) A 128-bit, 16-byte hex value for use with the key.
* `device_id` - (Optional) The device ID for SPEKE encryption.
* `region` - (Optional) The AWS Region for the API Gateway proxy endpoint (for SPEKE).
* `resource_id` - (Optional) An identifier for the content (for SPEKE).
* `url` - (Optional) The URL from the API Gateway proxy for the key server (for SPEKE).

### Gateway Bridge Source

* `arn` - (Required) The ARN of the bridge feeding this flow.
* `vpc_interface_attachment` - (Optional) The VPC interface attachment. See [VPC Interface Attachment](#vpc-interface-attachment) below.

### VPC Interface Attachment

* `name` - (Optional) The name of the VPC interface.

### Media Stream Source Configuration

* `encoding_name` - (Required) The encoding name. Valid values: `jxsv`, `raw`, `smpte291`, `pcm`.
* `name` - (Required) The name of the media stream.
* `input_configuration` - (Optional) The input configurations. See [Input Configuration](#input-configuration) below.

### Input Configuration

* `port` - (Required) The port for the incoming media stream.
* `ip` - (Optional) The IP address for the incoming media stream.
* `interface` - (Required) The VPC interface. See [Interface](#interface) below.

### Entitlement

* `name` - (Required, Forces new resource) The name of the entitlement.
* `subscriber` - (Required) The AWS account IDs that are allowed to subscribe to the flow.
* `data_transfer_subscriber_fee_percent` - (Optional, Forces new resource) Percentage (0-100) of the data transfer cost to be billed to the subscriber.
* `description` - (Optional) A description for the entitlement.
* `status` - (Optional) The status of the entitlement. Valid values: `ENABLED`, `DISABLED`. Defaults to `ENABLED`.
* `encryption` - (Optional) The encryption settings. See [Encryption](#encryption) above.

### Output

* `name` - (Optional) The name of the output.
* `description` - (Optional) A description for the output.
* `protocol` - (Optional) The protocol for the output. Valid values: `zixi-push`, `rtp-fec`, `rtp`, `zixi-pull`, `rist`, `st2110-jpegxs`, `cdi`, `srt-listener`, `srt-caller`, `udp`, `ndi-speed-hq`.
* `port` - (Optional) The port to use for the output.
* `destination` - (Optional) The IP address where content is sent.
* `cidr_allow_list` - (Optional) The range of IP addresses allowed to initiate output requests.
* `max_latency` - (Optional) The maximum latency in milliseconds.
* `min_latency` - (Optional) The minimum latency in milliseconds.
* `smoothing_latency` - (Optional) The smoothing latency in milliseconds.
* `stream_id` - (Optional) The stream ID for Zixi and SRT caller-based streams.
* `remote_id` - (Optional) The remote ID for the Zixi-pull stream.
* `sender_control_port` - (Optional) The port for outbound requests to initiate connection with the sender.
* `sender_ip_address` - (Optional) The IP address for outbound requests.
* `status` - (Optional) The status of the output. Valid values: `ENABLED`, `DISABLED`.
* `encryption` - (Optional) The encryption settings. See [Encryption](#encryption) above.
* `media_stream_output_configuration` - (Optional) The media stream output configurations. See [Media Stream Output Configuration](#media-stream-output-configuration) below.
* `vpc_interface_attachment` - (Optional) The VPC interface attachment. See [VPC Interface Attachment](#vpc-interface-attachment) above.

### Media Stream Output Configuration

* `encoding_name` - (Required) The encoding name. Valid values: `jxsv`, `raw`, `smpte291`, `pcm`.
* `name` - (Required) The name of the media stream.
* `destination_configuration` - (Optional) The destination configurations. See [Destination Configuration](#destination-configuration) below.
* `encoding_parameters` - (Optional) The encoding parameters. See [Encoding Parameters](#encoding-parameters) below.

### Destination Configuration

* `ip` - (Required) The IP address to send the media stream to.
* `port` - (Required) The port to send the media stream to.
* `interface` - (Required) The VPC interface. See [Interface](#interface) below.
* `outbound_ip` - The outbound IP address (computed).

### Encoding Parameters

* `compression_factor` - (Required) The compression factor (3.0 to 10.0).
* `encoder_profile` - (Required) The encoder profile. Valid values: `main`, `high`.

### Interface

* `name` - (Required) The name of the VPC interface.

### Media Stream

* `id` - (Required, Forces new resource) The unique identifier for the media stream.
* `name` - (Required) The name of the media stream.
* `type` - (Required) The type of media stream. Valid values: `video`, `audio`, `ancillary-data`.
* `clock_rate` - (Optional) The clock rate (Hz) of the media stream. Valid values: `48000`, `90000`, `96000`.
* `description` - (Optional) A description for the media stream.
* `video_format` - (Optional) The video format of the media stream.
* `attributes` - (Optional) The attributes of the media stream. See [Media Stream Attributes](#media-stream-attributes) below.

### Media Stream Attributes

* `lang` - (Optional) The language of the media stream.
* `fmtp` - (Optional) The format type parameters. See [FMTP](#fmtp) below.

### FMTP

* `channel_order` - (Optional) The format of the audio channel.
* `colorimetry` - (Optional) The colorimetry. Valid values: `BT601`, `BT709`, `BT2020`, `BT2100`, `ST2065-1`, `ST2065-3`, `XYZ`.
* `exact_framerate` - (Optional) The frame rate (e.g., `60000/1001`).
* `par` - (Optional) The pixel aspect ratio.
* `range` - (Optional) The range. Valid values: `NARROW`, `FULL`, `FULLPROTECT`.
* `scan_mode` - (Optional) The scan mode. Valid values: `progressive`, `interlace`, `progressive-segmented-frame`.
* `tcs` - (Optional) The transfer characteristic system. Valid values: `SDR`, `PQ`, `HLG`, `LINEAR`, `BT2100LINPQ`, `BT2100LINHLG`, `ST2065-1`, `ST428-1`, `DENSITY`.

### Source Failover Config

* `failover_mode` - (Optional) The type of failover. Valid values: `MERGE`, `FAILOVER`.
* `recovery_window` - (Optional) The size of the search window, in milliseconds.
* `state` - (Optional) Whether failover is enabled. Valid values: `ENABLED`, `DISABLED`.
* `source_priority` - (Optional) The priority settings for the source. See [Source Priority](#source-priority) below.

### Source Priority

* `primary_source` - (Optional) The name of the source to use as the primary source.

### Source Monitoring Config

* `content_quality_analysis_state` - (Optional) Whether content quality analysis is enabled. Valid values: `ENABLED`, `DISABLED`.
* `thumbnail_state` - (Optional) Whether thumbnail generation is enabled. Valid values: `ENABLED`, `DISABLED`.
* `audio_monitoring_setting` - (Optional) The audio monitoring settings. See [Audio Monitoring Setting](#audio-monitoring-setting) below.
* `video_monitoring_setting` - (Optional) The video monitoring settings. See [Video Monitoring Setting](#video-monitoring-setting) below.

### Audio Monitoring Setting

* `silent_audio` - (Optional) The silent audio detection settings. See [Silent Audio](#silent-audio) below.

### Silent Audio

* `state` - (Optional) Whether silent audio detection is enabled. Valid values: `ENABLED`, `DISABLED`.
* `threshold_seconds` - (Optional) The number of consecutive seconds of silence that triggers an event.

### Video Monitoring Setting

* `black_frames` - (Optional) The black frame detection settings. See [Black Frames](#black-frames) below.
* `frozen_frames` - (Optional) The frozen frame detection settings. See [Frozen Frames](#frozen-frames) below.

### Black Frames

* `state` - (Optional) Whether black frame detection is enabled. Valid values: `ENABLED`, `DISABLED`.
* `threshold_seconds` - (Optional) The number of consecutive seconds of black frames that triggers an event.

### Frozen Frames

* `state` - (Optional) Whether frozen frame detection is enabled. Valid values: `ENABLED`, `DISABLED`.
* `threshold_seconds` - (Optional) The number of consecutive seconds of frozen frames that triggers an event.

### Maintenance

* `day` - (Required) The day of the week for the maintenance window. Valid values: `Monday`, `Tuesday`, `Wednesday`, `Thursday`, `Friday`, `Saturday`, `Sunday`.
* `start_hour` - (Required) The hour to start maintenance, in 24-hour `HH:MM` format (e.g., `02:00`).
* `scheduled_date` - (Optional) The scheduled date for maintenance in ISO UTC format (YYYY-MM-DD).

### VPC Interface

* `name` - (Required) A unique name for the VPC interface.
* `role_arn` - (Required) The ARN of the role for creating ENIs in the VPC.
* `security_group_ids` - (Required) The security group IDs for the ENI (1-16 items).
* `subnet_id` - (Required) The subnet to place the ENI in.
* `network_interface_type` - (Optional) The type of network interface. Valid values: `ena`, `efa`. Defaults to `ena`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The ARN of the flow.
* `availability_zone` - The Availability Zone the flow was created in.
* `description` - A description of the flow.
* `egress_ip` - The IP address from which video will be sent to output destinations.
* `id` - The ARN of the flow.
* `status` - The current status of the flow (e.g., `STANDBY`, `ACTIVE`).
* `source.0.arn` - The ARN of the source.
* `source.0.data_transfer_subscriber_fee_percent` - The data transfer subscriber fee percent.
* `source.0.ingest_ip` - The IP address that the flow will be listening on for incoming content.
* `source.0.peer_ip_address` - The IP address of the device currently sending content to this source.
* `entitlement.*.arn` - The ARN of the entitlement.
* `output.*.arn` - The ARN of the output.
* `output.*.bridge_arn` - The ARN of the bridge attached to the output.
* `output.*.bridge_ports` - The bridge ports for the output.
* `output.*.data_transfer_subscriber_fee_percent` - The data transfer subscriber fee percent.
* `output.*.entitlement_arn` - The ARN of the entitlement on the output.
* `output.*.listener_address` - The listener address of the output.
* `output.*.media_live_input_arn` - The ARN of the MediaLive input attached to the output.
* `output.*.peer_ip_address` - The IP address of the device currently receiving content from this output.
* `media_stream.*.fmt` - The format type number of the media stream.
* `maintenance.0.deadline` - The deadline for the maintenance window.
* `maintenance.0.scheduled_date` - The scheduled date for maintenance.
* `vpc_interface.*.network_interface_ids` - The IDs of the ENIs created by MediaConnect.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

MediaConnect Flow can be imported using the flow ARN, e.g.,

```
$ terraform import aws_mediaconnect_flow.example arn:aws:mediaconnect:us-east-1:123456789012:flow:1-23aBC45dEF67hiJ8-12AbC34dEf56:example-flow
```
