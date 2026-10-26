# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_fis_safety_lever_state" "test" {
  state {
    status = var.status
    reason = "Managed by Terraform acceptance test"
  }
}

variable "status" {
  description = "Safety lever status to set"
  type        = string
  nullable    = false
}
