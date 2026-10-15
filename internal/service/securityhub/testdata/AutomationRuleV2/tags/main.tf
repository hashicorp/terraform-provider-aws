# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_securityhub_automation_rule_v2" "test" {
  rule_name   = var.rName
  description = "test description"
  rule_order  = 100
  rule_status = "ENABLED"

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

  tags = var.resource_tags

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

variable "resource_tags" {
  description = "Tags to set on resource. To specify no tags, set to `null`"
  type        = map(string)
  nullable    = true
}