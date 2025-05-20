# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

provider "null" {}

locals {
    var1 = "{"
    var2 = <<EOF
        "chromePolicies": {
        "DefaultDownloadDirectory": {
            "value": "/home/as2-streaming-user/MyFiles/TemporaryFiles1"
        }
    }
}
EOF
}
resource "aws_workspacesweb_browser_settings" "test" {
  browser_policy = "${local.var1} ${local.var2}"

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
