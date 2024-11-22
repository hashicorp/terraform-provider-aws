# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

provider "null" {}

resource "aws_ssm_patch_baseline" "test" {
  name                              = var.rName
  description                       = "Baseline containing all updates approved for production systems"
  approved_patches                  = ["KB123456"]
  approved_patches_compliance_level = "CRITICAL"

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
