terraform {
  required_version = ">= 0.12"
}

provider "aws" {
  region = var.aws_region
}

resource "aws_s3_bucket" "example" {
  bucket_prefix = "tf-ivs-stream-archive"
  force_destroy = true
}

resource "aws_ivs_recording_configuration" "example" {
  name = "tf-ivs-recording-configuration"
  destination_configuration {
    s3 {
      bucket_name = aws_s3_bucket.example.id
    }
  }
}

resource "aws_ivs_channel" "example" {
  name                        = "tf-ivs-channel"
  recording_configuration_arn = aws_ivs_recording_configuration.example.arn
}

data "aws_ivs_stream_key" "example" {
  channel_arn = aws_ivs_channel.example.arn
}
