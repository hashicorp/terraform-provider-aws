# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_s3control_bucket" "test" {
  bucket     = var.rName
  outpost_id = data.aws_outposts_outpost.test.id
}

data "aws_outposts_outposts" "test" {}

data "aws_outposts_outpost" "test" {
  id = tolist(data.aws_outposts_outposts.test.ids)[0]
}
variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
