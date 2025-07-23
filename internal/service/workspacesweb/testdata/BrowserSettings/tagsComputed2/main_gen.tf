# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

provider "null" {}

resource "aws_workspacesweb_browser_settings" "test" {
  browser_policy = jsonencode({
    chromePolicies = {
      DefaultDownloadDirectory = {
        value = "/home/as2-streaming-user/MyFiles/TemporaryFiles1"
      }
    }
  })

  tags = {
    (var.unknownTagKey) = null_resource.test.id
    (var.knownTagKey)   = var.knownTagValue
  }

}

resource "null_resource" "test" {}

variable "unknownTagKey" {
  type     = string
  nullable = false
}

variable "knownTagKey" {
  type     = string
  nullable = false
}

variable "knownTagValue" {
  type     = string
  nullable = false
}
