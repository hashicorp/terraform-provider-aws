# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

provider "null" {}

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
