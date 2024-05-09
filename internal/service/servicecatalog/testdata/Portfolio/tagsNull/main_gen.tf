# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_servicecatalog_portfolio" "test" {
  name          = var.rName
  description   = "test-b"
  provider_name = "test-c"

  tags = {
    (var.tagKey1) = null
  }
}


variable "rName" {
  type     = string
  nullable = false
}

variable "tagKey1" {
  type     = string
  nullable = false
}

