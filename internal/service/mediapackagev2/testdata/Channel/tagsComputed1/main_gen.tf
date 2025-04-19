# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

provider "null" {}

resource "aws_media_packagev2_channel_group" "test" {
  name = var.rName

  tags = {
    (var.unknownTagKey) = null_resource.test.id
  }
}

resource "aws_media_packagev2_channel" "test" {
  channel_group_name = aws_media_packagev2_channel_group.test.name
  name               = var.rName

  tags = {
    (var.unknownTagKey) = null_resource.test.id
  }
}

resource "null_resource" "test" {}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "unknownTagKey" {
  type     = string
  nullable = false
}
