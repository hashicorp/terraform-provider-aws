---
subcategory: "Elemental MediaConnect"
layout: "aws"
page_title: "AWS: aws_mediaconnect_flow"
description: |-
  Terraform resource for managing an AWS Elemental MediaConnect Flow.
---
# Resource: aws_mediaconnect_flow

Terraform resource for managing an AWS Elemental MediaConnect Flow.

## Example Usage

### Basic Usage

```terraform
resource "aws_mediaconnect_flow" "example" {
  name        = "example-flow"
  description = "description for example flow"

  source {
    description    = "A MediaConnect Flow"
    protocol       = "rtp"
    ingest_port    = 3010
    whitelist_cidr = "10.24.34.0/23"
  }
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) The name of the flow.
* `source` - (Required) The settings for the source of the flow. See [Sources](#sources) for more details.

The following arguments are optional:

* `availability_zone` - (Optional) The Availability Zone that you want to create the flow in. These options are limited to the Availability Zones within the current AWS Region.
* `entitlement` - (Optional) The entitlements that you want to grant on a flow. See [Entitlements](#entitlements) for more details.
* `maintenance` - (Optional) Create maintenance setting for a flow. See [Maintenance](#maintenance) for more details.
* `media_stream` - (Optional) The media streams that you want to add to the flow. You can associate these media streams with sources and outputs on the flow. See [Media Streams](#media-streams) for more details.
* `output` - (Optional) The outputs that you want to add to this flow. See [Outputs](#outputs) for more details.
* `source_failover_config` - (Optional) The settings for source failover. See [Source Failover Configuration](#source-failover-configuration) for more details.
* `vpc_interface` - (Optional) The VPC interfaces you want on the flow. See [VPC Interfaces](#vpc-interfaces) for more details.
* `start_flow` - (Optional) Whether to start/stop flow. Default: `false`
* `tags` - (Optional) A map of tags to assign to the flow. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### Sources

* `decryption` - (Optional) The type of encryption that is used on the content ingested from this source. Allowable encryption types: static-key. See [Encryption](#encryption) for more details.
* `description` - (Required) A description for the source. This value is not used or seen outside of the current AWS Elemental MediaConnect account.
* `gateway_bridge_source` - (Optional) The source configuration for cloud flows receiving a stream from a bridge. See [Gateway Bridge Source](#gateway-bridge-source) for more details.
* `ingest_port` - (Optional) The port that the flow will be listening on for incoming content.
* `max_bitrate` - (Optional) The smoothing max bitrate (in bps) for RIST, RTP, and RTP-FEC streams.
* `max_latency` - (Optional) The maximum latency in milliseconds. This parameter applies only to RIST-based, Zixi-based, and Fujitsu-based streams.
* `max_sync_buffer` - (Optional) The size of the buffer (in milliseconds) to use to sync incoming source data.
* `media_stream_source_configurations` - (Optional) The media streams that are associated with the source, and the parameters for those associations. See [Media Stream Source Configurations](#media-stream-source-configurations) for more details.
* `min_latency` - (Optional) The minimum latency in milliseconds for SRT-based streams. In streams that use the SRT protocol, this value that you set on your MediaConnect source or output represents the minimal potential latency of that connection. The latency of the stream is set to the highest number between the sender’s minimum latency and the receiver’s minimum latency.
* `name` - (Required) The name of the source.
* `protocol` - (Optional) The protocol that is used by the source.
* `sender_control_port` - (Optional) The port that the flow uses to send outbound requests to initiate connection with the sender.
* `sender_ip_address` - (Optional) The IP address that the flow communicates with to initiate connection with the sender.
* `listener_address` - (Optional) Source IP or domain name for SRT-caller protocol.
* `listener_port` - (Optional) Source port for SRT-caller protocol.
* `stream_id` - (Optional) The stream ID that you want to use for this transport. This parameter applies only to Zixi and SRT caller-based streams.
* `vpc_interface_name` - (Optional) The name of the VPC interface to use for this source.
* `whitelist_cidr` - (Optional) The range of IP addresses that should be allowed to contribute content to your source. These IP addresses should be in the form of a Classless Inter-Domain Routing (CIDR) block; for example, 10.0.0.0/16.

### Gateway Bridge Source

* `arn` - (Required) The ARN of the bridge feeding this flow.
* `vpc_interface_attachment` - (Optional) The name of the VPC interface attachment to use for this bridge source. See [VPC Interface Attachment](#vpc-interface-attachment) for more details.

### VPC Interface Attachment

* `vpc_interface_name` - (Optional) The name of the VPC interface to use for this resource.

### Media Stream Source Configurations

* `encoding_name` - (Required) The format you want to use to encode the data. For ancillary data streams, set the encoding name to smpte291. For audio streams, set the encoding name to pcm. For video, 2110 streams, set the encoding name to raw. For video, JPEG XS streams, set the encoding name to jxsv.
* `name` - (Required) The name of the media stream.
* `input_configurations` - (Optional) The transport parameters that you want to associate with the media stream. See [Input Configurations](#input-configurations) for more details.

### Input Configurations

* `port` - (Required) The port that you want the flow to listen on for an incoming media stream.
* `interface` - (Required) The VPC interface that you want to use for the incoming media stream. See [VPC Interface](#vpc-interface) for more details.

### Entitlements

* `subscribers` - (Required) The AWS account IDs that you want to share your content with. The receiving accounts (subscribers) will be allowed to create their own flows using your content as the source.
* `data_transfer_subscriber_fee_percent` - (Optional) Percentage from 0-100 of the data transfer cost to be billed to the subscriber.
* `description` - (Required) A description of the entitlement. This description appears only on the AWS Elemental MediaConnect console and will not be seen by the subscriber or end user.
* `encryption` - (Optional) The type of encryption that will be used on the output that is associated with this entitlement. Allowable encryption types: static-key, speke. See [Encryption](#encryption) for more details.
* `status` - (Optional) An indication of whether the new entitlement should be enabled or disabled as soon as it is created. If you don’t specify the entitlementStatus field in your request, MediaConnect sets it to ENABLED.
* `name` - (Required) The name of the entitlement. This value must be unique within the current flow.

### Encryption

* `role_arn` - (Required) The ARN of the role that you created during setup (when you set up AWS Elemental MediaConnect as a trusted entity).
* `algorithm` - (Required) The type of algorithm that is used for the encryption (such as aes128, aes192, or aes256).
* `constant_initialization_vector` - (Optional) A 128-bit, 16-byte hex value represented by a 32-character string, to be used with the key for encrypting content. This parameter is not valid for static key encryption.
* `device_id` - (Optional) The value of one of the devices that you configured with your digital rights management (DRM) platform key provider. This parameter is required for SPEKE encryption and is not valid for static key encryption.
* `key_type` - (Optional) The type of key that is used for the encryption. If no keyType is provided, the service will use the default setting (static-key).
* `region` - (Optional) The AWS Region that the API Gateway proxy endpoint was created in. This parameter is required for SPEKE encryption and is not valid for static key encryption.
* `resource_id` - (Optional) An identifier for the content. The service sends this value to the key server to identify the current endpoint. The resource ID is also known as the content ID. This parameter is required for SPEKE encryption and is not valid for static key encryption.
* `secret_arn` - (Optional) The ARN of the secret that you created in AWS Secrets Manager to store the encryption key. This parameter is required for static key encryption and is not valid for SPEKE encryption.
* `url` - (Optional) The URL from the API Gateway proxy that you set up to talk to your key server. This parameter is required for SPEKE encryption and is not valid for static key encryption.

### Maintenance

* `day` - (Required) The day of the week to use for maintenance.
* `start_time` - (Optional) The hour maintenance will start.

### Media Streams

* `id` - (Required) A unique identifier for the media stream.
* `name` - (Required) A name that helps you distinguish one media stream from another.
* `type` - (Required) The type of media stream.
* `attributes` - (Optional) The attributes that you want to assign to the new media stream. See [Media Stream Attributes](#media-stream-attributes) for more details.
* `clock_rate` - (Optional) The sample rate (in Hz) for the stream. If the media stream type is video or ancillary data, set this value to 90000. If the media stream type is audio, set this value to either 48000 or 96000.
* `description` - (Optional) A description that can help you quickly identify what your media stream is used for.
* `video_format` - (Optional) The resolution of the video.

### Media Stream Attributes

* `fmtp` - (Optional) The settings that you want to use to define the media stream. See [Fmtp](#fmtp) for more details.
* `lang` - (Optional) The audio language, in a format that is recognized by the receiver.

### Fmtp

* `channel_order` - (Optional) The format of the audio channel.
* `colorimetry` - (Optional) The format that is used for the representation of color.
* `exact_framerate` - (Optional) The frame rate for the video stream, in frames/second. For example: 60000/1001. If you specify a whole number, MediaConnect uses a ratio of N/1. For example, if you specify 60, MediaConnect uses 60/1 as the exactFramerate.
* `par` - (Optional) The pixel aspect ratio (PAR) of the video.
* `range` - (Optional) The encoding range of the video.
* `scan_mode` - (Optional) The type of compression that was used to smooth the video’s appearance.
* `tcs` - (Optional) The transfer characteristic system (TCS) that is used in the video.

### Outputs

* `protocol` - (Required) The protocol to use for the output.
* `cidr_allow_list` - (Optional) The range of IP addresses that should be allowed to initiate output requests to this flow. These IP addresses should be in the form of a Classless Inter-Domain Routing (CIDR) block; for example, 10.0.0.0/16.
* `description` - (Optional) A description of the output. This description appears only on the AWS Elemental MediaConnect console and will not be seen by the end user.
* `destination` - (Optional) The IP address from which video will be sent to output destinations.
* `encryption` - (Optional) The type of key used for the encryption. If no keyType is provided, the service will use the default setting (static-key). Allowable encryption types: static-key. See [Encryption](#encryption) for more details.
* `max_latency` - (Optional) The maximum latency in milliseconds. This parameter applies only to RIST-based, Zixi-based, and Fujitsu-based streams.
* `media_stream_output_configurations` - (Optional) The media streams that are associated with the output, and the parameters for those associations. See [Media Stream Output Configurations](#media-stream-output-configurations) for more details.
* `min_latency` - (Optional) The minimum latency in milliseconds for SRT-based streams. In streams that use the SRT protocol, this value that you set on your MediaConnect source or output represents the minimal potential latency of that connection. The latency of the stream is set to the highest number between the sender’s minimum latency and the receiver’s minimum latency.
* `name` - (Optional) The name of the output. This value must be unique within the current flow.
* `port` - (Optional) The port to use when content is distributed to this output.
* `remote_id` - (Optional) The remote ID for the Zixi-pull output stream.
* `sender_control_port` - (Optional) The port that the flow uses to send outbound requests to initiate connection with the sender.
* `smoothing_latency` - (Optional) The smoothing latency in milliseconds for RIST, RTP, and RTP-FEC streams.
* `stream_id` - (Optional) The stream ID that you want to use for this transport. This parameter applies only to Zixi and SRT caller-based streams.
* `vpc_interface_attachment` - (Optional) The name of the VPC interface attachment to use for this output. See [VPC Interface Attachment](#vpc-interface-attachment) for more details.

### Media Stream Output Configurations

* `encoding_name` - (Required) The format that will be used to encode the data. For ancillary data streams, set the encoding name to smpte291. For audio streams, set the encoding name to pcm. For video, 2110 streams, set the encoding name to raw. For video, JPEG XS streams, set the encoding name to jxsv.
* `name` - (Required) The name of the media stream that is associated with the output.
* `desination_configurations` - (Optional) The transport parameters that you want to associate with the media stream. See [Destination Configurations](#destination-configurations) for more details.
* `encoding_parameters` - (Optional) A collection of parameters that determine how MediaConnect will convert the content. These fields only apply to outputs on flows that have a CDI source. See [Encoding Parameters](#encoding-parameters) for more details.

### Destination Configurations

* `ip` - (Required) The IP address where you want MediaConnect to send contents of the media stream.
* `port` - (Required) The port that you want MediaConnect to use when it distributes the media stream to the output.
* `interface` - (Required) The VPC interface that you want to use for the media stream associated with the output. See [VPC Interface](#vpc-interface) for more details.

### VPC Interfaces

* `name` - (Required) The name of the VPC Interface. This value must be unique within the current flow.
* `role_arn` - (Required) Role ARN MediaConnect can assumes to create ENIs in customer's account
* `security_group_ids` - (Required) Security Group IDs to be used on ENI.
* `subnet_id` - (Required) Subnet must be in the AZ of the Flow.
* `network_interface_type` - (Optional) The type of network interface. If this value is not included in the request, MediaConnect uses ENA as the networkInterfaceType.

### VPC Interface

* `name` - (Required) The name of the VPC interface.

### Encoding Parameters

* `compression_factor` - (Required) A value that is used to calculate compression for an output. The bitrate of the output is calculated as follows: Output bitrate = (1 / compressionFactor) \* (source bitrate) This property only applies to outputs that use the ST 2110 JPEG XS protocol, with a flow source that uses the CDI protocol. Valid values are floating point numbers in the range of 3.0 to 10.0, inclusive.
* `encoder_profile` - (Required) A setting on the encoder that drives compression settings. This property only applies to video media streams associated with outputs that use the ST 2110 JPEG XS protocol, if at least one source on the flow uses the CDI protocol.

### Source Failover Configuration

* `failover_mode` - (Optional) The type of failover you choose for this flow. MERGE combines the source streams into a single stream, allowing graceful recovery from any single-source loss. FAILOVER allows switching between different streams.
* `recovery_window` - (Optional) Search window time to look for dash-7 packets.
* `source_priority` - (Optional) The priority you want to assign to a source. You can have a primary stream and a backup stream or two equally prioritized streams. See [Source Priority] for more details.
* `state` - (Optional)

### Source Priority

* `primary_source` - (Required) The name of the source you choose as the primary source for this flow.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Flow.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `60m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Elemental MediaConnect Flow using the `example_id_arg`. For example:

```terraform
import {
  to = aws_mediaconnect_flow.example
  id = "arn:aws:mediaconnect:us-west-2:123456789012:flow:1-7e7a28d2-163f-4b8f:tfflow"
}
```

Using `terraform import`, import Elemental MediaConnect Flow using the `example_id_arg`. For example:

```console
% terraform import aws_mediaconnect_flow.example arn:aws:mediaconnect:us-west-2:123456789012:flow:1-7e7a28d2-163f-4b8f:tfflow
```
