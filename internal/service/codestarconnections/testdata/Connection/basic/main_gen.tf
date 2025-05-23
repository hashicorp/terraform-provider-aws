# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_codestarconnections_connection" "test" {
  name          = var.rName
  provider_type = "Bitbucket"
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
