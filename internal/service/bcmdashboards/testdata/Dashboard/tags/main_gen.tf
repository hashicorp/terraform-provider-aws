# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_bcmdashboards_dashboard" "test" {
  name = var.rName

  widget {
    title = "example"

    configs {
      query_parameters {
        cost_and_usage {
          granularity = "MONTHLY"
          metrics     = ["UnblendedCost"]

          time_range {
            start_time {
              type  = "ABSOLUTE"
              value = "2025-01-01"
            }
            end_time {
              type  = "ABSOLUTE"
              value = "2025-03-31"
            }
          }
        }
      }

      display_config {
        graph {
          metric      = "UnblendedCost"
          visual_type = "BAR"
        }
      }
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
