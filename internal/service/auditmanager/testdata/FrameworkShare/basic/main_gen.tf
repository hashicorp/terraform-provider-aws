# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_auditmanager_framework_share" "test" {
  destination_account = data.aws_caller_identity.current.account_id
  destination_region  = var.secondary_region
  framework_id        = aws_auditmanager_framework.test.id
}

data "aws_caller_identity" "current" {}

resource "aws_auditmanager_control" "test" {
  name = var.rName

  control_mapping_sources {
    source_name          = var.rName
    source_set_up_option = "Procedural_Controls_Mapping"
    source_type          = "MANUAL"
  }
}

resource "aws_auditmanager_framework" "test" {
  name = var.rName

  control_sets {
    name = var.rName

    controls {
      id = aws_auditmanager_control.test.id
    }
  }
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "secondary_region" {
  description = "Secondary region"
  type        = string
  nullable    = false
}
