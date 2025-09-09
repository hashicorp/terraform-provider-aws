# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

provider "null" {}

resource "aws_networkmanager_link" "test" {
  global_network_id = aws_networkmanager_global_network.test.id
  site_id           = aws_networkmanager_site.test.id

  bandwidth {
    download_speed = 50
    upload_speed   = 10
  }

  tags = {
    (var.unknownTagKey) = null_resource.test.id
  }
}

resource "aws_networkmanager_site" "test" {
  global_network_id = aws_networkmanager_global_network.test.id
}

resource "aws_networkmanager_global_network" "test" {}

resource "null_resource" "test" {}

variable "unknownTagKey" {
  type     = string
  nullable = false
}
