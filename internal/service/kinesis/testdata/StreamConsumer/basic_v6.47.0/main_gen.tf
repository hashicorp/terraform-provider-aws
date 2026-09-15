# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_kinesis_stream_consumer" "test" {
  name       = var.rName
  stream_arn = aws_kinesis_stream.test.arn
}

resource "aws_kinesis_stream" "test" {
  name        = var.rName
  shard_count = 2
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
      version = "6.47.0"
    }
  }
}

provider "aws" {}
