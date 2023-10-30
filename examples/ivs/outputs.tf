# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

output "ingest_endpoint" {
  value = aws_ivs_channel.example.ingest_endpoint
}

output "stream_key" {
  value = data.aws_ivs_stream_key.example.value
}

output "playback_url" {
  value = aws_ivs_channel.example.playback_url
}
