resource "aws_budgets_budget" "test" {
  name         = var.rName
  budget_type  = "RI_UTILIZATION"
  limit_amount = "100.0"
  limit_unit   = "PERCENTAGE"
  time_unit    = "QUARTERLY"

  cost_filter {
    name   = "Service"
    values = ["Amazon Redshift"]
  }

{{- template "tags" . }}
}
