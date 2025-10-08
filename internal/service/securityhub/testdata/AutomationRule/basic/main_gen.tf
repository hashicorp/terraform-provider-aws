# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_securityhub_automation_rule" "test" {
  description = "test description"
  rule_name   = var.rName
  rule_order  = 1

  actions {
    finding_fields_update {
      severity {
        label   = "LOW"
        product = "0.0"
      }

      types = ["Software and Configuration Checks/Industry and Regulatory Standards"]

      user_defined_fields = {
        key = "value"
      }
    }
    type = "FINDING_FIELDS_UPDATE"
  }

  criteria {
    aws_account_id {
      comparison = "EQUALS"
      value      = "1234567890"
    }
  }

  depends_on = [aws_securityhub_account.test]
}

resource "aws_securityhub_account" "test" {}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
