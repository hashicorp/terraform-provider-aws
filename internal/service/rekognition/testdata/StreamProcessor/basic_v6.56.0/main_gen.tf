# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_rekognition_stream_processor" "test" {
  role_arn = aws_iam_role.test.arn
  name     = var.rName

  data_sharing_preference {
    opt_in = true
  }

  output {
    s3_destination {
      bucket = aws_s3_bucket.test.bucket
    }
  }

  settings {
    connected_home {
      labels = ["PERSON", "ALL"]
    }
  }

  input {
    kinesis_video_stream {
      arn = aws_kinesis_video_stream.test.arn
    }
  }

  notification_channel {
    sns_topic_arn = aws_sns_topic.test.arn
  }
}

resource "aws_iam_role" "test" {
  name = var.rName

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Sid    = ""
        Principal = {
          Service = "ec2.amazonaws.com"
        }
      },
    ]
  })
}

resource "aws_s3_bucket" "test" {
  bucket = var.rName
}

resource "aws_sns_topic" "test" {
  name = var.rName
}

resource "aws_kinesis_video_stream" "test" {
  name                    = var.rName
  data_retention_in_hours = 1
  device_name             = "kinesis-video-device-name"
  media_type              = "video/h264"
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "6.56.0"
    }
  }
}

provider "aws" {}
