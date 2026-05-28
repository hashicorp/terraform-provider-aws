# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

provider "null" {}

resource "aws_bedrockagentcore_policy_engine" "test" {
  name = replace(var.rName, "-", "_")

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
