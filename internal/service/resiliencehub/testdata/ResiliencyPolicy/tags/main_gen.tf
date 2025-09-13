# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_resiliencehub_resiliency_policy" "test" {
  name = var.rName

  tier = "NotApplicable"

  policy {
    az {
      rpo = "1h0m0s"
      rto = "1h0m0s"
    }
    hardware {
      rpo = "1h0m0s"
      rto = "1h0m0s"
    }
    software {
      rpo = "1h0m0s"
      rto = "1h0m0s"
    }
  }

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
