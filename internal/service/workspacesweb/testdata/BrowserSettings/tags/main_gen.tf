# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

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

  tags = var.resource_tags
}
variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "resource_tags" {
  description = "Tags to set on resource. To specify no tags, set to `null`"
  # Not setting a default, so that this must explicitly be set to `null` to specify no tags
  type     = map(string)
  nullable = true
}
