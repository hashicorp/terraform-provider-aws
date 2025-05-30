# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

provider "null" {}

resource "aws_workspacesweb_user_settings" "test" {
  copy_allowed     = "Enabled"
  download_allowed = "Enabled"
  paste_allowed    = "Enabled"
  print_allowed    = "Enabled"
  upload_allowed   = "Enabled"

  tags = {
    (var.unknownTagKey) = null_resource.test.id
  }

}
resource "null_resource" "test" {}

variable "unknownTagKey" {
  type     = string
  nullable = false
}
