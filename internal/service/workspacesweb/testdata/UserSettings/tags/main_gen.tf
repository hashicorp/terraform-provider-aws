# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_workspacesweb_user_settings" "test" {
  copy_allowed     = "Enabled"
  download_allowed = "Enabled"
  paste_allowed    = "Enabled"
  print_allowed    = "Enabled"
  upload_allowed   = "Enabled"

  tags = var.resource_tags

}
variable "resource_tags" {
  description = "Tags to set on resource. To specify no tags, set to `null`"
  # Not setting a default, so that this must explicitly be set to `null` to specify no tags
  type     = map(string)
  nullable = true
}
