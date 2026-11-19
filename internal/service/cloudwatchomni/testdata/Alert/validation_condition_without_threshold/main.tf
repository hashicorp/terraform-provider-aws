# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_cloudwatchomni_alert" "test" {
  space_id   = "11111111-2222-4333-8444-555555555555"
  profile_id = "example"
  name       = var.rName

  rule {
    telemetry_rule {
      query {
        expression = "up"
        language   = "PROMQL"
      }

      condition {
        comparator     = "GT"
        threshold_mode = "COUNT_OF_RESULTS"
      }

      evaluation {
        interval_seconds = 60
      }

      no_data {
        treat_as = "OK"
      }
    }
  }
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
