# Copyright IBM Corp. 2014, 2026
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
  }

}

resource "null_resource" "test" {}

variable "unknownTagKey" {
  type     = string
  nullable = false
}
