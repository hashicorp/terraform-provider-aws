# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_iam_virtual_mfa_device" "test" {
  virtual_mfa_device_name = var.rName

}


variable "rName" {
  type     = string
  nullable = false
}


