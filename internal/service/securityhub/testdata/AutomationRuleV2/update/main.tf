# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_securityhub_automation_rule_v2" "test" {
  rule_name   = var.rName
  description = "updated description"
  rule_order  = 200
  rule_status = "DISABLED"

  criteria {
    ocsf_finding_criteria_json = jsonencode({
      CompositeFilters = [
        {
          StringFilters = [
            {
              FieldName = "metadata.product.name"
              Filter = {
                Comparison = "EQUALS"
                Value      = "Inspector"
              }
            }
          ]
        }
      ]
      CompositeOperator = "AND"
    })
  }

  action {
    type = "FINDING_FIELDS_UPDATE"

    finding_fields_update {
      severity_id = 3
      status_id   = 2
      comment     = "Updated by automation rule"
    }
  }

  depends_on = [aws_securityhub_aggregator_v2.test]
}

resource "aws_securityhub_aggregator_v2" "test" {
  region_linking_mode = "SPECIFIED_REGIONS"
  linked_regions      = ["ap-southeast-1"]

  depends_on = [aws_securityhub_account_v2.test]
}

resource "aws_securityhub_account_v2" "test" {}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
