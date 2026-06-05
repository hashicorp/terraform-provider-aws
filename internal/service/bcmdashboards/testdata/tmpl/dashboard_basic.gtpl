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
{{- template "tags" . }}
}
