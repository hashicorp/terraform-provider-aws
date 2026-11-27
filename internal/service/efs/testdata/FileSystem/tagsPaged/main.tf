# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_efs_file_system" "test" {
  tags = {
    Name           = var.rName
    Another        = "tag"
    Test           = "yes"
    User           = "root"
    Page           = "1"
    Environment    = "prod"
    CostCenter     = "terraform"
    AcceptanceTest = "PagedTags"
    CreationToken  = "radek"
    PerfMode       = "max"
  }
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
