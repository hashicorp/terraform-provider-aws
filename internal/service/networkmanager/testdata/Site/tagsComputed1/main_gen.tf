# Copyright IBM Corp. 2014, 2025
# SPDX-License-Identifier: MPL-2.0

provider "null" {}

resource "aws_networkmanager_site" "test" {
  global_network_id = aws_networkmanager_global_network.test.id

  tags = {
    (var.unknownTagKey) = null_resource.test.id
  }
}

resource "aws_networkmanager_global_network" "test" {}

resource "null_resource" "test" {}

variable "unknownTagKey" {
  type     = string
  nullable = false
}
