# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_securityhub_automation_rule_v2" "test" {

  region = var.region

  rule_name   = var.rName
  description = "test description"
  rule_order  = 100

  criteria {
    ocsf_finding_criteria_json = jsonencode({
      CompositeFilters = [
        {
          StringFilters = [
            {
              FieldName = "metadata.product.name"
              Filter = {
                Comparison = "EQUALS"
                Value      = "GuardDuty"
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
      severity_id = 99
      status_id   = 3
      comment     = "Low severity GuardDuty finding suppressed"
    }
  }

  depends_on = [aws_securityhub_aggregator_v2.test]
}

resource "aws_securityhub_aggregator_v2" "test" {
  region = var.region

  region_linking_mode = "SPECIFIED_REGIONS"
  linked_regions      = ["ap-southeast-1"]

  depends_on = [aws_securityhub_account_v2.test]
}

resource "aws_securityhub_account_v2" "test" {
  region = var.region

}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
